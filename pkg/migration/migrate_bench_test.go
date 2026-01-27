package migration_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wizact/te-reo-bot/pkg/migration"
	"github.com/wizact/te-reo-bot/pkg/repository"
)

func setupBenchDB(b *testing.B) (*sql.DB, repository.WordRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatal(err)
	}

	err = repository.InitializeDatabase(db)
	if err != nil {
		b.Fatal(err)
	}

	repo := repository.NewSQLiteRepository(db)
	return db, repo
}

func BenchmarkMigrate(b *testing.B) {
	dict := createBenchDictionary(366)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		db, repo := setupBenchDB(b)
		migrator := migration.NewMigrator(repo)
		b.StartTimer()

		err := migrator.MigrateWords(dict)
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()
		db.Close()
		b.StartTimer()
	}
}

func BenchmarkParseDictionaryJSON(b *testing.B) {
	// Create sample JSON
	dict := createBenchDictionary(366)
	testJSON := `{"dictionary": [`
	for i, w := range dict.Words {
		if i > 0 {
			testJSON += ","
		}
		testJSON += fmt.Sprintf(`{"index":%d,"word":"%s","meaning":"%s","link":"","photo":"","photo_attribution":""}`,
			w.Index, w.Word, w.Meaning)
	}
	testJSON += `]}`

	data := []byte(testJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := migration.ParseDictionaryJSON(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func createBenchDictionary(count int) *migration.Dictionary {
	words := make([]migration.DictionaryWord, count)
	for i := 0; i < count; i++ {
		words[i] = migration.DictionaryWord{
			Index:            i + 1,
			Word:             fmt.Sprintf("Word%d", i+1),
			Meaning:          fmt.Sprintf("Meaning%d", i+1),
			Link:             "https://example.com",
			Photo:            "photo.jpg",
			PhotoAttribution: "Attribution",
		}
	}
	return &migration.Dictionary{Words: words}
}
