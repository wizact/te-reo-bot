package wotd

import (
	"encoding/json"
	"net/http"
)

// GetHealthCheck get all the facts
func GetHealthCheck(w http.ResponseWriter, r *http.Request) *AppError {
	json.NewEncoder(w).Encode("OK")
	return nil
}
