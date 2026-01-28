package backup_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wizact/te-reo-bot/pkg/backup"
	"github.com/wizact/te-reo-bot/pkg/entities"
)

func TestBackupFile(t *testing.T) {
	t.Run("creates backup successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		sourceFile := filepath.Join(tmpDir, "test.db")

		// Create source file
		testData := []byte("test data")
		err := os.WriteFile(sourceFile, testData, 0644)
		require.NoError(t, err)

		// Create backup
		backupPath, err := backup.BackupFile(sourceFile)
		require.NoError(t, err)
		assert.NotEmpty(t, backupPath)

		// Verify backup exists
		assert.FileExists(t, backupPath)

		// Verify backup content matches source
		backupData, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, testData, backupData)

		// Verify backup filename format
		assert.True(t, strings.HasPrefix(filepath.Base(backupPath), "test.db.backup."))
	})

	t.Run("returns empty string when source doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentFile := filepath.Join(tmpDir, "nonexistent.db")

		backupPath, err := backup.BackupFile(nonExistentFile)
		require.NoError(t, err)
		assert.Empty(t, backupPath)
	})

	t.Run("handles permission errors", func(t *testing.T) {
		tmpDir := t.TempDir()
		sourceFile := filepath.Join(tmpDir, "test.db")

		err := os.WriteFile(sourceFile, []byte("data"), 0644)
		require.NoError(t, err)

		// Make source unreadable
		err = os.Chmod(sourceFile, 0000)
		require.NoError(t, err)
		defer os.Chmod(sourceFile, 0644)

		_, err = backup.BackupFile(sourceFile)
		assert.Error(t, err)
		// Check that it's an AppError with the expected message
		appErr, ok := err.(*entities.AppError)
		require.True(t, ok, "Expected AppError")
		assert.Equal(t, "failed to open source", appErr.Message)
	})
}

func TestCleanupOldBackups(t *testing.T) {
	t.Run("removes old backups", func(t *testing.T) {
		tmpDir := t.TempDir()
		basePath := filepath.Join(tmpDir, "test.db")

		// Create old backups (8 days old)
		oldBackup1 := basePath + ".backup.20260119-100000"
		oldBackup2 := basePath + ".backup.20260118-100000"

		err := os.WriteFile(oldBackup1, []byte("old1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(oldBackup2, []byte("old2"), 0644)
		require.NoError(t, err)

		// Set modification time to 8 days ago
		oldTime := time.Now().AddDate(0, 0, -8)
		os.Chtimes(oldBackup1, oldTime, oldTime)
		os.Chtimes(oldBackup2, oldTime, oldTime)

		// Create recent backup (1 day old)
		recentBackup := basePath + ".backup.20260126-100000"
		err = os.WriteFile(recentBackup, []byte("recent"), 0644)
		require.NoError(t, err)

		// Cleanup backups older than 7 days
		err = backup.CleanupOldBackups(basePath, 7)
		require.NoError(t, err)

		// Verify old backups removed
		assert.NoFileExists(t, oldBackup1)
		assert.NoFileExists(t, oldBackup2)

		// Verify recent backup remains
		assert.FileExists(t, recentBackup)
	})

	t.Run("handles non-existent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		basePath := filepath.Join(tmpDir, "nonexistent", "test.db")

		// Should not error on non-existent directory
		err := backup.CleanupOldBackups(basePath, 7)
		assert.NoError(t, err)
	})

	t.Run("handles no backups to clean", func(t *testing.T) {
		tmpDir := t.TempDir()
		basePath := filepath.Join(tmpDir, "test.db")

		err := backup.CleanupOldBackups(basePath, 7)
		assert.NoError(t, err)
	})
}

func TestBackupFileNameFormat(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "words.db")

	err := os.WriteFile(sourceFile, []byte("data"), 0644)
	require.NoError(t, err)

	backupPath, err := backup.BackupFile(sourceFile)
	require.NoError(t, err)

	// Verify format: words.db.backup.YYYYMMDD-HHMMSS
	baseName := filepath.Base(backupPath)
	assert.True(t, strings.HasPrefix(baseName, "words.db.backup."))

	// Extract timestamp part
	parts := strings.Split(baseName, ".backup.")
	require.Len(t, parts, 2)
	timestamp := parts[1]

	// Verify timestamp format (length and basic structure)
	assert.Len(t, timestamp, 15) // YYYYMMDD-HHMMSS = 15 chars
	assert.Contains(t, timestamp, "-")
}
