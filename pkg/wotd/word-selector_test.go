package wotd_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/testutils"
	wotd "github.com/wizact/te-reo-bot/pkg/wotd"
)

func TestSelectWordByDay_EmptyDictionary(t *testing.T) {
	assert := assert.New(t)

	cleanup := testutils.SetupGlobalTestLogger()
	defer cleanup()
	ws := wotd.NewWordSelector()

	// Empty words slice
	emptyWords := map[int]wotd.Word{}

	word, err := ws.SelectWordByDay(emptyWords)

	assert.NotNil(err, "Expected error for empty dictionary")
	assert.Equal(wotd.Word{}, word, "Word should be empty on error")

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

	cleanup := testutils.SetupGlobalTestLogger()
	defer cleanup()
	ws := wotd.NewWordSelector()

	// Empty words slice
	emptyWords := make(map[int]wotd.Word)

	word, err := ws.SelectWordByIndex(emptyWords, 1)

	assert.NotNil(err, "Expected error for empty dictionary")
	assert.Equal(wotd.Word{}, word, "Word should be empty on error")

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

	cleanup := testutils.SetupGlobalTestLogger()
	defer cleanup()
	ws := wotd.NewWordSelector()

	// Valid words slice
	words := map[int]wotd.Word{
		1: {ID: 1, DayIndex: intPtr(1), Word: "āe", Meaning: "yes"},
		2: {ID: 2, DayIndex: intPtr(2), Word: "aha", Meaning: "what?"},
	}

	// Test with invalid indices
	testCases := []int{0, -1, -10}

	for _, invalidIndex := range testCases {
		word, err := ws.SelectWordByIndex(words, invalidIndex)

		assert.NotNil(err, "Expected error for invalid index %d", invalidIndex)
		assert.Equal(wotd.Word{}, word, "Word should be empty on error for index %d", invalidIndex)

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

	cleanup := testutils.SetupGlobalTestLogger()
	defer cleanup()
	ws := wotd.NewWordSelector()

	// Valid words slice
	words := map[int]wotd.Word{
		1: {ID: 1, DayIndex: intPtr(1), Word: "āe", Meaning: "yes"},
		2: {ID: 2, DayIndex: intPtr(2), Word: "aha", Meaning: "what?"},
	}

	// Test with valid index
	word, err := ws.SelectWordByIndex(words, 1)

	assert.Nil(err, "Should not have error for valid index")
	assert.NotEqual(wotd.Word{}, word, "Word should not be empty for valid index")
	assert.Equal("āe", word.Word)
	assert.Equal("yes", word.Meaning)
}

func TestSelectWordByDay_ValidWords(t *testing.T) {
	assert := assert.New(t)

	cleanup := testutils.SetupGlobalTestLogger()
	defer cleanup()
	ws := wotd.NewWordSelector()

	// Get current day of year to use as test data key
	doy := time.Now().YearDay()

	// Valid words map - include word for current day
	words := map[int]wotd.Word{
		doy: {ID: 1, DayIndex: intPtr(doy), Word: "āe", Meaning: "yes"},
	}

	// Test with valid words
	word, err := ws.SelectWordByDay(words)

	assert.Nil(err, "Should not have error for valid words")
	assert.NotEqual(wotd.Word{}, word, "Word should not be empty for valid words")
	assert.Equal("āe", word.Word, "Should select the word for current day")
}
