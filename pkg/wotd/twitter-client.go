package wotd

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/kelseyhightower/envconfig"
	ent "github.com/wizact/te-reo-bot/pkg/entities"
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

func Tweet(wo *Word, w http.ResponseWriter) *ent.AppError {
	var c TwitterCredential
	envconfig.Process("tereobot", &c)
	tc := NewTwitterClient(&c)

	t, tr, e := tc.SendTweet(wo.Word + " : " + wo.Meaning)

	if e == nil {
		json.NewEncoder(w).Encode(&ent.PostResponse{TwitterId: t.IDStr})
		return nil
	} else {
		return &ent.AppError{Error: e, Code: tr.StatusCode, Message: "Failed sending the tweet"}
	}
}

// TwitterCredential is a wrapper for consumer and access secrets
type TwitterCredential struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
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
	t, r, e := tc.client.Statuses.Update(message, nil)
	log.Println(r.Body, r.StatusCode, e.Error())
	return t, r, e
}
