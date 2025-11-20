package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wizact/te-reo-bot/pkg/generator"
	"github.com/wizact/te-reo-bot/pkg/migration"
	"github.com/wizact/te-reo-bot/pkg/repository"
	"github.com/wizact/te-reo-bot/pkg/validator"
)

const (
	defaultDBPath     = "./data/words.db"
	defaultOutputPath = "./cmd/server/dictionary.json"
)

func main() {
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

func runMigrate() {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	inputFile := fs.String("input", defaultOutputPath, "Path to input dictionary.json file")
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database file")
	fs.Parse(os.Args[2:])

	fmt.Println("🔄 Starting migration...")
	fmt.Printf("   Input: %s\n", *inputFile)
	fmt.Printf("   Database: %s\n", *dbPath)

	// Ensure database directory exists
	dbDir := filepath.Dir(*dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create database directory: %v\n", err)
		os.Exit(1)
	}

	// Open database
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize database schema
	if err := repository.InitializeDatabase(db); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	// Run migration
	repo := repository.NewSQLiteRepository(db)
	migrator := migration.NewMigrator(repo)

	if err := migrator.MigrateFromFile(*inputFile); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Migration failed: %v\n", err)
		os.Exit(1)
	}

	// Count imported words
	count, err := repo.GetWordCountByDayIndex()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to count words: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Migration complete!\n")
	fmt.Printf("   - %d words migrated\n", count)
	fmt.Printf("   - Database: %s\n", *dbPath)
	fmt.Println("\n💡 Next steps:")
	fmt.Println("   1. Run: dict-gen validate")
	fmt.Println("   2. Run: dict-gen generate")
}

func runGenerate() {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database file")
	outputFile := fs.String("output", defaultOutputPath, "Path to output dictionary.json file")
	compact := fs.Bool("compact", false, "Generate compact JSON (no indentation)")
	fs.Parse(os.Args[2:])

	fmt.Println("🔄 Generating dictionary.json...")
	fmt.Printf("   Database: %s\n", *dbPath)
	fmt.Printf("   Output: %s\n", *outputFile)

	// Open database
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewSQLiteRepository(db)

	// Validate before generating
	fmt.Println("\n🔍 Validating data...")
	v := validator.NewValidator(repo)
	report, err := v.Validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Validation error: %v\n", err)
		os.Exit(1)
	}

	if !report.IsValid {
		fmt.Println("❌ Validation failed!")
		fmt.Printf("   - Total words: %d (expected 366)\n", report.TotalWords)
		if len(report.MissingIndexes) > 0 {
			fmt.Printf("   - Missing indexes: %d\n", len(report.MissingIndexes))
			if len(report.MissingIndexes) <= 20 {
				fmt.Printf("     %v\n", report.MissingIndexes)
			} else {
				fmt.Printf("     First 20: %v\n", report.MissingIndexes[:20])
			}
		}
		fmt.Println("\n💡 Fix: Ensure all days 1-366 have exactly one word assigned")
		os.Exit(1)
	}

	fmt.Println("✓ Validation passed")

	// Generate JSON
	fmt.Println("\n📝 Generating JSON...")
	gen := generator.NewGenerator(repo)
	gen.SetPrettyPrint(!*compact)

	// Ensure output directory exists
	outputDir := filepath.Dir(*outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	if err := gen.GenerateToFile(*outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Generation failed: %v\n", err)
		os.Exit(1)
	}

	// Get file size
	fileInfo, err := os.Stat(*outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to stat output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Dictionary generated successfully!\n")
	fmt.Printf("   - Output: %s\n", *outputFile)
	fmt.Printf("   - Words: %d\n", report.TotalWords)
	fmt.Printf("   - Size: %.1f KB\n", float64(fileInfo.Size())/1024)
	
	format := "pretty (indented)"
	if *compact {
		format = "compact"
	}
	fmt.Printf("   - Format: %s\n", format)

	fmt.Println("\n💡 Next steps:")
	fmt.Println("   1. Review changes: git diff", *outputFile)
	fmt.Println("   2. Test server: go run cmd/server/main.go")
	fmt.Println("   3. Commit: git add", *outputFile)
}

func runValidate() {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database file")
	fs.Parse(os.Args[2:])

	fmt.Println("🔍 Validating database...")
	fmt.Printf("   Database: %s\n", *dbPath)

	// Open database
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewSQLiteRepository(db)
	v := validator.NewValidator(repo)

	report, err := v.Validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	if report.IsValid {
		fmt.Println("✅ Validation passed!")
		fmt.Printf("   - Total words: %d\n", report.TotalWords)
		fmt.Printf("   - Day index range: 1-366\n")
		fmt.Printf("   - All indexes unique: ✓\n")
	} else {
		fmt.Println("❌ Validation failed!")
		fmt.Printf("   - Total words: %d (expected 366)\n", report.TotalWords)
		
		if len(report.MissingIndexes) > 0 {
			fmt.Printf("   - Missing indexes: %d\n", len(report.MissingIndexes))
			if len(report.MissingIndexes) <= 20 {
				fmt.Printf("     %v\n", report.MissingIndexes)
			} else {
				fmt.Printf("     First 20: %v\n", report.MissingIndexes[:20])
			}
		}
		
		if len(report.DuplicateIndexes) > 0 {
			fmt.Printf("   - Duplicate indexes: %v\n", report.DuplicateIndexes)
		}
		
		fmt.Println("\n💡 Fix: Ensure all days 1-366 have exactly one word assigned")
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("dict-gen - Te Reo Bot Dictionary Generator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dict-gen <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  migrate   Import dictionary.json into SQLite database")
	fmt.Println("  generate  Generate dictionary.json from SQLite database")
	fmt.Println("  validate  Validate database integrity (366 unique indexes)")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Migrate existing dictionary.json")
	fmt.Println("  dict-gen migrate --input=./cmd/server/dictionary.json")
	fmt.Println()
	fmt.Println("  # Validate database")
	fmt.Println("  dict-gen validate")
	fmt.Println()
	fmt.Println("  # Generate dictionary.json")
	fmt.Println("  dict-gen generate --output=./cmd/server/dictionary.json")
	fmt.Println()
	fmt.Println("For more information, visit:")
	fmt.Println("  https://github.com/wizact/te-reo-bot")
}
