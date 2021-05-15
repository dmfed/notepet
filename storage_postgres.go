package notepet

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	initPostgresDBStatement = `
create table if not exists notes
(id char(64) primary key,
title varchar(150), 
body text,
tags varchar(150),
sticky boolean,
created timestamp,
lastedited timestamp)`
)

type PostgresStorage struct {
	db *sql.DB
}

func OpenPostgresStorage(host, port, username, password, dbname string) (Storage, error) {
	connString := fmt.Sprintf("user=%v password=%v host=%v port=%v database=%v sslmode=disable", username, password, host, port, dbname)
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, err
	}
	var psql PostgresStorage
	psql.db = db
	if _, err := psql.db.Exec(initPostgresDBStatement); err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return &psql, nil
}

func (psql *PostgresStorage) Get(ids ...NoteID) ([]Note, error) {
	var rows *sql.Rows
	var err error
	switch {
	case len(ids) > 0:
		statement := `select * from notes where id = ?`
		rows, err = psql.db.Query(statement, ids[0])
	default:
		statement := `select * from notes`
		rows, err = psql.db.Query(statement)
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

func (psql *PostgresStorage) Put(n Note) (NoteID, error) {
	if n.Title == "" && n.Body == "" {
		return BadNoteID, ErrCanNotAddEmptyNote
	}
	t := time.Now()
	n.TimeStamp = t
	n.LastEdited = t
	n.ID = generateID(n)
	statement := `insert into notes values (?, ?, ?, ?, ?, ?, ?)`
	_, err := psql.db.Exec(statement, n.ID, n.Title, n.Body, n.Tags, n.Sticky, n.TimeStamp, n.LastEdited)
	if err != nil {
		return BadNoteID, err
	}
	return n.ID, nil
}

func (psql *PostgresStorage) Upd(id NoteID, n Note) (NoteID, error) {
	/* if _, err := psql.Get(id); err != nil {
		return BadNoteID, err
	} */
	n.LastEdited = time.Now()
	statement := `update notes set title = ?, body = ?, tags = ?, sticky = ?, lastedited = ? where id = ?`
	_, err := psql.db.Exec(statement, n.Title, n.Body, n.Tags, n.Sticky, n.LastEdited, n.ID)
	if err != nil {
		id = BadNoteID
	}
	return id, err
}

func (psql *PostgresStorage) Del(id NoteID) error {
	statement := `delete from notes where id = ?`
	_, err := psql.db.Exec(statement, id)
	return err
}

func (psql *PostgresStorage) Search(query string) ([]Note, error) {
	var result []Note
	notes, err := psql.Get()
	if err != nil {
		return result, err
	}
	query = strings.ToLower(query)
	for _, note := range notes {
		if strings.Contains(strings.ToLower(note.String()), query) {
			result = append(result, note)
		}
	}
	if len(result) == 0 {
		return result, ErrNoNotesFound
	}
	sortNotes(result)
	return result, nil
}

func (psql *PostgresStorage) Close() error {
	return psql.db.Close()
}

func (psql *PostgresStorage) ExportJSON() ([]byte, error) {
	return []byte{}, nil
}
