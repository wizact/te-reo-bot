package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// BackupFile creates a timestamped backup of the given file
// Returns the backup path if created, or empty string if source doesn't exist
func BackupFile(filePath string) (string, error) {
	// Check if source exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", nil // No backup needed if file doesn't exist
	}

	// Create backup filename: file.db -> file.db.backup.20260127-150405
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup.%s", filePath, timestamp)

	// Copy file
	source, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source: %w", err)
	}
	defer source.Close()

	dest, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		os.Remove(backupPath)
		return "", fmt.Errorf("failed to copy: %w", err)
	}

	return backupPath, nil
}

// CleanupOldBackups removes backups older than specified days
func CleanupOldBackups(basePath string, keepDays int) error {
	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)
	pattern := base + ".backup.*"

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -keepDays)

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(match)
		}
	}

	return nil
}
