package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wizact/te-reo-bot/pkg/logger"
	"github.com/wizact/te-reo-bot/version"

	"github.com/wizact/yacli"
)

func main() {
	fmt.Printf("Te Reo Bot, Version: %s, Hash: %s", version.GetVersion(), version.GetGitCommit())
	fmt.Println()

	// Initialize global logger early in application startup
	err := logger.InitializeGlobalLogger(nil) // nil means load from environment
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Get the global logger for startup logging
	log := logger.GetGlobalLogger()
	log.Info("Te Reo Bot starting up",
		logger.String("version", version.GetVersion()),
		logger.String("git_commit", version.GetGitCommit()),
	)

	app := yacli.NewApplication()

	app.Name = "te reo bot"
	app.Description = "Te Reo Twitter & Mastodon bot"
	app.Version = version.VERSION

	app.AddCommand(&StartServerCommand{})

	ctx := context.Background()

	log.Info("Starting CLI application")
	app.Run(ctx)
}
