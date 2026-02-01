package migration

import (
	"database/sql"
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

// deduplicateDictWords returns unique words from dictionary, keeping first occurrence
func (m *Migrator) deduplicateDictWords(words []DictionaryWord) ([]DictionaryWord, int) {
	seen := make(map[string]bool)
	unique := make([]DictionaryWord, 0, len(words))
	duplicates := 0

	for _, w := range words {
		if !seen[w.Word] {
			seen[w.Word] = true
			unique = append(unique, w)
		} else {
			duplicates++
			m.logger.Debug("Skipping duplicate word in dictionary.json",
				logger.String("word", w.Word),
				logger.Int("day_index", w.Index))
		}
	}

	return unique, duplicates
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

	// Check existing word count
	existingCount, err := m.repo.GetWordCountByDayIndex()
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to check existing words").
			WithContext("operation", "migrate_check_existing").
			WithContext("word_count", wordCount)
		m.logger.ErrorWithStack(err, "Migration pre-check failed", logger.String("operation", "migrate_check_existing"))
		return appErr
	}

	m.logger.Info("Existing day_index assignments", logger.Int("existing_count", existingCount))

	// Begin transaction for all write operations
	tx, err := m.repo.BeginTx()
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to begin migration transaction").
			WithContext("operation", "migrate_begin_tx").
			WithContext("word_count", wordCount)
		m.logger.ErrorWithStack(err, "Migration transaction failed", logger.String("operation", "migrate_begin_tx"))
		return appErr
	}

	// Deduplicate database words (keep first occurrence)
	dbDuplicates, err := m.repo.DeduplicateWords(tx)
	if err != nil {
		m.repo.RollbackTx(tx)
		appErr := entities.NewAppError(err, 500, "Failed to deduplicate database words").
			WithContext("operation", "migrate_deduplicate_db")
		m.logger.ErrorWithStack(err, "Database deduplication failed", logger.String("operation", "migrate_deduplicate_db"))
		return appErr
	}
	if dbDuplicates > 0 {
		m.logger.Info("Removed duplicate words from database", logger.Int("duplicates_removed", dbDuplicates))
	}

	// Unset all day_index (preserves words, R1)
	if existingCount > 0 {
		m.logger.Info("Unsetting existing day_index assignments", logger.Int("count", existingCount))
		if err := m.repo.UnsetAllDayIndexes(tx); err != nil {
			m.repo.RollbackTx(tx)
			appErr := entities.NewAppError(err, 500, "Failed to unset day_index assignments").
				WithContext("operation", "migrate_unset_day_index").
				WithContext("existing_count", existingCount)
			m.logger.ErrorWithStack(err, "Unset day_index failed", logger.String("operation", "migrate_unset_day_index"))
			return appErr
		}
	}

	// Deduplicate dictionary.json words (keep first occurrence)
	uniqueWords, dictDuplicates := m.deduplicateDictWords(dict.Words)
	if dictDuplicates > 0 {
		m.logger.Info("Skipped duplicate entries in dictionary.json", logger.Int("duplicates_skipped", dictDuplicates))
	}

	// Process each unique word (update-or-insert pattern, R2, R3, R4)
	updated := 0
	inserted := 0
	m.logger.Info("Processing words", logger.Int("word_count", len(uniqueWords)))

	for i, dictWord := range uniqueWords {
		// Progress logging
		if (i+1)%50 == 0 {
			m.logger.Debug("Migration progress",
				logger.Int("processed", i+1),
				logger.Int("total", len(uniqueWords)))
		}

		// Lookup existing word by text (R2)
		existing, err := m.repo.GetWordByText(tx, dictWord.Word)
		if err != nil && err != sql.ErrNoRows {
			m.repo.RollbackTx(tx)
			appErr := entities.NewAppError(err, 500, "Failed to lookup existing word").
				WithContext("operation", "migrate_lookup_word").
				WithContext("word", dictWord.Word).
				WithContext("day_index", dictWord.Index)
			m.logger.ErrorWithStack(err, "Word lookup failed",
				logger.String("operation", "migrate_lookup_word"),
				logger.String("word", dictWord.Word))
			return appErr
		}

		if existing != nil {
			// Update existing word day_index (R3)
			if err := m.repo.UpdateWordDayIndex(tx, dictWord.Word, dictWord.Index); err != nil {
				m.repo.RollbackTx(tx)
				appErr := entities.NewAppError(err, 500, "Failed to update word day_index").
					WithContext("operation", "migrate_update_day_index").
					WithContext("word", dictWord.Word).
					WithContext("day_index", dictWord.Index)
				m.logger.ErrorWithStack(err, "Update day_index failed",
					logger.String("operation", "migrate_update_day_index"),
					logger.String("word", dictWord.Word),
					logger.Int("day_index", dictWord.Index))
				return appErr
			}
			updated++
		} else {
			// Insert new word (R4)
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
				appErr := entities.NewAppError(err, 500, fmt.Sprintf("Failed to insert word %q", dictWord.Word)).
					WithContext("operation", "migrate_insert_word").
					WithContext("word", dictWord.Word).
					WithContext("day_index", dictWord.Index)
				m.logger.ErrorWithStack(err, "Insert word failed",
					logger.String("operation", "migrate_insert_word"),
					logger.String("word", dictWord.Word),
					logger.Int("day_index", dictWord.Index))
				return appErr
			}
			inserted++
		}
	}

	// Commit transaction (R6)
	if err := m.repo.CommitTx(tx); err != nil {
		m.repo.RollbackTx(tx)
		appErr := entities.NewAppError(err, 500, "Failed to commit migration transaction").
			WithContext("operation", "migrate_commit_tx").
			WithContext("updated", updated).
			WithContext("inserted", inserted)
		m.logger.ErrorWithStack(err, "Migration commit failed", logger.String("operation", "migrate_commit_tx"))
		return appErr
	}

	// Final logging (R5, R7)
	preserved := existingCount - dbDuplicates - updated
	m.logger.Info("Migration completed successfully",
		logger.Int("duplicates_removed", dbDuplicates),
		logger.Int("duplicates_skipped", dictDuplicates),
		logger.Int("updated", updated),
		logger.Int("inserted", inserted),
		logger.Int("preserved", preserved),
		logger.Int("total_words", updated+inserted+preserved))

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
