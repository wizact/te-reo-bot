package handlers

import (
	"encoding/json"
	"net/http"

	ent "github.com/wizact/te-reo-bot/pkg/entities"
)

type HealthCheckRoute struct {
}

// GetHealthCheck returns OK when is called
func (hcr HealthCheckRoute) GetHealthCheck() appHandler {
	fn := func(w http.ResponseWriter, r *http.Request) *ent.AppError {
		json.NewEncoder(w).Encode("OK")
		return nil
	}

	return fn
}
