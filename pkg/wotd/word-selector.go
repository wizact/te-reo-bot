package wotd

import (
	"time"

	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
)

// WordSelector reads, parses, and selects the word-of-the-day
type WordSelector struct {
}

// NewWordSelector creates a new WordSelector
func NewWordSelector() *WordSelector {
	return &WordSelector{}
}

// getLogger returns the global logger instance
func (ws *WordSelector) getLogger() logger.Logger {
	return logger.GetGlobalLogger()
}

// SelectWordByDay selects a word from the provided array based on the day of the year
func (ws *WordSelector) SelectWordByDay(words map[int]Word) (Word, error) {
	if len(words) == 0 {
		err := entities.NewAppError(nil, 500, "Cannot select word from empty dictionary")
		err.WithContext("word_count", 0)
		err.WithContext("operation", "select_word_by_day")

		ws.getLogger().ErrorWithStack(err, "Cannot select word from empty dictionary",
			logger.Int("word_count", 0),
			logger.String("operation", "select_word_by_day"),
		)

		return Word{}, err
	}

	doy := time.Now().YearDay()

	sw, exists := words[doy]
	if !exists {
		err := entities.NewAppError(nil, 404, "No word found for current day of year")
		err.WithContext("day_of_year", doy)
		err.WithContext("word_count", len(words))
		err.WithContext("operation", "select_word_by_day")

		ws.getLogger().ErrorWithStack(err, "No word found for current day of year",
			logger.Int("day_of_year", doy),
			logger.Int("word_count", len(words)),
			logger.String("operation", "select_word_by_day"),
		)

		return Word{}, err
	}

	// Log the word selection
	ws.getLogger().Debug("Selected word by day",
		logger.Int("day_of_year", doy),
		logger.String("selected_word", sw.Word),
	)

	return sw, nil
}

// SelectWordByIndex selects a word from the provided array based on the provided index
func (ws *WordSelector) SelectWordByIndex(words map[int]Word, index int) (Word, error) {
	if len(words) == 0 {
		err := entities.NewAppError(nil, 500, "Cannot select word from empty dictionary")
		err.WithContext("word_count", 0)
		err.WithContext("requested_index", index)
		err.WithContext("operation", "select_word_by_index")

		ws.getLogger().ErrorWithStack(err, "Cannot select word from empty dictionary",
			logger.Int("word_count", 0),
			logger.Int("requested_index", index),
			logger.String("operation", "select_word_by_index"),
		)

		return Word{}, err
	}

	if index <= 0 {
		err := entities.NewAppError(nil, 400, "Invalid word index: must be greater than 0")
		err.WithContext("requested_index", index)
		err.WithContext("word_count", len(words))
		err.WithContext("operation", "select_word_by_index")

		ws.getLogger().ErrorWithStack(err, "Invalid word index provided",
			logger.Int("requested_index", index),
			logger.Int("word_count", len(words)),
			logger.String("operation", "select_word_by_index"),
		)

		return Word{}, err
	}

	sw, exists := words[index]
	if !exists {
		err := entities.NewAppError(nil, 404, "No word found for requested index")
		err.WithContext("requested_index", index)
		err.WithContext("word_count", len(words))
		err.WithContext("operation", "select_word_by_index")

		ws.getLogger().ErrorWithStack(err, "No word found for requested index",
			logger.Int("requested_index", index),
			logger.Int("word_count", len(words)),
			logger.String("operation", "select_word_by_index"),
		)

		return Word{}, err
	}

	ws.getLogger().Debug("Selected word by index",
		logger.Int("requested_index", index),
		logger.String("selected_word", sw.Word),
	)

	return sw, nil
}

// Word represents a Māori word with its meaning and metadata
type Word struct {
	ID               int       `json:"id" db:"id"`
	DayIndex         *int      `json:"index,omitempty" db:"day_index"`
	Word             string    `json:"word" db:"word"`
	Meaning          string    `json:"meaning" db:"meaning"`
	Link             string    `json:"link" db:"link"`
	Photo            string    `json:"photo" db:"photo"`
	PhotoAttribution string    `json:"photo_attribution" db:"photo_attribution"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	IsActive         bool      `json:"is_active" db:"is_active"`
}
