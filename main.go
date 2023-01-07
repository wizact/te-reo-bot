package main

import (
	"context"
	"fmt"
	"te-reo-bot/version"

	"github.com/wizact/yacli"
)

func main() {
	fmt.Printf("Te Reo Bot, Version: %s, Hash: %s", version.GetVersion(), version.GetGitCommit())
	fmt.Println()

	app := yacli.NewApplication()

	app.Name = "te reo bot"
	app.Description = "Te Reo Twitter & Mastodon bot"
	app.Version = version.VERSION

	app.AddCommand(&StartServerCommand{})

	ctx := context.Background()

	app.Run(ctx)
}
