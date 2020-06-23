package main

import (
	"encoding/json"
	"net/http"
)

// PostMessage post a message to a specific social channel
func PostMessage(w http.ResponseWriter, r *http.Request) *AppError {
	json.NewEncoder(w).Encode("OK")
	return nil
}
