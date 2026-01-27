package migration

import (
	"encoding/json"
	"fmt"
	"os"

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
	repo repository.WordRepository
}

// NewMigrator creates a new Migrator
func NewMigrator(repo repository.WordRepository) *Migrator {
	return &Migrator{repo: repo}
}

// MigrateWords imports words from a Dictionary into the database
func (m *Migrator) MigrateWords(dict *Dictionary) error {
	// Check if migration is needed (idempotent)
	existingCount, err := m.repo.GetWordCountByDayIndex()
	if err != nil {
		return fmt.Errorf("failed to check existing words: %w", err)
	}

	// Begin transaction for all write operations
	tx, err := m.repo.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Delete existing words if any (batch operation)
	if existingCount > 0 {
		if err := m.repo.DeleteAllWordsByDayIndex(tx); err != nil {
			m.repo.RollbackTx(tx)
			return fmt.Errorf("failed to delete existing words: %w", err)
		}
	}

	// Import all words from dictionary
	for _, dictWord := range dict.Words {
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
			return fmt.Errorf("failed to add word %q (index %d): %w",
				dictWord.Word, dictWord.Index, err)
		}
	}

	if err := m.repo.CommitTx(tx); err != nil {
		m.repo.RollbackTx(tx)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// MigrateFromFile reads a dictionary JSON file and imports it
func (m *Migrator) MigrateFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	dict, err := ParseDictionaryJSON(data)
	if err != nil {
		return err
	}

	return m.MigrateWords(dict)
}
