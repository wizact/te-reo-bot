package repository

import (
	"database/sql"
	"time"
)

// SQLiteRepository implements the WordRepository interface using SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite-based word repository
func NewSQLiteRepository(db *sql.DB) WordRepository {
	return &SQLiteRepository{db: db}
}

// BeginTx starts a new database transaction
func (r *SQLiteRepository) BeginTx() (*sql.Tx, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CommitTx commits the given transaction
func (r *SQLiteRepository) CommitTx(tx *sql.Tx) error {
	return tx.Commit()
}

// RollbackTx rolls back the given transaction
func (r *SQLiteRepository) RollbackTx(tx *sql.Tx) error {
	return tx.Rollback()
}

// GetAllWords returns all words from the database
func (r *SQLiteRepository) GetAllWords() ([]Word, error) {
	query := `
		SELECT id, day_index, word, meaning, link, photo, photo_attribution, 
		       created_at, updated_at, is_active
		FROM words
		ORDER BY day_index, id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []Word
	for rows.Next() {
		word, err := scanWord(rows)
		if err != nil {
			return nil, err
		}
		words = append(words, word)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return words, nil
}

// GetWordsByDayIndex returns words indexed by their day_index (1-366)
func (r *SQLiteRepository) GetWordsByDayIndex() (map[int]Word, error) {
	query := `
		SELECT id, day_index, word, meaning, link, photo, photo_attribution, 
		       created_at, updated_at, is_active
		FROM words
		WHERE day_index IS NOT NULL
		ORDER BY day_index
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wordsByDay := make(map[int]Word)
	for rows.Next() {
		word, err := scanWord(rows)
		if err != nil {
			return nil, err
		}
		if word.DayIndex != nil {
			wordsByDay[*word.DayIndex] = word
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return wordsByDay, nil
}

// GetWordByID retrieves a single word by its ID
func (r *SQLiteRepository) GetWordByID(id int) (*Word, error) {
	query := `
		SELECT id, day_index, word, meaning, link, photo, photo_attribution, 
		       created_at, updated_at, is_active
		FROM words
		WHERE id = ?
	`

	row := r.db.QueryRow(query, id)
	word, err := scanWord(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &word, nil
}

// GetWordByDayIndex retrieves a word by its day_index
func (r *SQLiteRepository) GetWordByDayIndex(dayIndex int) (*Word, error) {
	query := `
		SELECT id, day_index, word, meaning, link, photo, photo_attribution, 
		       created_at, updated_at, is_active
		FROM words
		WHERE day_index = ?
	`

	row := r.db.QueryRow(query, dayIndex)
	word, err := scanWord(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &word, nil
}

// AddWord inserts a new word into the database
func (r *SQLiteRepository) AddWord(tx *sql.Tx, word *Word) error {
	query := `
		INSERT INTO words (day_index, word, meaning, link, photo, photo_attribution)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := tx.Exec(query,
		word.DayIndex,
		word.Word,
		word.Meaning,
		word.Link,
		word.Photo,
		word.PhotoAttribution,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	word.ID = int(id)
	return nil
}

// UpdateWord updates an existing word in the database
func (r *SQLiteRepository) UpdateWord(word *Word) error {
	query := `
		UPDATE words
		SET day_index = ?, word = ?, meaning = ?, link = ?, photo = ?, 
		    photo_attribution = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		word.DayIndex,
		word.Word,
		word.Meaning,
		word.Link,
		word.Photo,
		word.PhotoAttribution,
		time.Now(),
		word.ID,
	)

	return err
}

// DeleteWord removes a word from the database (hard delete)
func (r *SQLiteRepository) DeleteWord(tx *sql.Tx, id int) error {
	query := `DELETE FROM words WHERE id = ?`
	_, err := tx.Exec(query, id)
	return err
}

// DeleteAllWordsByDayIndex deletes all words with day_index in single operation
func (r *SQLiteRepository) DeleteAllWordsByDayIndex(tx *sql.Tx) error {
	query := `DELETE FROM words WHERE day_index IS NOT NULL`
	_, err := tx.Exec(query)
	return err
}

// GetWordCount returns the total number of words in the database
func (r *SQLiteRepository) GetWordCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM words`
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

// GetWordCountByDayIndex returns the count of words with non-null day_index
func (r *SQLiteRepository) GetWordCountByDayIndex() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM words WHERE day_index IS NOT NULL`
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

// scanner is an interface for scanning database rows
type scanner interface {
	Scan(dest ...interface{}) error
}

// scanWord scans a database row into a Word struct
func scanWord(s scanner) (Word, error) {
	var word Word
	var dayIndex sql.NullInt64
	var link, photo, photoAttribution sql.NullString
	var createdAt, updatedAt sql.NullTime
	var isActive sql.NullBool

	err := s.Scan(
		&word.ID,
		&dayIndex,
		&word.Word,
		&word.Meaning,
		&link,
		&photo,
		&photoAttribution,
		&createdAt,
		&updatedAt,
		&isActive,
	)
	if err != nil {
		return Word{}, err
	}

	// Convert sql.Null types to appropriate Go types
	if dayIndex.Valid {
		val := int(dayIndex.Int64)
		word.DayIndex = &val
	}
	if link.Valid {
		word.Link = link.String
	}
	if photo.Valid {
		word.Photo = photo.String
	}
	if photoAttribution.Valid {
		word.PhotoAttribution = photoAttribution.String
	}
	if createdAt.Valid {
		word.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		word.UpdatedAt = updatedAt.Time
	}
	if isActive.Valid {
		word.IsActive = isActive.Bool
	}

	return word, nil
}
