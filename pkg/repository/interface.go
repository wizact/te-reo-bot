package repository

import "database/sql"

// WordRepository defines the interface for word data access operations
type WordRepository interface {
	// BeginTx starts a new database transaction
	BeginTx() (*sql.Tx, error)

	// CommitTx commits the given transaction
	CommitTx(tx *sql.Tx) error

	// RollbackTx rolls back the given transaction
	RollbackTx(tx *sql.Tx) error

	// GetAllWords returns all words from the database
	GetAllWords() ([]Word, error)

	// GetWordsByDayIndex returns words indexed by their day_index (1-366)
	// Only includes words with non-null day_index
	GetWordsByDayIndex() (map[int]Word, error)

	// GetWordByID retrieves a single word by its ID
	GetWordByID(id int) (*Word, error)

	// GetWordByDayIndex retrieves a word by its day_index
	GetWordByDayIndex(dayIndex int) (*Word, error)

	// AddWord inserts a new word into the database
	AddWord(tx *sql.Tx, word *Word) error

	// UpdateWord updates an existing word in the database
	UpdateWord(word *Word) error

	// DeleteWord removes a word from the database (hard delete)
	DeleteWord(tx *sql.Tx, id int) error

	// DeduplicateWords removes duplicate word entries from database, keeping first occurrence (lowest ID)
	// Returns count of deleted duplicate words
	// Used at start of migration to ensure data consistency
	DeduplicateWords(tx *sql.Tx) (int, error)

	// UnsetAllDayIndexes clears day_index for all words with non-null day_index
	// Used at start of migration to reset assignments before applying new ones
	UnsetAllDayIndexes(tx *sql.Tx) error

	// GetWordByText retrieves a word by exact text match (case-sensitive)
	// Returns sql.ErrNoRows if word doesn't exist
	// Can be called with or without transaction - if tx is nil, uses direct DB query
	GetWordByText(tx *sql.Tx, word string) (*Word, error)

	// UpdateWordDayIndex updates only the day_index field for an existing word
	// Preserves all other fields, updates updated_at timestamp
	// Uses word text for lookup (not ID)
	UpdateWordDayIndex(tx *sql.Tx, wordText string, dayIndex int) error

	// GetWordCount returns the total number of words in the database
	GetWordCount() (int, error)

	// GetWordCountByDayIndex returns the count of words with non-null day_index
	GetWordCountByDayIndex() (int, error)

	// DeleteAllWordsByDayIndex deletes all words with non-null day_index in a single query
	// Deprecated: Use UnsetAllDayIndexes instead. Will be removed in v2.0.0
	DeleteAllWordsByDayIndex(tx *sql.Tx) error
}
