package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/dmfed/notepet"
)

var TestNotes = []byte(`
{
	"notes": [
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
	]
}`)

type storagesim struct {
	Notes []notepet.Note
}

var (
	testfile    = "./test.json"
	samplesfile = "./samples.json"
)

func TestStoragePutsAndGetsSameBack(t *testing.T) {
	if _, err := os.Stat(testfile); err == nil {
		os.Remove(testfile)
	}

	st, err := OpenOrInitJSONFileStorage(testfile)
	if err != nil {
		fmt.Println("could not open or create file")
		t.FailNow()
	}
	defer os.Remove(testfile)
	defer st.Close()

	var sim storagesim
	if err := json.Unmarshal(TestNotes, &sim); err != nil {
		fmt.Println("could not unmarshal TestNotes")
		t.FailNow()
	}
	original := sim.Notes
	L := len(original)
	for i := len(original) - 1; i >= 0; i-- {
		// Putting in reverse order, because notes will be sorted by timestamp
		if _, err := st.Put(original[i]); err != nil {
			fmt.Println(err)
			t.Fail()
		}
	}
	if st.Len() != L {
		fmt.Printf("storage Len() in incorrect, expected: %v, got: %v\n", L, st.Len())
		t.Fail()
	}

	extracted, err := st.Get()
	if len(extracted) != L {
		fmt.Printf("extracted != original, expected: %v, got: %v\n", len(extracted), L)
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
}

func TestStorageExtractsNotesByIDsAndSearches(t *testing.T) {
	if _, err := os.Stat(samplesfile); os.IsNotExist(err) {
		file, err := os.Create(samplesfile)
		if err != nil {
			fmt.Println("could not write samples file can not proceed")
			t.FailNow()
		}
		file.Write(TestNotes)
		file.Close()
	}
	st, err := OpenJSONFileStorage(samplesfile)
	if err != nil {
		fmt.Println("could not open storage file")
		t.FailNow()
	}
	var sim storagesim
	if err := json.Unmarshal(TestNotes, &sim); err != nil {
		fmt.Println("could not unmarshal TestNotes data to storage")
		t.FailNow()
	}
	for _, note := range sim.Notes {
		n, err := st.Get(note.ID)
		if err != nil {
			fmt.Println("error getting existing note")
			t.Fail()
		}
		if n[0].Title != note.Title || n[0].Body != note.Body {
			fmt.Println("Get returned wrong note")
			t.Fail()
		}
	}
	for _, note := range sim.Notes {
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

// TODO add 1000+ notes at once with goroutines
