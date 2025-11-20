package repository

// WordRepository defines the interface for word data access operations
type WordRepository interface {
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
	AddWord(word *Word) error

	// UpdateWord updates an existing word in the database
	UpdateWord(word *Word) error

	// DeleteWord removes a word from the database (hard delete)
	DeleteWord(id int) error

	// GetWordCount returns the total number of words in the database
	GetWordCount() (int, error)

	// GetWordCountByDayIndex returns the count of words with non-null day_index
	GetWordCountByDayIndex() (int, error)
}
