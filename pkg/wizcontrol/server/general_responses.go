package server

import (
	"encoding/json"
	"net/http"
)

type GeneralResponse struct {
	Success bool    `json:"success"`
	Error   *string `json:"error,omitempty"`
}

func WriteError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	response := GeneralResponse{
		Success: false,
		Error:   &message,
	}
	json.NewEncoder(w).Encode(response) // nolint:errcheck
}

func WriteSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	response := GeneralResponse{
		Success: true,
		Error:   nil,
	}
	json.NewEncoder(w).Encode(response) // nolint:errcheck
}
