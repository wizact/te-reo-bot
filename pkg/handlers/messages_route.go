package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	ent "github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
	repo "github.com/wizact/te-reo-bot/pkg/repository"
	gcs "github.com/wizact/te-reo-bot/pkg/storage"
	wotd "github.com/wizact/te-reo-bot/pkg/wotd"
)

type MessagesRoute struct {
	bucketName string
	dryRun     bool
	db         *sql.DB
}

func (m MessagesRoute) SetupRoutes(routePath string, router *mux.Router) {
	router.Handle(routePath, appHandler(m.PostMessage())).Methods("POST")
	router.Handle(routePath, appHandler(m.GetImage())).Methods("GET")
}

// PostMessage post a message to a specific social channel
func (m MessagesRoute) PostMessage() appHandler {
	fn := func(w http.ResponseWriter, r *http.Request) *ent.AppError {
		var err error
		reqCtx := logger.ExtractRequestContext(r)
		log := getLogger()
		log.Info("Processing message request", reqCtx.ToFields()...)

		// Create WordSelector
		ws := wotd.NewWordSelector()

		// Create repository instance
		var wordsByDay map[int]wotd.Word
		wr := repo.NewSQLiteRepository(m.db)

		// Get words indexed by day_index
		wordsByDay, err = wr.GetWordsByDayIndex()
		if err != nil {
			if appErr, ok := err.(*ent.AppError); ok {
				appErr.WithContext("request_id", reqCtx.RequestID)
				appErr.WithContext("request_method", reqCtx.Method)
				appErr.WithContext("request_path", reqCtx.Path)
				return appErr
			}
			return ent.NewAppErrorWithContexts(err, 500, "Failed to get words", map[string]interface{}{
				"request_id":     reqCtx.RequestID,
				"request_method": reqCtx.Method,
				"request_path":   reqCtx.Path,
			})
		}

		wordIndex := r.URL.Query().Get("wordIndex")
		var wo wotd.Word
		if wind, eind := strconv.Atoi(wordIndex); eind == nil {
			log.Debug("Selecting word by index", logger.Int("index", wind))
			wo, err = ws.SelectWordByIndex(wordsByDay, wind)

		} else {
			log.Debug("Selecting word by day")
			wo, err = ws.SelectWordByDay(wordsByDay)
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

		if m.dryRun {
			log.Info("Dry run enabled - not posting to destination", logger.String("destination", dest))
			json.NewEncoder(w).Encode(&ent.PostResponse{Message: "Dry run - message not posted to destination"})
			return nil
		}

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
