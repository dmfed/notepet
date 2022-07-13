package sqlite

import (
	"database/sql"

	"github.com/dmfed/notepet/internal/storage"
	"github.com/dmfed/notepet/internal/storage/sqlqueries"
	"github.com/dmfed/notepet/model"
	_ "github.com/mattn/go-sqlite3"
)

const (
	sqliteDriverName = "sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
}

func New(filename string) (storage.NoteRepo, storage.UserRepo, storage.User2Note, error) {
	r, err := openSQLiteStorage(filename)
	return r, r, r, err
}

func openSQLiteStorage(filename string) (storage.NoteRepo, error) {
	db, err := sql.Open(sqliteDriverName, filename)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	_, err = db.Exec(sqlqueries.StatementCreateSQLITETableNotes)

	return &SQLiteStorage{db}, err
}

func (s *SQLiteStorage) Get(ids ...model.NoteID) ([]model.Note, error) {
	// TODO sortNotes(notes)
	return sqlqueries.GetNote(s.db, ids...)
}

func (s *SQLiteStorage) Put(n model.Note) (model.NoteID, error) {
	return sqlqueries.PutNote(s.db, n)
}

func (s *SQLiteStorage) Upd(id model.NoteID, n model.Note) (model.NoteID, error) {
	return sqlqueries.UpdNote(s.db, id, n)
}

func (s *SQLiteStorage) Del(id model.NoteID) error {
	return sqlqueries.DelNote(s.db, id)
}

func (s *SQLiteStorage) Search(ids []model.NoteID, query string) ([]model.Note, error) {
	// TODO sortNotes(result)
	return sqlqueries.Search(s.db, ids, query)
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

func (s *SQLiteStorage) GetID(username string) (model.UserID, error) {
	return sqlqueries.GetUsrID(s.db, username)
}

func (s *SQLiteStorage) NewUser(username string) (model.UserID, error) {
	return sqlqueries.NewUser(s.db, username)
}

// GetNotes returns IDs of notes belonging to user with UserID id.
func (s *SQLiteStorage) GetNotes(id model.UserID, offset, limit int) ([]model.NoteID, error) {
	return
}

// PutNotes adds IDs of notes belonging to user with id.
func (s *SQLiteStorage) PutNotes(id model.UserID, notes ...model.NoteID) error

// DelNotes deletes notes with IDs notes belonging to user with ID id.
func (s *SQLiteStorage) DelNotes(id model.UserID, notes ...model.NoteID)
