package model

type NoteList []Note

func (l *NoteList) Bytes() []byte {
	return noteListToBytes(*l)
}

func (l *NoteList) FromBytes(b []byte) (err error) {
	err = bytesToNoteList(b, l)
	return
}

func (l *NoteList) Append(n Note) {
	*l = append((*l), n)
}

func (l *NoteList) Sort() {
	sortNotes(*l)
}

func (l *NoteList) Len() int {
	return len(*l)
}
