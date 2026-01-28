package repository

import (
	"database/sql"

	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
)

// InitializeDatabase creates the necessary tables and indexes in the database
func InitializeDatabase(db *sql.DB) error {
	log := logger.GetGlobalLogger()
	log.Info("Initializing database schema")

	_, err := db.Exec(schema)
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to execute database schema")
		appErr.WithContext("operation", "init_db_schema")
		log.ErrorWithStack(err, "Schema initialization failed", logger.String("operation", "init_db_schema"))
		return appErr
	}

	// Enable WAL mode for better concurrency and resilience
	log.Debug("Enabling WAL mode")
	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to enable WAL mode")
		appErr.WithContext("operation", "init_db_wal")
		log.ErrorWithStack(err, "WAL mode initialization failed", logger.String("operation", "init_db_wal"))
		return appErr
	}

	// Enable foreign key constraints
	log.Debug("Enabling foreign key constraints")
	_, err = db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to enable foreign keys")
		appErr.WithContext("operation", "init_db_fk")
		log.ErrorWithStack(err, "Foreign key initialization failed", logger.String("operation", "init_db_fk"))
		return appErr
	}

	log.Info("Database initialized successfully")
	return nil
}
