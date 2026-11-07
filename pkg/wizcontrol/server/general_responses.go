package server

import (
	"encoding/json"
	"net/http"
)

// GeneralResponse represents the standard structure for general API responses.
type GeneralResponse struct {
	Success bool    `json:"success"`
	Error   *string `json:"error,omitempty"`
}

// WriteError writes an error response with the given status code and message.
func WriteError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	response := GeneralResponse{
		Success: false,
		Error:   &message,
	}
	json.NewEncoder(w).Encode(response) // nolint:errcheck
}
