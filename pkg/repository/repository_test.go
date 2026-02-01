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
// Tests for preserve migration words feature

func TestDeduplicateWords_NoDuplicates(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Setup: Add 3 unique words
	tx, err := repo.BeginTx()
	require.NoError(t, err)

	dayIndex1 := 1
	dayIndex2 := 2
	dayIndex3 := 3
	
	words := []*repository.Word{
		{DayIndex: &dayIndex1, Word: "kia ora", Meaning: "hello"},
		{DayIndex: &dayIndex2, Word: "aroha", Meaning: "love"},
		{DayIndex: &dayIndex3, Word: "whānau", Meaning: "family"},
	}

	for _, word := range words {
		err = repo.AddWord(tx, word)
		require.NoError(t, err)
	}

	// Execute: Deduplicate
	duplicatesRemoved, err := repo.DeduplicateWords(tx)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Verify: No duplicates removed, all 3 words remain
	assert.Equal(t, 0, duplicatesRemoved, "Should remove 0 duplicates")
	
	count, err := repo.GetWordCount()
	require.NoError(t, err)
	assert.Equal(t, 3, count, "Should have 3 words remaining")
}

func TestDeduplicateWords_WithDuplicates(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Setup: Add duplicate "kia ora" entries with different IDs
	// Insert manually to control IDs
	dayIndex1 := 1
	dayIndex2 := 2
	dayIndex3 := 3
	dayIndex4 := 4

	tx, err := repo.BeginTx()
	require.NoError(t, err)

	// Add first "kia ora" (ID will be 1)
	word1 := &repository.Word{DayIndex: &dayIndex1, Word: "kia ora", Meaning: "hello 1"}
	err = repo.AddWord(tx, word1)
	require.NoError(t, err)

	// Add "aroha" (ID will be 2)
	word2 := &repository.Word{DayIndex: &dayIndex2, Word: "aroha", Meaning: "love"}
	err = repo.AddWord(tx, word2)
	require.NoError(t, err)

	// Add second "kia ora" (ID will be 3)
	word3 := &repository.Word{DayIndex: &dayIndex3, Word: "kia ora", Meaning: "hello 2"}
	err = repo.AddWord(tx, word3)
	require.NoError(t, err)

	// Add third "kia ora" (ID will be 4)
	word4 := &repository.Word{DayIndex: &dayIndex4, Word: "kia ora", Meaning: "hello 3"}
	err = repo.AddWord(tx, word4)
	require.NoError(t, err)

	// Execute: Deduplicate
	duplicatesRemoved, err := repo.DeduplicateWords(tx)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Verify: 2 duplicates removed (IDs 3 and 4), only IDs 1 and 2 remain
	assert.Equal(t, 2, duplicatesRemoved, "Should remove 2 duplicate 'kia ora' entries")
	
	count, err := repo.GetWordCount()
	require.NoError(t, err)
	assert.Equal(t, 2, count, "Should have 2 unique words remaining")

	// Verify only first "kia ora" remains (lowest ID)
	remainingWord, err := repo.GetWordByID(word1.ID)
	require.NoError(t, err)
	assert.Equal(t, "kia ora", remainingWord.Word)
	assert.Equal(t, "hello 1", remainingWord.Meaning, "Should keep first occurrence")

	// Verify duplicate IDs are deleted
	_, err = repo.GetWordByID(word3.ID)
	assert.Equal(t, sql.ErrNoRows, err, "Duplicate ID 3 should be deleted")
	
	_, err = repo.GetWordByID(word4.ID)
	assert.Equal(t, sql.ErrNoRows, err, "Duplicate ID 4 should be deleted")
}

func TestUnsetAllDayIndexes(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Setup: Add 3 words with day_index and 2 words without
	tx, err := repo.BeginTx()
	require.NoError(t, err)

	dayIndex1 := 1
	dayIndex2 := 2
	dayIndex3 := 3

	words := []*repository.Word{
		{DayIndex: &dayIndex1, Word: "word1", Meaning: "meaning1"},
		{DayIndex: &dayIndex2, Word: "word2", Meaning: "meaning2"},
		{DayIndex: &dayIndex3, Word: "word3", Meaning: "meaning3"},
		{DayIndex: nil, Word: "word4", Meaning: "meaning4"},
		{DayIndex: nil, Word: "word5", Meaning: "meaning5"},
	}

	for _, word := range words {
		err = repo.AddWord(tx, word)
		require.NoError(t, err)
	}
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Execute: Unset all day indexes
	tx, err = repo.BeginTx()
	require.NoError(t, err)
	err = repo.UnsetAllDayIndexes(tx)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Verify: All 5 words have day_index=NULL
	allWords, err := repo.GetAllWords()
	require.NoError(t, err)
	assert.Len(t, allWords, 5, "Should have 5 words")
	
	for _, word := range allWords {
		assert.Nil(t, word.DayIndex, "All words should have NULL day_index")
	}

	// Verify count by day_index returns 0
	countByDayIndex, err := repo.GetWordCountByDayIndex()
	require.NoError(t, err)
	assert.Equal(t, 0, countByDayIndex, "Should have 0 words with non-null day_index")
}

func TestGetWordByText_Found(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Setup: Add a word
	tx, err := repo.BeginTx()
	require.NoError(t, err)

	dayIndex := 1
	originalWord := &repository.Word{
		DayIndex:         &dayIndex,
		Word:             "kia ora",
		Meaning:          "hello, be well",
		Link:             "https://example.com",
		Photo:            "photo.jpg",
		PhotoAttribution: "John Doe",
	}
	err = repo.AddWord(tx, originalWord)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Execute: Get word by text
	tx, err = repo.BeginTx()
	require.NoError(t, err)
	foundWord, err := repo.GetWordByText(tx, "kia ora")
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Verify: Returns word with correct data
	require.NoError(t, err)
	require.NotNil(t, foundWord)
	assert.Equal(t, originalWord.ID, foundWord.ID)
	assert.Equal(t, "kia ora", foundWord.Word)
	assert.Equal(t, "hello, be well", foundWord.Meaning)
	assert.Equal(t, "https://example.com", foundWord.Link)
	assert.Equal(t, "photo.jpg", foundWord.Photo)
	assert.Equal(t, "John Doe", foundWord.PhotoAttribution)
	assert.Equal(t, 1, *foundWord.DayIndex)
}

func TestGetWordByText_NotFound(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Setup: Empty database (no words)
	
	// Execute: Get non-existent word
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	foundWord, err := repo.GetWordByText(tx, "missing word")
	repo.CommitTx(tx)

	// Verify: Returns sql.ErrNoRows
	assert.Equal(t, sql.ErrNoRows, err, "Should return sql.ErrNoRows for missing word")
	assert.Nil(t, foundWord)
}

func TestGetWordByText_CaseSensitive(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Setup: Add word with lowercase
	tx, err := repo.BeginTx()
	require.NoError(t, err)

	dayIndex := 1
	word := &repository.Word{
		DayIndex: &dayIndex,
		Word:     "kia ora",
		Meaning:  "hello",
	}
	err = repo.AddWord(tx, word)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Execute: Search with different case
	tx, err = repo.BeginTx()
	require.NoError(t, err)
	foundWord, err := repo.GetWordByText(tx, "Kia Ora")
	repo.CommitTx(tx)

	// Verify: Returns sql.ErrNoRows (case-sensitive)
	assert.Equal(t, sql.ErrNoRows, err, "Should be case-sensitive")
	assert.Nil(t, foundWord)

	// Verify: Exact match works
	tx, err = repo.BeginTx()
	require.NoError(t, err)
	foundWord, err = repo.GetWordByText(tx, "kia ora")
	repo.CommitTx(tx)
	require.NoError(t, err)
	assert.NotNil(t, foundWord)
	assert.Equal(t, "kia ora", foundWord.Word)
}

func TestUpdateWordDayIndex(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	// Setup: Add word with day_index=1
	tx, err := repo.BeginTx()
	require.NoError(t, err)

	dayIndex := 1
	originalWord := &repository.Word{
		DayIndex:         &dayIndex,
		Word:             "kia ora",
		Meaning:          "hello, be well",
		Link:             "https://example.com",
		Photo:            "photo.jpg",
		PhotoAttribution: "John Doe",
	}
	err = repo.AddWord(tx, originalWord)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	originalID := originalWord.ID

	// Get original word to capture initial state
	wordBeforeUpdate, err := repo.GetWordByID(originalID)
	require.NoError(t, err)
	originalCreatedAt := wordBeforeUpdate.CreatedAt
	originalUpdatedAt := wordBeforeUpdate.UpdatedAt

	// Execute: Update day_index to 5
	tx, err = repo.BeginTx()
	require.NoError(t, err)
	err = repo.UpdateWordDayIndex(tx, "kia ora", 5)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	// Verify: day_index changed, other fields preserved
	updatedWord, err := repo.GetWordByID(originalID)
	require.NoError(t, err)
	
	assert.Equal(t, 5, *updatedWord.DayIndex, "day_index should be updated to 5")
	assert.Equal(t, originalID, updatedWord.ID, "ID should not change")
	assert.Equal(t, "kia ora", updatedWord.Word, "Word text should not change")
	assert.Equal(t, "hello, be well", updatedWord.Meaning, "Meaning should not change")
	assert.Equal(t, "https://example.com", updatedWord.Link, "Link should not change")
	assert.Equal(t, "photo.jpg", updatedWord.Photo, "Photo should not change")
	assert.Equal(t, "John Doe", updatedWord.PhotoAttribution, "Photo attribution should not change")
	assert.Equal(t, originalCreatedAt, updatedWord.CreatedAt, "CreatedAt should not change")
	assert.True(t, updatedWord.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be newer")
}
