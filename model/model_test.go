package model

import (
	"fmt"
	"testing"
	"time"
)

var testNotes = []Note{
	{
		ID:         "1",
		Title:      "testtitle1",
		Body:       "testbody1",
		Tags:       "tags1",
		Sticky:     false,
		TimeStamp:  time.Now(),
		LastEdited: time.Now(),
	},
	{
		ID:         "2",
		Title:      "testtitle2",
		Body:       "testbody2",
		Tags:       "tags2",
		Sticky:     false,
		TimeStamp:  time.Now(),
		LastEdited: time.Now(),
	},
	{
		ID:         "3",
		Title:      "testtitle3",
		Body:       "testbody3",
		Tags:       "tags3",
		Sticky:     false,
		TimeStamp:  time.Now(),
		LastEdited: time.Now(),
	},
}

func TestHashGenerator(t *testing.T) {
	stamp := time.Now()
	h1 := generateID(Note{Title: "hello", TimeStamp: stamp})
	h2 := generateID(Note{Title: "hello", TimeStamp: time.Now()})
	h3 := generateID(Note{Title: "hello", TimeStamp: stamp})
	if h1 == h2 {
		fmt.Println("hash is not unique")
		t.Fail()
	}
	if h1 != h3 {
		fmt.Println("hash should be the same for one note")
		t.Fail()
	}
}

func TestNoteMethods(t *testing.T) {
	n := testNotes[0]

	b := n.Bytes()
	// t.Log(string(b))

	var n2 Note
	err := n2.FromBytes(b)
	if err != nil {
		t.Error(err)
	}

	// t.Log(string(n2.Bytes()))

	if n2.Title != n.Title || n2.ID != n.ID ||
		!n2.TimeStamp.Equal(n.TimeStamp) {
		t.Errorf("incorrect unmarshalling")
	}
}

func TestNoteListMethods(t *testing.T) {
	var nl NoteList
	for i := range testNotes {
		nl.Append(testNotes[i])
	}
	if nl.Len() != 3 {
		t.Errorf("Len is less than expected")
	}
	nl.Sort()
	if nl[0].ID != testNotes[2].ID {
		// the list should reverse after calling Sort
		// (timestamps were added with time.Now())
		// newest notes should come first after Sort()
		t.Errorf("Sort does not rearrange list in correct order")
	}

	b := nl.Bytes()
	// t.Log(string(b))

	var nl2 NoteList
	err := nl2.FromBytes(b)
	if err != nil {
		t.Error(err)
	}
	// t.Log(string(nl2.Bytes()))
}
