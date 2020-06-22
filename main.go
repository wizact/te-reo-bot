package main

import (
	"context"
	"fmt"
	"github.com/wizact/yacli"
	"te-reo-bot/version"
)

func main() {
	fmt.Printf("Te Reo Bot, Version: %s, Hash: %s", version.VERSION, version.GITCOMMIT)
	fmt.Println()

	app := yacli.NewApplication()

	app.Name = "te reo bot"
	app.Description = "Te Reo Twitter bot"
	app.Version = version.VERSION

	app.AddCommand(&StartServerCommand{})

	ctx := context.Background()

	app.Run(ctx)
}
