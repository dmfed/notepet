package notepet

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// NoteID is a unique Id of each note as recorded by Storage
type NoteID string

func (id NoteID) String() string {
	return string(id)
}

// BadNoteID is an invalid id. It is returned when method signature
// requres to return NoteID but there is no actual valid data to return
var BadNoteID = NoteID("nil")

// Note is a basic structure keeping note data. Fields are self-explanatory.
type Note struct {
	ID         NoteID    `json:"id,omitempty"`
	Title      string    `json:"title,omitempty"`
	Body       string    `json:"body,omitempty"`
	Tags       string    `json:"tags,omitempty"`
	Sticky     bool      `json:"sticky,omitempty"`
	TimeStamp  time.Time `json:"timestamp,omitempty"`
	LastEdited time.Time `json:"lastedited,omitempty"`
}

func (n Note) GenerateID() NoteID {
	sum := sha256.New()
	sum.Write([]byte(n.Title))
	sum.Write([]byte(n.TimeStamp.String()))
	s := string(sum.Sum(nil))
	return NoteID(fmt.Sprintf("%x", s))
}

func (n Note) String() (out string) {
	out += fmt.Sprintf("%v %v", n.ID, n.TimeStamp.Format("02/01/2006 15:04:05"))
	if n.Sticky {
		out += " STICKY"
	}
	if n.Title != "" {
		out += " " + strings.TrimRight(n.Title, "\n")
	}
	out += "\n"
	out += n.Body + "\n"
	if n.Tags != "" {
		out += fmt.Sprintf("> %v <", strings.Trim(n.Tags, " \n"))
	}
	return
}

// StringLong string with verbose output
func (n Note) StringLong() (out string) {
	out = "ID: " + fmt.Sprint(n.ID)
	if n.Sticky {
		out += " STICKY " + "\n"
	} else {
		out += "\n"
	}
	if n.Title != "" {
		out += "Title: " + strings.TrimRight(n.Title, "\n") + "\n"
	}
	out += strings.TrimRight(n.Body, "\n") + "\n"
	if n.Tags != "" {
		out += "Tags: " + strings.Trim(n.Tags, " \n") + "\n"
	}
	out += n.TimeStamp.Format("02/01/2006 15:04:05") + "\n"
	return
}

//Storage interface represents any type of storage for Note objects.
type Storage interface {
	Get(...NoteID) ([]Note, error)
	Put(Note) (NoteID, error)
	Upd(NoteID, Note) (NoteID, error)
	Del(NoteID) error
	Search(string) ([]Note, error)
	Close() error
	ExportJSON() ([]byte, error) // debugging. Remove this.
}
