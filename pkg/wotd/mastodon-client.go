package wotd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/mattn/go-mastodon"
	ent "github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
	gcs "github.com/wizact/te-reo-bot/pkg/storage"
)

type MastodonClient struct {
	mastodonServerName  string
	mastodonClientID    string
	mastodonAccessToken string
}

func (mclient *MastodonClient) NewClient() *MastodonClient {
	var mc MastodonCredential
	envconfig.Process("tereobot", &mc)

	mclient.mastodonServerName = mc.MastodonServerName
	mclient.mastodonClientID = mc.MastodonClientID
	mclient.mastodonAccessToken = mc.MastodonAccessToken

	return mclient
}

// getLogger returns the global logger instance
func (mclient *MastodonClient) getLogger() logger.Logger {
	return logger.GetGlobalLogger()
}

func (mclient *MastodonClient) client() *mastodon.Client {
	c := mastodon.NewClient(&mastodon.Config{
		Server:      mclient.mastodonServerName,
		ClientID:    mclient.mastodonClientID,
		AccessToken: mclient.mastodonAccessToken,
	})

	return c
}

func (mclient *MastodonClient) Toot(wo *Word, w http.ResponseWriter, bucketName string) *ent.AppError {
	var att *mastodon.Attachment
	mids := []mastodon.ID{}
	tc := mclient.client()
	tootContent := wo.Word + ": " + wo.Meaning

	// check if the wo has a photo
	if hasMedia(wo) {
		media, err := acquireMedia(bucketName, wo.Photo)
		if err != nil {
			return err
		}

		var e error
		if wo.Attribution != "" {
			att, e = tc.UploadMediaFromMedia(context.Background(), &mastodon.Media{File: bytes.NewReader(media), Description: wo.Attribution})
		} else {
			att, e = tc.UploadMediaFromBytes(context.Background(), media)
		}

		if e != nil {
			// Create enhanced AppError with context
			appErr := ent.NewAppError(e, 500, "Failed sending the toot with media")
			appErr.WithContext("word", wo.Word)
			appErr.WithContext("toot_content", tootContent)
			appErr.WithContext("bucket_name", bucketName)
			appErr.WithContext("photo", wo.Photo)
			appErr.WithContext("attribution", wo.Attribution)
			appErr.WithContext("operation", "mastodon_media_upload")

			// Log the error with stack trace and context
			mclient.getLogger().ErrorWithStack(e, "Failed to upload media to Mastodon",
				logger.String("word", wo.Word),
				logger.String("toot_content", tootContent),
				logger.String("bucket_name", bucketName),
				logger.String("photo", wo.Photo),
				logger.String("attribution", wo.Attribution),
				logger.String("operation", "mastodon_media_upload"),
			)

			return appErr
		}

		// Log successful media upload
		mclient.getLogger().Debug("Successfully uploaded media to Mastodon",
			logger.String("word", wo.Word),
			logger.String("photo", wo.Photo),
			logger.String("attribution", wo.Attribution),
			logger.String("attachment_id", string(att.ID)),
			logger.String("operation", "mastodon_media_upload"),
		)
	}

	if att != nil && len(att.ID) > 0 {
		mids = []mastodon.ID{att.ID}
	}

	ms, e := tc.PostStatus(context.Background(), &mastodon.Toot{Status: tootContent, MediaIDs: mids})

	if e == nil {
		// Log successful toot
		mclient.getLogger().Info("Successfully sent toot to Mastodon",
			logger.String("toot_id", string(ms.ID)),
			logger.String("word", wo.Word),
			logger.String("toot_content", tootContent),
			logger.Bool("has_media", len(mids) > 0),
			logger.String("operation", "mastodon_post"),
		)
		json.NewEncoder(w).Encode(&ent.PostResponse{TootId: string(ms.ID)})
		return nil
	} else {
		// Create enhanced AppError with context
		appErr := ent.NewAppError(e, 500, "Failed sending the toot")
		appErr.WithContext("word", wo.Word)
		appErr.WithContext("toot_content", tootContent)
		appErr.WithContext("has_media", len(mids) > 0)
		appErr.WithContext("operation", "mastodon_post")

		// Log the error with stack trace and context
		mclient.getLogger().ErrorWithStack(e, "Failed to send toot to Mastodon API",
			logger.String("word", wo.Word),
			logger.String("toot_content", tootContent),
			logger.Bool("has_media", len(mids) > 0),
			logger.String("operation", "mastodon_post"),
		)

		return appErr
	}
}

func acquireMedia(bucketName, objectName string) ([]byte, *ent.AppError) {
	log := logger.GetGlobalLogger()
	cscw := gcs.NewGoogleCloudStorageClientWrapper(log)
	err := cscw.Client(context.Background())

	if err != nil {
		// Check if it's already an AppError, if not wrap it
		if appErr, ok := err.(*ent.AppError); ok {
			return nil, appErr
		}

		// Create enhanced AppError with context
		appErr := ent.NewAppError(err, 500, "Failed to acquire image")
		appErr.WithContext("bucket_name", bucketName)
		appErr.WithContext("object_name", objectName)
		appErr.WithContext("operation", "gcs_client_init")

		// Log the error with stack trace and context
		log.ErrorWithStack(err, "Failed to initialize Google Cloud Storage client",
			logger.String("bucket_name", bucketName),
			logger.String("object_name", objectName),
			logger.String("operation", "gcs_client_init"),
		)

		return nil, appErr
	}

	media, err := cscw.GetObject(context.Background(), bucketName, objectName)

	if err != nil {
		// Check if it's already an AppError, if not wrap it
		if appErr, ok := err.(*ent.AppError); ok {
			return nil, appErr
		}

		// Create enhanced AppError with context
		appErr := ent.NewAppError(err, 500, "Failed to acquire image")
		appErr.WithContext("bucket_name", bucketName)
		appErr.WithContext("object_name", objectName)
		appErr.WithContext("operation", "gcs_get_object")

		// Log the error with stack trace and context
		log.ErrorWithStack(err, "Failed to get object from Google Cloud Storage",
			logger.String("bucket_name", bucketName),
			logger.String("object_name", objectName),
			logger.String("operation", "gcs_get_object"),
		)

		return nil, appErr
	}

	// Log successful media acquisition
	log.Debug("Successfully acquired media from Google Cloud Storage",
		logger.String("bucket_name", bucketName),
		logger.String("object_name", objectName),
		logger.Int("media_size", len(media)),
		logger.String("operation", "gcs_get_object"),
	)

	return media, nil
}

func hasMedia(wo *Word) bool {
	return len(wo.Photo) > 0
}

// MastodonCredential is a wrapper for consumer and access secrets
type MastodonCredential struct {
	MastodonServerName  string
	MastodonClientID    string
	MastodonAccessToken string
}
