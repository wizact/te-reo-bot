package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	ent "github.com/wizact/te-reo-bot/pkg/entities"
	wotd "github.com/wizact/te-reo-bot/pkg/wotd"
)

type MessagesRoute struct {
	bucketName string
}

func (m MessagesRoute) SetupRoutes(routePath string, router *mux.Router) {
	router.Handle(messagesRoute, appHandler(m.PostMessage())).Methods("POST")
	router.Handle(messagesRoute, appHandler(m.GetImage())).Methods("GET")
}

// PostMessage post a message to a specific social channel
func (m MessagesRoute) PostMessage() appHandler {
	fn := func(w http.ResponseWriter, r *http.Request) *ent.AppError {
		ws := wotd.WordSelector{}
		f, erf := ws.ReadFile()

		if erf != nil {
			return &ent.AppError{Error: erf, Code: 500, Message: "Failed sending the word of the day"}
		}

		d, epf := ws.ParseFile(f)
		if epf != nil {
			return &ent.AppError{Error: epf, Code: 500, Message: "Failed sending the word of the day"}
		}

		var wo *wotd.Word
		wordIndex := r.URL.Query().Get("wordIndex")
		if wind, eind := strconv.Atoi(wordIndex); eind == nil {
			wo = ws.SelectWordByIndex(d.Words, wind)
		} else {
			wo = ws.SelectWordByDay(d.Words)
		}

		dest := r.URL.Query().Get("dest")
		if strings.ToLower(dest) == "twitter" {
			return wotd.Tweet(wo, w)
		} else if strings.ToLower(dest) == "mastodon" {
			mastodonClient := wotd.MastodonClient{}
			return mastodonClient.Toot(wo, w, m.bucketName)
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
		var cscw wotd.CloudStorageClientWrapper
		err := cscw.Client(context.Background())

		if err != nil {
			return &ent.AppError{Error: err, Code: 500, Message: "Failed to acquire image"}
		}

		b, err := cscw.GetObject(context.Background(), m.bucketName, fn)

		if err != nil {
			return &ent.AppError{Error: err, Code: 500, Message: "Failed to acquire image"}
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(b)

		return nil
	}

	return fn
}
