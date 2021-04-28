package notepet

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	statementDBcreate string = `create table notes (id text primary key unique, title text, body text, tags text, sticky boolean, timestamp datetime, lastedited datetime)`
	statementDBput           = `insert into notes values (?, ?, ?, ?, ?, ?, ?)`
	statementDBgetOne        = `select * from notes where id = ?`
	statementDBgetAll        = `select * from notes`
	statementDBupd           = `update notes set title = ?, body = ?, tags = ?, sticky = ?, lastedited = ? where id = ?`
	statementDBdel           = `delete from notes where id = ?`
	statementDBsearch        = `select distinct * from notes where title like ? or body like ? or tags like ? or lastedited like ? or timestamp like ?`
)

type SQLiteStorage struct {
	// mu sync.Mutex
	db *sql.DB
}

func OpenOrInitSQLiteStorage(filename string) (Storage, error) {
	if st, err := openSQLiteStorage(filename, false); err == nil {
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
		if _, err := sqls.db.Exec(statementDBcreate); err != nil {
			return nil, err
		}
	}
	db.SetMaxOpenConns(1)
	return &sqls, nil
}

func (sqls *SQLiteStorage) Get(ids ...NoteID) ([]Note, error) {
	var rows *sql.Rows
	var err error
	switch {
	case len(ids) > 0:
		rows, err = sqls.db.Query(statementDBgetOne, ids[0])
	default:
		rows, err = sqls.db.Query(statementDBgetAll)
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
	_, err := sqls.db.Exec(statementDBput, n.ID, n.Title, n.Body, n.Tags, n.Sticky, n.TimeStamp, n.LastEdited)
	if err != nil {
		return BadNoteID, err
	}
	return n.ID, nil
}

func (sqls *SQLiteStorage) Upd(id NoteID, n Note) (NoteID, error) {
	return BadNoteID, nil
}

func (sqls *SQLiteStorage) Del(id NoteID) error {
	return nil
}

func (sqls *SQLiteStorage) Search(query string) ([]Note, error) {
	return []Note{}, nil
}

func (sqls *SQLiteStorage) Close() error {
	return sqls.db.Close()
}

func (sqls *SQLiteStorage) ExportJSON() ([]byte, error) {
	return []byte{}, nil
}
