package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"encoding/csv"
	"encoding/json"

	"github.com/wizact/te-reo-bot/pkg/wotd"
)

func main() {
	file, err := os.Open("te reo words.csv")

	// Checks for the error
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	// Closes the file
	defer file.Close()

	reader := csv.NewReader(file)

	r, err := reader.ReadAll()

	// Checks for the error
	if err != nil {
		fmt.Println("Error reading records")
	}

	td := &wotd.Dictionary{}

	for _, v := range r {
		i, e := strconv.Atoi(v[1])
		if e != nil {
			panic(e)
		}

		w := wotd.Word{Index: i, Word: v[2], Meaning: v[3], Attribution: v[4], Photo: v[5], Link: ""}
		td.Words = append(td.Words, w)
	}

	j, e := json.Marshal(td)
	if e != nil {
		panic(e)
	}

	f, e := os.Create("temp.json")
	if e != nil {
		panic(e)
	}
	defer f.Close()

	f.Write(j)
}
