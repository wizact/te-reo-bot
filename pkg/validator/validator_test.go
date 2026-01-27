package validator_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wizact/te-reo-bot/pkg/repository"
	"github.com/wizact/te-reo-bot/pkg/validator"
)

func setupTestDB(t *testing.T) (*sql.DB, repository.WordRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	err = repository.InitializeDatabase(db)
	require.NoError(t, err)

	repo := repository.NewSQLiteRepository(db)
	return db, repo
}

func addTestWords(t *testing.T, repo repository.WordRepository, indexes []int) {
	for _, idx := range indexes {
		word := &repository.Word{
			DayIndex: &idx,
			Word:     "Word" + string(rune(idx)),
			Meaning:  "Meaning" + string(rune(idx)),
		}
		tx, err := repo.BeginTx()
		require.NoError(t, err)
		err = repo.AddWord(tx, word)
		require.NoError(t, err)
		err = repo.CommitTx(tx)
		require.NoError(t, err)
	}
}

func TestValidateComplete366Words(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add all 366 words (1-366)
	indexes := make([]int, 366)
	for i := range indexes {
		indexes[i] = i + 1
	}
	addTestWords(t, repo, indexes)

	v := validator.NewValidator(repo)
	report, err := v.Validate()

	assert.NoError(t, err)
	assert.True(t, report.IsValid, "Should be valid with all 366 words")
	assert.Empty(t, report.MissingIndexes, "No missing indexes")
	assert.Empty(t, report.DuplicateIndexes, "No duplicate indexes")
	assert.Equal(t, 366, report.TotalWords)
}

func TestValidateMissingIndexes(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add words but skip index 5, 10, and 100
	indexes := []int{1, 2, 3, 4, 6, 7, 8, 9, 11, 12}
	addTestWords(t, repo, indexes)

	v := validator.NewValidator(repo)
	report, err := v.Validate()

	assert.NoError(t, err)
	assert.False(t, report.IsValid, "Should be invalid with missing indexes")
	assert.Contains(t, report.MissingIndexes, 5)
	assert.Contains(t, report.MissingIndexes, 10)
	assert.Contains(t, report.MissingIndexes, 100)
}

func TestValidateFewerThan366(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add only 100 words
	indexes := make([]int, 100)
	for i := range indexes {
		indexes[i] = i + 1
	}
	addTestWords(t, repo, indexes)

	v := validator.NewValidator(repo)
	report, err := v.Validate()

	assert.NoError(t, err)
	assert.False(t, report.IsValid, "Should be invalid with < 366 words")
	assert.Equal(t, 100, report.TotalWords)
	assert.Len(t, report.MissingIndexes, 266, "Should have 266 missing indexes")
}

func TestValidateEmpty(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	v := validator.NewValidator(repo)
	report, err := v.Validate()

	assert.NoError(t, err)
	assert.False(t, report.IsValid, "Should be invalid with no words")
	assert.Equal(t, 0, report.TotalWords)
	assert.Len(t, report.MissingIndexes, 366, "All 366 indexes should be missing")
}

func TestValidationReport(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add some words with gaps
	indexes := []int{1, 2, 5, 10}
	addTestWords(t, repo, indexes)

	v := validator.NewValidator(repo)
	report, err := v.Validate()

	assert.NoError(t, err)
	assert.NotNil(t, report)

	// Verify report structure
	assert.False(t, report.IsValid)
	assert.Equal(t, 4, report.TotalWords)
	assert.NotEmpty(t, report.MissingIndexes)

	// Check that we have proper error messages
	assert.NotEmpty(t, report.Errors)
}

func TestValidateWithUnassignedWords(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add complete 366 words
	indexes := make([]int, 366)
	for i := range indexes {
		indexes[i] = i + 1
	}
	addTestWords(t, repo, indexes)

	// Add some words without day_index
	unassigned := &repository.Word{
		Word:    "Unassigned word",
		Meaning: "Not assigned to a day",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, unassigned)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	v := validator.NewValidator(repo)
	report, err := v.Validate()

	assert.NoError(t, err)
	assert.True(t, report.IsValid, "Should be valid - unassigned words don't affect validation")
	assert.Equal(t, 366, report.TotalWords, "Should count only words with day_index")
}

func TestValidationReportMessages(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add only 5 words
	indexes := []int{1, 2, 3, 4, 5}
	addTestWords(t, repo, indexes)

	v := validator.NewValidator(repo)
	report, err := v.Validate()

	assert.NoError(t, err)
	assert.NotEmpty(t, report.Errors)

	// Check for helpful error message
	hasCountError := false
	for _, errMsg := range report.Errors {
		if len(errMsg) > 0 {
			hasCountError = true
			break
		}
	}
	assert.True(t, hasCountError, "Should have error messages")
}

func TestValidateIndexRange(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add all 366 indexes
	indexes := make([]int, 366)
	for i := range indexes {
		indexes[i] = i + 1
	}
	addTestWords(t, repo, indexes)

	v := validator.NewValidator(repo)
	report, err := v.Validate()

	assert.NoError(t, err)
	assert.True(t, report.IsValid)

	// Verify no indexes are out of range
	assert.Empty(t, report.MissingIndexes)
}
