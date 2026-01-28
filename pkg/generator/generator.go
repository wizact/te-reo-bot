package generator

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
	"github.com/wizact/te-reo-bot/pkg/repository"
)

// Generator handles generation of dictionary.json from database
type Generator struct {
	repo        repository.WordRepository
	logger      logger.Logger
	prettyPrint bool
}

// NewGenerator creates a new Generator
func NewGenerator(repo repository.WordRepository) *Generator {
	return &Generator{
		repo:        repo,
		logger:      logger.GetGlobalLogger(),
		prettyPrint: true, // Default to pretty-print for human readability
	}
}

// SetPrettyPrint sets whether to format JSON with indentation
func (g *Generator) SetPrettyPrint(pretty bool) {
	g.prettyPrint = pretty
}

// DictionaryWord represents a word in the output JSON format
type DictionaryWord struct {
	Index            int    `json:"index"`
	Word             string `json:"word"`
	Meaning          string `json:"meaning"`
	Link             string `json:"link"`
	Photo            string `json:"photo"`
	PhotoAttribution string `json:"photo_attribution"`
}

// Dictionary represents the root structure of dictionary.json
type Dictionary struct {
	Words []DictionaryWord `json:"dictionary"`
}

// GenerateJSON generates dictionary.json content from database
// Only includes words with day_index assigned
func (g *Generator) GenerateJSON() ([]byte, error) {
	return g.generateJSON(false)
}

// GenerateAllJSON generates dictionary.json content from database
// Includes ALL words, using 0 for words without day_index
func (g *Generator) GenerateAllJSON() ([]byte, error) {
	return g.generateJSON(true)
}

// generateJSON is the internal implementation that handles both modes
func (g *Generator) generateJSON(includeAll bool) ([]byte, error) {
	g.logger.Info("Starting JSON generation", logger.Bool("include_all", includeAll))

	var words []DictionaryWord
	var wordCount int

	if includeAll {
		// Get ALL words from database
		allWords, err := g.repo.GetAllWords()
		if err != nil {
			appErr := entities.NewAppErrorWithContexts(err, 500, "Failed to get all words for generation", map[string]interface{}{
				"operation":   "generate_get_all_words",
				"include_all": true,
			})
			g.logger.ErrorWithStack(err, "Generation query failed", logger.String("operation", "generate_get_all_words"))
			return nil, appErr
		}

		wordCount = len(allWords)
		g.logger.Info("Retrieved all words for generation", logger.Int("word_count", wordCount))

		// Convert to dictionary words
		words = make([]DictionaryWord, 0, len(allWords))
		for _, word := range allWords {
			index := 0
			if word.DayIndex != nil {
				index = *word.DayIndex
			}
			dictWord := DictionaryWord{
				Index:            index,
				Word:             word.Word,
				Meaning:          word.Meaning,
				Link:             word.Link,
				Photo:            word.Photo,
				PhotoAttribution: word.PhotoAttribution,
			}
			words = append(words, dictWord)
		}

		// Sort by index (0s will be first), then by ID order
		sort.Slice(words, func(i, j int) bool {
			if words[i].Index == words[j].Index {
				return words[i].Word < words[j].Word // Secondary sort by word name
			}
			return words[i].Index < words[j].Index
		})
	} else {
		// Get only words with day_index (original behavior)
		wordsByDay, err := g.repo.GetWordsByDayIndex()
		if err != nil {
			appErr := entities.NewAppErrorWithContexts(err, 500, "Failed to get words for generation", map[string]interface{}{
				"operation":   "generate_get_words",
				"include_all": false,
			})
			g.logger.ErrorWithStack(err, "Generation query failed", logger.String("operation", "generate_get_words"))
			return nil, appErr
		}

		wordCount = len(wordsByDay)
		g.logger.Info("Retrieved words for generation", logger.Int("word_count", wordCount))

		// Convert to slice and sort by day_index
		words = make([]DictionaryWord, 0, len(wordsByDay))
		for dayIndex, word := range wordsByDay {
			dictWord := DictionaryWord{
				Index:            dayIndex,
				Word:             word.Word,
				Meaning:          word.Meaning,
				Link:             word.Link,
				Photo:            word.Photo,
				PhotoAttribution: word.PhotoAttribution,
			}
			words = append(words, dictWord)
		}

		// Sort by index
		sort.Slice(words, func(i, j int) bool {
			return words[i].Index < words[j].Index
		})
	}

	// Create dictionary structure
	dict := Dictionary{
		Words: words,
	}

	// Marshal to JSON
	var jsonBytes []byte
	var err error
	if g.prettyPrint {
		jsonBytes, err = json.MarshalIndent(dict, "", "    ")
	} else {
		jsonBytes, err = json.Marshal(dict)
	}

	if err != nil {
		appErr := entities.NewAppErrorWithContexts(err, 500, "Failed to marshal JSON", map[string]interface{}{
			"operation":  "generate_marshal",
			"word_count": wordCount,
		})
		g.logger.ErrorWithStack(err, "JSON marshal failed", logger.String("operation", "generate_marshal"))
		return nil, appErr
	}

	g.logger.Info("JSON generation completed", logger.Int("byte_count", len(jsonBytes)))
	return jsonBytes, nil
}

// GenerateToFile generates dictionary.json and writes it to a file
// Only includes words with day_index assigned
func (g *Generator) GenerateToFile(filePath string) error {
	return g.generateToFile(filePath, false)
}

// GenerateAllToFile generates dictionary.json with ALL words and writes it to a file
func (g *Generator) GenerateAllToFile(filePath string) error {
	return g.generateToFile(filePath, true)
}

// generateToFile is the internal implementation
func (g *Generator) generateToFile(filePath string, includeAll bool) error {
	g.logger.Info("Generating dictionary file",
		logger.String("file_path", filePath),
		logger.Bool("include_all", includeAll))

	var jsonBytes []byte
	var err error

	if includeAll {
		jsonBytes, err = g.GenerateAllJSON()
	} else {
		jsonBytes, err = g.GenerateJSON()
	}

	if err != nil {
		return err
	}

	// Write to temporary file first (atomic write)
	tmpFile := filePath + ".tmp"
	g.logger.Debug("Writing temporary file", logger.String("tmp_file", tmpFile))

	err = os.WriteFile(tmpFile, jsonBytes, 0644)
	if err != nil {
		appErr := entities.NewAppErrorWithContexts(err, 500, "Failed to write temporary file", map[string]interface{}{
			"operation":  "generate_write_tmp",
			"file_path":  tmpFile,
		})
		g.logger.ErrorWithStack(err, "Temporary file write failed", logger.String("operation", "generate_write_tmp"), logger.String("file_path", tmpFile))
		return appErr
	}

	// Rename temporary file to final name (atomic operation)
	g.logger.Debug("Renaming temporary file to final destination", logger.String("dest", filePath))
	err = os.Rename(tmpFile, filePath)
	if err != nil {
		// Clean up temporary file on error
		os.Remove(tmpFile)
		appErr := entities.NewAppErrorWithContexts(err, 500, "Failed to rename temporary file", map[string]interface{}{
			"operation":  "generate_rename",
			"tmp_file":   tmpFile,
			"dest_file":  filePath,
		})
		g.logger.ErrorWithStack(err, "File rename failed", logger.String("operation", "generate_rename"), logger.String("dest", filePath))
		return appErr
	}

	g.logger.Info("Dictionary file generated successfully", logger.String("file_path", filePath))
	return nil
}
