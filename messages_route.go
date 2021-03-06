package main

import (
	"encoding/json"
	"net/http"

	"github.com/kelseyhightower/envconfig"
)

//PostMessage post a message to a specific social channel
func PostMessage(w http.ResponseWriter, r *http.Request) *AppError {
	var c TwitterCredential
	envconfig.Process("tereobot", &c)
	tc := NewTwitterClient(&c)

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

	t, tr, e := tc.SendTweet(wo.Word + " : " + wo.Meaning)

	if e == nil {
		json.NewEncoder(w).Encode(&TwitterResponse{TwitterId: t.IDStr})
		return nil
	} else {
		return &AppError{Error: e, Code: tr.StatusCode, Message: "Failed sending the tweet"}
	}
}

// TwitterResponse is the tweet Id after a successful update operation
type TwitterResponse struct {
	TwitterId string `json:"tweetId"`
}

// TwitterCredential is a wrapper for consumer and access secrets
type TwitterCredential struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}
