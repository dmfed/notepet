package sqlqueries

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/dmfed/notepet/internal/storage"
	"github.com/dmfed/notepet/model"
)

// Create tables
var (
	StatementCreateSQLITETableNotes = `
	create table if not exists notes (id text primary key unique, title text, body text, tags text, sticky boolean, timestamp datetime, lastedited datetime)`

	StatementCreateSQLITEUsersTable = `
	create table if not exists users (id integer primary key autoincrement, username text unique)`

	StatementCreateSQLITEUser2NotesTable = `
	create table if not exists user2notes (userid integer, noteid text unique)`
)

// Note repo
var (
	StatementSelectNotes = `
	select id, title, body, tags, sticky, timestamp, lastedited from notes where id in ($1)`

	StatementInsertNote = `
	insert into notes values ($1, $2, $3, $4, $5, $6, $7)`

	StatementUpdateNote = `
	update update notes set title = $1, body = $2, tags = $3, sticky = $4, lastedited = $5 where id = $6`

	StatementDeleteNote = `
	delete from notes where id = $1`
)

// UsersRepo
var (
	StatementSelectUserID = `
	select id from users where username = $1`

	StatementInsertUser = `
	insert into users (username) values ($1) returning id`
)

var (
	StatementSelectUserNotes = `
	select noteid from user2notes where userid = $1`

	StatementInsertUserNotes = `
	insert into user2notes values($1, $2)`  // TODO
)

func GetNote(db *sql.DB, ids ...model.NoteID) ([]model.Note, error) {
	var (
		rows  *sql.Rows
		notes []model.Note
		err   error
	)

	rows, err = db.Query(StatementSelectNotes, ids)
	if err != nil {
		return notes, err
	}
	defer rows.Close()

	for rows.Next() {
		var n model.Note
		if rows.Scan(&n.ID, &n.Title, &n.Body, &n.Tags, &n.Sticky, &n.TimeStamp, &n.LastEdited); err == nil {
			notes = append(notes, n)
		} else {
			log.Println(err)
		}
	}
	// TODO sortNotes(notes)
	return notes, nil
}

func PutNote(db *sql.DB, n model.Note) (model.NoteID, error) {
	if n.Title == "" && n.Body == "" {
		return model.BadNoteID, storage.ErrCanNotAddEmptyNote
	}
	t := time.Now()
	n.TimeStamp = t
	n.LastEdited = t
	// TODO n.ID = generateID(n)
	_, err := db.Exec(StatementInsertNote, n.ID, n.Title, n.Body, n.Tags, n.Sticky, n.TimeStamp, n.LastEdited)
	if err != nil {
		return model.BadNoteID, err
	}
	return n.ID, nil
}

func UpdNote(db *sql.DB, id model.NoteID, n model.Note) (model.NoteID, error) {
	n.LastEdited = time.Now()
	_, err := db.Exec(StatementUpdateNote, n.Title, n.Body, n.Tags, n.Sticky, n.LastEdited, id)
	if err != nil {
		id = model.BadNoteID
	}
	return id, err
}

func DelNote(db *sql.DB, id model.NoteID) error {
	_, err := db.Exec(StatementDeleteNote, id)
	return err
}

func GetUsrID(db *sql.DB, username string) (id model.UserID, err error) {
	row := db.QueryRow(StatementSelectUserID, username)
	err = row.Scan(&id)
	return
}

func NewUser(db *sql.DB, username string) (id model.UserID, err error) {
	row := db.QueryRow(StatementInsertUser, username)
	err = row.Scan(&id)
	return
}

func InsertUserNotes(db *sql.DB, id model.UserID, notes []model.NoteID) {
	// TODO
}

func Search(db *sql.DB, ids []model.NoteID, query string) ([]model.Note, error) {
	var result []model.Note

	notes, err := GetNote(db, ids...)
	if err != nil {
		return result, err
	}
	query = strings.ToLower(query)
	for _, note := range notes {
		if strings.Contains(strings.ToLower(note.String()), query) {
			result = append(result, note)
		}
	}

	// TODO sortNotes(result)
	return result, nil
}
