package generator_test

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wizact/te-reo-bot/pkg/generator"
	"github.com/wizact/te-reo-bot/pkg/repository"
)

func setupTestDB(t *testing.T) (*sql.DB, repository.WordRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	err = repository.InitializeDatabase(db)
	require.NoError(t, err)

	repo := repository.NewSQLiteRepository(db)
	return db, repo
}

func addTestWords(t *testing.T, repo repository.WordRepository) {
	words := []struct {
		dayIndex int
		word     string
		meaning  string
		link     string
		photo    string
		attr     string
	}{
		{1, "Kia ora", "Hello, be well", "https://example.com", "kia-ora.jpg", "Photo 1"},
		{2, "Kōtiro", "Girl", "", "kotiro.jpg", "Photo 2"},
		{3, "Tama", "Boy", "", "", ""},
	}

	tx, err := repo.BeginTx()
	require.NoError(t, err)
	for _, w := range words {
		word := &repository.Word{
			DayIndex:         &w.dayIndex,
			Word:             w.word,
			Meaning:          w.meaning,
			Link:             w.link,
			Photo:            w.photo,
			PhotoAttribution: w.attr,
		}
		err := repo.AddWord(tx, word)
		require.NoError(t, err)
	}
	err = repo.CommitTx(tx)
	require.NoError(t, err)
}

func TestGenerateJSON(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	addTestWords(t, repo)

	gen := generator.NewGenerator(repo)
	jsonBytes, err := gen.GenerateJSON()

	assert.NoError(t, err)
	assert.NotEmpty(t, jsonBytes)

	// Verify it's valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NoError(t, err)
}

func TestGenerateJSONStructure(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	addTestWords(t, repo)

	gen := generator.NewGenerator(repo)
	jsonBytes, err := gen.GenerateJSON()
	require.NoError(t, err)

	// Parse JSON to verify structure
	var dict struct {
		Dictionary []map[string]interface{} `json:"dictionary"`
	}
	err = json.Unmarshal(jsonBytes, &dict)
	require.NoError(t, err)

	assert.Len(t, dict.Dictionary, 3)
	
	// Verify first word
	firstWord := dict.Dictionary[0]
	assert.Equal(t, float64(1), firstWord["index"])
	assert.Equal(t, "Kia ora", firstWord["word"])
	assert.Equal(t, "Hello, be well", firstWord["meaning"])
}

func TestGenerateJSONSorted(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add words in random order
	indexes := []int{3, 1, 2}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	for _, idx := range indexes {
		word := &repository.Word{
			DayIndex: &idx,
			Word:     "Word" + string(rune(idx)),
			Meaning:  "Meaning" + string(rune(idx)),
		}
		err := repo.AddWord(tx, word)
		require.NoError(t, err)
	}
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	gen := generator.NewGenerator(repo)
	jsonBytes, err := gen.GenerateJSON()
	require.NoError(t, err)

	// Parse and verify order
	var dict struct {
		Dictionary []struct {
			Index int `json:"index"`
		} `json:"dictionary"`
	}
	err = json.Unmarshal(jsonBytes, &dict)
	require.NoError(t, err)

	// Should be sorted by day_index
	assert.Equal(t, 1, dict.Dictionary[0].Index)
	assert.Equal(t, 2, dict.Dictionary[1].Index)
	assert.Equal(t, 3, dict.Dictionary[2].Index)
}

func TestGenerateJSONPreservesUTF8(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add word with Māori characters
	dayIndex := 1
	word := &repository.Word{
		DayIndex: &dayIndex,
		Word:     "Kākāpō",
		Meaning:  "The world's only flightless parrot",
		Photo:    "kakapo.jpg",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	gen := generator.NewGenerator(repo)
	jsonBytes, err := gen.GenerateJSON()
	require.NoError(t, err)

	// Verify UTF-8 encoding is preserved
	var dict struct {
		Dictionary []struct {
			Word string `json:"word"`
		} `json:"dictionary"`
	}
	err = json.Unmarshal(jsonBytes, &dict)
	require.NoError(t, err)

	assert.Equal(t, "Kākāpō", dict.Dictionary[0].Word)
}

func TestGenerateJSONEmptyFields(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add word with empty optional fields
	dayIndex := 1
	word := &repository.Word{
		DayIndex: &dayIndex,
		Word:     "Test",
		Meaning:  "Test meaning",
		Link:     "",
		Photo:    "",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	gen := generator.NewGenerator(repo)
	jsonBytes, err := gen.GenerateJSON()
	require.NoError(t, err)

	// Verify empty fields are included
	var dict struct {
		Dictionary []struct {
			Link  string `json:"link"`
			Photo string `json:"photo"`
		} `json:"dictionary"`
	}
	err = json.Unmarshal(jsonBytes, &dict)
	require.NoError(t, err)

	assert.Equal(t, "", dict.Dictionary[0].Link)
	assert.Equal(t, "", dict.Dictionary[0].Photo)
}

func TestGenerateToFile(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	addTestWords(t, repo)

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "dictionary.json")

	gen := generator.NewGenerator(repo)
	err := gen.GenerateToFile(outputFile)
	assert.NoError(t, err)

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	var dict map[string]interface{}
	err = json.Unmarshal(data, &dict)
	assert.NoError(t, err)
}

func TestGenerateJSONPrettyFormat(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	addTestWords(t, repo)

	gen := generator.NewGenerator(repo)
	gen.SetPrettyPrint(true)
	
	jsonBytes, err := gen.GenerateJSON()
	require.NoError(t, err)

	// Pretty-printed JSON should have newlines
	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, "\n")
	assert.Contains(t, jsonStr, "  ") // indentation
}

func TestGenerateJSONCompactFormat(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	addTestWords(t, repo)

	gen := generator.NewGenerator(repo)
	gen.SetPrettyPrint(false)
	
	jsonBytes, err := gen.GenerateJSON()
	require.NoError(t, err)

	// Compact JSON should not have unnecessary whitespace
	jsonStr := string(jsonBytes)
	assert.NotContains(t, jsonStr, "\n  ") // no indentation
}

func TestGenerateOnlyDayIndexedWords(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add words with day_index
	dayIndex1 := 1
	word1 := &repository.Word{
		DayIndex: &dayIndex1,
		Word:     "Assigned",
		Meaning:  "Has day index",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word1)
	require.NoError(t, err)

	// Add word without day_index
	word2 := &repository.Word{
		Word:    "Unassigned",
		Meaning: "No day index",
	}
	err = repo.AddWord(tx, word2)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	gen := generator.NewGenerator(repo)
	jsonBytes, err := gen.GenerateJSON()
	require.NoError(t, err)

	// Parse and verify only day-indexed words are included
	var dict struct {
		Dictionary []struct {
			Word string `json:"word"`
		} `json:"dictionary"`
	}
	err = json.Unmarshal(jsonBytes, &dict)
	require.NoError(t, err)

	assert.Len(t, dict.Dictionary, 1, "Only words with day_index should be generated")
	assert.Equal(t, "Assigned", dict.Dictionary[0].Word)
}

// Error scenario tests

func TestGeneratorErrorScenarios(t *testing.T) {
	t.Run("Write to invalid directory", func(t *testing.T) {
		db, repo := setupTestDB(t)
		defer db.Close()

		addTestWords(t, repo)

		gen := generator.NewGenerator(repo)
		// Try to write to non-existent directory without creating it
		err := gen.GenerateToFile("/nonexistent/directory/output.json")
		assert.Error(t, err)
	})

	t.Run("Empty database", func(t *testing.T) {
		db, repo := setupTestDB(t)
		defer db.Close()

		gen := generator.NewGenerator(repo)
		jsonBytes, err := gen.GenerateJSON()

		assert.NoError(t, err)
		// Verify it contains an empty dictionary array (format may vary)
		var dict struct {
			Dictionary []interface{} `json:"dictionary"`
		}
		json.Unmarshal(jsonBytes, &dict)
		assert.Len(t, dict.Dictionary, 0)
	})

	t.Run("Generate with closed database", func(t *testing.T) {
		db, repo := setupTestDB(t)
		addTestWords(t, repo)
		db.Close() // Close database first

		gen := generator.NewGenerator(repo)
		_, err := gen.GenerateJSON()
		assert.Error(t, err, "Should fail with closed database")
	})
}
