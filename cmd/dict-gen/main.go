package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wizact/te-reo-bot/pkg/backup"
	"github.com/wizact/te-reo-bot/pkg/generator"
	"github.com/wizact/te-reo-bot/pkg/migration"
	"github.com/wizact/te-reo-bot/pkg/repository"
	"github.com/wizact/te-reo-bot/pkg/validator"
)

const (
	defaultDBPath     = "./data/words.db"
	defaultOutputPath = "./cmd/server/dictionary.json"
	maxFileSize       = 100 * 1024 * 1024 // 100MB
)

var (
	verboseMode bool
	quietMode   bool
)

// logInfo prints message if not in quiet mode
func logInfo(format string, args ...interface{}) {
	if !quietMode {
		fmt.Printf(format, args...)
	}
}

// logVerbose prints message only in verbose mode
func logVerbose(format string, args ...interface{}) {
	if verboseMode {
		fmt.Printf("[VERBOSE] "+format, args...)
	}
}

// logError always prints errors to stderr
func logError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func main() {
	// Parse global flags and filter them out
	filteredArgs := []string{os.Args[0]}
	for _, arg := range os.Args[1:] {
		if arg == "--verbose" || arg == "-v" {
			verboseMode = true
		} else if arg == "--quiet" || arg == "-q" {
			quietMode = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	os.Args = filteredArgs

	// Validate flag combination
	if verboseMode && quietMode {
		logError("Error: cannot use --verbose and --quiet together\n")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "migrate":
		runMigrate()
	case "generate":
		runGenerate()
	case "validate":
		runValidate()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// validatePath checks if the given file path exists as is accessible
func validatePath(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	// Checks if path is accessible
	if os.IsPermission(err) {
		return fmt.Errorf("permission denied for file: %s", path)
	}

	return err
}

// validateFileSize checks if the file size is within acceptable limits
func validateFileSize(path string, maxSize int64) error {
	s, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stat input file: %v\n", err)
		os.Exit(1)
	}
	if s.Size() > maxSize {
		fmt.Fprintf(os.Stderr, "Input file is too large (over 100MB): %s\n", path)
		os.Exit(1)
	}
	return nil
}

// validatePathTraversal ensures the path does not contain traversal sequences
func validatePathTraversal(path string) error {
	cleanedPath := filepath.Clean(path)

	// Check if path contains .. after cleaning
	if strings.Contains(cleanedPath, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}

	// Reject absolute paths
	if filepath.IsAbs(cleanedPath) {
		return fmt.Errorf("absolute paths not allowed: %s", path)
	}

	return nil
}

func runMigrate() {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	inputFile := fs.String("input", defaultOutputPath, "Path to input dictionary.json file")
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database file")
	dryRun := fs.Bool("dry-run", false, "Preview migration without modifying database")
	fs.Parse(os.Args[2:])

	if *dryRun {
		logInfo("DRY RUN MODE - No changes will be made\n")
	}
	logInfo("Starting migration...\n")
	logInfo("   Input: %s\n", *inputFile)
	logInfo("   Database: %s\n", *dbPath)

	// Validate input file path
	logVerbose("Validating input file path: %s\n", *inputFile)
	if err := validatePath(*inputFile); err != nil {
		logError("Invalid input file: %v\n", err)
		os.Exit(1)
	}

	// Validate path traversal
	logVerbose("Checking for path traversal...\n")
	if err := validatePathTraversal(*inputFile); err != nil {
		logError("Invalid input file path: %v\n", err)
		os.Exit(1)
	}

	// Validate file size (max 100MB)
	logVerbose("Validating file size...\n")
	if err := validateFileSize(*inputFile, maxFileSize); err != nil {
		logError("Invalid input file: %v\n", err)
		os.Exit(1)
	}

	// Parse dictionary file for dry-run or migration
	logVerbose("Reading input file...\n")
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		logError("Failed to read input file: %v\n", err)
		os.Exit(1)
	}

	logVerbose("Parsing JSON...\n")
	dict, err := migration.ParseDictionaryJSON(data)
	if err != nil {
		logError("Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	// Dry-run preview
	if *dryRun {
		logInfo("\nMigration Preview:\n")
		logInfo("   Words to import: %d\n", len(dict.Words))

		// Show sample words
		sampleCount := 5
		if len(dict.Words) < sampleCount {
			sampleCount = len(dict.Words)
		}
		logInfo("\n   Sample words:\n")
		for i := 0; i < sampleCount; i++ {
			w := dict.Words[i]
			logInfo("      [%d] %s - %s\n", w.Index, w.Word, w.Meaning)
		}

		// Check for potential issues
		dayIndexes := make(map[int]bool)
		duplicates := []int{}
		missing := []int{}

		for _, w := range dict.Words {
			if dayIndexes[w.Index] {
				duplicates = append(duplicates, w.Index)
			}
			dayIndexes[w.Index] = true
		}

		for i := 1; i <= 366; i++ {
			if !dayIndexes[i] {
				missing = append(missing, i)
			}
		}

		if len(duplicates) > 0 {
			logInfo("\n   WARNING: Duplicate day indexes: %v\n", duplicates)
		}

		if len(missing) > 0 {
			logInfo("\n   WARNING: Missing day indexes: %d\n", len(missing))
			if len(missing) <= 20 {
				logInfo("      Missing: %v\n", missing)
			} else {
				logInfo("      First 20 missing: %v\n", missing[:20])
			}
		}

		if len(duplicates) == 0 && len(missing) == 0 {
			logInfo("\n   Validation: All 366 day indexes present and unique\n")
		}

		logInfo("\n Dry-run complete. Run without --dry-run to apply changes.\n")
		return
	}

	// Backup existing database if it exists
	if _, err := os.Stat(*dbPath); err == nil {
		logInfo("\nBacking up existing database...\n")
		logVerbose("Creating backup of: %s\n", *dbPath)
		backupPath, err := backup.BackupFile(*dbPath)
		if err != nil {
			logError("Failed to create backup: %v\n", err)
			os.Exit(1)
		}
		if backupPath != "" {
			logInfo("   Backup created: %s\n", backupPath)

			// Cleanup old backups (keep last 7 days)
			logVerbose("Cleaning up old backups...\n")
			if err := backup.CleanupOldBackups(*dbPath, 7); err != nil {
				logError("Warning: failed to cleanup old backups: %v\n", err)
				// Don't exit - this is non-critical
			}
		}
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(*dbPath)
	logVerbose("Ensuring database directory exists: %s\n", dbDir)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		logError("Failed to create database directory: %v\n", err)
		os.Exit(1)
	}

	// Open database
	logVerbose("Opening database: %s\n", *dbPath)
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		logError("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize database schema
	logVerbose("Initializing database schema...\n")
	if err := repository.InitializeDatabase(db); err != nil {
		logError("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	// Run migration
	repo := repository.NewSQLiteRepository(db)
	migrator := migration.NewMigrator(repo)

	logVerbose("Starting migration transaction...\n")
	if err := migrator.MigrateFromFile(*inputFile); err != nil {
		logError("Migration failed: %v\n", err)
		os.Exit(1)
	}
	logVerbose("Migration transaction committed\n")

	// Count imported words
	count, err := repo.GetWordCountByDayIndex()
	if err != nil {
		logError("Failed to count words: %v\n", err)
		os.Exit(1)
	}

	logInfo("Migration complete!\n")
	logInfo("   - %d words migrated\n", count)
	logInfo("   - Database: %s\n", *dbPath)
	logInfo("\n Next steps:\n")
	logInfo("   1. Run: dict-gen validate\n")
	logInfo("   2. Run: dict-gen generate\n")
}

func runGenerate() {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database file")
	outputFile := fs.String("output", defaultOutputPath, "Path to output dictionary.json file")
	compact := fs.Bool("compact", false, "Generate compact JSON (no indentation)")
	all := fs.Bool("all", false, "Export ALL words (including those without day_index, using 0 for index)")
	fs.Parse(os.Args[2:])

	if err := validatePath(*dbPath); err != nil {
		logError("Invalid database file: %v\n", err)
		os.Exit(1)
	}

	logInfo("Generating dictionary.json...\n")
	logInfo("   Database: %s\n", *dbPath)
	logInfo("   Output: %s\n", *outputFile)
	if *all {
		logInfo("   Mode: Export ALL words (including those without day_index)\n")
	} else {
		logInfo("   Mode: Export only words with day_index (1-366)\n")
	}

	// Open database
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		logError("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewSQLiteRepository(db)

	// Validate before generating (only for day_index mode)
	var wordCount int
	if !*all {
		logInfo("\n Validating data...\n")
		v := validator.NewValidator(repo)
		report, err := v.Validate()
		if err != nil {
			logError("Validation error: %v\n", err)
			os.Exit(1)
		}

		if !report.IsValid {
			logInfo("Validation failed!\n")
			logInfo("   - Total words: %d (expected 366)\n", report.TotalWords)
			if len(report.MissingIndexes) > 0 {
				logInfo("   - Missing indexes: %d\n", len(report.MissingIndexes))
				if len(report.MissingIndexes) <= 20 {
					logInfo("     %v\n", report.MissingIndexes)
				} else {
					logInfo("     First 20: %v\n", report.MissingIndexes[:20])
				}
			}
			logInfo("\n Fix: Ensure all days 1-366 have exactly one word assigned\n")
			os.Exit(1)
		}

		logInfo("Validation passed\n")
		wordCount = report.TotalWords
	} else {
		// Get total word count for --all mode
		count, err := repo.GetWordCount()
		if err != nil {
			logError("Failed to count words: %v\n", err)
			os.Exit(1)
		}
		wordCount = count
	}

	// Backup existing dictionary.json if it exists
	if _, err := os.Stat(*outputFile); err == nil {
		logInfo("\nBacking up existing dictionary.json...\n")
		backupPath, err := backup.BackupFile(*outputFile)
		if err != nil {
			logError("Failed to create backup: %v\n", err)
			os.Exit(1)
		}
		if backupPath != "" {
			logInfo("   Backup created: %s\n", backupPath)
		}
	}

	// Generate JSON
	logInfo("\n Generating JSON...\n")
	gen := generator.NewGenerator(repo)
	gen.SetPrettyPrint(!*compact)

	// Ensure output directory exists
	outputDir := filepath.Dir(*outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logError("Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate based on mode
	if *all {
		if err := gen.GenerateAllToFile(*outputFile); err != nil {
			logError("Generation failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := gen.GenerateToFile(*outputFile); err != nil {
			logError("Generation failed: %v\n", err)
			os.Exit(1)
		}
	}

	// Get file size
	fileInfo, err := os.Stat(*outputFile)
	if err != nil {
		logError("Failed to stat output file: %v\n", err)
		os.Exit(1)
	}

	logInfo("Dictionary generated successfully!\n")
	logInfo("   - Output: %s\n", *outputFile)
	logInfo("   - Words: %d\n", wordCount)
	logInfo("   - Size: %.1f KB\n", float64(fileInfo.Size())/1024)

	format := "pretty (indented)"
	if *compact {
		format = "compact"
	}
	logInfo("   - Format: %s\n", format)

	logInfo("\n Next steps:\n")
	logInfo("   1. Review changes: git diff %s\n", *outputFile)
	logInfo("   2. Test server: go run cmd/server/main.go\n")
	logInfo("   3. Commit: git add %s\n", *outputFile)
}

func runValidate() {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database file")
	fs.Parse(os.Args[2:])

	logInfo(" Validating database...\n")
	logInfo("   Database: %s\n", *dbPath)

	// Open database
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		logError("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewSQLiteRepository(db)
	v := validator.NewValidator(repo)

	report, err := v.Validate()
	if err != nil {
		logError("Validation error: %v\n", err)
		os.Exit(1)
	}

	logInfo("\n")
	if report.IsValid {
		logInfo("Validation passed!\n")
		logInfo("   - Total words: %d\n", report.TotalWords)
		logInfo("   - Day index range: 1-366\n")
		logInfo("   - All indexes unique: Checked\n")
	} else {
		logInfo("Validation failed!\n")
		logInfo("   - Total words: %d (expected 366)\n", report.TotalWords)

		if len(report.MissingIndexes) > 0 {
			logInfo("   - Missing indexes: %d\n", len(report.MissingIndexes))
			if len(report.MissingIndexes) <= 20 {
				logInfo("     %v\n", report.MissingIndexes)
			} else {
				logInfo("     First 20: %v\n", report.MissingIndexes[:20])
			}
		}

		if len(report.DuplicateIndexes) > 0 {
			logInfo("   - Duplicate indexes: %v\n", report.DuplicateIndexes)
		}

		logInfo("\n Fix: Ensure all days 1-366 have exactly one word assigned\n")
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("dict-gen - Te Reo Bot Dictionary Generator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dict-gen [--verbose | --quiet] <command> [flags]")
	fmt.Println()
	fmt.Println("Global Flags:")
	fmt.Println("  --verbose, -v  Enable verbose output")
	fmt.Println("  --quiet, -q    Suppress all output except errors")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  migrate   Import dictionary.json into SQLite database")
	fmt.Println("  generate  Generate dictionary.json from SQLite database")
	fmt.Println("  validate  Validate database integrity (366 unique indexes)")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Preview migration without making changes")
	fmt.Println("  dict-gen migrate --input=./dictionary.json --dry-run")
	fmt.Println()
	fmt.Println("  # Migrate existing dictionary.json")
	fmt.Println("  dict-gen migrate --input=./cmd/server/dictionary.json")
	fmt.Println()
	fmt.Println("  # Validate database")
	fmt.Println("  dict-gen validate")
	fmt.Println()
	fmt.Println("  # Generate dictionary.json (only 366 words with day_index)")
	fmt.Println("  dict-gen generate --output=./cmd/server/dictionary.json")
	fmt.Println()
	fmt.Println("  # Generate dictionary.json with ALL words (including extras)")
	fmt.Println("  dict-gen generate --all --output=./backup-all-words.json")
	fmt.Println()
	fmt.Println("For more information, visit:")
	fmt.Println("  https://github.com/wizact/te-reo-bot")
}
