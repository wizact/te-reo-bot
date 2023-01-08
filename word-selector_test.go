package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseFile(t *testing.T) {
	assert := assert.New(t)

	ws := WordSelector{}

	jc := `{
			"dictionary": [
				{ "index":	1	, "word": "āe", "meaning": "yes", "link": "", "photo": ""},
				{ "index":	2	, "word": "aha", "meaning": "what?", "link": "", "photo": ""}
		]}`

	a, e := ws.ParseFile(bytes.NewBufferString(jc).Bytes())

	assert.Nil(e, "Failed parsing dictionary")
	assert.NotNil(a)
	assert.True(a != nil && a.Words != nil && len(a.Words) > 0)

}

func TestReadFile(t *testing.T) {
	assert := assert.New(t)

	ws := WordSelector{}

	f, e := ws.ReadFile()

	assert.Nil(e, "Failed reading dictionary file")
	assert.NotNil(f)
	assert.True(len(f) > 0)
}
