package main

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

//WordSelector reads, parses, and selects the word-of-the-day
type WordSelector struct {
}

//SelectWordByDay selects a word from the provided array based on the day of the year
func (ws *WordSelector) SelectWordByDay(words []Word) *Word {
	doy := time.Now().YearDay()
	low := len(words)
	if doy <= low {
		return &words[doy-1]
	} else {
		return &words[(doy-((doy/low)*low))-1]
	}
}

//SelectWordByIndex selects a word from the provided array based on the day of the year
func (ws *WordSelector) SelectWordByIndex(words []Word, index int) *Word {
	low := len(words)
	if index <= low {
		return &words[index-1]
	} else {
		return &words[(index-((index/low)*low))-1]
	}
}

//ParseFile unmarshal a json string to the struct type
func (ws *WordSelector) ParseFile(f []byte) (*Dictionary, error) {
	wd := Dictionary{}

	err := json.Unmarshal(f, &wd)

	if err != nil {
		return nil, err
	}

	return &wd, nil

}

//ReadFile reads dictionary json file
func (ws *WordSelector) ReadFile() ([]byte, error) {
	f, err := ioutil.ReadFile("dictionary.json")

	if err != nil {
		return nil, err
	}

	return f, nil
}

//Dictionary is the parent element of json file
type Dictionary struct {
	Words []Word `json:"dictionary"`
}

//Word is the wrapper around each word and it's meaning
type Word struct {
	Index       int    `json:"index"`
	Word        string `json:"word"`
	Meaning     string `json:"meaning"`
	Link        string `json:"link"`
	Photo       string `json:"photo"`
	Attribution string `json:"photo_attribution"`
}
