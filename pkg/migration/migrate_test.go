package migration_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wizact/te-reo-bot/pkg/migration"
	"github.com/wizact/te-reo-bot/pkg/repository"
)

const testJSON = `{
    "dictionary": [
        {
            "index": 1,
            "word": "Kia ora",
            "meaning": "Hello, be well",
            "link": "https://example.com",
            "photo": "kia-ora.jpg",
            "photo_attribution": "Photo of greeting"
        },
        {
            "index": 2,
            "word": "Korimako",
            "meaning": "The New Zealand bellbird",
            "link": "",
            "photo": "bellbird.jpeg",
            "photo_attribution": "Photo of bellbird"
        },
        {
            "index": 3,
            "word": "Test word",
            "meaning": "Test meaning",
            "link": "",
            "photo": "",
            "photo_attribution": ""
        }
    ]
}`

func setupTestDB(t *testing.T) (*sql.DB, repository.WordRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	err = repository.InitializeDatabase(db)
	require.NoError(t, err)

	repo := repository.NewSQLiteRepository(db)
	return db, repo
}

func TestParseDictionaryJSON(t *testing.T) {
	dict, err := migration.ParseDictionaryJSON([]byte(testJSON))
	
	assert.NoError(t, err, "Parsing valid JSON should succeed")
	assert.NotNil(t, dict)
	assert.Len(t, dict.Words, 3, "Should parse 3 words")
}

func TestParseDictionaryJSONInvalid(t *testing.T) {
	invalidJSON := `{"dictionary": "not an array"}`
	
	dict, err := migration.ParseDictionaryJSON([]byte(invalidJSON))
	
	assert.Error(t, err, "Parsing invalid JSON should fail")
	assert.Nil(t, dict)
}

func TestMigrateWords(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	migrator := migration.NewMigrator(repo)
	dict, err := migration.ParseDictionaryJSON([]byte(testJSON))
	require.NoError(t, err)

	err = migrator.MigrateWords(dict)
	assert.NoError(t, err, "Migration should succeed")

	// Verify words were imported
	count, err := repo.GetWordCount()
	assert.NoError(t, err)
	assert.Equal(t, 3, count, "Should have 3 words in database")

	// Verify a specific word
	word, err := repo.GetWordByDayIndex(1)
	assert.NoError(t, err)
	assert.Equal(t, "Kia ora", word.Word)
	assert.Equal(t, "Hello, be well", word.Meaning)
}

func TestMigrateFromFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_dictionary.json")
	
	err := os.WriteFile(testFile, []byte(testJSON), 0644)
	require.NoError(t, err)

	db, repo := setupTestDB(t)
	defer db.Close()

	migrator := migration.NewMigrator(repo)
	err = migrator.MigrateFromFile(testFile)
	assert.NoError(t, err, "Migration from file should succeed")

	// Verify migration
	count, err := repo.GetWordCount()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestMigrateIdempotent(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	migrator := migration.NewMigrator(repo)
	dict, err := migration.ParseDictionaryJSON([]byte(testJSON))
	require.NoError(t, err)

	// First migration
	err = migrator.MigrateWords(dict)
	assert.NoError(t, err)

	// Second migration should not create duplicates
	err = migrator.MigrateWords(dict)
	assert.NoError(t, err, "Second migration should succeed (idempotent)")

	// Should still have 3 words (not 6)
	count, err := repo.GetWordCount()
	assert.NoError(t, err)
	assert.Equal(t, 3, count, "Should not create duplicates")
}

func TestMigratePreservesAllFields(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	migrator := migration.NewMigrator(repo)
	dict, err := migration.ParseDictionaryJSON([]byte(testJSON))
	require.NoError(t, err)

	err = migrator.MigrateWords(dict)
	require.NoError(t, err)

	// Verify all fields are preserved
	word, err := repo.GetWordByDayIndex(1)
	require.NoError(t, err)
	
	assert.Equal(t, 1, *word.DayIndex)
	assert.Equal(t, "Kia ora", word.Word)
	assert.Equal(t, "Hello, be well", word.Meaning)
	assert.Equal(t, "https://example.com", word.Link)
	assert.Equal(t, "kia-ora.jpg", word.Photo)
	assert.Equal(t, "Photo of greeting", word.PhotoAttribution)
}

func TestMigrateEmptyFields(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	migrator := migration.NewMigrator(repo)
	dict, err := migration.ParseDictionaryJSON([]byte(testJSON))
	require.NoError(t, err)

	err = migrator.MigrateWords(dict)
	require.NoError(t, err)

	// Verify word with empty optional fields
	word, err := repo.GetWordByDayIndex(2)
	require.NoError(t, err)
	
	assert.Equal(t, "", word.Link, "Empty link should be preserved")
}

func TestMigrateDictionaryStructure(t *testing.T) {
	dict, err := migration.ParseDictionaryJSON([]byte(testJSON))
	require.NoError(t, err)

	// Verify structure matches
	assert.Len(t, dict.Words, 3)
	
	// Check first word fields
	firstWord := dict.Words[0]
	assert.Equal(t, 1, firstWord.Index)
	assert.Equal(t, "Kia ora", firstWord.Word)
	assert.Equal(t, "Hello, be well", firstWord.Meaning)
}
