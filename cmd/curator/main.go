package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wizact/te-reo-bot/pkg/curator"
	"github.com/wizact/te-reo-bot/pkg/logger"
	"github.com/wizact/te-reo-bot/pkg/repository"
)

func main() {
	var dbPath string
	var validateOnly bool

	flag.StringVar(&dbPath, "db", "", "path to SQLite database (defaults to ./data/words.db)")
	flag.BoolVar(&validateOnly, "validate", false, "run curator validation and exit")
	flag.Parse()

	// In TUI mode, route all log output to /dev/null so that background Info
	// messages do not corrupt the terminal rendering managed by gocui.
	if !validateOnly {
		devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not redirect stdout to /dev/null: %v\n", err)
		} else {
			defer devNull.Close()
			os.Stdout = devNull
		}
	}

	if err := logger.InitializeGlobalLogger(&logger.LoggerConfig{
		EnableStackTraces: false,
		LogLevel:          "fatal",
		Environment:       "test",
		LogFormat:         "text",
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	db, err := repository.OpenSQLiteDB(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	service := curator.NewService(repository.NewSQLiteRepository(db))

	if validateOnly {
		report, err := service.Validate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(curator.FormatValidationReport(report))
		if report.HasIssues() {
			os.Exit(1)
		}
		return
	}

	app := curator.NewTUI(service)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "curator UI failed: %v\n", err)
		os.Exit(1)
	}
}
