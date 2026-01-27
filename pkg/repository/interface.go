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

	// GetWordCount returns the total number of words in the database
	GetWordCount() (int, error)

	// GetWordCountByDayIndex returns the count of words with non-null day_index
	GetWordCountByDayIndex() (int, error)

	// DeleteAllWordsByDayIndex deletes all words with non-null day_index in a single query
	DeleteAllWordsByDayIndex(tx *sql.Tx) error
}
