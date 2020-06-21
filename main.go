package main

import (
	"context"
 	"github.com/wizact/yacli"
)

func main() {
	app := yacli.NewApplication()

	app.Name = "te reo bot"
	app.Description = "Te Reo Twitter bot"
	app.Version = "0.0.0"

	app.AddCommand(&StartServerCommand{})

	ctx := context.Background()

	app.Run(ctx)
}
