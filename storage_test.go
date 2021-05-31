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

func testStorage(t *testing.T, st Storage) {
	defer st.Close()
	notes := getTestNotes()
	var ids []NoteID
	for _, note := range notes {
		id, err := st.Put(note)
		if err != nil {
			fmt.Println("storage failed to put note:", err)
			t.Fail()
		} else {
			ids = append(ids, id)
		}
	}
	received, err := st.Get()
	if err != nil {
		fmt.Println("storage failed to get all notes:", err)
		t.Fail()
	}
	for i, note := range received {
		if note.Title != notes[len(notes)-1-i].Title || note.Body != notes[len(notes)-1-i].Body {
			fmt.Println("notes recevied from storage do not match original")
			t.Fail()
		}
	}
	for i, id := range ids {
		n, err := st.Get(id)
		if err != nil {
			fmt.Println("error getting existing note ny ID:", err)
			t.Fail()
		}
		if n[0].Title != notes[i].Title || n[0].Body != notes[i].Body || n[0].Tags != notes[i].Tags {
			fmt.Println("got wrong note by ID")
			t.Fail()
		}
	}
	for _, note := range notes {
		if _, err := st.Search(note.Title); err != nil {
			fmt.Println("storage failed to search existing note by Title")
			t.Fail()
		}
		if _, err := st.Search(note.Body); err != nil {
			fmt.Println("storage failed to search existing note by Body")
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
	updID, err := st.Upd(received[0].ID, received[1])
	if err != nil {
		fmt.Println("failed to update existing note with err:", err)
		t.Fail()
	}
	updated, err := st.Get(updID)
	if err != nil {
		fmt.Println("failed to get updated note by new id", err)
		t.Fail()
	}
	if updated[0].Title != received[1].Title || updated[0].Body != received[1].Body {
		fmt.Println("updated note differs from original")
		fmt.Println("updated:", updated[0])
		fmt.Println("expected:", received[1])
		t.Fail()
	}
	toDelete, _ := st.Get()
	var deleteErr error
	for _, n := range toDelete {
		if err := st.Del(n.ID); err != nil {
			deleteErr = err
		}
	}
	if deleteErr != nil {
		fmt.Println("storage failed to delete one or more existing notes")
		t.Fail()
	}
}

func Test_JSONStoragePutsAndGetsSameBack(t *testing.T) {
	fmt.Println("Testing JSON Storage")
	testfile := "./test.json"
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
	testStorage(t, st)
}

func Test_SQLiteStorage(t *testing.T) {
	fmt.Println("Testing SQLite Storage")
	testDBfile := "./test.db"
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
	defer os.Remove(testDBfile)
	defer st.Close()
	testStorage(t, st)
}

func Test_PostgresStorage(t *testing.T) {
	fmt.Println("Testing Postgres Storage")
	st, err := OpenPostgresStorage("127.0.0.1", "5432", "notepet", "notepet", "notepet")
	if err != nil {
		fmt.Println("could not connect to database:", err)
		t.Fail()
	}
	defer st.Close()
	testStorage(t, st)
}

/*
func TestSqliteStorageConcurretly(t *testing.T) {
	st, err := OpenOrInitSQLiteStorage(testDBfile)
	if err != nil {
		fmt.Println("could not create test.db:", err)
		t.Fail()
	}
	defer os.Remove(testDBfile)
	defer st.Close()
}
*/
