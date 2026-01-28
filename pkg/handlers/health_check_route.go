package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	ent "github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
)

type HealthCheckRoute struct {
	logger logger.Logger
}

// NewHealthCheckRoute creates a new HealthCheckRoute
func NewHealthCheckRoute() *HealthCheckRoute {
	return &HealthCheckRoute{
		logger: logger.GetGlobalLogger(),
	}
}

func (hcr HealthCheckRoute) SetupRoutes(routePath string, router *mux.Router) {
	router.Handle(routePath, appHandler(hcr.GetHealthCheck())).Methods("GET")
}

// GetHealthCheck returns OK when is called
func (hcr HealthCheckRoute) GetHealthCheck() appHandler {
	fn := func(w http.ResponseWriter, r *http.Request) *ent.AppError {
		reqCtx := logger.ExtractRequestContext(r)
		hcr.logger.Debug("Health check request", reqCtx.ToFields()...)
		json.NewEncoder(w).Encode("OK")
		return nil
	}

	return fn
}
