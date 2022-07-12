package model

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

func sortNotes(notes []Note) {
	sort.Slice(notes, func(i, j int) bool {
		if notes[i].Sticky && notes[j].Sticky {
			return notes[i].TimeStamp.After(notes[j].TimeStamp)
		} else if notes[i].Sticky {
			return true
		} else if notes[j].Sticky {
			return false
		}
		return notes[i].TimeStamp.After(notes[j].TimeStamp)
	})
}

func generateID(n Note) NoteID {
	sum := sha256.New()
	sum.Write([]byte(n.Title))
	sum.Write([]byte(n.TimeStamp.String()))
	s := string(sum.Sum(nil))
	return NoteID(fmt.Sprintf("%x", s))
}

func noteToBytes(n Note) []byte {
	data, _ := json.Marshal(n)
	return data
}

func noteListToBytes(notes []Note) []byte {
	data, _ := json.Marshal(notes)
	return data
}

func bytesToNote(data []byte, n *Note) (err error) {
	err = json.Unmarshal(data, &n)
	return
}

func bytesToNoteList(data []byte, notes *NoteList) (err error) {
	err = json.Unmarshal(data, notes)
	return
}
