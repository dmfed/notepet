package notepet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//JSONFileStorage reads from json file and keeps objects in memory while the program is running
//Each change of objects list in memory is immediately flushed back to disk.
//This struct implements the Storage interface.
type JSONFileStorage struct {
	mu        sync.Mutex
	Notes     []Note
	idToIndex map[NoteID]int
	filename  string
	changed   bool
}

// OpenOrInitJSONFileStorage returns Storage interface is file exists
// or initializes new storage with requested path
// If function returns an error this is an indication that
// file with requested path and name could not be created
func OpenOrInitJSONFileStorage(filename string) (*JSONFileStorage, error) {
	if st, err := OpenJSONFileStorage(filename); err == nil {
		return st, err
	}
	return CreateJSONFileStorage(filename)
}

//OpenJSONFileStorage opens an existing sotrage and returns pointer to it
func OpenJSONFileStorage(filename string) (*JSONFileStorage, error) {
	// TODO: Create lockfile on open. Check lockfile exists before actually opening.
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var st JSONFileStorage
	st.filename = filename
	st.idToIndex = make(map[NoteID]int)
	if err = json.Unmarshal(data, &st.Notes); err != nil {
		return nil, err
	}
	st.reindex()
	st.startSyncDaemon(time.Minute * 2)
	return &st, nil
}

//CreateJSONFileStorage initializes empty json file then creates and returns new
//Storage interface
func CreateJSONFileStorage(filename string) (*JSONFileStorage, error) {
	if _, err := os.Stat(filename); err == nil {
		return nil, fmt.Errorf("could not create new storage: file %v already exists", filename)
	}
	dir, _ := filepath.Split(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) && dir != "" {
		if err = os.MkdirAll(dir, 0755); err != nil { // 755 needs to be changed to 0700
			return nil, err
		}
	}
	var st JSONFileStorage
	st.Notes = []Note{}
	st.filename = filename
	if err := st.syncToDisk(); err != nil {
		// this is probably a write / permissions error
		// we shouldn't return an unusable storage
		return nil, err
	}
	return OpenJSONFileStorage(st.filename)
}

//Get checks if Note with index i is present in the storage and returns relevant Note
func (st *JSONFileStorage) Get(ids ...NoteID) ([]Note, error) {
	notesToReturn := []Note{}
	st.mu.Lock()
	defer st.mu.Unlock()
	switch len(ids) {
	case 0:
		notesToReturn = append(notesToReturn, st.Notes...)
	default:
		for _, id := range ids {
			if index, ok := st.idToIndex[id]; ok {
				notesToReturn = append(notesToReturn, st.Notes[index])
			}
		}
	}
	if len(notesToReturn) == 0 {
		return notesToReturn, ErrNoNotesFound
	}
	return notesToReturn, nil
}

// Put adds Note to Storage
func (st *JSONFileStorage) Put(note Note) (NoteID, error) {
	if note.Title == "" && note.Body == "" {
		return BadNoteID, ErrCanNotAddEmptyNote
	}
	t := time.Now()
	note.TimeStamp = t
	note.LastEdited = t
	note.ID = generateID(note)
	st.mu.Lock()
	defer st.mu.Unlock()
	st.Notes = append(st.Notes, note)
	st.changed = true
	defer st.reindex()
	return note.ID, nil
}

// Upd replaces Note with id with supplied Note note.
// Returns error if underlying io operation is unsucessful
func (st *JSONFileStorage) Upd(id NoteID, note Note) (NoteID, error) {
	if note.Title == "" && note.Body == "" {
		return BadNoteID, ErrCanNotAddEmptyNote
	}
	note.LastEdited = time.Now()
	st.mu.Lock()
	defer st.mu.Unlock()
	index, ok := st.idToIndex[id]
	if !ok {
		return BadNoteID, ErrNoNotesFound
	}
	note.ID = id // ID won't change when replacing, only note.TimeEdited
	note.TimeStamp = st.Notes[index].TimeStamp
	st.Notes[index] = note
	st.changed = true
	defer st.reindex()
	return note.ID, nil

}

// Del removes Note from Storage
func (st *JSONFileStorage) Del(id NoteID) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	index, ok := st.idToIndex[id]
	if !ok {
		return ErrNoNotesFound
	}
	st.Notes = append(st.Notes[:index], st.Notes[index+1:]...)
	st.changed = true
	defer st.reindex()
	return nil
}

//Search removes leading and trailing spaces from request and matches the resulting substring
//against each note in the storage, checking body and title. Search is case insensitive.
func (st *JSONFileStorage) Search(want string) ([]Note, error) {
	want = strings.ToLower(strings.Trim(want, " \n"))
	var result []Note
	st.mu.Lock()
	defer st.mu.Unlock()
	for i := 0; i < len(st.Notes); i++ {
		if strings.Contains(strings.ToLower(st.Notes[i].String()), want) {
			result = append(result, st.Notes[i])
		}
	}
	if len(result) == 0 {
		return result, ErrNoNotesFound
	}
	return result, nil
}

// Close flushes all notes to disk
func (st *JSONFileStorage) Close() (err error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.changed {
		err = st.syncToDisk()
	}
	return nil
}

// ExportJSON returns a byte array of all notes in JSON format
func (st *JSONFileStorage) ExportJSON() ([]byte, error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	return json.MarshalIndent(st.Notes, "", "    ")
}

//syncToDisk rebuilds json and flushes it do disk.
func (st *JSONFileStorage) syncToDisk() error {
	data, err := json.MarshalIndent(st.Notes, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(st.filename, data, 0664)
}

func (st *JSONFileStorage) reindex() {
	sortNotes(st.Notes)
	for index, note := range st.Notes {
		st.idToIndex[note.ID] = index
	}
}

func (st *JSONFileStorage) startSyncDaemon(d time.Duration) {
	go func() {
		for {
			timer := time.NewTimer(d)
			<-timer.C
			st.mu.Lock()
			if st.changed {
				st.syncToDisk()
				st.changed = false
			}
			st.mu.Unlock()
		}
	}()
}
