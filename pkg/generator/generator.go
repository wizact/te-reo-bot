package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/wizact/te-reo-bot/pkg/repository"
)

// Generator handles generation of dictionary.json from database
type Generator struct {
	repo        repository.WordRepository
	prettyPrint bool
}

// NewGenerator creates a new Generator
func NewGenerator(repo repository.WordRepository) *Generator {
	return &Generator{
		repo:        repo,
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
func (g *Generator) GenerateJSON() ([]byte, error) {
	// Get all words with day_index
	wordsByDay, err := g.repo.GetWordsByDayIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get words by day index: %w", err)
	}

	// Convert to slice and sort by day_index
	words := make([]DictionaryWord, 0, len(wordsByDay))
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

	// Create dictionary structure
	dict := Dictionary{
		Words: words,
	}

	// Marshal to JSON
	var jsonBytes []byte
	if g.prettyPrint {
		jsonBytes, err = json.MarshalIndent(dict, "", "    ")
	} else {
		jsonBytes, err = json.Marshal(dict)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return jsonBytes, nil
}

// GenerateToFile generates dictionary.json and writes it to a file
func (g *Generator) GenerateToFile(filePath string) error {
	jsonBytes, err := g.GenerateJSON()
	if err != nil {
		return err
	}

	// Write to temporary file first (atomic write)
	tmpFile := filePath + ".tmp"
	err = os.WriteFile(tmpFile, jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Rename temporary file to final name (atomic operation)
	err = os.Rename(tmpFile, filePath)
	if err != nil {
		// Clean up temporary file on error
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}
