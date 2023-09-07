package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
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

// StartServerCommand is struct for info required to start an http server
type StartServerCommand struct {
	port    string
	address string
	tls     bool
}

// Flags returns the flag sets
func (fc *StartServerCommand) Flags() *flag.FlagSet {
	f := &flag.FlagSet{}

	f.StringVar(&fc.address, "address", "localhost", "-address=localhost")
	f.StringVar(&fc.port, "port", "8080", "-port=8080")
	f.BoolVar(&fc.tls, "tls", false, "-tls=true")

	return f
}

func (fc *StartServerCommand) Port() string {
	return fc.port
}

func (fc *StartServerCommand) Address() string {
	return fc.address
}

// Name of the command
func (fc *StartServerCommand) Name() string {
	return "start-server"
}

// HelpString is the string shown as usage
func (fc *StartServerCommand) HelpString() string {
	return "Start the server using provided address and port"
}

// Run a command
func (fc *StartServerCommand) Run(ctx context.Context, args []string) error {
	var serverAddress string
	if fc.address == "localhost" {
		fc.address = ""
	}

	serverAddress = fmt.Sprintf("%s:%s", fc.address, fc.port)

	fmt.Println("Listening to requests from: " + serverAddress)

	router := mux.NewRouter()
	router.Use(CommonMiddleware)

	// HealthCheck route setup
	hcr := HealthCheckRoute{}
	hcr.SetupRoutes(healthCheckRoute, router)

	// MessageRoute route setup
	bn, err := getMediaBucketName()
	if err != nil {
		log.Fatal("Cannot get the bucket name from environment variables")
	}

	mr := MessagesRoute{bucketName: bn}
	mr.SetupRoutes(messagesRoute, router)

	if fc.tls {
		log.Fatal(http.ListenAndServeTLS(serverAddress,
			"certs/server.crt",
			"certs/server.key",
			router))
	} else {
		log.Fatal(http.ListenAndServe(serverAddress, router))
	}

	return nil
}

// CommonMiddleware the generic middleware
func CommonMiddleware(next http.Handler) http.Handler {
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
		ee := json.NewEncoder(w).Encode(&friendlyError{Message: e.Message})
		if ee != nil {
			log.Fatal(ee.Error())
		}
	}
}

// friendlyError is sanitised error message sent back to the user
type friendlyError struct {
	Message string `json:"message"`
}

// ServerConfig to wrap configuration
type ServerConfig struct {
	ApiKey string
}

// StorageConfig stores information required for storage service
type StorageConfig struct {
	BucketName string
}

func getMediaBucketName() (string, error) {
	var s StorageConfig
	err := envconfig.Process("tereobot", &s)
	if err != nil {
		return "nil", err
	}

	return s.BucketName, nil
}
