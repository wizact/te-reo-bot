package wotd

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/mattn/go-mastodon"
	ent "github.com/wizact/te-reo-bot/pkg/entities"
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

func Toot(wo *Word, w http.ResponseWriter) *ent.AppError {
	var att *mastodon.Attachment
	mids := []mastodon.ID{}
	tc := newMastodonClient()

	// check if the wo has a photo
	if hasMedia(wo) {
		media, err := acquireMedia(wo.Photo)
		if err != nil {
			return err
		}

		var e error
		att, e = tc.UploadMediaFromBytes(context.Background(), media)
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

func acquireMedia(fn string) ([]byte, *ent.AppError) {
	gsc, err := NewCloudStorageClient()

	if err != nil {
		return nil, err
	}

	media, err := GetObject(gsc, fn)

	if err != nil {
		return nil, err
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
