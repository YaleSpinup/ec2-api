package api

import (
	"net/http"
)

// AccountsHandler responds to account list requests
func (s *server) AccountsHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}

	accounts := []string{}
	for k := range s.accountsMap {
		accounts = append(accounts, k)
	}

	handleResponseOk(w, accounts)
}
