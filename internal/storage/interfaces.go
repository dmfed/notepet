package storage

import (
	"errors"

	"github.com/dmfed/notepet/model"
)

var (
	// ErrNoNotesFound when no model.Notes with requested model.NoteID are found in the storage
	ErrNotFound = errors.New("error: no model.Notes with such model.NoteID")
	// ErrCanNotAddEmptymodel.Note is returned when trying to Put() model.Note with empty body and title
	// into Storage
	ErrCanNotAddEmptyNote = errors.New("error: can not add empty model.Note")
	// ErrStorageIsNil is returned when nil pointer is passed to NewAPIHandler or
	// Newmodel.NotepetServer
	ErrStorageIsNil = errors.New("can not use storage: storage is nil")
)

//Storage interface represents any type of storage for model.Note objects.
type NoteRepo interface {
	// Get signature is intended to accept zero or one model.NoteID
	// if more than one model.NoteIDs specified the method may or may not
	// return second and subsequent ids depending on implementation
	Get(...model.NoteID) ([]model.Note, error)

	// Put accepts model.Note and should return model.NoteID if model.Note has been
	// successfully added to Storage
	Put(model.Note) (model.NoteID, error)

	// Upd accepts model.Note and should return model.NoteID if model.Note has been
	// successfully modified in Storage
	Upd(model.NoteID, model.Note) (model.NoteID, error)

	// Del deletes model.Note with specified model.NoteID. If delete was successful
	// it should return nil, error otherwise.
	Del(model.NoteID) error

	// Close should be used when Storage is no longer needed (to close
	// network connection, flush file to disk etc.)
	// If Close has not been called data is not guaranteed to be consistent.
	Close() error
}

type UserRepo interface {
	// GetID returns ID of user with username
	GetID(username string) (model.UserID, error)

	// NewUser adds a new use
	NewUser(username string) (model.UserID, error)
}

type User2Note interface {
	// GetNotes returns IDs of notes belonging to user with UserID id.
	GetNotes(id model.UserID, offset, limit int) ([]model.NoteID, error)

	// PutNotes adds IDs of notes belonging to user with id.
	PutNotes(id model.UserID, notes ...model.NoteID) error

	// DelNotes deletes notes with IDs notes belonging to user with ID id.
	DelNotes(id model.UserID, notes ...model.NoteID)
}
