package model

import (
	"fmt"
	"strings"
	"time"
)

// BadNoteID is an invalid id. It is returned when method signature
// requires to return NoteID but there is no actual valid data to return.
var BadNoteID = NoteID("nil")

// NoteID is a unique Id of each note as recorded by Storage
type NoteID string

func (id NoteID) String() string {
	return string(id)
}

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

func (n *Note) String() (out string) {
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

func (n *Note) Bytes() []byte {
	return noteToBytes(*n)
}

func (n *Note) FromBytes(b []byte) (err error) {
	err = bytesToNote(b, n)
	return
}
