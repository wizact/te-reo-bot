package repository_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wotd "github.com/wizact/te-reo-bot/pkg/wotd"
)

func TestUpdateWordTx(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	dayIndex := 1
	word := &wotd.Word{
		DayIndex: &dayIndex,
		Word:     "taketake",
		Meaning:  "original",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	word.Meaning = "source"
	word.Photo = "image.png"
	tx, err = repo.BeginTx()
	require.NoError(t, err)
	err = repo.UpdateWordTx(tx, word)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	updated, err := repo.GetWordByID(word.ID)
	require.NoError(t, err)
	assert.Equal(t, "source", updated.Meaning)
	assert.Equal(t, "image.png", updated.Photo)
}

func TestUpdateWordDayIndexByID(t *testing.T) {
	db, repo := setupTestRepository(t)
	defer db.Close()

	dayIndex := 1
	word := &wotd.Word{
		DayIndex: &dayIndex,
		Word:     "hurihuri",
		Meaning:  "rotate",
	}
	tx, err := repo.BeginTx()
	require.NoError(t, err)
	err = repo.AddWord(tx, word)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	tx, err = repo.BeginTx()
	require.NoError(t, err)
	err = repo.UpdateWordDayIndexByID(tx, word.ID, nil)
	require.NoError(t, err)
	err = repo.CommitTx(tx)
	require.NoError(t, err)

	updated, err := repo.GetWordByID(word.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.DayIndex)
}
