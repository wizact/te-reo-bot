package wotd_test

import (
	"bytes"
	"database/sql"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
	"github.com/wizact/te-reo-bot/pkg/repository"
	"github.com/wizact/te-reo-bot/pkg/wotd"
)

// TestWordSelectorIntegrationErrors tests word selector error scenarios with logging
func TestWordSelectorIntegrationErrors(t *testing.T) {
	tests := []struct {
		name             string
		setupTest        func() (*wotd.WordSelector, *bytes.Buffer)
		testOperation    func(*wotd.WordSelector) error
		expectedError    string
		logShouldContain []string
	}{
		{
			name: "SelectWordByDay - empty dictionary",
			setupTest: func() (*wotd.WordSelector, *bytes.Buffer) {
				var logBuffer bytes.Buffer
				config := &logger.LoggerConfig{
					EnableStackTraces: true,
					LogLevel:          "debug",
					Environment:       "test",
					LogFormat:         "json",
				}
				testLogger := logger.NewLoggerWithWriter(config, &logBuffer)
				logger.SetGlobalLogger(testLogger)
				ws := wotd.NewWordSelector()
				return ws, &logBuffer
			},
			testOperation: func(ws *wotd.WordSelector) error {
				emptyWords := make(map[int]wotd.Word)
				_, err := ws.SelectWordByDay(emptyWords)
				return err
			},
			expectedError: "Cannot select word from empty dictionary",
			logShouldContain: []string{
				"Cannot select word from empty dictionary",
				"stack_trace",
				"word_count",
				"operation",
				"select_word_by_day",
			},
		},
		{
			name: "SelectWordByDay - word not found for day",
			setupTest: func() (*wotd.WordSelector, *bytes.Buffer) {
				var logBuffer bytes.Buffer
				config := &logger.LoggerConfig{
					EnableStackTraces: true,
					LogLevel:          "debug",
					Environment:       "test",
					LogFormat:         "json",
				}
				testLogger := logger.NewLoggerWithWriter(config, &logBuffer)
				logger.SetGlobalLogger(testLogger)
				ws := wotd.NewWordSelector()
				return ws, &logBuffer
			},
			testOperation: func(ws *wotd.WordSelector) error {
				// Create map with only one word at index 100
				words := map[int]wotd.Word{
					100: {
						ID:       1,
						DayIndex: intPtr(100),
						Word:     "test",
						Meaning:  "test meaning",
					},
				}
				// Current day won't match index 100
				_, err := ws.SelectWordByDay(words)
				return err
			},
			expectedError: "No word found for current day of year",
			logShouldContain: []string{
				"No word found for current day of year",
				"stack_trace",
				"day_of_year",
				"word_count",
				"operation",
				"select_word_by_day",
			},
		},
		{
			name: "SelectWordByIndex - invalid index (zero)",
			setupTest: func() (*wotd.WordSelector, *bytes.Buffer) {
				var logBuffer bytes.Buffer
				config := &logger.LoggerConfig{
					EnableStackTraces: true,
					LogLevel:          "debug",
					Environment:       "test",
					LogFormat:         "json",
				}
				testLogger := logger.NewLoggerWithWriter(config, &logBuffer)
				logger.SetGlobalLogger(testLogger)
				ws := wotd.NewWordSelector()
				return ws, &logBuffer
			},
			testOperation: func(ws *wotd.WordSelector) error {
				words := map[int]wotd.Word{
					1: {
						ID:       1,
						DayIndex: intPtr(1),
						Word:     "test",
						Meaning:  "test meaning",
					},
				}
				_, err := ws.SelectWordByIndex(words, 0)
				return err
			},
			expectedError: "Invalid word index: must be greater than 0",
			logShouldContain: []string{
				"Invalid word index provided",
				"stack_trace",
				"requested_index",
				"word_count",
				"operation",
				"select_word_by_index",
			},
		},
		{
			name: "SelectWordByIndex - invalid index (negative)",
			setupTest: func() (*wotd.WordSelector, *bytes.Buffer) {
				var logBuffer bytes.Buffer
				config := &logger.LoggerConfig{
					EnableStackTraces: true,
					LogLevel:          "debug",
					Environment:       "test",
					LogFormat:         "json",
				}
				testLogger := logger.NewLoggerWithWriter(config, &logBuffer)
				logger.SetGlobalLogger(testLogger)
				ws := wotd.NewWordSelector()
				return ws, &logBuffer
			},
			testOperation: func(ws *wotd.WordSelector) error {
				words := map[int]wotd.Word{
					1: {
						ID:       1,
						DayIndex: intPtr(1),
						Word:     "test",
						Meaning:  "test meaning",
					},
				}
				_, err := ws.SelectWordByIndex(words, -1)
				return err
			},
			expectedError: "Invalid word index: must be greater than 0",
			logShouldContain: []string{
				"Invalid word index provided",
				"stack_trace",
				"requested_index",
				"word_count",
				"operation",
				"select_word_by_index",
			},
		},
		{
			name: "SelectWordByIndex - word not found for index",
			setupTest: func() (*wotd.WordSelector, *bytes.Buffer) {
				var logBuffer bytes.Buffer
				config := &logger.LoggerConfig{
					EnableStackTraces: true,
					LogLevel:          "debug",
					Environment:       "test",
					LogFormat:         "json",
				}
				testLogger := logger.NewLoggerWithWriter(config, &logBuffer)
				logger.SetGlobalLogger(testLogger)
				ws := wotd.NewWordSelector()
				return ws, &logBuffer
			},
			testOperation: func(ws *wotd.WordSelector) error {
				words := map[int]wotd.Word{
					1: {
						ID:       1,
						DayIndex: intPtr(1),
						Word:     "test",
						Meaning:  "test meaning",
					},
				}
				_, err := ws.SelectWordByIndex(words, 100) // Index 100 doesn't exist
				return err
			},
			expectedError: "No word found for requested index",
			logShouldContain: []string{
				"No word found for requested index",
				"stack_trace",
				"requested_index",
				"word_count",
				"operation",
				"select_word_by_index",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws, logBuffer := tt.setupTest()

			// Execute the test operation
			err := tt.testOperation(ws)

			// Verify error occurred
			assert.NotNil(t, err, "Expected error to occur")

			// Verify it's an AppError with proper message
			appErr, ok := err.(*entities.AppError)
			assert.True(t, ok, "Error should be an AppError")
			assert.Equal(t, tt.expectedError, appErr.Message, "Unexpected error message")

			// Verify stack trace is captured
			assert.True(t, appErr.HasStackTrace(), "AppError should have stack trace")

			// Verify logging occurred
			logOutput := logBuffer.String()
			assert.NotEmpty(t, logOutput, "Error should be logged")

			for _, shouldContain := range tt.logShouldContain {
				assert.Contains(t, logOutput, shouldContain,
					"Log should contain: %s", shouldContain)
			}

			// Verify stack trace is in logs
			assert.Contains(t, logOutput, "stack_trace", "Stack trace should be logged")
		})
	}
}

// TestTwitterClientIntegrationErrors tests Twitter client error scenarios
func TestTwitterClientIntegrationErrors(t *testing.T) {
	// Skip if no Twitter credentials are available
	if os.Getenv("TEREOBOT_CONSUMERKEY") == "" {
		t.Skip("Skipping Twitter integration test - no credentials available")
	}

	tests := []struct {
		name             string
		setupTest        func() (*bytes.Buffer, logger.Logger)
		testOperation    func(logger.Logger) error
		expectedError    string
		logShouldContain []string
	}{
		{
			name: "Tweet with invalid credentials",
			setupTest: func() (*bytes.Buffer, logger.Logger) {
				var logBuffer bytes.Buffer
				config := &logger.LoggerConfig{
					EnableStackTraces: true,
					LogLevel:          "debug",
					Environment:       "test",
					LogFormat:         "json",
				}
				testLogger := logger.NewLoggerWithWriter(config, &logBuffer)
				return &logBuffer, testLogger
			},
			testOperation: func(log logger.Logger) error {
				// Set invalid credentials
				os.Setenv("TEREOBOT_CONSUMERKEY", "invalid")
				os.Setenv("TEREOBOT_CONSUMERSECRET", "invalid")
				os.Setenv("TEREOBOT_ACCESSTOKEN", "invalid")
				os.Setenv("TEREOBOT_ACCESSSECRET", "invalid")

				defer func() {
					os.Unsetenv("TEREOBOT_CONSUMERKEY")
					os.Unsetenv("TEREOBOT_CONSUMERSECRET")
					os.Unsetenv("TEREOBOT_ACCESSTOKEN")
					os.Unsetenv("TEREOBOT_ACCESSSECRET")
				}()

				// Create test word
				testWord := wotd.Word{
					ID:               1,
					DayIndex:         intPtr(1),
					Word:             "test",
					Meaning:          "test meaning",
					Link:             "https://example.com",
					Photo:            "test.jpg",
					PhotoAttribution: "test attribution",
				}

				// Create response recorder
				rr := httptest.NewRecorder()

				// Attempt to tweet (this should fail with invalid credentials)
				appErr := wotd.Tweet(testWord, rr)
				if appErr != nil {
					return appErr
				}
				return nil
			},
			expectedError: "Failed sending the tweet",
			logShouldContain: []string{
				"Failed to send tweet to Twitter API",
				"stack_trace",
				"word",
				"test",
				"message",
				"operation",
				"twitter_post",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuffer, testLogger := tt.setupTest()

			// Execute the test operation
			err := tt.testOperation(testLogger)

			// For Twitter API errors, we expect an error
			if tt.expectedError != "" {
				assert.NotNil(t, err, "Expected error to occur")

				// Verify it's an AppError with proper message
				appErr, ok := err.(*entities.AppError)
				assert.True(t, ok, "Error should be an AppError")
				assert.Equal(t, tt.expectedError, appErr.Message, "Unexpected error message")

				// Verify logging occurred
				logOutput := logBuffer.String()
				assert.NotEmpty(t, logOutput, "Error should be logged")

				for _, shouldContain := range tt.logShouldContain {
					assert.Contains(t, logOutput, shouldContain,
						"Log should contain: %s", shouldContain)
				}
			}
		})
	}
}

// TestMastodonClientIntegrationErrors tests Mastodon client error scenarios
func TestMastodonClientIntegrationErrors(t *testing.T) {
	// Skip if no Mastodon credentials are available
	if os.Getenv("TEREOBOT_MASTODONSERVERNAME") == "" {
		t.Skip("Skipping Mastodon integration test - no credentials available")
	}

	tests := []struct {
		name             string
		setupTest        func() (*bytes.Buffer, *wotd.MastodonClient)
		testOperation    func(*wotd.MastodonClient) error
		expectedError    string
		logShouldContain []string
	}{
		{
			name: "Toot with invalid credentials",
			setupTest: func() (*bytes.Buffer, *wotd.MastodonClient) {
				var logBuffer bytes.Buffer
				config := &logger.LoggerConfig{
					EnableStackTraces: true,
					LogLevel:          "debug",
					Environment:       "test",
					LogFormat:         "json",
				}
				testLogger := logger.NewLoggerWithWriter(config, &logBuffer)

				// Set invalid credentials
				os.Setenv("TEREOBOT_MASTODONSERVERNAME", "https://invalid.mastodon.server")
				os.Setenv("TEREOBOT_MASTODONCLIENTID", "invalid")
				os.Setenv("TEREOBOT_MASTODONACCESSTOKEN", "invalid")

				logger.SetGlobalLogger(testLogger)
				mc := &wotd.MastodonClient{}
				mc = mc.NewClient()

				return &logBuffer, mc
			},
			testOperation: func(mc *wotd.MastodonClient) error {
				defer func() {
					os.Unsetenv("TEREOBOT_MASTODONSERVERNAME")
					os.Unsetenv("TEREOBOT_MASTODONCLIENTID")
					os.Unsetenv("TEREOBOT_MASTODONACCESSTOKEN")
				}()

				// Create test word
				testWord := wotd.Word{
					ID:               1,
					DayIndex:         intPtr(1),
					Word:             "test",
					Meaning:          "test meaning",
					Link:             "https://example.com",
					Photo:            "test.jpg",
					PhotoAttribution: "test attribution",
				}

				// Create response recorder
				rr := httptest.NewRecorder()

				// Attempt to toot (this should fail with invalid credentials)
				appErr := mc.Toot(testWord, rr, "test-bucket")
				if appErr != nil {
					return appErr
				}
				return nil
			},
			expectedError: "Failed sending the toot",
			logShouldContain: []string{
				"Failed to send toot to Mastodon API",
				"stack_trace",
				"word",
				"test",
				"toot_content",
				"operation",
				"mastodon_post",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuffer, mc := tt.setupTest()

			// Execute the test operation
			err := tt.testOperation(mc)

			// For Mastodon API errors, we expect an error
			if tt.expectedError != "" {
				assert.NotNil(t, err, "Expected error to occur")

				// Verify it's an AppError with proper message
				appErr, ok := err.(*entities.AppError)
				assert.True(t, ok, "Error should be an AppError")
				assert.Equal(t, tt.expectedError, appErr.Message, "Unexpected error message")

				// Verify logging occurred
				logOutput := logBuffer.String()
				assert.NotEmpty(t, logOutput, "Error should be logged")

				for _, shouldContain := range tt.logShouldContain {
					assert.Contains(t, logOutput, shouldContain,
						"Log should contain: %s", shouldContain)
				}
			}
		})
	}
}

// TestEndToEndWordProcessing tests the complete word processing flow with database
func TestEndToEndWordProcessing(t *testing.T) {
	// Create logger for testing
	var logBuffer bytes.Buffer
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLoggerWithWriter(config, &logBuffer)

	// Test complete flow: database -> select word
	t.Run("Successful word processing flow from database", func(t *testing.T) {
		logger.SetGlobalLogger(testLogger)
		defer logger.ResetGlobalLogger()

		// Create in-memory test database
		db, err := sql.Open("sqlite3", ":memory:")
		assert.Nil(t, err, "Should create in-memory database")
		defer db.Close()

		// Initialize schema
		err = repository.InitializeDatabase(db)
		assert.Nil(t, err, "Should initialize database schema")

		// Create repository and add test words
		repo := repository.NewSQLiteRepository(db)
		tx, err := repo.BeginTx()
		assert.Nil(t, err, "Should begin transaction")

		// Add test words with various day indices
		testWords := []wotd.Word{
			{
				DayIndex:         intPtr(1),
				Word:             "kia ora",
				Meaning:          "hello, hi",
				Link:             "https://maoridictionary.co.nz/search?idiom=&phrase=&proverb=&loan=&histLoanWords=&keywords=kia+ora",
				Photo:            "kia-ora.jpg",
				PhotoAttribution: "Test Attribution",
			},
			{
				DayIndex:         intPtr(2),
				Word:             "tēnā koe",
				Meaning:          "hello (to one person)",
				Link:             "https://maoridictionary.co.nz/search?idiom=&phrase=&proverb=&loan=&histLoanWords=&keywords=tena+koe",
				Photo:            "tena-koe.jpg",
				PhotoAttribution: "Test Attribution",
			},
			{
				DayIndex:         intPtr(150),
				Word:             "whānau",
				Meaning:          "family",
				Link:             "https://maoridictionary.co.nz/search?idiom=&phrase=&proverb=&loan=&histLoanWords=&keywords=whanau",
				Photo:            "whanau.jpg",
				PhotoAttribution: "Test Attribution",
			},
		}

		for _, word := range testWords {
			wordCopy := word
			err = repo.AddWord(tx, &wordCopy)
			assert.Nil(t, err, "Should add word to database")
		}

		err = repo.CommitTx(tx)
		assert.Nil(t, err, "Should commit transaction")

		// Get words from repository
		wordsByDay, err := repo.GetWordsByDayIndex()
		assert.Nil(t, err, "Should get words by day index")
		assert.Equal(t, 3, len(wordsByDay), "Should have 3 words")

		// Create word selector and test
		ws := wotd.NewWordSelector()

		// Select word by index
		wordByIndex, err := ws.SelectWordByIndex(wordsByDay, 1)
		assert.Nil(t, err, "Should successfully select word by index")
		assert.Equal(t, "kia ora", wordByIndex.Word, "Should select correct word")

		// Select another word by different index
		wordByIndex2, err := ws.SelectWordByIndex(wordsByDay, 150)
		assert.Nil(t, err, "Should successfully select word by index 150")
		assert.Equal(t, "whānau", wordByIndex2.Word, "Should select correct word")

		// Verify logging for successful operations
		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "Selected word by index", "Should log word selection by index")
	})

	t.Run("Error handling with empty database", func(t *testing.T) {
		// Reset log buffer
		logBuffer.Reset()

		logger.SetGlobalLogger(testLogger)
		defer logger.ResetGlobalLogger()

		// Create empty word map
		emptyWords := make(map[int]wotd.Word)

		ws := wotd.NewWordSelector()

		// Try to select from empty dictionary
		_, err := ws.SelectWordByDay(emptyWords)
		assert.NotNil(t, err, "Should fail with empty dictionary")

		// Verify error is properly logged with stack trace
		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "Cannot select word from empty dictionary", "Should log empty dictionary error")
		assert.Contains(t, logOutput, "stack_trace", "Should include stack trace in error log")

		// Verify it's an AppError with context
		appErr, ok := err.(*entities.AppError)
		assert.True(t, ok, "Error should be an AppError")
		assert.True(t, appErr.HasStackTrace(), "AppError should have stack trace")
	})
}
