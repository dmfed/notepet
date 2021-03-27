package notepet

import (
	"encoding/json"
	"fmt"
	"testing"
)

var testnotes = []byte(`
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

func TestNotesMarshal(t *testing.T) {
	notes := []Note{
		{ID: NoteID("first"), Body: "hello"},
		{ID: NoteID("second"), Body: "hello2"},
	}
	data, err := json.Marshal(notes)
	if err != nil {
		fmt.Println(err)
	}
	var newnotes []Note
	err = json.Unmarshal(data, &newnotes)
	if err != nil {
		fmt.Println(err)
	}
}

type NoteList struct {
	Notes []Note
}

func TestNotesUnmarshal(t *testing.T) {
	var c NoteList
	err := json.Unmarshal(testnotes, &c)
	if err != nil {
		fmt.Println(err)
	}
}
