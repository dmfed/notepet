package notepet

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

var TestNotes = []byte(`
[
	{
		"id": "f36112adc4cb98655790146781553eb71491e439f067ec7d109f9017132c5307",
		"title": "Test4",
		"body": "Body4",
		"tags": "tst4",
		"timestamp": "2021-03-26T17:56:45.378249509+03:00",
		"lastedited": "2021-03-26T17:56:45.378249509+03:00"
	},
	{
		"id": "ad433d4e7fcc9fa8cf0495e480e08d529f3122dab276443fd98d75ecfff1adc3",
		"title": "Test3",
		"body": "Body3",
		"tags": "tst3",
		"timestamp": "2021-03-26T17:56:45.378231157+03:00",
		"lastedited": "2021-03-26T17:56:45.378231157+03:00"
	},
	{
		"id": "5a19ae6f255ee382ff0d49e63d9d88969974b9db547fe128050dc7479cff1d8d",
		"title": "Test2",
		"body": "Body2",
		"tags": "tst2",
		"timestamp": "2021-03-26T17:56:45.378211361+03:00",
		"lastedited": "2021-03-26T17:56:45.378211361+03:00"
	},
	{
		"id": "7f0d826270b30a3ff6f27a4612133ede217e4f3a57ae502ea70ded1d49bee799",
		"title": "Test1",
		"body": "Body1",
		"tags": "tst",
		"timestamp": "2021-03-26T17:56:45.378169121+03:00",
		"lastedited": "2021-03-26T17:56:45.378169121+03:00"
	}
]`)

func getTestNotes() []Note {
	var notes []Note
	json.Unmarshal(TestNotes, &notes)
	return notes
}

var (
	testfile   = "./test.json"
	testDBfile = "./test.db"
)

func TestJSONStoragePutsAndGetsSameBack(t *testing.T) {
	if _, err := os.Stat(testfile); err == nil {
		os.Remove(testfile)
	}

	st, err := OpenOrInitJSONFileStorage(testfile)
	if err != nil {
		fmt.Println("could not open or create file:", err)
		t.FailNow()
	}
	defer os.Remove(testfile)
	defer st.Close()

	original := getTestNotes()
	var ids []NoteID
	for i := len(original) - 1; i >= 0; i-- {
		// Putting in reverse order, because notes will be sorted by timestamp
		id, err := st.Put(original[i])
		if err != nil {
			fmt.Println(err)
			t.Fail()
		} else {
			ids = append(ids, id)
		}
	}

	extracted, err := st.Get()
	if len(extracted) != len(original) {
		fmt.Printf("extracted != original, expected: %v, got: %v\n", len(extracted), len(original))
		t.Fail()
	}
	if err != nil {
		fmt.Printf("st.Get() returned: %v\n", err)
		t.Fail()
	}

	for i := 0; i < len(original); i++ {
		if original[i].Body != extracted[i].Body || original[i].Title != extracted[i].Title {
			fmt.Println("failed comparing original and extracted")
			t.Fail()
		}
	}

	for i, id := range ids {
		n, err := st.Get(id)
		if err != nil {
			fmt.Println("error getting existing note")
			t.Fail()
		}
		if n[0].Title != original[len(original)-i-1].Title || n[0].Body != original[len(original)-i-1].Body {
			fmt.Println("Get returned wrong note")
			t.Fail()
		}
	}
	for _, note := range original {
		if _, err := st.Search(note.Title); err != nil {
			fmt.Println("Failed to search existing note by Title")
			t.Fail()
		}
		if _, err := st.Search(note.Body); err != nil {
			fmt.Println("Failed to search existing note by Body")
			t.Fail()
		}
	}
}

func Test_SQLiteStorage(t *testing.T) {
	if _, err := OpenSQLiteStorage("noexistent"); err == nil {
		fmt.Println("OpenSQLiteStorage creates file when not needed")
		t.Fail()
	}
	if _, err := os.Stat(testDBfile); err == nil {
		os.Remove(testDBfile)
	}
	st, err := OpenOrInitSQLiteStorage(testDBfile)
	if err != nil {
		fmt.Println("could not create test.db:", err)
		t.Fail()
	}
	defer os.Remove("test.db")
	defer st.Close()

	notes := getTestNotes()
	var ids []NoteID
	for _, note := range notes {
		id, err := st.Put(note)
		if err != nil {
			fmt.Println("Failed to put note into SQLite Storage:", err)
			t.Fail()
		} else {
			ids = append(ids, id)
		}
	}
	received, err := st.Get()
	if err != nil {
		fmt.Println("Could not get all notes:", err)
		t.Fail()
	}
	for i, note := range received {
		if note.Title != notes[len(notes)-1-i].Title || note.Body != notes[len(notes)-1-i].Body {
			fmt.Println("Notes do not match")
			t.Fail()
		}
	}
	for i, id := range ids {
		n, err := st.Get(id)
		if err != nil {
			fmt.Println("error getting note ny ID:", err)
			t.Fail()
		}
		if n[0].Title != notes[i].Title || n[0].Body != notes[i].Body || n[0].Tags != notes[i].Tags {
			fmt.Println("got wrong note by ID")
			t.Fail()
		}
	}
	for _, note := range notes {
		if _, err := st.Search(note.Title); err != nil {
			fmt.Println("SQLite storage failed to search existing note by Title")
			t.Fail()
		}
		if _, err := st.Search(note.Body); err != nil {
			fmt.Println("SQLite storage failed to search existing note by Body")
			t.Fail()
		}
	}
	queries := []string{"Test", "test", "tst", "Body", "body"}
	for _, q := range queries {
		if res, err := st.Search(q); err != nil || len(res) < 1 {
			fmt.Println("failed looking for:", q)
			t.Fail()
		}
	}
}
