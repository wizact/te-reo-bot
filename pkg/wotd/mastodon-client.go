package wotd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/mattn/go-mastodon"
	ent "github.com/wizact/te-reo-bot/pkg/entities"
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
			return &ent.AppError{Error: e, Code: 500, Message: "Failed sending the toot with media"}
		}
	}

	if att != nil && len(att.ID) > 0 {
		mids = []mastodon.ID{att.ID}
	}

	ms, e := tc.PostStatus(context.Background(), &mastodon.Toot{Status: wo.Word + ": " + wo.Meaning + " #aotearoa #newzealand", MediaIDs: mids})

	if e == nil {
		json.NewEncoder(w).Encode(&ent.PostResponse{TootId: string(ms.ID)})
		return nil
	} else {
		return &ent.AppError{Error: e, Code: 500, Message: "Failed sending the toot"}
	}
}

func acquireMedia(bucketName, objectName string) ([]byte, *ent.AppError) {

	var cscw gcs.GoogleCloudStorageClientWrapper
	err := cscw.Client(context.Background())

	if err != nil {
		return nil, &ent.AppError{Error: err, Code: 500, Message: "Failed to acquire image"}
	}

	media, err := cscw.GetObject(context.Background(), bucketName, objectName)

	if err != nil {
		return nil, &ent.AppError{Error: err, Code: 500, Message: "Failed to acquire image"}
	}

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
