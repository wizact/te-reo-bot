package curator

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/wizact/te-reo-bot/pkg/repository"
	wotd "github.com/wizact/te-reo-bot/pkg/wotd"
)

const (
	// MaxDayIndex is the highest valid day assignment in the SQLite dataset.
	MaxDayIndex = 366
	// MinDayIndex is the lowest valid day assignment in the SQLite dataset.
	MinDayIndex = 1
)

// SortColumn identifies a supported curator sort column.
type SortColumn string

const (
	SortByDayIndex SortColumn = "day"
	SortByID       SortColumn = "id"
	SortByWord     SortColumn = "word"
	SortByMeaning  SortColumn = "meaning"
	SortByCreated  SortColumn = "created"
	SortByUpdated  SortColumn = "updated"
)

// ListOptions configures how curator word rows are loaded.
type ListOptions struct {
	Filter         string
	SortColumn     SortColumn
	SortDescending bool
}

// WordInput contains fields required to create or edit words.
type WordInput struct {
	Word             string
	Meaning          string
	Link             string
	Photo            string
	PhotoAttribution string
	DayIndex         *int
	AutoAssign       bool
}

// ValidationReport summarizes curator lint results.
type ValidationReport struct {
	TotalWords      int
	AssignedWords   int
	UnassignedWords int
	MissingDays     []int
	DuplicateDays   []int
	OutOfRangeDays  []int
	EmptyWordIDs    []int
	EmptyMeaningIDs []int
}

// HasIssues returns true when the report contains curator-visible problems.
func (r ValidationReport) HasIssues() bool {
	return len(r.MissingDays) > 0 ||
		len(r.DuplicateDays) > 0 ||
		len(r.OutOfRangeDays) > 0 ||
		len(r.EmptyWordIDs) > 0 ||
		len(r.EmptyMeaningIDs) > 0
}

// Service contains curator business logic.
type Service struct {
	repo repository.WordRepository
}

// NewService creates a curator service backed by the repository layer.
func NewService(repo repository.WordRepository) *Service {
	return &Service{repo: repo}
}

// ListWords returns words filtered and sorted for the curator UI.
func (s *Service) ListWords(options ListOptions) ([]wotd.Word, error) {
	words, err := s.repo.GetAllWords()
	if err != nil {
		return nil, err
	}

	filtered := filterWords(words, options.Filter)
	sortWords(filtered, normalizeSortColumn(options.SortColumn), options.SortDescending)
	return filtered, nil
}

// AddWord creates a new word and optionally assigns the next available day.
func (s *Service) AddWord(input WordInput) (*wotd.Word, error) {
	normalized, err := normalizeWordInput(input)
	if err != nil {
		return nil, err
	}

	words, err := s.repo.GetAllWords()
	if err != nil {
		return nil, err
	}

	if normalized.AutoAssign {
		nextDay, err := nextAvailableDay(words)
		if err != nil {
			return nil, err
		}
		normalized.DayIndex = &nextDay
	}

	if normalized.DayIndex != nil {
		if err := ensureDayAvailable(words, 0, *normalized.DayIndex); err != nil {
			return nil, err
		}
	}

	word := &wotd.Word{
		DayIndex:         normalized.DayIndex,
		Word:             normalized.Word,
		Meaning:          normalized.Meaning,
		Link:             normalized.Link,
		Photo:            normalized.Photo,
		PhotoAttribution: normalized.PhotoAttribution,
	}

	tx, err := s.repo.BeginTx()
	if err != nil {
		return nil, err
	}

	if err := s.repo.AddWord(tx, word); err != nil {
		s.repo.RollbackTx(tx)
		return nil, err
	}

	if err := s.repo.CommitTx(tx); err != nil {
		s.repo.RollbackTx(tx)
		return nil, err
	}

	return word, nil
}

// UpdateWord updates curator-editable metadata for an existing word.
func (s *Service) UpdateWord(id int, input WordInput) (*wotd.Word, error) {
	normalized, err := normalizeWordInput(input)
	if err != nil {
		return nil, err
	}

	word, err := s.repo.GetWordByID(id)
	if err != nil {
		return nil, err
	}

	word.Word = normalized.Word
	word.Meaning = normalized.Meaning
	word.Link = normalized.Link
	word.Photo = normalized.Photo
	word.PhotoAttribution = normalized.PhotoAttribution

	tx, err := s.repo.BeginTx()
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateWordTx(tx, word); err != nil {
		s.repo.RollbackTx(tx)
		return nil, err
	}

	if err := s.repo.CommitTx(tx); err != nil {
		s.repo.RollbackTx(tx)
		return nil, err
	}

	return s.repo.GetWordByID(id)
}

// AssignDayIndex assigns, unassigns, or swaps day ownership for the given word.
func (s *Service) AssignDayIndex(wordID int, dayIndex *int) error {
	if dayIndex != nil {
		if err := validateDayIndex(*dayIndex); err != nil {
			return err
		}
	}

	words, err := s.repo.GetAllWords()
	if err != nil {
		return err
	}

	var source *wotd.Word
	var target *wotd.Word
	for i := range words {
		if words[i].ID == wordID {
			source = &words[i]
		}
		if dayIndex != nil && words[i].DayIndex != nil && *words[i].DayIndex == *dayIndex && words[i].ID != wordID {
			target = &words[i]
		}
	}

	if source == nil {
		return sql.ErrNoRows
	}

	if sameDay(source.DayIndex, dayIndex) {
		return nil
	}

	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}

	if target != nil {
		if err := s.repo.UpdateWordDayIndexByID(tx, source.ID, nil); err != nil {
			s.repo.RollbackTx(tx)
			return err
		}
		if err := s.repo.UpdateWordDayIndexByID(tx, target.ID, source.DayIndex); err != nil {
			s.repo.RollbackTx(tx)
			return err
		}
	}

	if err := s.repo.UpdateWordDayIndexByID(tx, source.ID, dayIndex); err != nil {
		s.repo.RollbackTx(tx)
		return err
	}

	if err := s.repo.CommitTx(tx); err != nil {
		s.repo.RollbackTx(tx)
		return err
	}

	return nil
}

// AutoAssignNextDay assigns the next unallocated day to the given word.
func (s *Service) AutoAssignNextDay(wordID int) (int, error) {
	words, err := s.repo.GetAllWords()
	if err != nil {
		return 0, err
	}

	nextDay, err := nextAvailableDay(words)
	if err != nil {
		return 0, err
	}

	if err := s.AssignDayIndex(wordID, &nextDay); err != nil {
		return 0, err
	}

	return nextDay, nil
}

// Validate checks curator-visible day assignment integrity.
func (s *Service) Validate() (*ValidationReport, error) {
	words, err := s.repo.GetAllWords()
	if err != nil {
		return nil, err
	}

	report := &ValidationReport{
		TotalWords: len(words),
	}

	seenDays := make(map[int]int)
	for _, word := range words {
		if strings.TrimSpace(word.Word) == "" {
			report.EmptyWordIDs = append(report.EmptyWordIDs, word.ID)
		}
		if strings.TrimSpace(word.Meaning) == "" {
			report.EmptyMeaningIDs = append(report.EmptyMeaningIDs, word.ID)
		}

		if word.DayIndex == nil {
			report.UnassignedWords++
			continue
		}

		report.AssignedWords++
		day := *word.DayIndex
		if day < MinDayIndex || day > MaxDayIndex {
			report.OutOfRangeDays = append(report.OutOfRangeDays, day)
			continue
		}

		seenDays[day]++
	}

	for day := MinDayIndex; day <= MaxDayIndex; day++ {
		switch seenDays[day] {
		case 0:
			report.MissingDays = append(report.MissingDays, day)
		case 1:
		default:
			report.DuplicateDays = append(report.DuplicateDays, day)
		}
	}

	sort.Ints(report.OutOfRangeDays)
	sort.Ints(report.EmptyWordIDs)
	sort.Ints(report.EmptyMeaningIDs)

	return report, nil
}

// FormatValidationReport renders a CLI-friendly validation report.
func FormatValidationReport(report *ValidationReport) string {
	if report == nil {
		return "no validation report"
	}

	lines := []string{
		fmt.Sprintf("total words: %d", report.TotalWords),
		fmt.Sprintf("assigned words: %d", report.AssignedWords),
		fmt.Sprintf("unassigned words: %d", report.UnassignedWords),
		fmt.Sprintf("missing day indexes: %s", formatDayList(report.MissingDays)),
		fmt.Sprintf("duplicate day indexes: %s", formatDayList(report.DuplicateDays)),
		fmt.Sprintf("out-of-range day indexes: %s", formatDayList(report.OutOfRangeDays)),
		fmt.Sprintf("empty word ids: %s", formatIDList(report.EmptyWordIDs)),
		fmt.Sprintf("empty meaning ids: %s", formatIDList(report.EmptyMeaningIDs)),
	}

	if report.HasIssues() {
		lines = append(lines, "status: invalid")
	} else {
		lines = append(lines, "status: valid")
	}

	return strings.Join(lines, "\n")
}

func normalizeWordInput(input WordInput) (WordInput, error) {
	input.Word = strings.TrimSpace(input.Word)
	input.Meaning = strings.TrimSpace(input.Meaning)
	input.Link = strings.TrimSpace(input.Link)
	input.Photo = strings.TrimSpace(input.Photo)
	input.PhotoAttribution = strings.TrimSpace(input.PhotoAttribution)

	if input.Word == "" {
		return input, fmt.Errorf("word is required")
	}
	if input.Meaning == "" {
		return input, fmt.Errorf("meaning is required")
	}
	if input.DayIndex != nil {
		if err := validateDayIndex(*input.DayIndex); err != nil {
			return input, err
		}
	}
	if input.AutoAssign && input.DayIndex != nil {
		return input, fmt.Errorf("day index cannot be set when auto-assign is enabled")
	}

	return input, nil
}

func filterWords(words []wotd.Word, query string) []wotd.Word {
	normalizedQuery := normalizeSearch(query)
	if normalizedQuery == "" {
		return append([]wotd.Word(nil), words...)
	}

	filtered := make([]wotd.Word, 0, len(words))
	for _, word := range words {
		if strings.Contains(normalizeSearch(buildSearchText(word)), normalizedQuery) {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

func buildSearchText(word wotd.Word) string {
	parts := []string{
		strconv.Itoa(word.ID),
		word.Word,
		word.Meaning,
		word.Link,
		word.Photo,
		word.PhotoAttribution,
	}
	if word.DayIndex != nil {
		parts = append(parts, strconv.Itoa(*word.DayIndex))
	}
	return strings.Join(parts, " ")
}

func normalizeSearch(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeSortColumn(column SortColumn) SortColumn {
	switch column {
	case SortByID, SortByWord, SortByMeaning, SortByCreated, SortByUpdated:
		return column
	default:
		return SortByDayIndex
	}
}

func sortWords(words []wotd.Word, column SortColumn, desc bool) {
	sort.SliceStable(words, func(i, j int) bool {
		if desc {
			return compareWords(words[j], words[i], column)
		}
		return compareWords(words[i], words[j], column)
	})
}

func compareWords(left, right wotd.Word, column SortColumn) bool {
	switch column {
	case SortByID:
		if left.ID == right.ID {
			return compareWordTieBreak(left, right)
		}
		return left.ID < right.ID
	case SortByWord:
		return compareStrings(left.Word, right.Word, left, right)
	case SortByMeaning:
		return compareStrings(left.Meaning, right.Meaning, left, right)
	case SortByCreated:
		if left.CreatedAt.Equal(right.CreatedAt) {
			return compareWordTieBreak(left, right)
		}
		return left.CreatedAt.Before(right.CreatedAt)
	case SortByUpdated:
		if left.UpdatedAt.Equal(right.UpdatedAt) {
			return compareWordTieBreak(left, right)
		}
		return left.UpdatedAt.Before(right.UpdatedAt)
	default:
		leftDay := sortableDayIndex(left.DayIndex)
		rightDay := sortableDayIndex(right.DayIndex)
		if leftDay == rightDay {
			return compareWordTieBreak(left, right)
		}
		return leftDay < rightDay
	}
}

func compareStrings(leftValue, rightValue string, left, right wotd.Word) bool {
	leftValue = strings.ToLower(leftValue)
	rightValue = strings.ToLower(rightValue)
	if leftValue == rightValue {
		return compareWordTieBreak(left, right)
	}
	return leftValue < rightValue
}

func compareWordTieBreak(left, right wotd.Word) bool {
	leftDay := sortableDayIndex(left.DayIndex)
	rightDay := sortableDayIndex(right.DayIndex)
	if leftDay != rightDay {
		return leftDay < rightDay
	}
	if strings.ToLower(left.Word) != strings.ToLower(right.Word) {
		return strings.ToLower(left.Word) < strings.ToLower(right.Word)
	}
	return left.ID < right.ID
}

func sortableDayIndex(dayIndex *int) int {
	if dayIndex == nil {
		return MaxDayIndex + 1
	}
	return *dayIndex
}

func validateDayIndex(dayIndex int) error {
	if dayIndex < MinDayIndex || dayIndex > MaxDayIndex {
		return fmt.Errorf("day index must be between %d and %d", MinDayIndex, MaxDayIndex)
	}
	return nil
}

func nextAvailableDay(words []wotd.Word) (int, error) {
	used := make(map[int]struct{})
	for _, word := range words {
		if word.DayIndex == nil {
			continue
		}
		if *word.DayIndex >= MinDayIndex && *word.DayIndex <= MaxDayIndex {
			used[*word.DayIndex] = struct{}{}
		}
	}

	for day := MinDayIndex; day <= MaxDayIndex; day++ {
		if _, exists := used[day]; !exists {
			return day, nil
		}
	}

	return 0, fmt.Errorf("no free day indexes remain")
}

func ensureDayAvailable(words []wotd.Word, ignoreWordID int, dayIndex int) error {
	for _, word := range words {
		if word.ID == ignoreWordID || word.DayIndex == nil {
			continue
		}
		if *word.DayIndex == dayIndex {
			return fmt.Errorf("day index %d is already assigned to %q", dayIndex, word.Word)
		}
	}
	return nil
}

func sameDay(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func formatDayList(days []int) string {
	if len(days) == 0 {
		return "none"
	}

	values := make([]string, 0, len(days))
	for _, day := range days {
		values = append(values, strconv.Itoa(day))
	}
	return strings.Join(values, ", ")
}

func formatIDList(ids []int) string {
	if len(ids) == 0 {
		return "none"
	}

	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, strconv.Itoa(id))
	}
	return strings.Join(values, ", ")
}
