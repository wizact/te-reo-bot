package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	ent "github.com/wizact/te-reo-bot/pkg/entities"
	gcs "github.com/wizact/te-reo-bot/pkg/storage"
	wotd "github.com/wizact/te-reo-bot/pkg/wotd"
)

type MessagesRoute struct {
	bucketName string
}

func (m MessagesRoute) SetupRoutes(routePath string, router *mux.Router) {
	router.Handle(routePath, appHandler(m.PostMessage())).Methods("POST")
	router.Handle(routePath, appHandler(m.GetImage())).Methods("GET")
}

// PostMessage post a message to a specific social channel
func (m MessagesRoute) PostMessage() appHandler {
	fn := func(w http.ResponseWriter, r *http.Request) *ent.AppError {
		// Create WordSelector
		ws := wotd.NewWordSelector()
		f, erf := ws.ReadFile("./dictionary.json")

		if erf != nil {
			// Check if it's already an AppError, if not wrap it
			if appErr, ok := erf.(*ent.AppError); ok {
				return appErr
			}
			return &ent.AppError{Err: erf, Code: 500, Message: "Failed sending the word of the day"}
		}

		d, epf := ws.ParseFile(f, "./dictionary.json")
		if epf != nil {
			// Check if it's already an AppError, if not wrap it
			if appErr, ok := epf.(*ent.AppError); ok {
				return appErr
			}
			return &ent.AppError{Err: epf, Code: 500, Message: "Failed sending the word of the day"}
		}

		var wo *wotd.Word
		var err error
		wordIndex := r.URL.Query().Get("wordIndex")
		if wind, eind := strconv.Atoi(wordIndex); eind == nil {
			wo, err = ws.SelectWordByIndex(d.Words, wind)
		} else {
			wo, err = ws.SelectWordByDay(d.Words)
		}

		if err != nil {
			// Check if it's already an AppError, if not wrap it
			if appErr, ok := err.(*ent.AppError); ok {
				return appErr
			}
			return &ent.AppError{Err: err, Code: 500, Message: "Failed selecting word of the day"}
		}

		dest := r.URL.Query().Get("dest")
		if strings.ToLower(dest) == "twitter" {
			return wotd.Tweet(wo, w)
		} else if strings.ToLower(dest) == "mastodon" {
			mastodonClient := wotd.MastodonClient{}
			return mastodonClient.NewClient().Toot(wo, w, m.bucketName)
		} else {
			json.NewEncoder(w).Encode(&ent.PostResponse{Message: "No destination has been selected"})
			return nil
		}
	}

	return fn
}

// GetImage gets the image based on the provided name from the cloud storage
func (m MessagesRoute) GetImage() appHandler {
	fn := func(w http.ResponseWriter, r *http.Request) *ent.AppError {
		fn := r.URL.Query().Get("fn")
		cscw := gcs.NewGoogleCloudStorageClientWrapper(getLogger())
		err := cscw.Client(context.Background())

		if err != nil {
			// Check if it's already an AppError, if not wrap it
			if appErr, ok := err.(*ent.AppError); ok {
				return appErr
			}
			return &ent.AppError{Err: err, Code: 500, Message: "Failed to acquire image"}
		}

		b, err := cscw.GetObject(context.Background(), m.bucketName, fn)

		if err != nil {
			// Check if it's already an AppError, if not wrap it
			if appErr, ok := err.(*ent.AppError); ok {
				return appErr
			}
			return &ent.AppError{Err: err, Code: 500, Message: "Failed to acquire image"}
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(b)

		return nil
	}

	return fn
}
