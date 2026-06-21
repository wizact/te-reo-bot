package curator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wotd "github.com/wizact/te-reo-bot/pkg/wotd"
)

// Tests for pure helper functions in tui.go that do not require a terminal.

func TestFormatDay_Nil(t *testing.T) {
	assert.Equal(t, "—", formatDay(nil))
}

func TestFormatDay_Value(t *testing.T) {
	day := 42
	assert.Equal(t, "42", formatDay(&day))
}

func TestFormatDayInput_Nil(t *testing.T) {
	assert.Equal(t, "", formatDayInput(nil))
}

func TestFormatDayInput_Value(t *testing.T) {
	day := 100
	assert.Equal(t, "100", formatDayInput(&day))
}

func TestEmptyPlaceholder_Empty(t *testing.T) {
	assert.Equal(t, "—", emptyPlaceholder(""))
	assert.Equal(t, "—", emptyPlaceholder("   "))
}

func TestEmptyPlaceholder_NonEmpty(t *testing.T) {
	assert.Equal(t, "https://example.com", emptyPlaceholder("https://example.com"))
}

func TestTruncate_FitsWithin(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
}

func TestTruncate_Exact(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 5))
}

func TestTruncate_Truncated(t *testing.T) {
	result := truncate("hello world", 6)
	assert.Equal(t, "hello…", result)
	assert.Equal(t, 6, len([]rune(result)))
}

func TestTruncate_WidthOne(t *testing.T) {
	assert.Equal(t, "h", truncate("hello", 1))
}

func TestTruncate_ZeroWidth(t *testing.T) {
	assert.Equal(t, "", truncate("hello", 0))
}

func TestTruncate_NegativeWidth(t *testing.T) {
	assert.Equal(t, "", truncate("hello", -1))
}

func TestTruncate_Unicode(t *testing.T) {
	// "whānau" has 6 runes; truncating to 4 should produce "whā…"
	result := truncate("whānau", 4)
	assert.Equal(t, 4, len([]rune(result)))
}

func TestVisibleWidth_ASCII(t *testing.T) {
	assert.Equal(t, 5, visibleWidth("hello"))
}

func TestVisibleWidth_Unicode(t *testing.T) {
	assert.Equal(t, 6, visibleWidth("whānau"))
}

func TestFormatListRow_Width(t *testing.T) {
	row := formatListRow("1", "42", "aroha", "love")
	assert.NotEmpty(t, row)
	// Row should contain the supplied values
	assert.Contains(t, row, "aroha")
	assert.Contains(t, row, "love")
}

func TestFormatListRow_LongValues_Truncated(t *testing.T) {
	longWord := "averylongwordthatexceedscolumnwidth"
	row := formatListRow("1", "1", longWord, "meaning")
	// The long word column is 22 wide; truncation should apply
	assert.NotContains(t, row, longWord)
}

func TestFormatWordRow_Basic(t *testing.T) {
	day := 7
	word := wotd.Word{ID: 1, DayIndex: &day, Word: "aroha", Meaning: "love"}
	row := formatWordRow(word)
	assert.Contains(t, row, "aroha")
	assert.Contains(t, row, "love")
	assert.Contains(t, row, "7")
}

func TestFormatWordRow_NoDayIndex(t *testing.T) {
	word := wotd.Word{ID: 2, Word: "kai", Meaning: "food"}
	row := formatWordRow(word)
	assert.Contains(t, row, "kai")
	assert.Contains(t, row, "food")
	assert.Contains(t, row, "—")
}

func TestWordInputFromForm_Basic(t *testing.T) {
	values := map[string]string{
		"word":              "aroha",
		"meaning":           "love",
		"link":              "https://example.com",
		"photo":             "",
		"photo_attribution": "",
		"day_index":         "",
	}
	input, err := wordInputFromForm(values)
	require.NoError(t, err)
	assert.Equal(t, "aroha", input.Word)
	assert.Equal(t, "love", input.Meaning)
	assert.Equal(t, "https://example.com", input.Link)
	assert.Nil(t, input.DayIndex)
	assert.False(t, input.AutoAssign)
}

func TestWordInputFromForm_AutoAssign(t *testing.T) {
	values := map[string]string{
		"word":              "rangi",
		"meaning":           "sky",
		"link":              "",
		"photo":             "",
		"photo_attribution": "",
		"day_index":         "auto",
	}
	input, err := wordInputFromForm(values)
	require.NoError(t, err)
	assert.True(t, input.AutoAssign)
	assert.Nil(t, input.DayIndex)
}

func TestWordInputFromForm_AutoAssignCaseInsensitive(t *testing.T) {
	values := map[string]string{
		"word":      "rangi",
		"meaning":   "sky",
		"day_index": "AUTO",
	}
	input, err := wordInputFromForm(values)
	require.NoError(t, err)
	assert.True(t, input.AutoAssign)
}

func TestWordInputFromForm_NumericDay(t *testing.T) {
	values := map[string]string{
		"word":      "moana",
		"meaning":   "ocean",
		"day_index": "42",
	}
	input, err := wordInputFromForm(values)
	require.NoError(t, err)
	require.NotNil(t, input.DayIndex)
	assert.Equal(t, 42, *input.DayIndex)
}

func TestWordInputFromForm_InvalidDay(t *testing.T) {
	values := map[string]string{
		"word":      "wai",
		"meaning":   "water",
		"day_index": "notanumber",
	}
	_, err := wordInputFromForm(values)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "day index must be an integer or auto")
}

func TestCurrentWord_EmptyList(t *testing.T) {
	tui := &TUI{}
	assert.Nil(t, tui.currentWord())
}

func TestCurrentWord_OutOfBounds(t *testing.T) {
	tui := &TUI{words: []wotd.Word{{Word: "kai"}}, selected: 5}
	assert.Nil(t, tui.currentWord())
}

func TestCurrentWord_Valid(t *testing.T) {
	tui := &TUI{words: []wotd.Word{{Word: "ahi"}, {Word: "ua"}}, selected: 1}
	w := tui.currentWord()
	require.NotNil(t, w)
	assert.Equal(t, "ua", w.Word)
}

func TestHandleError_SetsStatus(t *testing.T) {
	tui := &TUI{}
	result := tui.handleError(assert.AnError)
	assert.Nil(t, result)
	assert.Equal(t, assert.AnError.Error(), tui.status)
}

func TestSortWords_ByCreated_NonEqual(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)
	words := []wotd.Word{
		{ID: 1, Word: "rangi", CreatedAt: now},
		{ID: 2, Word: "ahi", CreatedAt: earlier},
	}
	sortWords(words, SortByCreated, false)
	assert.Equal(t, "ahi", words[0].Word)
	assert.Equal(t, "rangi", words[1].Word)
}

func TestSortWords_ByUpdated_NonEqual(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)
	words := []wotd.Word{
		{ID: 1, Word: "rangi", UpdatedAt: now},
		{ID: 2, Word: "ahi", UpdatedAt: earlier},
	}
	sortWords(words, SortByUpdated, false)
	assert.Equal(t, "ahi", words[0].Word)
	assert.Equal(t, "rangi", words[1].Word)
}

func TestSortWords_ByID_TieBreakSameIDSameWord(t *testing.T) {
	// Exercises the ID tie-break (left.ID < right.ID) path in compareWordTieBreak
	// when day and word are identical — IDs differ
	day := 1
	words := []wotd.Word{
		{ID: 2, Word: "rangi", DayIndex: &day},
		{ID: 1, Word: "rangi", DayIndex: &day},
	}
	sortWords(words, SortByID, false)
	assert.Equal(t, 1, words[0].ID)
	assert.Equal(t, 2, words[1].ID)
}
