package validator

import (
	"fmt"
	"sort"

	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
	"github.com/wizact/te-reo-bot/pkg/repository"
)

const requiredWordCount = 366

// Validator handles validation of dictionary data
type Validator struct {
	repo   repository.WordRepository
	logger logger.Logger
}

// NewValidator creates a new Validator
func NewValidator(repo repository.WordRepository) *Validator {
	return &Validator{
		repo:   repo,
		logger: logger.GetGlobalLogger(),
	}
}

// ValidationReport contains the results of validation
type ValidationReport struct {
	IsValid          bool     `json:"is_valid"`
	TotalWords       int      `json:"total_words"`
	MissingIndexes   []int    `json:"missing_indexes,omitempty"`
	DuplicateIndexes []int    `json:"duplicate_indexes,omitempty"`
	Errors           []string `json:"errors,omitempty"`
}

// Validate checks if the database contains exactly 366 unique day indexes (1-366)
func (v *Validator) Validate() (*ValidationReport, error) {
	v.logger.Info("Starting validation")

	report := &ValidationReport{
		IsValid:          true,
		MissingIndexes:   []int{},
		DuplicateIndexes: []int{},
		Errors:           []string{},
	}

	// Get all words by day index
	wordsByDay, err := v.repo.GetWordsByDayIndex()
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to get words for validation")
		appErr.WithContext("operation", "validate_get_words")
		v.logger.ErrorWithStack(err, "Validation query failed", logger.String("operation", "validate_get_words"))
		return nil, appErr
	}

	report.TotalWords = len(wordsByDay)
	v.logger.Info("Retrieved words for validation", logger.Int("word_count", report.TotalWords))

	// Check total count
	if report.TotalWords != requiredWordCount {
		report.IsValid = false
		report.Errors = append(report.Errors,
			fmt.Sprintf("Expected %d words with day_index, but found %d",
				requiredWordCount, report.TotalWords))
		v.logger.Error(nil, "Word count mismatch",
			logger.Int("expected", requiredWordCount),
			logger.Int("found", report.TotalWords))
	}

	// Find missing indexes (1-366)
	for i := 1; i <= requiredWordCount; i++ {
		if _, exists := wordsByDay[i]; !exists {
			report.MissingIndexes = append(report.MissingIndexes, i)
		}
	}

	if len(report.MissingIndexes) > 0 {
		report.IsValid = false
		report.Errors = append(report.Errors,
			fmt.Sprintf("Missing day indexes: %v", report.MissingIndexes))
		v.logger.Error(nil, "Missing day indexes", logger.Int("count", len(report.MissingIndexes)))
	}

	// Note: Duplicates are prevented by UNIQUE constraint in database,
	// but we check here for completeness
	if len(report.DuplicateIndexes) > 0 {
		report.IsValid = false
		report.Errors = append(report.Errors,
			fmt.Sprintf("Duplicate day indexes found: %v", report.DuplicateIndexes))
		v.logger.Error(nil, "Duplicate day indexes found", logger.Int("count", len(report.DuplicateIndexes)))
	}

	if report.IsValid {
		v.logger.Info("Validation passed")
	} else {
		v.logger.Error(nil, "Validation failed", logger.Int("error_count", len(report.Errors)))
	}

	return report, nil
}

// GetMissingIndexesRange returns missing indexes in a more compact format
func (r *ValidationReport) GetMissingIndexesRange() []string {
	if len(r.MissingIndexes) == 0 {
		return []string{}
	}

	// Sort indexes
	sorted := make([]int, len(r.MissingIndexes))
	copy(sorted, r.MissingIndexes)
	sort.Ints(sorted)

	ranges := []string{}
	start := sorted[0]
	end := sorted[0]

	for i := 1; i < len(sorted); i++ {
		if sorted[i] == end+1 {
			end = sorted[i]
		} else {
			if start == end {
				ranges = append(ranges, fmt.Sprintf("%d", start))
			} else {
				ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
			}
			start = sorted[i]
			end = sorted[i]
		}
	}

	// Add the last range
	if start == end {
		ranges = append(ranges, fmt.Sprintf("%d", start))
	} else {
		ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
	}

	return ranges
}
