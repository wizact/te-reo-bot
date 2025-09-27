package wotd_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
	wotd "github.com/wizact/te-reo-bot/pkg/wotd"
)

func TestParseFile(t *testing.T) {
	assert := assert.New(t)

	// Create logger config for testing
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLogger(config)
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()
	ws := wotd.NewWordSelector()

	jc := `{
			"dictionary": [
				{ "index":	1	, "word": "āe", "meaning": "yes", "link": "", "photo": ""},
				{ "index":	2	, "word": "aha", "meaning": "what?", "link": "", "photo": ""}
		]}`

	a, e := ws.ParseFile(bytes.NewBufferString(jc).Bytes(), "test-dictionary.json")

	assert.Nil(e, "Failed parsing dictionary")
	assert.NotNil(a)
	assert.True(a != nil && a.Words != nil && len(a.Words) > 0)
}

func TestReadFile(t *testing.T) {
	assert := assert.New(t)

	// Create logger config for testing
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLogger(config)
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()
	ws := wotd.NewWordSelector()

	f, e := ws.ReadFile("../../cmd/server/dictionary.json")

	assert.Nil(e, "Failed reading dictionary file")
	assert.NotNil(f)
	assert.True(len(f) > 0)
}

// Error scenario tests

func TestParseFile_InvalidJSON(t *testing.T) {
	assert := assert.New(t)

	// Create logger config for testing
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLogger(config)
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()
	ws := wotd.NewWordSelector()

	// Invalid JSON content
	invalidJSON := `{
		"dictionary": [
			{ "index": 1, "word": "āe", "meaning": "yes" // Missing closing brace
		]
	}`

	dictionary, err := ws.ParseFile([]byte(invalidJSON), "invalid-dictionary.json")

	assert.NotNil(err, "Expected error for invalid JSON")
	assert.Nil(dictionary, "Dictionary should be nil on parse error")

	// Check if it's an AppError with proper context
	appErr, ok := err.(*entities.AppError)
	assert.True(ok, "Error should be an AppError")
	assert.Equal("Failed to parse dictionary file", appErr.Message)
	assert.Equal(500, appErr.Code)

	// Check context
	filePath, exists := appErr.GetContext("file_path")
	assert.True(exists, "Context should contain file_path")
	assert.Equal("invalid-dictionary.json", filePath)

	fileSize, exists := appErr.GetContext("file_size")
	assert.True(exists, "Context should contain file_size")
	assert.Equal(len(invalidJSON), fileSize)

	operation, exists := appErr.GetContext("operation")
	assert.True(exists, "Context should contain operation")
	assert.Equal("json_unmarshal", operation)

	// Check stack trace
	assert.True(appErr.HasStackTrace(), "AppError should have stack trace")
}

func TestReadFile_NonExistentFile(t *testing.T) {
	assert := assert.New(t)

	// Create logger config for testing
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLogger(config)
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()
	ws := wotd.NewWordSelector()

	// Try to read a non-existent file
	nonExistentFile := "/path/to/non-existent-file.json"
	content, err := ws.ReadFile(nonExistentFile)

	assert.NotNil(err, "Expected error for non-existent file")
	assert.Nil(content, "Content should be nil on read error")

	// Check if it's an AppError with proper context
	appErr, ok := err.(*entities.AppError)
	assert.True(ok, "Error should be an AppError")
	assert.Equal("Failed to read dictionary file", appErr.Message)
	assert.Equal(500, appErr.Code)

	// Check context
	filePath, exists := appErr.GetContext("file_path")
	assert.True(exists, "Context should contain file_path")
	assert.Equal(nonExistentFile, filePath)

	operation, exists := appErr.GetContext("operation")
	assert.True(exists, "Context should contain operation")
	assert.Equal("file_read", operation)

	// Check stack trace
	assert.True(appErr.HasStackTrace(), "AppError should have stack trace")
}

func TestSelectWordByDay_EmptyDictionary(t *testing.T) {
	assert := assert.New(t)

	// Create logger config for testing
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLogger(config)
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()
	ws := wotd.NewWordSelector()

	// Empty words slice
	emptyWords := []wotd.Word{}

	word, err := ws.SelectWordByDay(emptyWords)

	assert.NotNil(err, "Expected error for empty dictionary")
	assert.Nil(word, "Word should be nil on error")

	// Check if it's an AppError with proper context
	appErr, ok := err.(*entities.AppError)
	assert.True(ok, "Error should be an AppError")
	assert.Equal("Cannot select word from empty dictionary", appErr.Message)
	assert.Equal(500, appErr.Code)

	// Check context
	wordCount, exists := appErr.GetContext("word_count")
	assert.True(exists, "Context should contain word_count")
	assert.Equal(0, wordCount)

	operation, exists := appErr.GetContext("operation")
	assert.True(exists, "Context should contain operation")
	assert.Equal("select_word_by_day", operation)

	// Check stack trace
	assert.True(appErr.HasStackTrace(), "AppError should have stack trace")
}

func TestSelectWordByIndex_EmptyDictionary(t *testing.T) {
	assert := assert.New(t)

	// Create logger config for testing
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLogger(config)
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()
	ws := wotd.NewWordSelector()

	// Empty words slice
	emptyWords := []wotd.Word{}

	word, err := ws.SelectWordByIndex(emptyWords, 1)

	assert.NotNil(err, "Expected error for empty dictionary")
	assert.Nil(word, "Word should be nil on error")

	// Check if it's an AppError with proper context
	appErr, ok := err.(*entities.AppError)
	assert.True(ok, "Error should be an AppError")
	assert.Equal("Cannot select word from empty dictionary", appErr.Message)
	assert.Equal(500, appErr.Code)

	// Check context
	wordCount, exists := appErr.GetContext("word_count")
	assert.True(exists, "Context should contain word_count")
	assert.Equal(0, wordCount)

	requestedIndex, exists := appErr.GetContext("requested_index")
	assert.True(exists, "Context should contain requested_index")
	assert.Equal(1, requestedIndex)

	operation, exists := appErr.GetContext("operation")
	assert.True(exists, "Context should contain operation")
	assert.Equal("select_word_by_index", operation)

	// Check stack trace
	assert.True(appErr.HasStackTrace(), "AppError should have stack trace")
}

func TestSelectWordByIndex_InvalidIndex(t *testing.T) {
	assert := assert.New(t)

	// Create logger config for testing
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLogger(config)
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()
	ws := wotd.NewWordSelector()

	// Valid words slice
	words := []wotd.Word{
		{Index: 1, Word: "āe", Meaning: "yes"},
		{Index: 2, Word: "aha", Meaning: "what?"},
	}

	// Test with invalid indices
	testCases := []int{0, -1, -10}

	for _, invalidIndex := range testCases {
		word, err := ws.SelectWordByIndex(words, invalidIndex)

		assert.NotNil(err, "Expected error for invalid index %d", invalidIndex)
		assert.Nil(word, "Word should be nil on error for index %d", invalidIndex)

		// Check if it's an AppError with proper context
		appErr, ok := err.(*entities.AppError)
		assert.True(ok, "Error should be an AppError for index %d", invalidIndex)
		assert.Equal("Invalid word index: must be greater than 0", appErr.Message)
		assert.Equal(400, appErr.Code)

		// Check context
		requestedIndex, exists := appErr.GetContext("requested_index")
		assert.True(exists, "Context should contain requested_index for index %d", invalidIndex)
		assert.Equal(invalidIndex, requestedIndex)

		wordCount, exists := appErr.GetContext("word_count")
		assert.True(exists, "Context should contain word_count for index %d", invalidIndex)
		assert.Equal(len(words), wordCount)

		operation, exists := appErr.GetContext("operation")
		assert.True(exists, "Context should contain operation for index %d", invalidIndex)
		assert.Equal("select_word_by_index", operation)

		// Check stack trace
		assert.True(appErr.HasStackTrace(), "AppError should have stack trace for index %d", invalidIndex)
	}
}

func TestSelectWordByIndex_ValidIndex(t *testing.T) {
	assert := assert.New(t)

	// Create logger config for testing
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLogger(config)
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()
	ws := wotd.NewWordSelector()

	// Valid words slice
	words := []wotd.Word{
		{Index: 1, Word: "āe", Meaning: "yes"},
		{Index: 2, Word: "aha", Meaning: "what?"},
	}

	// Test with valid index
	word, err := ws.SelectWordByIndex(words, 1)

	assert.Nil(err, "Should not have error for valid index")
	assert.NotNil(word, "Word should not be nil for valid index")
	assert.Equal("āe", word.Word)
	assert.Equal("yes", word.Meaning)
}

func TestSelectWordByDay_ValidWords(t *testing.T) {
	assert := assert.New(t)

	// Create logger config for testing
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLogger(config)
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()
	ws := wotd.NewWordSelector()

	// Valid words slice
	words := []wotd.Word{
		{Index: 1, Word: "āe", Meaning: "yes"},
		{Index: 2, Word: "aha", Meaning: "what?"},
	}

	// Test with valid words
	word, err := ws.SelectWordByDay(words)

	assert.Nil(err, "Should not have error for valid words")
	assert.NotNil(word, "Word should not be nil for valid words")
	assert.True(word.Word == "āe" || word.Word == "aha", "Should select one of the available words")
}

// Test integration scenario: read and parse file with error handling
func TestReadAndParseFile_Integration(t *testing.T) {
	assert := assert.New(t)

	// Create logger config for testing
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLogger(config)
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()
	ws := wotd.NewWordSelector()

	// Test successful read and parse
	content, err := ws.ReadFile("../../cmd/server/dictionary.json")
	assert.Nil(err, "Should successfully read dictionary file")
	assert.NotNil(content, "Content should not be nil")

	dictionary, err := ws.ParseFile(content, "../../cmd/server/dictionary.json")
	assert.Nil(err, "Should successfully parse dictionary file")
	assert.NotNil(dictionary, "Dictionary should not be nil")
	assert.True(len(dictionary.Words) > 0, "Dictionary should contain words")

	// Test word selection with parsed dictionary
	word, err := ws.SelectWordByIndex(dictionary.Words, 1)
	assert.Nil(err, "Should successfully select word by index")
	assert.NotNil(word, "Selected word should not be nil")

	wordByDay, err := ws.SelectWordByDay(dictionary.Words)
	assert.Nil(err, "Should successfully select word by day")
	assert.NotNil(wordByDay, "Selected word by day should not be nil")
}
