package main

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"net/http"
)

// TwitterClient is a wrapper for twitter client implementation
type TwitterClient struct {
	client *twitter.Client
}

// NewTwitterClient returns an authenticated instance of Twitter client
func NewTwitterClient(credential *TwitterCredential) *TwitterClient {
	tc := &TwitterClient{}
	tc.authenticate(credential)

	return tc
}

func (tc *TwitterClient) authenticate(credential *TwitterCredential) {
	config := oauth1.NewConfig(credential.ConsumerKey, credential.ConsumerSecret)
	token := oauth1.NewToken(credential.AccessToken, credential.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	tc.client = twitter.NewClient(httpClient)
}

// SendTweet updates the authenticated account with a new tweet
func (tc *TwitterClient) SendTweet(message string) (*twitter.Tweet, *http.Response, error) {
	t, r ,e := tc.client.Statuses.Update(message, nil)
	return t, r, e
}
