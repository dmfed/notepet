package notepet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	// ErrStorageIsNil is returned when nil pointer is passed to NewAPIHandler or
	// NewNotepetServer
	ErrStorageIsNil = errors.New("can not initialize handler: stirage is nil")
	// ErrNoNotesFound is returned when error occurs accessing notes storage or
	ErrNoNotesFound = errors.New("error: no notes found")
)

// APIHandler implements http.Handler ready to serve requests to API
type APIHandler struct {
	Storage Storage
	Tokens  map[string]struct{}
}

// NewAPIHandler returns instance of http.Handler ready to run
func NewAPIHandler(st Storage, tokens ...string) (*APIHandler, error) {
	if st == nil {
		return nil, ErrStorageIsNil
	}
	var handler APIHandler
	handler.Storage = st
	handler.Tokens = make(map[string]struct{})
	for _, token := range tokens {
		handler.Tokens[token] = struct{}{}
	}
	return &handler, nil
}

// NewNotepetServer returns instance of http.Server ready to run on ListenAndServe call
func NewNotepetServer(ip, port string, st Storage, tokens ...string) (*http.Server, error) {
	srv := &http.Server{Addr: ip + ":" + port}
	apihandler, err := NewAPIHandler(st, tokens...)
	if err != nil {
		return nil, err
	}
	http.Handle("/api", apihandler)
	// http.Handle("/notes", http.HandlerFunc(HandleWeb))
	//Handling OS signals
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-interrupts
		log.Printf("exiting on signal: %v\n", sig)

		if err := st.Close(); err != nil {
			log.Printf("error closing storage: %v\n", err)
		} else {
			log.Printf("storage closed")
		}

		/* ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel() */
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("server shutdown error: %v\n", err)
		} else {
			log.Println("shut down gracefully")
		}
	}()
	return srv, nil
}

// RegisterStorage makes APIHandler use the supplied Storage
func (ah *APIHandler) RegisterStorage(st Storage) error {
	if st == nil {
		return fmt.Errorf("can not register nil storage")
	}
	ah.Storage = st
	return nil
}

// RegisterToken adds token to globalValidTokens map
// so that server may use them for authentication
func (ah *APIHandler) RegisterToken(token string) {
	if ah.Tokens == nil {
		ah.Tokens = make(map[string]struct{})
	}
	ah.Tokens[token] = struct{}{}
}

// ServerHTTP implements http.Handler interface
func (ah *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handler http.HandlerFunc
	switch r.URL.Query().Get("action") {
	case "new":
		handler = methodPut(ah.authenticate(ah.handleAPINew))
	case "get":
		handler = methodGet(ah.authenticate(ah.handleAPIGet))
	case "upd":
		handler = methodPost(ah.authenticate(ah.handleAPIUpd))
	case "del":
		handler = methodDelete(ah.authenticate(ah.handleAPIDel))
	case "search":
		handler = methodGet(ah.authenticate(ah.handleAPISearch))
	default:
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	handler(w, r)
}

func (ah *APIHandler) authenticate(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Notepet-Token")
		if token == "" {
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}
		if _, ok := ah.Tokens[token]; !ok {
			http.Error(w, "403 Forbidden", http.StatusForbidden)
			return
		}
		h(w, r)
	}
}

func methodGet(h http.HandlerFunc) http.HandlerFunc {
	return allowMethod(h, "GET")
}

func methodPut(h http.HandlerFunc) http.HandlerFunc {
	return allowMethod(h, "PUT")
}

func methodPost(h http.HandlerFunc) http.HandlerFunc {
	return allowMethod(h, "POST")
}

func methodDelete(h http.HandlerFunc) http.HandlerFunc {
	return allowMethod(h, "DELETE")
}

// Handlers

func allowMethod(h http.HandlerFunc, method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			http.Error(w, "405 Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}

func (ah *APIHandler) handleAPINew(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "400 could not read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	note, err := bytesToNote(data)
	if err != nil {
		http.Error(w, "400 could not parse request body", http.StatusBadRequest)
		return
	}
	id, err := ah.Storage.Put(note)
	if err != nil {
		http.Error(w, "500 error putting note", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(201)
	w.Write([]byte(id.String()))
}

func (ah *APIHandler) handleAPIGet(w http.ResponseWriter, r *http.Request) {
	reqid := r.URL.Query().Get("id")
	// start, end
	var notes []Note
	var err error
	switch reqid {
	case "":
		notes, err = ah.Storage.Get()
	default:
		notes, err = ah.Storage.Get(NoteID(reqid))
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(noteListToBytes(notes))
}

func (ah *APIHandler) handleAPIUpd(w http.ResponseWriter, r *http.Request) {
	reqid := r.URL.Query().Get("id")
	if reqid == "" {
		http.Error(w, "400 no id requested", http.StatusBadRequest)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "400 could not read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	note, err := bytesToNote(data)
	if err != nil {
		http.Error(w, "400 could not parse request body", http.StatusBadRequest)
		return
	}
	newID, err := ah.Storage.Upd(NoteID(reqid), note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(202)
	w.Write([]byte(newID.String()))
}

func (ah *APIHandler) handleAPIDel(w http.ResponseWriter, r *http.Request) {
	reqid := r.URL.Query().Get("id")
	if reqid == "" {
		http.Error(w, "400 no id requested", http.StatusBadRequest)
		return
	}
	if err := ah.Storage.Del(NoteID(reqid)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("200 note deleted " + reqid))
}

func (ah *APIHandler) handleAPISearch(w http.ResponseWriter, r *http.Request) {
	searchquery := r.URL.Query().Get("q")
	if searchquery == "" {
		http.Error(w, "400 no search query provided", http.StatusBadRequest)
		return
	}
	notelist, err := ah.Storage.Search(searchquery)
	if err != nil || len(notelist) == 0 {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(noteListToBytes(notelist))
}

// HandleFavicon is intended to be used to handle request to /favicon.ico
func HandleFavicon(w http.ResponseWriter, r *http.Request) {

}
