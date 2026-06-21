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

func TestAddWord_RejectsMeaningRequired(t *testing.T) {
	db, _, service := setupCuratorService(t)
	defer db.Close()

	_, err := service.AddWord(curator.WordInput{Word: "aho", Meaning: ""})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "meaning is required")
}

func TestAddWord_RejectsAutoAssignWithExplicitDay(t *testing.T) {
	db, _, service := setupCuratorService(t)
	defer db.Close()

	dayFive := 5
	_, err := service.AddWord(curator.WordInput{
		Word:       "manu",
		Meaning:    "bird",
		AutoAssign: true,
		DayIndex:   &dayFive,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "day index cannot be set when auto-assign is enabled")
}

func TestListWords_SortByID_Descending(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	addTestWord(t, repo, &wotd.Word{Word: "ahi", Meaning: "fire"})
	addTestWord(t, repo, &wotd.Word{Word: "wai", Meaning: "water"})
	addTestWord(t, repo, &wotd.Word{Word: "rangi", Meaning: "sky"})

	words, err := service.ListWords(curator.ListOptions{SortColumn: curator.SortByID, SortDescending: true})
	require.NoError(t, err)
	require.Len(t, words, 3)
	// Descending by ID: rangi, wai, ahi
	assert.Equal(t, "rangi", words[0].Word)
	assert.Equal(t, "ahi", words[2].Word)
}

func TestListWords_SortByMeaning_TieBreak(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	// Two words with same meaning, different words — exercises compareStrings tie-break
	dayOne := 1
	dayTwo := 2
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayTwo, Word: "rangi", Meaning: "sky"})
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayOne, Word: "ātea", Meaning: "sky"})

	words, err := service.ListWords(curator.ListOptions{SortColumn: curator.SortByMeaning})
	require.NoError(t, err)
	require.Len(t, words, 2)
	// Same meaning, tie-break by day: dayOne (ātea) first
	assert.Equal(t, "ātea", words[0].Word)
}

func TestListWords_SortByDayIndex_TieBreak(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	// Two unassigned words — exercises compareWordTieBreak by word
	addTestWord(t, repo, &wotd.Word{Word: "rangi", Meaning: "sky"})
	addTestWord(t, repo, &wotd.Word{Word: "ahi", Meaning: "fire"})

	words, err := service.ListWords(curator.ListOptions{SortColumn: curator.SortByDayIndex})
	require.NoError(t, err)
	require.Len(t, words, 2)
	// Both unassigned, tie-break by word alphabetically: ahi < rangi
	assert.Equal(t, "ahi", words[0].Word)
}

func TestUpdateWord_UpdatesFields(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	dayOne := 1
	w := &wotd.Word{DayIndex: &dayOne, Word: "rangi", Meaning: "sky"}
	addTestWord(t, repo, w)

	updated, err := service.UpdateWord(w.ID, curator.WordInput{
		Word:    "rangi",
		Meaning: "sky and weather",
		Link:    "https://example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, "rangi", updated.Word)
	assert.Equal(t, "sky and weather", updated.Meaning)
	assert.Equal(t, "https://example.com", updated.Link)
}

func TestUpdateWord_RejectsEmptyWord(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	w := &wotd.Word{Word: "wai", Meaning: "water"}
	addTestWord(t, repo, w)

	_, err := service.UpdateWord(w.ID, curator.WordInput{Word: "", Meaning: "water"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "word is required")
}

func TestAutoAssignNextDay_AssignsFirstFreeDay(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	dayOne := 1
	dayTwo := 2
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayOne, Word: "ahi", Meaning: "fire"})
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayTwo, Word: "ua", Meaning: "rain"})
	w := &wotd.Word{Word: "hau", Meaning: "wind"}
	addTestWord(t, repo, w)

	day, err := service.AutoAssignNextDay(w.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, day)

	fetched, err := repo.GetWordByID(w.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched.DayIndex)
	assert.Equal(t, 3, *fetched.DayIndex)
}

func TestFormatValidationReport_NilReport(t *testing.T) {
	result := curator.FormatValidationReport(nil)
	assert.Equal(t, "no validation report", result)
}

func TestFormatValidationReport_ValidReport(t *testing.T) {
	report := &curator.ValidationReport{
		TotalWords:      2,
		AssignedWords:   2,
		UnassignedWords: 0,
	}
	result := curator.FormatValidationReport(report)
	assert.Contains(t, result, "total words: 2")
	assert.Contains(t, result, "assigned words: 2")
	assert.Contains(t, result, "unassigned words: 0")
	assert.Contains(t, result, "status: valid")
}

func TestFormatValidationReport_InvalidReport(t *testing.T) {
	report := &curator.ValidationReport{
		TotalWords:      1,
		AssignedWords:   1,
		MissingDays:     []int{2, 3},
		DuplicateDays:   []int{1},
		EmptyWordIDs:    []int{5},
		EmptyMeaningIDs: []int{7},
	}
	result := curator.FormatValidationReport(report)
	assert.Contains(t, result, "missing day indexes: 2, 3")
	assert.Contains(t, result, "duplicate day indexes: 1")
	assert.Contains(t, result, "empty word ids: 5")
	assert.Contains(t, result, "empty meaning ids: 7")
	assert.Contains(t, result, "status: invalid")
}

func TestListWords_SortByWord(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	addTestWord(t, repo, &wotd.Word{Word: "rangi", Meaning: "sky"})
	addTestWord(t, repo, &wotd.Word{Word: "ahi", Meaning: "fire"})
	addTestWord(t, repo, &wotd.Word{Word: "moana", Meaning: "ocean"})

	words, err := service.ListWords(curator.ListOptions{SortColumn: curator.SortByWord})
	require.NoError(t, err)
	require.Len(t, words, 3)
	assert.Equal(t, "ahi", words[0].Word)
	assert.Equal(t, "moana", words[1].Word)
	assert.Equal(t, "rangi", words[2].Word)
}

func TestListWords_SortByWordDescending(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	addTestWord(t, repo, &wotd.Word{Word: "rangi", Meaning: "sky"})
	addTestWord(t, repo, &wotd.Word{Word: "ahi", Meaning: "fire"})
	addTestWord(t, repo, &wotd.Word{Word: "moana", Meaning: "ocean"})

	words, err := service.ListWords(curator.ListOptions{SortColumn: curator.SortByWord, SortDescending: true})
	require.NoError(t, err)
	require.Len(t, words, 3)
	assert.Equal(t, "rangi", words[0].Word)
	assert.Equal(t, "moana", words[1].Word)
	assert.Equal(t, "ahi", words[2].Word)
}

func TestListWords_SortByMeaning(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	addTestWord(t, repo, &wotd.Word{Word: "wai", Meaning: "water"})
	addTestWord(t, repo, &wotd.Word{Word: "ahi", Meaning: "fire"})

	words, err := service.ListWords(curator.ListOptions{SortColumn: curator.SortByMeaning})
	require.NoError(t, err)
	require.Len(t, words, 2)
	assert.Equal(t, "ahi", words[0].Word)
	assert.Equal(t, "wai", words[1].Word)
}

func TestListWords_SortByID(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	addTestWord(t, repo, &wotd.Word{Word: "rangi", Meaning: "sky"})
	addTestWord(t, repo, &wotd.Word{Word: "ahi", Meaning: "fire"})

	words, err := service.ListWords(curator.ListOptions{SortColumn: curator.SortByID})
	require.NoError(t, err)
	require.Len(t, words, 2)
	assert.Equal(t, "rangi", words[0].Word)
	assert.Equal(t, "ahi", words[1].Word)
}

func TestListWords_SortByDayIndex_UnassignedLast(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	dayThree := 3
	dayOne := 1
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayThree, Word: "rangi", Meaning: "sky"})
	addTestWord(t, repo, &wotd.Word{Word: "ahi", Meaning: "fire"})
	addTestWord(t, repo, &wotd.Word{DayIndex: &dayOne, Word: "wai", Meaning: "water"})

	words, err := service.ListWords(curator.ListOptions{SortColumn: curator.SortByDayIndex})
	require.NoError(t, err)
	require.Len(t, words, 3)
	assert.Equal(t, "wai", words[0].Word)
	assert.Equal(t, "rangi", words[1].Word)
	assert.Equal(t, "ahi", words[2].Word)
}

func TestListWords_SortByUpdated(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	addTestWord(t, repo, &wotd.Word{Word: "rangi", Meaning: "sky"})
	addTestWord(t, repo, &wotd.Word{Word: "ahi", Meaning: "fire"})

	words, err := service.ListWords(curator.ListOptions{SortColumn: curator.SortByUpdated})
	require.NoError(t, err)
	require.Len(t, words, 2)
	// both have same timestamp, tie-break by day index then word
	assert.Len(t, words, 2)
}

func TestAddWord_WithExplicitDayIndex(t *testing.T) {
	db, _, service := setupCuratorService(t)
	defer db.Close()

	dayFive := 5
	word, err := service.AddWord(curator.WordInput{
		Word:     "tūī",
		Meaning:  "native bird",
		DayIndex: &dayFive,
	})
	require.NoError(t, err)
	require.NotNil(t, word.DayIndex)
	assert.Equal(t, 5, *word.DayIndex)
}

func TestAddWord_NoDay(t *testing.T) {
	db, _, service := setupCuratorService(t)
	defer db.Close()

	word, err := service.AddWord(curator.WordInput{
		Word:    "manu",
		Meaning: "bird",
	})
	require.NoError(t, err)
	assert.Nil(t, word.DayIndex)
}

func TestAddWord_RejectsInvalidDayIndex(t *testing.T) {
	db, _, service := setupCuratorService(t)
	defer db.Close()

	invalid := 400
	_, err := service.AddWord(curator.WordInput{
		Word:     "aho",
		Meaning:  "cord",
		DayIndex: &invalid,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "day index must be between")
}

func TestAssignDayIndex_SameDay_NoOp(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	dayOne := 1
	w := &wotd.Word{DayIndex: &dayOne, Word: "koru", Meaning: "loop"}
	addTestWord(t, repo, w)

	err := service.AssignDayIndex(w.ID, &dayOne)
	require.NoError(t, err)

	fetched, err := repo.GetWordByID(w.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched.DayIndex)
	assert.Equal(t, 1, *fetched.DayIndex)
}

func TestAssignDayIndex_RejectsOutOfRangeDay(t *testing.T) {
	db, repo, service := setupCuratorService(t)
	defer db.Close()

	w := &wotd.Word{Word: "ngahere", Meaning: "forest"}
	addTestWord(t, repo, w)

	invalid := 400
	err := service.AssignDayIndex(w.ID, &invalid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "day index must be between")
}
