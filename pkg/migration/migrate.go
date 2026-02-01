package migration

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
	"github.com/wizact/te-reo-bot/pkg/repository"
)

// Dictionary represents the root structure of dictionary.json
type Dictionary struct {
	Words []DictionaryWord `json:"dictionary"`
}

// DictionaryWord represents a word entry in dictionary.json
type DictionaryWord struct {
	Index            int    `json:"index"`
	Word             string `json:"word"`
	Meaning          string `json:"meaning"`
	Link             string `json:"link"`
	Photo            string `json:"photo"`
	PhotoAttribution string `json:"photo_attribution"`
}

// ParseDictionaryJSON parses a dictionary JSON file
func ParseDictionaryJSON(data []byte) (*Dictionary, error) {
	var dict Dictionary
	if err := json.Unmarshal(data, &dict); err != nil {
		return nil, fmt.Errorf("failed to parse dictionary JSON: %w", err)
	}
	return &dict, nil
}

// Migrator handles migration of dictionary data to SQLite
type Migrator struct {
	repo   repository.WordRepository
	logger logger.Logger
}

// NewMigrator creates a new Migrator
func NewMigrator(repo repository.WordRepository) *Migrator {
	return &Migrator{
		repo:   repo,
		logger: logger.GetGlobalLogger(),
	}
}

// MigrateWords imports words from a Dictionary into the database
func (m *Migrator) MigrateWords(dict *Dictionary) error {
	wordCount := len(dict.Words)
	m.logger.Info("Starting migration", logger.Int("word_count", wordCount))

	// Check if migration is needed (idempotent)
	existingCount, err := m.repo.GetWordCountByDayIndex()
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to check existing words")
		appErr.WithContext("operation", "migrate_check_existing")
		appErr.WithContext("word_count", wordCount)
		m.logger.ErrorWithStack(err, "Migration pre-check failed", logger.String("operation", "migrate_check_existing"))
		return appErr
	}

	m.logger.Info("Existing words count", logger.Int("existing_count", existingCount))

	// Begin transaction for all write operations
	tx, err := m.repo.BeginTx()
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to begin migration transaction")
		appErr.WithContext("operation", "migrate_begin_tx")
		appErr.WithContext("word_count", wordCount)
		m.logger.ErrorWithStack(err, "Migration transaction failed", logger.String("operation", "migrate_begin_tx"))
		return appErr
	}

	// Delete existing words if any (batch operation)
	if existingCount > 0 {
		m.logger.Info("Deleting existing words", logger.Int("count", existingCount))
		if err := m.repo.DeleteAllWordsByDayIndex(tx); err != nil {
			m.repo.RollbackTx(tx)
			appErr := entities.NewAppError(err, 500, "Failed to delete existing words")
			appErr.WithContext("operation", "migrate_delete_existing")
			appErr.WithContext("word_count", existingCount)
			m.logger.ErrorWithStack(err, "Migration delete failed", logger.String("operation", "migrate_delete_existing"))
			return appErr
		}
	}

	// Import all words from dictionary
	m.logger.Info("Importing words", logger.Int("word_count", wordCount))
	for i, dictWord := range dict.Words {
		if (i+1)%50 == 0 {
			m.logger.Debug("Migration progress", logger.Int("processed", i+1), logger.Int("total", wordCount))
		}

		word := &repository.Word{
			DayIndex:         &dictWord.Index,
			Word:             dictWord.Word,
			Meaning:          dictWord.Meaning,
			Link:             dictWord.Link,
			Photo:            dictWord.Photo,
			PhotoAttribution: dictWord.PhotoAttribution,
		}

		if err := m.repo.AddWord(tx, word); err != nil {
			m.repo.RollbackTx(tx)
			appErr := entities.NewAppError(err, 500, fmt.Sprintf("Failed to add word %q", dictWord.Word))
			appErr.WithContext("operation", "migrate_add_word")
			appErr.WithContext("word", dictWord.Word)
			appErr.WithContext("day_index", dictWord.Index)
			m.logger.ErrorWithStack(err, "Migration add word failed",
				logger.String("operation", "migrate_add_word"),
				logger.String("word", dictWord.Word),
				logger.Int("day_index", dictWord.Index))
			return appErr
		}
	}

	if err := m.repo.CommitTx(tx); err != nil {
		m.repo.RollbackTx(tx)
		appErr := entities.NewAppError(err, 500, "Failed to commit migration transaction")
		appErr.WithContext("operation", "migrate_commit_tx")
		appErr.WithContext("word_count", wordCount)
		m.logger.ErrorWithStack(err, "Migration commit failed", logger.String("operation", "migrate_commit_tx"))
		return appErr
	}

	m.logger.Info("Migration completed successfully", logger.Int("word_count", wordCount))
	return nil
}

// MigrateFromFile reads a dictionary JSON file and imports it
func (m *Migrator) MigrateFromFile(filePath string) error {
	m.logger.Info("Reading dictionary file", logger.String("file_path", filePath))

	data, err := os.ReadFile(filePath)
	if err != nil {
		appErr := entities.NewAppError(err, 500, "failed to read file")
		appErr.WithContext("operation", "migrate_read_file")
		appErr.WithContext("file_path", filePath)
		m.logger.ErrorWithStack(err, "File read failed", logger.String("operation", "migrate_read_file"), logger.String("file_path", filePath))
		return appErr
	}

	dict, err := ParseDictionaryJSON(data)
	if err != nil {
		appErr := entities.NewAppErrorWithContexts(err, 500, "Failed to parse dictionary JSON", map[string]interface{}{
			"operation": "migrate_parse_json",
			"file_path": filePath,
		})
		m.logger.ErrorWithStack(err, "JSON parse failed", logger.String("operation", "migrate_parse_json"), logger.String("file_path", filePath))
		return appErr
	}

	return m.MigrateWords(dict)
}
