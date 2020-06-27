package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadFile(t *testing.T) {
	assert := assert.New(t)

	ws := WordSelector{}

	a, e := ws.ParseFile()

	assert.Nil(e, "Failed parsing dictionary file")
	assert.NotNil(a)
	assert.True(a != nil && a.Words != nil && len(a.Words) > 0)

}
