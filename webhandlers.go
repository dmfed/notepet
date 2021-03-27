package notepet

import "net/http"

func HandleWeb(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("web interface is currently down"))
}
