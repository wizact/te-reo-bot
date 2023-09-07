package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	ent "github.com/wizact/te-reo-bot/pkg/entities"
)

const (
	healthCheckRoute = "/__health-check"
	messagesRoute    = "/messages"
)

// StartServer starts the http server
func StartServer(address, port string, tls bool) {
	serverAddress := fmt.Sprintf("%s:%s", address, port)

	fmt.Println("Listening to requests from: " + serverAddress)

	router := mux.NewRouter()
	router.Use(commonMiddleware)

	// HealthCheck route setup
	hcr := HealthCheckRoute{}
	hcr.SetupRoutes(healthCheckRoute, router)

	// MessageRoute route setup
	bn, err := (&StorageConfig{}).BucketName()
	if err != nil {
		log.Fatal("Cannot get the bucket name from environment variables")
	}

	mr := MessagesRoute{bucketName: bn}
	mr.SetupRoutes(messagesRoute, router)

	if tls {
		log.Fatal(http.ListenAndServeTLS(serverAddress,
			"certs/server.crt",
			"certs/server.key",
			router))
	} else {
		log.Fatal(http.ListenAndServe(serverAddress, router))
	}
}

// commonMiddleware the generic middleware
func commonMiddleware(next http.Handler) http.Handler {
	var s ServerConfig
	err := envconfig.Process("tereobot", &s)

	if err != nil {
		panic("Cannot read configuration")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Index(r.RequestURI, healthCheckRoute) != 0 {
			rak, err := findCaseInsensitiveHeader("X-Api-Key", r)

			if err != nil {
				http.Error(w, "authentication failed", http.StatusUnauthorized)
				return
			}

			if rak != s.ApiKey {
				http.Error(w, "authentication failed", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func findCaseInsensitiveHeader(headerName string, r *http.Request) (string, error) {
	if strings.Trim(headerName, "") == "" {
		return "", errors.New("auth header is missing")
	}

	for s := range r.Header {
		if strings.EqualFold(s, headerName) {
			apiKeyHeader := r.Header[headerName]
			if len(apiKeyHeader) > 0 {
				return apiKeyHeader[0], nil
			}
		}
	}

	return "", errors.New("auth header is missing")

}

type appHandler func(http.ResponseWriter, *http.Request) *ent.AppError

// ServeHTTP to serve requests but respond with a friendly error message if any
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.

		fmt.Println(e.Error)

		w.WriteHeader(e.Code)
		ee := json.NewEncoder(w).Encode(&ent.FriendlyError{Message: e.Message})
		if ee != nil {
			log.Fatal(ee.Error())
		}
	}
}

// ServerConfig to wrap configuration
type ServerConfig struct {
	ApiKey string
}

// StorageConfig stores information required for storage service
type StorageConfig struct {
	bucketName string
}

func (s *StorageConfig) BucketName() (string, error) {
	err := envconfig.Process("tereobot", s)
	if err != nil {
		return "nil", err
	}

	return s.bucketName, nil
}
