package wotd

import (
	"encoding/json"
	"io/ioutil"
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
func (ws *WordSelector) SelectWordByDay(words []Word) (*Word, error) {
	if len(words) == 0 {
		err := entities.NewAppError(nil, 500, "Cannot select word from empty dictionary")
		err.WithContext("word_count", 0)
		err.WithContext("operation", "select_word_by_day")

		ws.getLogger().ErrorWithStack(nil, "Attempted to select word from empty dictionary",
			logger.Field{Key: "word_count", Value: 0},
			logger.Field{Key: "operation", Value: "select_word_by_day"},
		)

		return nil, err
	}

	doy := time.Now().YearDay()
	low := len(words)
	var selectedIndex int

	if doy <= low {
		selectedIndex = doy - 1
	} else {
		// Use modulo to wrap around, ensuring we never get -1
		remainder := doy % low
		if remainder == 0 {
			selectedIndex = low - 1 // Use last word when evenly divisible
		} else {
			selectedIndex = remainder - 1
		}
	}

	// Log the word selection
	ws.getLogger().Debug("Selected word by day",
		logger.Field{Key: "day_of_year", Value: doy},
		logger.Field{Key: "selected_index", Value: selectedIndex},
		logger.Field{Key: "word_count", Value: low},
		logger.Field{Key: "selected_word", Value: words[selectedIndex].Word},
	)

	return &words[selectedIndex], nil
}

// SelectWordByIndex selects a word from the provided array based on the provided index
func (ws *WordSelector) SelectWordByIndex(words []Word, index int) (*Word, error) {
	if len(words) == 0 {
		err := entities.NewAppError(nil, 500, "Cannot select word from empty dictionary")
		err.WithContext("word_count", 0)
		err.WithContext("requested_index", index)
		err.WithContext("operation", "select_word_by_index")

		ws.getLogger().ErrorWithStack(nil, "Attempted to select word from empty dictionary",
			logger.Field{Key: "word_count", Value: 0},
			logger.Field{Key: "requested_index", Value: index},
			logger.Field{Key: "operation", Value: "select_word_by_index"},
		)

		return nil, err
	}

	if index <= 0 {
		err := entities.NewAppError(nil, 400, "Invalid word index: must be greater than 0")
		err.WithContext("requested_index", index)
		err.WithContext("word_count", len(words))
		err.WithContext("operation", "select_word_by_index")

		ws.getLogger().ErrorWithStack(nil, "Invalid word index provided",
			logger.Field{Key: "requested_index", Value: index},
			logger.Field{Key: "word_count", Value: len(words)},
			logger.Field{Key: "operation", Value: "select_word_by_index"},
		)

		return nil, err
	}

	low := len(words)
	var selectedIndex int

	if index <= low {
		selectedIndex = index - 1
	} else {
		// Use modulo to wrap around, ensuring we never get -1
		remainder := index % low
		if remainder == 0 {
			selectedIndex = low - 1 // Use last word when evenly divisible
		} else {
			selectedIndex = remainder - 1
		}
	}

	// Log the word selection
	ws.getLogger().Debug("Selected word by index",
		logger.Field{Key: "requested_index", Value: index},
		logger.Field{Key: "selected_index", Value: selectedIndex},
		logger.Field{Key: "word_count", Value: low},
		logger.Field{Key: "selected_word", Value: words[selectedIndex].Word},
	)

	return &words[selectedIndex], nil
}

// ParseFile unmarshal a json string to the struct type
func (ws *WordSelector) ParseFile(f []byte, filePath string) (*Dictionary, error) {
	wd := Dictionary{}

	err := json.Unmarshal(f, &wd)

	if err != nil {
		// Log the parsing error with context
		ws.getLogger().ErrorWithStack(err, "Failed to parse dictionary JSON file",
			logger.Field{Key: "file_path", Value: filePath},
			logger.Field{Key: "file_size", Value: len(f)},
		)

		// Return enhanced AppError with context
		appErr := entities.NewAppError(err, 500, "Failed to parse dictionary file")
		appErr.WithContext("file_path", filePath)
		appErr.WithContext("file_size", len(f))
		appErr.WithContext("operation", "json_unmarshal")
		return nil, appErr
	}

	// Log successful parsing
	ws.getLogger().Info("Successfully parsed dictionary file",
		logger.Field{Key: "file_path", Value: filePath},
		logger.Field{Key: "word_count", Value: len(wd.Words)},
	)

	return &wd, nil
}

// ReadFile reads dictionary json file
func (ws *WordSelector) ReadFile(filePath string) ([]byte, error) {
	f, err := ioutil.ReadFile(filePath)

	if err != nil {
		// Log the file reading error with context
		ws.getLogger().ErrorWithStack(err, "Failed to read dictionary file",
			logger.Field{Key: "file_path", Value: filePath},
			logger.Field{Key: "operation", Value: "file_read"},
		)

		// Return enhanced AppError with context
		appErr := entities.NewAppError(err, 500, "Failed to read dictionary file")
		appErr.WithContext("file_path", filePath)
		appErr.WithContext("operation", "file_read")
		return nil, appErr
	}

	// Log successful file reading
	ws.getLogger().Debug("Successfully read dictionary file",
		logger.Field{Key: "file_path", Value: filePath},
		logger.Field{Key: "file_size", Value: len(f)},
	)

	return f, nil
}

// Dictionary is the parent element of json file
type Dictionary struct {
	Words []Word `json:"dictionary"`
}

// Word is the wrapper around each word and it's meaning
type Word struct {
	Index       int    `json:"index"`
	Word        string `json:"word"`
	Meaning     string `json:"meaning"`
	Link        string `json:"link"`
	Photo       string `json:"photo"`
	Attribution string `json:"photo_attribution"`
}
