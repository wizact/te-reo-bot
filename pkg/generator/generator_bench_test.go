package generator_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wizact/te-reo-bot/pkg/generator"
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

func addBenchWords(b *testing.B, repo repository.WordRepository, count int) {
	tx, err := repo.BeginTx()
	if err != nil {
		b.Fatal(err)
	}
	for i := 1; i <= count; i++ {
		word := &repository.Word{
			DayIndex:         &i,
			Word:             fmt.Sprintf("Word%d", i),
			Meaning:          fmt.Sprintf("Meaning for word %d", i),
			Link:             "https://example.com",
			Photo:            "photo.jpg",
			PhotoAttribution: "Attribution",
		}
		err := repo.AddWord(tx, word)
		if err != nil {
			b.Fatal(err)
		}
	}
	err = repo.CommitTx(tx)
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkGenerateJSON(b *testing.B) {
	db, repo := setupBenchDB(b)
	defer db.Close()

	addBenchWords(b, repo, 366)

	gen := generator.NewGenerator(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}


func BenchmarkGenerateJSONPrettyPrint(b *testing.B) {
	db, repo := setupBenchDB(b)
	defer db.Close()

	addBenchWords(b, repo, 366)

	gen := generator.NewGenerator(repo)
	gen.SetPrettyPrint(true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateJSONCompact(b *testing.B) {
	db, repo := setupBenchDB(b)
	defer db.Close()

	addBenchWords(b, repo, 366)

	gen := generator.NewGenerator(repo)
	gen.SetPrettyPrint(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}
