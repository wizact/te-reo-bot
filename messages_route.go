package main

import (
	"encoding/json"
	"net/http"

	"github.com/kelseyhightower/envconfig"
)

// PostMessage post a message to a specific social channel
func PostMessage(w http.ResponseWriter, r *http.Request) *AppError {
	var c TwitterCredential
	envconfig.Process("tereobot", &c)
	tc := NewTwitterClient(&c)
	tc.SendTweet("Hi")

	json.NewEncoder(w).Encode("OK")
	return nil
}

// TwitterCredential is a wrapper for consumer and access secrets
type TwitterCredential struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}
