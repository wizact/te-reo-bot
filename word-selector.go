package main

import (
	"encoding/json"
	"io/ioutil"
)

//WordSelector reads, parses, and selects the word-of-the-day
type WordSelector struct {
}

// SelectWordByDay selects a word from the provided array based on the day of the year
func (ws *WordSelector) SelectWordByDay(words []Word) (*Word, error) {
	w := &Word{}
	return w, nil
}

func (ws *WordSelector) ParseFile() (*Dictionary, error) {
	f, err := ioutil.ReadFile("dictionary.json")

	if err != nil {
		return nil, err
	}

	wd := Dictionary{}

	err = json.Unmarshal(f, &wd)

	if err != nil {
		return nil, err
	}

	return &wd, nil

}

//Dictionary is the parent element of json file
type Dictionary struct {
	Words []Word `json:"dictionary"`
}

//Word is the wrapper around each word and it's meaning
type Word struct {
	Index   int    `json:"index"`
	Word    string `json:"word"`
	Meaning string `json:"meaning"`
	Link    string `json:"link"`
	Photo   string `json:"photo"`
}
