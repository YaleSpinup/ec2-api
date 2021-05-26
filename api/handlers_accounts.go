package api

import (
	"net/http"
)

// AccountsHandler responds to account list requests
func (s *server) AccountsHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not Implemented"))
}
