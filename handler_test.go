package notepet

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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

type fakeStorage struct {
	Notes []Note
}

func (s *fakeStorage) Get(id ...NoteID) ([]Note, error) {
	return s.Notes, nil
}

func (s *fakeStorage) Put(n Note) (NoteID, error) {
	s.Notes = append(s.Notes, n)
	return "abcdef", nil
}

func (s *fakeStorage) Upd(id NoteID, n Note) (NoteID, error) {
	return NoteID("abc"), nil
}

func (s *fakeStorage) Del(id NoteID) error {
	return nil
}
func (s *fakeStorage) Search(want string) ([]Note, error) {
	switch want {
	case "test":
		return s.Notes, nil
	default:
		return []Note{}, ErrNoNotesFound
	}

}
func (s *fakeStorage) Close() error {
	return nil
}
func (s *fakeStorage) ExportJSON() ([]byte, error) {
	return []byte{}, nil
}

func initFakeStorage() (*fakeStorage, error) {
	var s fakeStorage
	err := json.Unmarshal(TestNotes, &s)
	return &s, err
}

func initTestHandler(s Storage) http.Handler {
	tokens := make(map[string]struct{})
	tokens["test"] = struct{}{}
	hndlr := APIHandler{Storage: s, Tokens: tokens}
	return &hndlr
}

func Test_APIHandlerSearch(t *testing.T) {
	s, err := initFakeStorage()
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	hndlr := initTestHandler(s)
	req := httptest.NewRequest("GET", "http://example.com/api?action=search&q=test", nil)
	req.Header.Add("Notepet-Token", "test")
	w := httptest.NewRecorder()
	hndlr.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	var notes []Note
	if err := json.Unmarshal(body, &notes); err != nil {
		fmt.Println("failed to unmarshal response notes")
		t.Fail()
	}
	for i := range notes {
		if notes[i] != s.Notes[i] {
			fmt.Println(resp.StatusCode)
			fmt.Println(resp.Header.Get("Content-Type"))
			fmt.Println(string(body))
			t.Fail()
		}
	}
}
