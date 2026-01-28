package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
)

// BackupFile creates a timestamped backup of the given file
// Returns the backup path if created, or empty string if source doesn't exist
func BackupFile(filePath string) (string, error) {
	log := logger.GetGlobalLogger()
	log.Info("Starting backup", logger.String("file_path", filePath))

	// Check if source exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Debug("Source file does not exist, skipping backup", logger.String("file_path", filePath))
		return "", nil // No backup needed if file doesn't exist
	}

	// Create backup filename: file.db -> file.db.backup.20260127-150405
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup.%s", filePath, timestamp)

	log.Debug("Creating backup", logger.String("backup_path", backupPath))

	// Copy file
	source, err := os.Open(filePath)
	if err != nil {
		appErr := entities.NewAppError(err, 500, "failed to open source")
		appErr.WithContext("operation", "backup_open_source")
		appErr.WithContext("file_path", filePath)
		log.ErrorWithStack(err, "Backup source open failed", logger.String("operation", "backup_open_source"), logger.String("file_path", filePath))
		return "", appErr
	}
	defer source.Close()

	dest, err := os.Create(backupPath)
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to create backup file")
		appErr.WithContext("operation", "backup_create_dest")
		appErr.WithContext("backup_path", backupPath)
		log.ErrorWithStack(err, "Backup dest create failed", logger.String("operation", "backup_create_dest"), logger.String("backup_path", backupPath))
		return "", appErr
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		os.Remove(backupPath)
		appErr := entities.NewAppError(err, 500, "Failed to copy file for backup")
		appErr.WithContext("operation", "backup_copy")
		appErr.WithContext("file_path", filePath)
		appErr.WithContext("backup_path", backupPath)
		log.ErrorWithStack(err, "Backup copy failed", logger.String("operation", "backup_copy"), logger.String("backup_path", backupPath))
		return "", appErr
	}

	log.Info("Backup created successfully", logger.String("backup_path", backupPath))
	return backupPath, nil
}

// CleanupOldBackups removes backups older than specified days
func CleanupOldBackups(basePath string, keepDays int) error {
	log := logger.GetGlobalLogger()
	log.Info("Starting backup cleanup", logger.String("base_path", basePath), logger.Int("keep_days", keepDays))

	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)
	pattern := base + ".backup.*"

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to glob backup files")
		appErr.WithContext("operation", "backup_cleanup_glob")
		appErr.WithContext("base_path", basePath)
		appErr.WithContext("pattern", pattern)
		log.ErrorWithStack(err, "Backup cleanup glob failed", logger.String("operation", "backup_cleanup_glob"), logger.String("pattern", pattern))
		return appErr
	}

	log.Debug("Found backup files", logger.Int("count", len(matches)))

	cutoff := time.Now().AddDate(0, 0, -keepDays)
	deletedCount := 0

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			log.Debug("Cannot stat backup file, skipping", logger.String("file", match))
			continue
		}
		if info.ModTime().Before(cutoff) {
			log.Debug("Removing old backup", logger.String("file", match))
			if err := os.Remove(match); err != nil {
				log.Error(err, "Failed to remove old backup", logger.String("file", match))
			} else {
				deletedCount++
			}
		}
	}

	log.Info("Backup cleanup completed", logger.Int("deleted_count", deletedCount))
	return nil
}
