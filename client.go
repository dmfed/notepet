package notepet

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// APIClient represents client fetching notes from notepet server
// It implements Storage interface
type APIClient struct {
	Token      string
	HTTPClient *http.Client
	URL        url.URL
}

// NewAPIClient returns instance of APIClient configured
// to send requests to specified ip address
func NewAPIClient(ip, port, token string) (Storage, error) {
	var c APIClient
	c.Token = token
	c.URL = url.URL{Scheme: "http",
		Host: ip + ":" + port,
		Path: "/api"}
	c.HTTPClient = &http.Client{}
	return &c, nil
}

// Get implements Storage
func (ac *APIClient) Get(ids ...NoteID) ([]Note, error) {
	url := ac.URL
	q := url.Query()
	q.Add("action", "get")
	if len(ids) > 0 {
		q.Add("id", ids[0].String())
	}
	url.RawQuery = q.Encode()
	req, _ := http.NewRequest("GET", url.String(), nil)
	req.Header.Add("Notepet-Token", ac.Token)
	resp, err := ac.HTTPClient.Do(req)
	if err != nil {
		return []Note{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []Note{}, fmt.Errorf("server returned: %v", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Note{}, err
	}
	return bytesToNoteList(data)
}

// Put implements Storage
func (ac *APIClient) Put(n Note) (NoteID, error) {
	url := ac.URL
	q := url.Query()
	q.Add("action", "new")
	url.RawQuery = q.Encode()
	data := noteToBytes(n)
	body := io.NopCloser(bytes.NewReader(data))
	req, _ := http.NewRequest("PUT", url.String(), body)
	req.Header.Add("Notepet-Token", ac.Token)
	resp, err := ac.HTTPClient.Do(req)
	if err != nil {
		return BadNoteID, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return BadNoteID, fmt.Errorf("server returned: %v", resp.Status)
	}
	id, err := io.ReadAll(resp.Body)
	if err != nil {
		return BadNoteID, err
	}
	return NoteID(id), nil
}

// Upd implements Storage
func (ac *APIClient) Upd(id NoteID, n Note) (NoteID, error) {
	url := ac.URL
	q := url.Query()
	q.Add("action", "upd")
	q.Add("id", id.String())
	url.RawQuery = q.Encode()
	data := noteToBytes(n)
	body := io.NopCloser(bytes.NewReader(data))
	req, _ := http.NewRequest("POST", url.String(), body)
	req.Header.Add("Notepet-Token", ac.Token)
	resp, err := ac.HTTPClient.Do(req)
	if err != nil {
		return BadNoteID, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return BadNoteID, fmt.Errorf("server returned: %v", resp.Status)
	}
	returnid, err := io.ReadAll(resp.Body)
	if err != nil {
		return BadNoteID, err
	}
	return NoteID(returnid), nil
}

// Del implements Storage
func (ac *APIClient) Del(id NoteID) error {
	url := ac.URL
	q := url.Query()
	q.Add("action", "del")
	q.Add("id", id.String())
	url.RawQuery = q.Encode()
	req, _ := http.NewRequest("DELETE", url.String(), nil)
	req.Header.Add("Notepet-Token", ac.Token)
	resp, err := ac.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", resp.Status)
	}
	return nil
}

// Search implements Storage
func (ac *APIClient) Search(query string) ([]Note, error) {
	url := ac.URL
	q := url.Query()
	q.Add("action", "search")
	q.Add("q", query)
	req, _ := http.NewRequest("GET", url.String(), nil)
	req.Header.Add("Notepet-Token", ac.Token)
	resp, err := ac.HTTPClient.Do(req)
	if err != nil {
		return []Note{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Note{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return []Note{}, fmt.Errorf("server returned: %v", resp.Status)
	}
	return bytesToNoteList(data)
}

//ExportJSON implements Storage
func (ac *APIClient) ExportJSON() ([]byte, error) {
	url := ac.URL
	q := url.Query()
	q.Add("action", "get")
	url.RawQuery = q.Encode()
	req, _ := http.NewRequest("GET", url.String(), nil)
	req.Header.Add("Notepet-Token", ac.Token)
	resp, err := ac.HTTPClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return []byte{}, fmt.Errorf("server returned: %v", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// Close implements Storage
func (ac *APIClient) Close() error {
	return nil
}
