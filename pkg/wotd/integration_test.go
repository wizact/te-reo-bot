package wotd_test

import (
	"bytes"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
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
			name: "ReadFile - non-existent file",
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
				_, err := ws.ReadFile("/non/existent/path/dictionary.json")
				return err
			},
			expectedError: "Failed to read dictionary file",
			logShouldContain: []string{
				"Failed to read dictionary file",
				"stack_trace",
				"file_path",
				"/non/existent/path/dictionary.json",
				"operation",
				"file_read",
			},
		},
		{
			name: "ParseFile - invalid JSON",
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
				invalidJSON := []byte(`{"dictionary": [{"index": 1, "word": "test" // invalid}`)
				_, err := ws.ParseFile(invalidJSON, "test-invalid.json")
				return err
			},
			expectedError: "Failed to parse dictionary file",
			logShouldContain: []string{
				"Failed to parse dictionary JSON file",
				"stack_trace",
				"file_path",
				"test-invalid.json",
				"file_size",
			},
		},
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
				emptyWords := []wotd.Word{}
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
			name: "SelectWordByIndex - invalid index",
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
				words := []wotd.Word{{Index: 1, Word: "test", Meaning: "test meaning"}}
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
				testWord := &wotd.Word{
					Index:   1,
					Word:    "test",
					Meaning: "test meaning",
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
				testWord := &wotd.Word{
					Index:   1,
					Word:    "test",
					Meaning: "test meaning",
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

// TestEndToEndWordProcessing tests the complete word processing flow with error handling
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

	// Test complete flow: read file -> parse -> select word
	t.Run("Successful word processing flow", func(t *testing.T) {
		logger.SetGlobalLogger(testLogger)
		defer logger.ResetGlobalLogger()
		ws := wotd.NewWordSelector()

		// Read actual dictionary file
		content, err := ws.ReadFile("../../cmd/server/dictionary.json")
		assert.Nil(t, err, "Should successfully read dictionary file")

		// Parse the content
		dictionary, err := ws.ParseFile(content, "../../cmd/server/dictionary.json")
		assert.Nil(t, err, "Should successfully parse dictionary file")
		assert.NotNil(t, dictionary, "Dictionary should not be nil")
		assert.True(t, len(dictionary.Words) > 0, "Dictionary should contain words")

		// Select word by day
		wordByDay, err := ws.SelectWordByDay(dictionary.Words)
		assert.Nil(t, err, "Should successfully select word by day")
		assert.NotNil(t, wordByDay, "Selected word should not be nil")

		// Select word by index
		wordByIndex, err := ws.SelectWordByIndex(dictionary.Words, 1)
		assert.Nil(t, err, "Should successfully select word by index")
		assert.NotNil(t, wordByIndex, "Selected word should not be nil")

		// Verify logging for successful operations
		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "Successfully read dictionary file", "Should log successful file read")
		assert.Contains(t, logOutput, "Successfully parsed dictionary file", "Should log successful file parse")
		assert.Contains(t, logOutput, "Selected word by day", "Should log word selection by day")
		assert.Contains(t, logOutput, "Selected word by index", "Should log word selection by index")
	})

	t.Run("Error propagation in word processing flow", func(t *testing.T) {
		// Reset log buffer
		logBuffer.Reset()

		logger.SetGlobalLogger(testLogger)
		defer logger.ResetGlobalLogger()
		ws := wotd.NewWordSelector()

		// Try to read non-existent file
		_, err := ws.ReadFile("/non/existent/file.json")
		assert.NotNil(t, err, "Should fail to read non-existent file")

		// Verify error is properly logged with stack trace
		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "Failed to read dictionary file", "Should log file read error")
		assert.Contains(t, logOutput, "stack_trace", "Should include stack trace in error log")

		// Verify it's an AppError with context
		appErr, ok := err.(*entities.AppError)
		assert.True(t, ok, "Error should be an AppError")
		assert.True(t, appErr.HasStackTrace(), "AppError should have stack trace")

		// Check context
		filePath, exists := appErr.GetContext("file_path")
		assert.True(t, exists, "Context should contain file_path")
		assert.Equal(t, "/non/existent/file.json", filePath)
	})
}
