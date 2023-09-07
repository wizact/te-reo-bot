package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	ent "github.com/wizact/te-reo-bot/pkg/entities"
)

type HealthCheckRoute struct {
}

func (hcr HealthCheckRoute) SetupRoutes(routePath string, router *mux.Router) {
	router.Handle(routePath, appHandler(hcr.GetHealthCheck())).Methods("GET")
}

// GetHealthCheck returns OK when is called
func (hcr HealthCheckRoute) GetHealthCheck() appHandler {
	fn := func(w http.ResponseWriter, r *http.Request) *ent.AppError {
		json.NewEncoder(w).Encode("OK")
		return nil
	}

	return fn
}
