package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	ent "github.com/wizact/te-reo-bot/pkg/entities"
	wotd "github.com/wizact/te-reo-bot/pkg/wotd"
)

// PostMessage post a message to a specific social channel
func PostMessage(w http.ResponseWriter, r *http.Request) *ent.AppError {
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
		return wotd.Toot(wo, w)
	} else {
		json.NewEncoder(w).Encode(&ent.PostResponse{Message: "No destination has been selected"})
		return nil
	}
}

// GetImage gets the image based on the provided name from the cloud storage
func GetImage(w http.ResponseWriter, r *http.Request) *ent.AppError {
	fn := r.URL.Query().Get("fn")
	gsc, err := wotd.NewCloudStorageClient()

	if err != nil {
		return err
	}

	b, err := wotd.GetObject(gsc, fn)

	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(b)

	return nil
}
