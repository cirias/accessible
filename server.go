package accessible

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func handleCheck(w http.ResponseWriter, req *http.Request) {
	url := req.URL.Query().Get("url")
	r := Check(url)

	enc := json.NewEncoder(w)
	if err := enc.Encode(r); err != nil {
		http.Error(w, fmt.Sprintf("could not encode check result: %s", err), http.StatusInternalServerError)
	}
}

var CheckHandler = http.HandlerFunc(handleCheck)

func BasicAuthHandler(username, password string, h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		u, p, ok := req.BasicAuth()
		if !ok || u != username || p != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Access private resources"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, req)
	}

	return http.HandlerFunc(handler)
}
