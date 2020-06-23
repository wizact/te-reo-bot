package main

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
)

const (
	healthCheckRoute = "/__health-check"
	messagesRoute    = "/messages"

	apiKeyValue = "placeholder"
)

// StartServerCommand is struct for info required to start an http server
type StartServerCommand struct {
	port    string
	address string
}

// Flags returns the flag sets
func (fc *StartServerCommand) Flags() *flag.FlagSet {
	f := &flag.FlagSet{}

	f.StringVar(&fc.address, "address", "localhost", "-address=localhost")
	f.StringVar(&fc.port, "port", "8080", "-port=8080")

	return f
}

func (fc *StartServerCommand) getPort() string {
	return fc.port
}

func (fc *StartServerCommand) getAddress() string {
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
	serverAddress := fmt.Sprintf("%s:%s", fc.address, fc.port)
	fmt.Println("Listening to requests from: " + serverAddress)

	router := mux.NewRouter()
	router.Use(CommonMiddleware)

	router.Handle(healthCheckRoute, appHandler(GetHealthCheck)).Methods("GET")
	router.Handle(messagesRoute, appHandler(PostMessage)).Methods("POST")

	log.Fatal(http.ListenAndServe(serverAddress, router))

	return nil
}

// CommonMiddleware the generic middleware
func CommonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Index(r.RequestURI, healthCheckRoute) != 0 {
			rak, err := findCaseInsensitiveHeader("X-Api-Key", r)

			if err != nil {
				http.Error(w, "authentication failed", 401)
				return
			}

			if rak != apiKeyValue {
				http.Error(w, "authentication failed", 401)
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
		if strings.ToLower(s) == strings.ToLower(headerName) {
			apiKeyHeader := r.Header[headerName]
			if len(apiKeyHeader) > 0 {
				return apiKeyHeader[0], nil
			}
		}
	}

	return "", errors.New("auth header is missing")

}

type appHandler func(http.ResponseWriter, *http.Request) *AppError

// ServeHTTP to serve requests but respond with a friendly error message if any
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		ee := json.NewEncoder(w).Encode(&friendlyError{Message: e.Message})
		if ee != nil {
			log.Fatal(ee.Error())
		}
	}
}

// AppError as app error container
type AppError struct {
	Error   error  `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// friendlyError is sanitised error message sent back to the user
type friendlyError struct {
	Message string `json:"message"`
}
