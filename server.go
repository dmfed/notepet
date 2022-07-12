package notepet

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

// APIHandler implements http.Handler ready to serve requests to API
type APIHandler struct {
	Storage Storage
}

// NewAPIHandler returns instance of http.Handler ready to run
func NewHTTPHandler(st Storage) http.Handler {
	return &APIHandler{st}
}

// NewNotepetServer returns instance of http.Server ready to listen on
// addr and run on ListenAndServe call.
func NewNotepetServer(addr string, st Storage) (*http.Server, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("Invalid URL provided as addr: %w", err)
	}
	srv := &http.Server{Addr: u.Host}
	http.Handle(u.Path, NewHTTPHandler(st))

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

// ServerHTTP implements http.Handler interface
func (ah *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handler http.HandlerFunc
	switch r.URL.Query().Get("action") {
	case "new":
		handler = methodPut(ah.handleAPINew)
	case "get":
		handler = methodGet(ah.handleAPIGet)
	case "upd":
		handler = methodPost(ah.handleAPIUpd)
	case "del":
		handler = methodDelete(ah.handleAPIDel)
	case "search":
		handler = methodGet(ah.handleAPISearch)
	default:
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	handler(w, r)
}

func methodGet(h http.HandlerFunc) http.HandlerFunc {
	return allowMethod(h, http.MethodGet)
}

func methodPut(h http.HandlerFunc) http.HandlerFunc {
	return allowMethod(h, http.MethodPut)
}

func methodPost(h http.HandlerFunc) http.HandlerFunc {
	return allowMethod(h, http.MethodPost)
}

func methodDelete(h http.HandlerFunc) http.HandlerFunc {
	return allowMethod(h, http.MethodDelete)
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
	w.WriteHeader(http.StatusOK)
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
	w.WriteHeader(http.StatusAccepted)
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
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 note deleted " + reqid))
}

func (ah *APIHandler) handleAPISearch(w http.ResponseWriter, r *http.Request) {
	searchquery := r.URL.Query().Get("q")
	if searchquery == "" {
		http.Error(w, "400 no search query provided", http.StatusBadRequest)
		return
	}
	notelist, err := ah.Storage.Search(searchquery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(noteListToBytes(notelist))
}

// HandleFavicon is intended to be used to handle request to /favicon.ico
func handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Bad luck this time, Chrome :)")
}
