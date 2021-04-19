package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dmfed/notepet"
)

var (
	// ErrNoNotesFound when no notes with requested NoteID are found in the storage
	ErrNoNotesFound = errors.New("error: no notes with such NoteID")
	// ErrCanNotAddEmptyNote is returned when trying to Put() note with empty body and title
	ErrCanNotAddEmptyNote = errors.New("error: can not add empty note")
)

//JSONFileStorage reads from json file and keeps objects in memory while the program is running
//Each change of objects list in memory is immediately flushed back to disk.
//This struct implements the Storage interface.
type JSONFileStorage struct {
	Notes     []notepet.Note `json:"notes,omitempty"`
	idToIndex map[notepet.NoteID]int
	filename  string
	changed   bool
	mu        sync.Mutex
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
	st.idToIndex = make(map[notepet.NoteID]int)
	if err = json.Unmarshal(data, &st); err != nil {
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
	st.Notes = []notepet.Note{}
	st.filename = filename
	st.Close()
	return OpenJSONFileStorage(st.filename)
}

//Get checks if Note with index i is present in the storage and returns relevant Note
func (st *JSONFileStorage) Get(ids ...notepet.NoteID) ([]notepet.Note, error) {
	notesToReturn := []notepet.Note{}
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
func (st *JSONFileStorage) Put(note notepet.Note) (notepet.NoteID, error) {
	if note.Title == "" && note.Body == "" {
		return notepet.BadNoteID, ErrCanNotAddEmptyNote
	}
	t := time.Now()
	note.TimeStamp = t
	note.LastEdited = t
	note.ID = note.GenerateID()
	st.mu.Lock()
	defer st.mu.Unlock()
	st.Notes = append(st.Notes, note)
	st.changed = true
	defer st.reindex()
	return note.ID, nil
}

// Upd removes Note at index i of Storage and places supplied object in its place.
// Returns error if underlying io operation is unsucessful
func (st *JSONFileStorage) Upd(id notepet.NoteID, note notepet.Note) (notepet.NoteID, error) {
	if note.Title == "" && note.Body == "" {
		return notepet.BadNoteID, ErrCanNotAddEmptyNote
	}
	note.LastEdited = time.Now()
	st.mu.Lock()
	defer st.mu.Unlock()
	index, ok := st.idToIndex[id]
	if !ok {
		return notepet.BadNoteID, ErrNoNotesFound
	}
	note.ID = id // ID won't change when replacing, only note.TimeEdited
	note.TimeStamp = st.Notes[index].TimeStamp
	st.Notes[index] = note
	st.changed = true
	defer st.reindex()
	return note.ID, nil

}

// Del removes Note from Storage
func (st *JSONFileStorage) Del(id notepet.NoteID) error {
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
func (st *JSONFileStorage) Search(want string) ([]notepet.Note, error) {
	want = strings.ToLower(strings.Trim(want, " \n"))
	var result []notepet.Note
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
	return json.MarshalIndent(st, "", "    ")
}

// Len returns the number of objects in the Storage
func (st *JSONFileStorage) Len() int {
	return len(st.Notes)
}

// Less implements sort.Interface from standard library
func (st *JSONFileStorage) Less(i, j int) bool {
	if st.Notes[i].Sticky && st.Notes[j].Sticky {
		return st.Notes[i].TimeStamp.After(st.Notes[j].TimeStamp)
	} else if st.Notes[i].Sticky {
		return true
	} else if st.Notes[j].Sticky {
		return false
	}
	return st.Notes[i].TimeStamp.After(st.Notes[j].TimeStamp)
}

func (st *JSONFileStorage) Swap(i, j int) {
	st.Notes[i], st.Notes[j] = st.Notes[j], st.Notes[i]
}

//syncToDisk rebuilds json and flushes it do disk.
func (st *JSONFileStorage) syncToDisk() error {
	data, err := json.MarshalIndent(st, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(st.filename, data, 0664)
}

func (st *JSONFileStorage) reindex() {
	sort.Sort(st)
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
