package repository

import (
	"database/sql"
)

// InitializeDatabase creates the necessary tables and indexes in the database
func InitializeDatabase(db *sql.DB) error {
	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	// Enable WAL mode for better concurrency and resilience
	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return err
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		return err
	}

	return nil
}
