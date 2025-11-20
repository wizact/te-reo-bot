package repository_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wizact/te-reo-bot/pkg/repository"
)

func TestInitializeDatabase(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = repository.InitializeDatabase(db)
	assert.NoError(t, err, "Database initialization should succeed")
}

func TestSchemaCreatesWordsTable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = repository.InitializeDatabase(db)
	require.NoError(t, err)

	// Check that words table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='words'").Scan(&tableName)
	assert.NoError(t, err, "words table should exist")
	assert.Equal(t, "words", tableName)
}

func TestSchemaCreatesIndexes(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = repository.InitializeDatabase(db)
	require.NoError(t, err)

	// Check that indexes exist
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='words'")
	require.NoError(t, err)
	defer rows.Close()

	indexes := make(map[string]bool)
	for rows.Next() {
		var indexName string
		err = rows.Scan(&indexName)
		require.NoError(t, err)
		indexes[indexName] = true
	}

	// SQLite creates automatic indexes for PRIMARY KEY and UNIQUE constraints
	// We should have our custom indexes
	assert.Contains(t, indexes, "idx_day_index", "idx_day_index should exist")
	assert.Contains(t, indexes, "idx_active", "idx_active should exist")
}

func TestUniqueDayIndexConstraint(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = repository.InitializeDatabase(db)
	require.NoError(t, err)

	// Insert first word with day_index 1
	_, err = db.Exec("INSERT INTO words (day_index, word, meaning) VALUES (1, 'test1', 'meaning1')")
	assert.NoError(t, err, "First insert should succeed")

	// Try to insert another word with the same day_index
	_, err = db.Exec("INSERT INTO words (day_index, word, meaning) VALUES (1, 'test2', 'meaning2')")
	assert.Error(t, err, "Duplicate day_index should fail")
}

func TestDayIndexCheckConstraint(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = repository.InitializeDatabase(db)
	require.NoError(t, err)

	// Test valid day_index values
	validIndexes := []int{1, 100, 366}
	for _, idx := range validIndexes {
		_, err = db.Exec("INSERT INTO words (day_index, word, meaning) VALUES (?, ?, ?)",
			idx, "word"+string(rune(idx)), "meaning"+string(rune(idx)))
		assert.NoError(t, err, "day_index %d should be valid", idx)
	}

	// Test invalid day_index values (0 and 367)
	invalidIndexes := []int{0, 367}
	for _, idx := range invalidIndexes {
		_, err = db.Exec("INSERT INTO words (day_index, word, meaning) VALUES (?, ?, ?)",
			idx, "invalid", "invalid")
		assert.Error(t, err, "day_index %d should be invalid", idx)
	}
}

func TestNullableDayIndex(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = repository.InitializeDatabase(db)
	require.NoError(t, err)

	// Insert word without day_index (NULL)
	result, err := db.Exec("INSERT INTO words (word, meaning) VALUES (?, ?)", "unassigned", "meaning")
	assert.NoError(t, err, "Inserting word without day_index should succeed")

	id, err := result.LastInsertId()
	require.NoError(t, err)

	// Verify day_index is NULL
	var dayIndex sql.NullInt64
	err = db.QueryRow("SELECT day_index FROM words WHERE id = ?", id).Scan(&dayIndex)
	assert.NoError(t, err)
	assert.False(t, dayIndex.Valid, "day_index should be NULL")
}

func TestIdempotentInitialization(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Initialize twice
	err = repository.InitializeDatabase(db)
	assert.NoError(t, err, "First initialization should succeed")

	err = repository.InitializeDatabase(db)
	assert.NoError(t, err, "Second initialization should succeed (idempotent)")
}

func TestRequiredFields(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = repository.InitializeDatabase(db)
	require.NoError(t, err)

	// Try to insert without word (required field)
	_, err = db.Exec("INSERT INTO words (day_index, meaning) VALUES (?, ?)", 1, "meaning")
	assert.Error(t, err, "Insert without word should fail")

	// Try to insert without meaning (required field)
	_, err = db.Exec("INSERT INTO words (day_index, word) VALUES (?, ?)", 2, "word")
	assert.Error(t, err, "Insert without meaning should fail")
}
