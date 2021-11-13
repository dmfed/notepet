package notepet

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
	err := json.Unmarshal(TestNotes, &s.Notes)
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
		t.Log(err)
		t.Fail()
	}
	hndlr := initTestHandler(s)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api?action=search&q=test", nil)
	req.Header.Add("Notepet-Token", "test")
	w := httptest.NewRecorder()
	hndlr.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	var notes []Note
	if err := json.Unmarshal(body, &notes); err != nil {
		t.Log("failed to unmarshal response notes")
		t.Fail()
	}
	for i := range notes {
		if notes[i] != s.Notes[i] {
			t.Log(resp.StatusCode)
			t.Log(resp.Header.Get("Content-Type"))
			t.Log(string(body))
			t.Fail()
		}
	}
}
