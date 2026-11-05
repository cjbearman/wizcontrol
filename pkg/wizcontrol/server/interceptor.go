package server

import (
	"net/http"
	"strings"
)

func intercept(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// All responses get no-cache headers
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// If auth is configured, authenticate the request
		if AuthUsername != "" && AuthPassword != "" {
			username, password, ok := r.BasicAuth()
			if !ok || username != AuthUsername || password != AuthPassword {
				w.Header().Set("WWW-Authenticate", `Basic realm="Wizcontrol"`)
				WriteError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
		}

		// We only send JSON
		w.Header().Set("Content-Type", "application/json")

		// Do we have an Accept header?
		if accept := r.Header.Get("Accept"); accept != "" {
			accept = strings.ToLower(accept)

			if !strings.Contains(accept, "application/json") && !strings.Contains(accept, "*/*") {
				WriteError(w, http.StatusNotAcceptable, "Accept header must allow application/json")
				return
			}
		}

		// How about content-type
		if r.Method != http.MethodGet {
			if contentType := r.Header.Get("Content-Type"); contentType != "application/json" && contentType != "" {
				WriteError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
				return
			}
		}

		next(w, r)
	}
}
