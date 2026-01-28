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
	log := logger.GetGlobalLogger()

	err := envconfig.Process("tereobot", &c)
	if err != nil {
		appErr := ent.NewAppErrorWithContexts(err, 500, "Failed to load Twitter config", map[string]interface{}{
			"operation":     "load_twitter_config",
			"config_prefix": "tereobot",
		})
		log.ErrorWithStack(err, "Config error",
			logger.String("operation", "load_twitter_config"),
			logger.String("config_prefix", "tereobot"))
		return appErr
	}

	// Validate required fields
	if c.ConsumerKey == "" || c.ConsumerSecret == "" || c.AccessToken == "" || c.AccessSecret == "" {
		appErr := ent.NewAppErrorWithContexts(nil, 500, "Missing Twitter credentials", map[string]interface{}{
			"operation":           "validate_twitter_config",
			"has_consumer_key":    c.ConsumerKey != "",
			"has_consumer_secret": c.ConsumerSecret != "",
			"has_access_token":    c.AccessToken != "",
			"has_access_secret":   c.AccessSecret != "",
		})
		log.Error(appErr, "Credentials missing",
			logger.String("operation", "validate_twitter_config"),
			logger.Bool("has_consumer_key", c.ConsumerKey != ""),
			logger.Bool("has_consumer_secret", c.ConsumerSecret != ""),
			logger.Bool("has_access_token", c.AccessToken != ""),
			logger.Bool("has_access_secret", c.AccessSecret != ""))
		return appErr
	}

	log.Debug("Twitter config loaded", logger.Bool("has_credentials", true))

	tc := NewTwitterClient(&c)

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
		appErr := ent.NewAppErrorWithContexts(e, tr.StatusCode, "Failed sending the tweet", map[string]interface{}{
			"word":        wo.Word,
			"message":     message,
			"status_code": tr.StatusCode,
			"operation":   "twitter_post",
		})

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
