package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	ent "github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
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
		reqCtx := logger.ExtractRequestContext(r)
		log := getLogger()
		log.Info("Processing message request", reqCtx.ToFields()...)

		// Create WordSelector
		ws := wotd.NewWordSelector()
		log.Debug("Reading dictionary file", logger.String("file_path", "./dictionary.json"))
		f, erf := ws.ReadFile("./dictionary.json")

		if erf != nil {
			// Check if it's already an AppError, if not wrap it
			if appErr, ok := erf.(*ent.AppError); ok {
				appErr.WithContext("request_id", reqCtx.RequestID)
				appErr.WithContext("request_method", reqCtx.Method)
				appErr.WithContext("request_path", reqCtx.Path)
				return appErr
			}
			return ent.NewAppErrorWithContexts(erf, 500, "Failed sending the word of the day", map[string]interface{}{
				"request_id":     reqCtx.RequestID,
				"request_method": reqCtx.Method,
				"request_path":   reqCtx.Path,
			})
		}

		log.Debug("Parsing dictionary file")
		d, epf := ws.ParseFile(f, "./dictionary.json")
		if epf != nil {
			// Check if it's already an AppError, if not wrap it
			if appErr, ok := epf.(*ent.AppError); ok {
				appErr.WithContext("request_id", reqCtx.RequestID)
				appErr.WithContext("request_method", reqCtx.Method)
				appErr.WithContext("request_path", reqCtx.Path)
				return appErr
			}
			return ent.NewAppErrorWithContexts(epf, 500, "Failed sending the word of the day", map[string]interface{}{
				"request_id":     reqCtx.RequestID,
				"request_method": reqCtx.Method,
				"request_path":   reqCtx.Path,
			})
		}

		var wo *wotd.Word
		var err error
		wordIndex := r.URL.Query().Get("wordIndex")
		if wind, eind := strconv.Atoi(wordIndex); eind == nil {
			log.Debug("Selecting word by index", logger.Int("index", wind))
			wo, err = ws.SelectWordByIndex(d.Words, wind)
		} else {
			log.Debug("Selecting word by day")
			wo, err = ws.SelectWordByDay(d.Words)
		}

		if err != nil {
			// Check if it's already an AppError, if not wrap it
			if appErr, ok := err.(*ent.AppError); ok {
				appErr.WithContext("request_id", reqCtx.RequestID)
				appErr.WithContext("request_method", reqCtx.Method)
				appErr.WithContext("request_path", reqCtx.Path)
				return appErr
			}
			return ent.NewAppErrorWithContexts(err, 500, "Failed selecting word of the day", map[string]interface{}{
				"request_id":     reqCtx.RequestID,
				"request_method": reqCtx.Method,
				"request_path":   reqCtx.Path,
			})
		}

		dest := r.URL.Query().Get("dest")
		log.Info("Posting message", logger.String("destination", dest), logger.String("word", wo.Word))
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
		reqCtx := logger.ExtractRequestContext(r)
		ctx := r.Context()
		log := getLogger()

		fileName := r.URL.Query().Get("fn")
		fields := append(reqCtx.ToFields(), logger.String("file_name", fileName))
		log.Info("Getting image", fields...)

		cscw := gcs.NewGoogleCloudStorageClientWrapper(getLogger())
		err := cscw.Client(ctx)

		if err != nil {
			// Check if it's already an AppError, if not wrap it
			if appErr, ok := err.(*ent.AppError); ok {
				appErr.WithContext("request_id", reqCtx.RequestID)
				appErr.WithContext("request_method", reqCtx.Method)
				appErr.WithContext("request_path", reqCtx.Path)
				return appErr
			}
			return ent.NewAppErrorWithContexts(err, 500, "Failed to acquire image", map[string]interface{}{
				"request_id":     reqCtx.RequestID,
				"request_method": reqCtx.Method,
				"request_path":   reqCtx.Path,
			})
		}

		b, err := cscw.GetObject(ctx, m.bucketName, fileName)

		if err != nil {
			// Check if it's already an AppError, if not wrap it
			if appErr, ok := err.(*ent.AppError); ok {
				appErr.WithContext("request_id", reqCtx.RequestID)
				appErr.WithContext("request_method", reqCtx.Method)
				appErr.WithContext("request_path", reqCtx.Path)
				return appErr
			}
			return ent.NewAppErrorWithContexts(err, 500, "Failed to acquire image", map[string]interface{}{
				"request_id":     reqCtx.RequestID,
				"request_method": reqCtx.Method,
				"request_path":   reqCtx.Path,
				"file_name":      fileName,
				"bucket_name":    m.bucketName,
			})
		}

		log.Info("Image retrieved successfully", logger.String("file_name", fileName), logger.Int("size_bytes", len(b)))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(b)

		return nil
	}

	return fn
}
