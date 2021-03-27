package notepet

import "encoding/json"

func noteToBytes(n Note) []byte {
	data, _ := json.MarshalIndent(n, "", "    ")
	return data
}

func noteListToBytes(notes []Note) []byte {
	data, _ := json.MarshalIndent(notes, "", "    ")
	return data
}

func bytesToNote(data []byte) (n Note, err error) {
	err = json.Unmarshal(data, &n)
	return
}

func bytesToNoteList(data []byte) (notes []Note, err error) {
	err = json.Unmarshal(data, &notes)
	return
}
