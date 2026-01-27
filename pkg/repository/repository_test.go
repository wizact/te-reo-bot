package repository_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wizact/te-reo-bot/pkg/repository"
)

func setupTestRepository(t *testing.T) (*sql.DB, repository.WordRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	err = repository.InitializeDatabase(db)
	require.NoError(t, err)

	repo := repository.NewSQLiteRepository(db)
	return db, repo
}

func TestAddWord(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	dayIndex := 1
	word := &repository.Word{
		DayIndex: &dayIndex,
		Word:     "Kia ora",
		Meaning:  "Hello, be well",
		Link:     "https://example.com",
		Photo:    "kia-ora.jpg",
	}

	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word)
	assert.NoError(t, err, "Adding word should succeed")
	err = repo.CommitTx(tx)
	require.NoError(t, err)
	assert.NotZero(t, word.ID, "Word ID should be set after insert")
}

func TestAddWordWithoutDayIndex(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	word := &repository.Word{
		Word:    "Unassigned",
		Meaning: "A word not yet assigned to a day",
	}

	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)
	assert.NoError(t, err, "Adding word without day_index should succeed")
	assert.NotZero(t, word.ID)
}

func TestGetWordByID(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Add a word first
	dayIndex := 5
	originalWord := &repository.Word{
		DayIndex: &dayIndex,
		Word:     "Test word",
		Meaning:  "Test meaning",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, originalWord)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Retrieve it
	retrievedWord, err := repo.GetWordByID(originalWord.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedWord)
	assert.Equal(t, originalWord.Word, retrievedWord.Word)
	assert.Equal(t, originalWord.Meaning, retrievedWord.Meaning)
	assert.Equal(t, *originalWord.DayIndex, *retrievedWord.DayIndex)
}

func TestGetWordByIDNotFound(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	word, err := repo.GetWordByID(999)
	assert.Error(t, err, "Should return error for non-existent ID")
	assert.Nil(t, word)
}

func TestGetWordByDayIndex(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Add a word with day_index
	dayIndex := 10
	originalWord := &repository.Word{
		DayIndex: &dayIndex,
		Word:     "Day 10 word",
		Meaning:  "Day 10 meaning",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, originalWord)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Retrieve by day_index
	retrievedWord, err := repo.GetWordByDayIndex(10)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedWord)
	assert.Equal(t, originalWord.Word, retrievedWord.Word)
}

func TestGetAllWords(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Add multiple words
	dayIndex1 := 1
	dayIndex2 := 2
	words := []*repository.Word{
		{DayIndex: &dayIndex1, Word: "Word1", Meaning: "Meaning1"},
		{DayIndex: &dayIndex2, Word: "Word2", Meaning: "Meaning2"},
		{Word: "Word3", Meaning: "Meaning3"}, // No day_index
	}

	tx, err := repo.BeginTx()
	require.NoError(t, err)
	for _, word := range words {
		err = repo.AddWord(tx, word)
		require.NoError(t, err)
	}
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Get all words
	allWords, err := repo.GetAllWords()
	assert.NoError(t, err)
	assert.Len(t, allWords, 3)
}

func TestGetWordsByDayIndex(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Add words with and without day_index
	dayIndex1 := 1
	dayIndex2 := 2
	words := []*repository.Word{
		{DayIndex: &dayIndex1, Word: "Word1", Meaning: "Meaning1"},
		{DayIndex: &dayIndex2, Word: "Word2", Meaning: "Meaning2"},
		{Word: "Word3", Meaning: "Meaning3"}, // No day_index - should not be in map
	}

	tx, err := repo.BeginTx()
	require.NoError(t, err)
	for _, word := range words {
		err := repo.AddWord(tx, word)
		require.NoError(t, err)
	}
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Get words by day_index
	wordsByDay, err := repo.GetWordsByDayIndex()
	assert.NoError(t, err)
	assert.Len(t, wordsByDay, 2, "Only words with day_index should be returned")
	assert.Contains(t, wordsByDay, 1)
	assert.Contains(t, wordsByDay, 2)
	assert.Equal(t, "Word1", wordsByDay[1].Word)
	assert.Equal(t, "Word2", wordsByDay[2].Word)
}

func TestUpdateWord(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Add a word
	dayIndex := 7
	word := &repository.Word{
		DayIndex: &dayIndex,
		Word:     "Original",
		Meaning:  "Original meaning",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Update it
	word.Word = "Updated"
	word.Meaning = "Updated meaning"
	newDayIndex := 8
	word.DayIndex = &newDayIndex

	err = repo.UpdateWord(word)
	assert.NoError(t, err)

	// Verify update
	updated, err := repo.GetWordByID(word.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Word)
	assert.Equal(t, "Updated meaning", updated.Meaning)
	assert.Equal(t, 8, *updated.DayIndex)
}

func TestDeleteWord(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Add a word
	dayIndex := 15
	word := &repository.Word{
		DayIndex: &dayIndex,
		Word:     "To be deleted",
		Meaning:  "Will be removed",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	tx, err = repo.BeginTx()
	require.NoError(t, err)
	// Delete it
	err = repo.DeleteWord(tx, word.ID)
	assert.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetWordByID(word.ID)
	assert.Error(t, err, "Word should not be found after deletion")
}

func TestGetWordCount(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Initially empty
	count, err := repo.GetWordCount()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// Add words
	dayIndex := 1
	words := []*repository.Word{
		{DayIndex: &dayIndex, Word: "Word1", Meaning: "Meaning1"},
		{Word: "Word2", Meaning: "Meaning2"},
	}

	tx, err := repo.BeginTx()
	require.NoError(t, err)
	for _, word := range words {
		err := repo.AddWord(tx, word)
		require.NoError(t, err)
	}
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	count, err = repo.GetWordCount()
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestGetWordCountByDayIndex(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Add words with and without day_index
	dayIndex1 := 1
	dayIndex2 := 2
	words := []*repository.Word{
		{DayIndex: &dayIndex1, Word: "Word1", Meaning: "Meaning1"},
		{DayIndex: &dayIndex2, Word: "Word2", Meaning: "Meaning2"},
		{Word: "Word3", Meaning: "Meaning3"}, // No day_index
	}

	tx, err := repo.BeginTx()
	require.NoError(t, err)
	for _, word := range words {
		err := repo.AddWord(tx, word)
		require.NoError(t, err)
	}
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	count, err := repo.GetWordCountByDayIndex()
	assert.NoError(t, err)
	assert.Equal(t, 2, count, "Only words with day_index should be counted")
}

func TestDuplicateDayIndexError(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Add first word with day_index 1
	dayIndex := 1
	word1 := &repository.Word{
		DayIndex: &dayIndex,
		Word:     "Word1",
		Meaning:  "Meaning1",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word1)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Try to add another word with the same day_index
	word2 := &repository.Word{
		DayIndex: &dayIndex,
		Word:     "Word2",
		Meaning:  "Meaning2",
	}
	tx, err = repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word2)
	assert.Error(t, err, "Adding duplicate day_index should fail")
	repo.RollbackTx(tx)
}

func TestRequiredFieldsValidation(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Note: SQLite doesn't enforce NOT NULL on empty strings, only on actual NULL values
	// Application layer should validate non-empty strings
	// Here we test that we can add words with required fields present

	dayIndex := 1
	word := &repository.Word{
		DayIndex: &dayIndex,
		Word:     "Valid word",
		Meaning:  "Valid meaning",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word)
	assert.NoError(t, err, "Adding word with all required fields should succeed")
	err = repo.CommitTx(tx)
	require.NoError(t, err)
	assert.NotZero(t, word.ID)
}
