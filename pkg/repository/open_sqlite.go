package repository

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const defaultSQLiteDBPath = "./data/words.db"

// OpenSQLiteDB opens the configured SQLite database, verifies connectivity, and initializes schema.
func OpenSQLiteDB(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = defaultSQLiteDBPath
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	if err := InitializeDatabase(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
