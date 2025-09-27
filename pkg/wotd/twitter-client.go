package wotd

import (
	"encoding/json"
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/kelseyhightower/envconfig"
	ent "github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
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

// getLogger returns the global logger instance
func (tc *TwitterClient) getLogger() logger.Logger {
	return logger.GetGlobalLogger()
}

func Tweet(wo *Word, w http.ResponseWriter) *ent.AppError {
	var c TwitterCredential
	envconfig.Process("tereobot", &c)
	tc := NewTwitterClient(&c)
	log := logger.GetGlobalLogger()

	message := wo.Word + " : " + wo.Meaning
	t, tr, e := tc.SendTweet(message)

	if e == nil {
		// Log successful tweet
		log.Info("Successfully sent tweet",
			logger.String("twitter_id", t.IDStr),
			logger.String("word", wo.Word),
			logger.String("message", message),
			logger.Int("status_code", tr.StatusCode),
		)
		json.NewEncoder(w).Encode(&ent.PostResponse{TwitterId: t.IDStr})
		return nil
	} else {
		// Create enhanced AppError with context
		appErr := ent.NewAppError(e, tr.StatusCode, "Failed sending the tweet")
		appErr.WithContext("word", wo.Word)
		appErr.WithContext("message", message)
		appErr.WithContext("status_code", tr.StatusCode)
		appErr.WithContext("operation", "twitter_post")

		// Log the error with stack trace and context
		log.ErrorWithStack(e, "Failed to send tweet to Twitter API",
			logger.String("word", wo.Word),
			logger.String("message", message),
			logger.Int("status_code", tr.StatusCode),
			logger.String("operation", "twitter_post"),
		)

		return appErr
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

	if e != nil {
		// Log the API error with contextual information
		tc.getLogger().ErrorWithStack(e, "Twitter API call failed",
			logger.String("message", message),
			logger.Int("status_code", r.StatusCode),
			logger.String("operation", "statuses_update"),
		)
	} else {
		// Log successful API call
		tc.getLogger().Debug("Twitter API call successful",
			logger.String("message", message),
			logger.String("tweet_id", t.IDStr),
			logger.Int("status_code", r.StatusCode),
			logger.String("operation", "statuses_update"),
		)
	}

	return t, r, e
}
