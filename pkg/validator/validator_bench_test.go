package validator_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wizact/te-reo-bot/pkg/repository"
	"github.com/wizact/te-reo-bot/pkg/validator"
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
			DayIndex: &i,
			Word:     fmt.Sprintf("Word%d", i),
			Meaning:  fmt.Sprintf("Meaning for word %d", i),
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

func BenchmarkValidate(b *testing.B) {
	db, repo := setupBenchDB(b)
	defer db.Close()

	addBenchWords(b, repo, 366)

	v := validator.NewValidator(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := v.Validate()
		if err != nil {
			b.Fatal(err)
		}
	}
}

