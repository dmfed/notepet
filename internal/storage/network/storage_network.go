package network

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/dmfed/notepet/internal/storage"
	"github.com/dmfed/notepet/model"
)

// APIClient represents http client fetching notes from notepet server.
// It implements Storage interface.
type APIClient struct {
	Username   string
	Password   string
	HTTPClient *http.Client
	URL        *url.URL
}

var hostnameRE = regexp.MustCompile(`http://|https://`)

// NewAPIClient returns instance of APIClient configured
// to send requests to specified ip address
func NewAPIClient(addr, username, password string) (storage.NoteRepo, error) {
	var (
		a   APIClient
		err error
	)
	a.Username = username
	a.Password = password
	a.HTTPClient = &http.Client{}
	a.URL, err = url.Parse(addr)
	return &a, err
}

// Get implements Storage
func (ac *APIClient) Get(ids ...model.NoteID) ([]model.Note, error) {
	var req *http.Request
	if len(ids) > 0 {
		req = ac.formRequest(http.MethodGet, map[string]string{"action": "get", "id": ids[0].String()}, nil)
	} else {
		req = ac.formRequest(http.MethodGet, map[string]string{"action": "get"}, nil)
	}
	data, err := ac.doRequest(req, http.StatusOK)
	if err != nil {
		return []model.Note{}, err
	}
	return bytesToNoteList(data)
}

// Put implements Storage
func (ac *APIClient) Put(n model.Note) (model.NoteID, error) {
	body := bytes.NewReader(noteToBytes(n))
	req := ac.formRequest(http.MethodPut, map[string]string{"action": "new"}, body)
	data, err := ac.doRequest(req, http.StatusCreated)
	if err != nil {
		return BadNoteID, err
	}
	return model.NoteID(data), nil
}

// Upd implements Storage
func (ac *APIClient) Upd(id model.NoteID, n model.Note) (model.NoteID, error) {
	body := bytes.NewReader(noteToBytes(n))
	req := ac.formRequest(http.MethodPost, map[string]string{"action": "upd", "id": id.String()}, body)
	data, err := ac.doRequest(req, http.StatusAccepted)
	if err != nil {
		return BadNoteID, err
	}
	return model.NoteID(data), nil
}

// Del implements Storage
func (ac *APIClient) Del(id model.NoteID) error {
	req := ac.formRequest(http.MethodDelete, map[string]string{"action": "del", "id": id.String()}, nil)
	_, err := ac.doRequest(req, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}

// Search implements Storage
func (ac *APIClient) Search(query string) ([]model.Note, error) {
	req := ac.formRequest(http.MethodGet, map[string]string{"action": "search", "q": query}, nil)
	data, err := ac.doRequest(req, http.StatusOK)
	if err != nil {
		return []model.Note{}, err
	}
	return bytesToNoteList(data)
}

//ExportJSON implements Storage
func (ac *APIClient) ExportJSON() ([]byte, error) {
	req := ac.formRequest(http.MethodGet, map[string]string{"action": "get"}, nil)
	data, err := ac.doRequest(req, http.StatusOK)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

// Close implements Storage
func (ac *APIClient) Close() error {
	return nil
}

func (ac *APIClient) formRequest(method string, params map[string]string, body io.Reader) *http.Request {
	url := ac.formUrlFromMap(params)
	req, _ := http.NewRequest(method, url.String(), body)
	req.Header.Add("Notepet-Token", ac.Token)
	return req
}

func (ac *APIClient) doRequest(r *http.Request, needstatus int) ([]byte, error) {
	resp, err := ac.HTTPClient.Do(r)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != needstatus {
		return []byte{}, fmt.Errorf("server returned: %v", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (ac *APIClient) formUrlFromMap(params map[string]string) url.URL {
	url := ac.URL
	q := url.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	url.RawQuery = q.Encode()
	return url
}
