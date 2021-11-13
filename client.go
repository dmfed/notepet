package notepet

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// APIClient represents http client fetching notes from notepet server.
// It implements Storage interface.
type APIClient struct {
	Token      string
	HTTPClient *http.Client
	URL        url.URL
}

// NewAPIClient returns instance of APIClient configured
// to send requests to specified ip address
func NewAPIClient(ip, port, path, apptoken string) (Storage, error) {
	var ac APIClient
	ac.Token = apptoken
	ac.HTTPClient = &http.Client{}
	ip = strings.TrimLeft(ip, "htps:/") // make sure hostname is not prefixed with http:// or https://
	ac.URL = url.URL{Scheme: "https",
		Host: ip + ":" + port,
		Path: "/" + strings.Trim(path, "/")}
	return &ac, nil
}

// Get implements Storage
func (ac *APIClient) Get(ids ...NoteID) ([]Note, error) {
	var req *http.Request
	if len(ids) > 0 {
		req = ac.formRequest(http.MethodGet, map[string]string{"action": "get", "id": ids[0].String()}, nil)
	} else {
		req = ac.formRequest(http.MethodGet, map[string]string{"action": "get"}, nil)
	}
	data, err := ac.doRequest(req, http.StatusOK)
	if err != nil {
		return []Note{}, err
	}
	return bytesToNoteList(data)
}

// Put implements Storage
func (ac *APIClient) Put(n Note) (NoteID, error) {
	body := bytes.NewReader(noteToBytes(n))
	req := ac.formRequest(http.MethodPut, map[string]string{"action": "new"}, body)
	data, err := ac.doRequest(req, http.StatusCreated)
	if err != nil {
		return BadNoteID, err
	}
	return NoteID(data), nil
}

// Upd implements Storage
func (ac *APIClient) Upd(id NoteID, n Note) (NoteID, error) {
	body := bytes.NewReader(noteToBytes(n))
	req := ac.formRequest(http.MethodPost, map[string]string{"action": "upd", "id": id.String()}, body)
	data, err := ac.doRequest(req, http.StatusAccepted)
	if err != nil {
		return BadNoteID, err
	}
	return NoteID(data), nil
}

// Del implements Storage
func (ac *APIClient) Del(id NoteID) error {
	req := ac.formRequest(http.MethodDelete, map[string]string{"action": "del", "id": id.String()}, nil)
	_, err := ac.doRequest(req, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}

// Search implements Storage
func (ac *APIClient) Search(query string) ([]Note, error) {
	req := ac.formRequest(http.MethodGet, map[string]string{"action": "search", "q": query}, nil)
	data, err := ac.doRequest(req, http.StatusOK)
	if err != nil {
		return []Note{}, err
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
