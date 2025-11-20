package repository

import (
	"time"
)

// Word represents a Māori word with its meaning and metadata
type Word struct {
	ID               int       `json:"id" db:"id"`
	DayIndex         *int      `json:"index,omitempty" db:"day_index"`
	Word             string    `json:"word" db:"word"`
	Meaning          string    `json:"meaning" db:"meaning"`
	Link             string    `json:"link" db:"link"`
	Photo            string    `json:"photo" db:"photo"`
	PhotoAttribution string    `json:"photo_attribution" db:"photo_attribution"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	IsActive         bool      `json:"is_active" db:"is_active"`
}
