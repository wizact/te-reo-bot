package curator_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wizact/te-reo-bot/pkg/curator"
	"github.com/wizact/te-reo-bot/pkg/repository"
	wotd "github.com/wizact/te-reo-bot/pkg/wotd"
)

func setupCuratorService(t *testing.T) (*sql.DB, repository.WordRepository, *curator.Service) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	err = repository.InitializeDatabase(db)
	require.NoError(t, err)

	repo := repository.NewSQLiteRepository(db)
	service := curator.NewService(repo)
	return db, repo, service
}

func addTestWord(t *testing.T, repo repository.WordRepository, word *wotd.Word) {
	t.Helper()
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)
}

func TestListWords_FiltersUnicodeText(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	dayOne := 1
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayOne, Word: "whānau", Meaning: "family"})
	addTestWord(t, repo, &wotd.Word{Word: "kai", Meaning: "food"})

	words, err := service.ListWords(curator.ListOptions{
		Filter:     "WHĀ",
		SortColumn: curator.SortByWord,
	})
	require.NoError(t, err)
	require.Len(t, words, 1)
	assert.Equal(t, "whānau", words[0].Word)
}

func TestAddWord_AutoAssignsNextDay(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	dayOne := 1
	dayThree := 3
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayOne, Word: "kia ora", Meaning: "hello"})
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayThree, Word: "aroha", Meaning: "love"})

	word, err := service.AddWord(curator.WordInput{
		Word:       "marae",
		Meaning:    "meeting grounds",
		AutoAssign: true,
	})
	require.NoError(t, err)
	require.NotNil(t, word.DayIndex)
	assert.Equal(t, 2, *word.DayIndex)
}

func TestAddWord_RejectsOccupiedDay(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	dayOne := 1
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayOne, Word: "kia ora", Meaning: "hello"})

	_, err := service.AddWord(curator.WordInput{
		Word:     "mōrena",
		Meaning:  "good morning",
		DayIndex: &dayOne,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already assigned")
}

func TestAssignDayIndex_SwapsAssignedWords(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	dayOne := 1
	dayTwo := 2
	first := &wotd.Word{DayIndex: &dayOne, Word: "tuatahi", Meaning: "first"}
	second := &wotd.Word{DayIndex: &dayTwo, Word: "tuarua", Meaning: "second"}
	addTestWord(t, repo, first)
	addTestWord(t, repo, second)

	err := service.AssignDayIndex(first.ID, &dayTwo)
	require.NoError(t, err)

	updatedFirst, err := repo.GetWordByID(first.ID)
	require.NoError(t, err)
	updatedSecond, err := repo.GetWordByID(second.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedFirst.DayIndex)
	require.NotNil(t, updatedSecond.DayIndex)
	assert.Equal(t, 2, *updatedFirst.DayIndex)
	assert.Equal(t, 1, *updatedSecond.DayIndex)
}

func TestAssignDayIndex_MovesAssignedWordToUnassigned(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	dayOne := 1
	word := &wotd.Word{DayIndex: &dayOne, Word: "wai", Meaning: "water"}
	addTestWord(t, repo, word)

	err := service.AssignDayIndex(word.ID, nil)
	require.NoError(t, err)

	updated, err := repo.GetWordByID(word.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.DayIndex)
}

func TestValidate_ReportsMissingDaysAndEmptyFields(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	dayOne := 1
	dayTwo := 2
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayOne, Word: "kia ora", Meaning: "hello"})
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayTwo, Word: "", Meaning: ""})
	addTestWord(t, repo, &wotd.Word{Word: "whare", Meaning: "house"})

	report, err := service.Validate()
	require.NoError(t, err)
	assert.Equal(t, 3, report.TotalWords)
	assert.Equal(t, 2, report.AssignedWords)
	assert.Equal(t, 1, report.UnassignedWords)
	assert.Contains(t, report.MissingDays, 3)
	assert.NotEmpty(t, report.EmptyWordIDs)
	assert.NotEmpty(t, report.EmptyMeaningIDs)
	assert.True(t, report.HasIssues())
}
