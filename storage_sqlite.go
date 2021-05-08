package notepet

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
	mu sync.Mutex
}

func OpenOrInitSQLiteStorage(filename string) (Storage, error) {
	if st, err := OpenSQLiteStorage(filename); err == nil {
		return st, nil
	}
	return CreateSQLiteStorage(filename)
}

func OpenSQLiteStorage(filename string) (Storage, error) {
	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}
	return openSQLiteStorage(filename, false)
}

func CreateSQLiteStorage(filename string) (Storage, error) {
	if _, err := os.Stat(filename); err == nil {
		return nil, fmt.Errorf("could not create new storage: file %v already exists", filename)
	}
	dir, _ := filepath.Split(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) && dir != "" {
		if err = os.MkdirAll(dir, 0755); err != nil { // 755 needs to be changed to 0700
			return nil, err
		}
	}
	return openSQLiteStorage(filename, true)
}

func openSQLiteStorage(filename string, initDB bool) (Storage, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	var sqls SQLiteStorage
	sqls.db = db
	if initDB {
		statement := `create table notes (id text primary key unique, title text, body text, tags text, sticky boolean, timestamp datetime, lastedited datetime)`
		if _, err := sqls.db.Exec(statement); err != nil {
			return nil, err
		}
	}
	db.SetMaxOpenConns(1)
	return &sqls, nil
}

func (sqls *SQLiteStorage) Get(ids ...NoteID) ([]Note, error) {
	sqls.mu.Lock()
	defer sqls.mu.Unlock()
	var rows *sql.Rows
	var err error
	switch {
	case len(ids) > 0:
		statement := `select * from notes where id = ?`
		rows, err = sqls.db.Query(statement, ids[0])
	default:
		statement := `select * from notes`
		rows, err = sqls.db.Query(statement)
	}
	notes := []Note{}
	if err != nil {
		return notes, err
	}
	defer rows.Close()
	for rows.Next() {
		var n Note
		//var id string
		if rows.Scan(&n.ID, &n.Title, &n.Body, &n.Tags, &n.Sticky, &n.TimeStamp, &n.LastEdited); err == nil {
			//n.ID = NoteID(id)
			notes = append(notes, n)
		} else {
			log.Println(err)
		}
	}
	sortNotes(notes)
	return notes, nil
}

func (sqls *SQLiteStorage) Put(n Note) (NoteID, error) {
	if n.Title == "" && n.Body == "" {
		return BadNoteID, ErrCanNotAddEmptyNote
	}
	t := time.Now()
	n.TimeStamp = t
	n.LastEdited = t
	n.ID = generateID(n)
	sqls.mu.Lock()
	defer sqls.mu.Unlock()
	statement := `insert into notes values (?, ?, ?, ?, ?, ?, ?)`
	_, err := sqls.db.Exec(statement, n.ID, n.Title, n.Body, n.Tags, n.Sticky, n.TimeStamp, n.LastEdited)
	if err != nil {
		return BadNoteID, err
	}
	return n.ID, nil
}

func (sqls *SQLiteStorage) Upd(id NoteID, n Note) (NoteID, error) {
	/* if _, err := sqls.Get(id); err != nil {
		return BadNoteID, err
	} */
	n.LastEdited = time.Now()
	sqls.mu.Lock()
	defer sqls.mu.Unlock()
	statement := `update notes set title = ?, body = ?, tags = ?, sticky = ?, lastedited = ? where id = ?`
	_, err := sqls.db.Exec(statement, n.Title, n.Body, n.Tags, n.Sticky, n.LastEdited, n.ID)
	if err != nil {
		id = BadNoteID
	}
	return id, err
}

func (sqls *SQLiteStorage) Del(id NoteID) error {
	statement := `delete from notes where id = ?`
	sqls.mu.Lock()
	defer sqls.mu.Unlock()
	_, err := sqls.db.Exec(statement, id)
	return err
}

func (sqls *SQLiteStorage) Search(query string) ([]Note, error) {
	var result []Note
	notes, err := sqls.Get()
	if err != nil {
		return result, err
	}
	query = strings.ToLower(query)
	for _, note := range notes {
		if strings.Contains(strings.ToLower(note.String()), query) {
			result = append(result, note)
		}
	}
	sortNotes(result)
	return result, nil
}

func (sqls *SQLiteStorage) Close() error {
	return sqls.db.Close()
}

func (sqls *SQLiteStorage) ExportJSON() ([]byte, error) {
	return []byte{}, nil
}
