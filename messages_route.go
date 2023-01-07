package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// PostMessage post a message to a specific social channel
func PostMessage(w http.ResponseWriter, r *http.Request) *AppError {
	ws := WordSelector{}
	f, erf := ws.ReadFile()

	if erf != nil {
		return &AppError{Error: erf, Code: 500, Message: "Failed sending the tweet"}
	}

	d, epf := ws.ParseFile(f)
	if epf != nil {
		return &AppError{Error: epf, Code: 500, Message: "Failed sending the tweet"}
	}

	wo := ws.SelectWordByDay(d.Words)

	dest := r.URL.Query().Get("dest")
	if strings.ToLower(dest) == "twitter" {
		return tweet(wo, w)
	} else if strings.ToLower(dest) == "mastodon" {
		return toot(wo, w)
	} else {
		json.NewEncoder(w).Encode(&PostResponse{Message: "No destination has been selected"})
		return nil
	}
}

// PostResponse is the tweet/mastodon Id after a successful update operation
type PostResponse struct {
	TwitterId string `json:"tweetId"`
	TootId    string `json:"tootId"`
	Message   string `json:"message"`
}
