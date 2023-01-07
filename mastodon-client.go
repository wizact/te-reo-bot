package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/mattn/go-mastodon"
)

func newMastodonClient() *mastodon.Client {
	var mc MastodonCredential
	envconfig.Process("tereobot", &mc)

	var serverName string = mc.MastodonServerName
	var clientName string = mc.MastodonClientID
	var accessToken string = mc.MastodonAccessToken

	c := mastodon.NewClient(&mastodon.Config{
		Server:      serverName,
		ClientID:    clientName,
		AccessToken: accessToken,
	})

	return c
}

func toot(wo *Word, w http.ResponseWriter) *AppError {

	tc := newMastodonClient()

	ms, e := tc.PostStatus(context.Background(), &mastodon.Toot{Status: wo.Word + ": " + wo.Meaning})

	if e == nil {
		json.NewEncoder(w).Encode(&PostResponse{TootId: string(ms.ID)})
		return nil
	} else {
		return &AppError{Error: e, Code: 500, Message: "Failed sending the toot"}
	}
}

// MastodonCredential is a wrapper for consumer and access secrets
type MastodonCredential struct {
	MastodonServerName  string
	MastodonClientID    string
	MastodonAccessToken string
}
