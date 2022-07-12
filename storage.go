package notepet

import "errors"

var (
	// ErrNoNotesFound when no notes with requested NoteID are found in the storage
	ErrNoNotesFound = errors.New("error: no notes with such NoteID")
	// ErrCanNotAddEmptyNote is returned when trying to Put() note with empty body and title
	// into Storage
	ErrCanNotAddEmptyNote = errors.New("error: can not add empty note")
	// ErrStorageIsNil is returned when nil pointer is passed to NewAPIHandler or
	// NewNotepetServer
	ErrStorageIsNil = errors.New("can not use storage: storage is nil")
)

//Storage interface represents any type of storage for Note objects.
type Storage interface {
	// Get signature is intended to accept zero or one NoteID
	// if more than one NoteIDs specified the method may or may not
	// return second and subsequent ids depending on implementation
	Get(...NoteID) ([]Note, error)
	// Put accepts Note and should return NoteID if Note has been
	// successfully added to Storage
	Put(Note) (NoteID, error)
	// Upd accepts Note and should return NoteID if Note has been
	// successfully modified in Storage
	Upd(NoteID, Note) (NoteID, error)
	// Del deletes Note with specified NoteID. If delete was successful
	// it should return nil, error otherwise.
	Del(NoteID) error
	// Search looks up Notes containing specified string. It should
	// return error if no Notes have been found or if other error occured.
	Search(string) ([]Note, error)
	// Close should be used when Storage is no longer needed (to close
	// network connection, flush file to disk etc.)
	// If Close has not been called data is not guaranteed to be consistent.
	Close() error
}

func NewStorage(conf StorageConfig) (Storage, error) {
	return nil, nil
}

// Migrate copies all notes from src (source) Storage
// to dst (destination) Storage. If succesful the returned
// error in nil.
func Migrate(dst, src Storage) error {
	if src == nil || dst == nil {
		return ErrStorageIsNil
	}
	sourcenotes, err := src.Get()
	if err != nil {
		return err
	}
	for i := len(sourcenotes) - 1; i >= 0; i-- {
		if _, err := dst.Put(sourcenotes[i]); err != nil {
			return err
		}
	}
	return nil
}

// ExportJSON requests all Notes from st Storage, serializes to
// JSON and returns byte array. Just use string(output) if string type
// is required.
func ExportJSON(st Storage) ([]byte, error) {
	if st == nil {
		return []byte{}, ErrStorageIsNil
	}
	notes, err := st.Get()
	if err != nil {
		return []byte{}, ErrStorageIsNil
	}
	b := noteListToBytes(notes)
	return b, nil
}
