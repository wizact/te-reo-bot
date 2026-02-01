package repository

import (
	"database/sql"
	"time"

	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
)

// SQLiteRepository implements the WordRepository interface using SQLite
type SQLiteRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewSQLiteRepository creates a new SQLite-based word repository
func NewSQLiteRepository(db *sql.DB) WordRepository {
	return &SQLiteRepository{
		db:     db,
		logger: logger.GetGlobalLogger(),
	}
}

// BeginTx starts a new database transaction
func (r *SQLiteRepository) BeginTx() (*sql.Tx, error) {
	r.logger.Debug("Starting database transaction")
	tx, err := r.db.Begin()
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to begin transaction")
		appErr.WithContext("operation", "begin_tx")
		appErr.WithContext("table", "words")
		r.logger.ErrorWithStack(err, "Transaction begin failed", logger.String("operation", "begin_tx"))
		return nil, appErr
	}
	return tx, nil
}

// CommitTx commits the given transaction
func (r *SQLiteRepository) CommitTx(tx *sql.Tx) error {
	r.logger.Debug("Committing database transaction")
	err := tx.Commit()
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to commit transaction")
		appErr.WithContext("operation", "commit_tx")
		appErr.WithContext("table", "words")
		r.logger.ErrorWithStack(err, "Transaction commit failed", logger.String("operation", "commit_tx"))
		return appErr
	}
	return nil
}

// RollbackTx rolls back the given transaction
func (r *SQLiteRepository) RollbackTx(tx *sql.Tx) error {
	r.logger.Debug("Rolling back database transaction")
	err := tx.Rollback()
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to rollback transaction")
		appErr.WithContext("operation", "rollback_tx")
		appErr.WithContext("table", "words")
		r.logger.ErrorWithStack(err, "Transaction rollback failed", logger.String("operation", "rollback_tx"))
		return appErr
	}
	return nil
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
		appErr := entities.NewAppError(err, 500, "Failed to query all words")
		appErr.WithContext("operation", "get_all_words")
		appErr.WithContext("table", "words")
		appErr.WithContext("query_type", "select")
		r.logger.ErrorWithStack(err, "Query failed", logger.String("operation", "get_all_words"))
		return nil, appErr
	}
	defer rows.Close()

	var words []Word
	for rows.Next() {
		word, err := scanWord(rows)
		if err != nil {
			appErr := entities.NewAppError(err, 500, "Failed to scan word row")
			appErr.WithContext("operation", "get_all_words")
			appErr.WithContext("table", "words")
			appErr.WithContext("query_type", "select")
			r.logger.ErrorWithStack(err, "Row scan failed", logger.String("operation", "get_all_words"))
			return nil, appErr
		}
		words = append(words, word)
	}

	if err = rows.Err(); err != nil {
		appErr := entities.NewAppError(err, 500, "Error iterating word rows")
		appErr.WithContext("operation", "get_all_words")
		appErr.WithContext("table", "words")
		r.logger.ErrorWithStack(err, "Row iteration failed", logger.String("operation", "get_all_words"))
		return nil, appErr
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
		appErr := entities.NewAppError(err, 500, "Failed to query words by day index")
		appErr.WithContext("operation", "get_words_by_day_index")
		appErr.WithContext("table", "words")
		appErr.WithContext("query_type", "select")
		r.logger.ErrorWithStack(err, "Query failed", logger.String("operation", "get_words_by_day_index"))
		return nil, appErr
	}
	defer rows.Close()

	wordsByDay := make(map[int]Word)
	for rows.Next() {
		word, err := scanWord(rows)
		if err != nil {
			appErr := entities.NewAppError(err, 500, "Failed to scan word row")
			appErr.WithContext("operation", "get_words_by_day_index")
			appErr.WithContext("table", "words")
			r.logger.ErrorWithStack(err, "Row scan failed", logger.String("operation", "get_words_by_day_index"))
			return nil, appErr
		}
		if word.DayIndex != nil {
			wordsByDay[*word.DayIndex] = word
		}
	}

	if err = rows.Err(); err != nil {
		appErr := entities.NewAppError(err, 500, "Error iterating word rows")
		appErr.WithContext("operation", "get_words_by_day_index")
		appErr.WithContext("table", "words")
		r.logger.ErrorWithStack(err, "Row iteration failed", logger.String("operation", "get_words_by_day_index"))
		return nil, appErr
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
		appErr := entities.NewAppError(err, 500, "Failed to get word by ID")
		appErr.WithContext("operation", "get_word_by_id")
		appErr.WithContext("table", "words")
		appErr.WithContext("word_id", id)
		r.logger.ErrorWithStack(err, "Query failed", logger.String("operation", "get_word_by_id"), logger.Int("word_id", id))
		return nil, appErr
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
		appErr := entities.NewAppError(err, 500, "Failed to get word by day index")
		appErr.WithContext("operation", "get_word_by_day_index")
		appErr.WithContext("table", "words")
		appErr.WithContext("day_index", dayIndex)
		r.logger.ErrorWithStack(err, "Query failed", logger.String("operation", "get_word_by_day_index"), logger.Int("day_index", dayIndex))
		return nil, appErr
	}

	return &word, nil
}

// AddWord inserts a new word into the database
func (r *SQLiteRepository) AddWord(tx *sql.Tx, word *Word) error {
	query := `
		INSERT INTO words (day_index, word, meaning, link, photo, photo_attribution)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	r.logger.Debug("Adding word to database", logger.String("word", word.Word), logger.String("operation", "insert"))

	result, err := tx.Exec(query,
		word.DayIndex,
		word.Word,
		word.Meaning,
		word.Link,
		word.Photo,
		word.PhotoAttribution,
	)
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to insert word")
		appErr.WithContext("operation", "insert_word")
		appErr.WithContext("table", "words")
		appErr.WithContext("word", word.Word)
		if word.DayIndex != nil {
			appErr.WithContext("day_index", *word.DayIndex)
		}
		r.logger.ErrorWithStack(err, "Insert failed", logger.String("operation", "insert_word"), logger.String("word", word.Word))
		return appErr
	}

	id, err := result.LastInsertId()
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to get inserted word ID")
		appErr.WithContext("operation", "insert_word")
		appErr.WithContext("table", "words")
		appErr.WithContext("word", word.Word)
		r.logger.ErrorWithStack(err, "LastInsertId failed", logger.String("operation", "insert_word"), logger.String("word", word.Word))
		return appErr
	}

	word.ID = int(id)
	r.logger.Info("Word added successfully", logger.Int("word_id", word.ID), logger.String("word", word.Word))
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

	r.logger.Debug("Updating word in database", logger.Int("word_id", word.ID), logger.String("word", word.Word))

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

	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to update word")
		appErr.WithContext("operation", "update_word")
		appErr.WithContext("table", "words")
		appErr.WithContext("word_id", word.ID)
		appErr.WithContext("word", word.Word)
		r.logger.ErrorWithStack(err, "Update failed", logger.String("operation", "update_word"), logger.Int("word_id", word.ID))
		return appErr
	}

	r.logger.Info("Word updated successfully", logger.Int("word_id", word.ID), logger.String("word", word.Word))
	return nil
}

// DeleteWord removes a word from the database (hard delete)
func (r *SQLiteRepository) DeleteWord(tx *sql.Tx, id int) error {
	r.logger.Debug("Deleting word from database", logger.Int("word_id", id))

	query := `DELETE FROM words WHERE id = ?`
	_, err := tx.Exec(query, id)
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to delete word")
		appErr.WithContext("operation", "delete_word")
		appErr.WithContext("table", "words")
		appErr.WithContext("word_id", id)
		r.logger.ErrorWithStack(err, "Delete failed", logger.String("operation", "delete_word"), logger.Int("word_id", id))
		return appErr
	}

	r.logger.Info("Word deleted successfully", logger.Int("word_id", id))
	return nil
}

// DeleteAllWordsByDayIndex deletes all words with day_index in single operation
func (r *SQLiteRepository) DeleteAllWordsByDayIndex(tx *sql.Tx) error {
	r.logger.Debug("Deleting all words with day_index")

	query := `DELETE FROM words WHERE day_index IS NOT NULL`
	_, err := tx.Exec(query)
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to delete words by day index")
		appErr.WithContext("operation", "delete_words_by_day_index")
		appErr.WithContext("table", "words")
		r.logger.ErrorWithStack(err, "Bulk delete failed", logger.String("operation", "delete_words_by_day_index"))
		return appErr
	}

	r.logger.Info("All words with day_index deleted successfully")
	return nil
}

// GetWordCount returns the total number of words in the database
func (r *SQLiteRepository) GetWordCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM words`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to get word count")
		appErr.WithContext("operation", "get_word_count")
		appErr.WithContext("table", "words")
		r.logger.ErrorWithStack(err, "Count query failed", logger.String("operation", "get_word_count"))
		return 0, appErr
	}
	return count, nil
}

// GetWordCountByDayIndex returns the count of words with non-null day_index
func (r *SQLiteRepository) GetWordCountByDayIndex() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM words WHERE day_index IS NOT NULL`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to get word count by day index")
		appErr.WithContext("operation", "get_word_count_by_day_index")
		appErr.WithContext("table", "words")
		r.logger.ErrorWithStack(err, "Count query failed", logger.String("operation", "get_word_count_by_day_index"))
		return 0, appErr
	}
	return count, nil
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
