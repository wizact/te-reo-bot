package main

import (
	"context"
	"flag"

	hndl "github.com/wizact/te-reo-bot/pkg/handlers"
)

// StartServerCommand is struct for info required to start an http server
type StartServerCommand struct {
	port    string
	address string
	tls     bool
}

// Flags returns the flag sets
func (fc *StartServerCommand) Flags() *flag.FlagSet {
	f := &flag.FlagSet{}

	f.StringVar(&fc.address, "address", "localhost", "-address=localhost")
	f.StringVar(&fc.port, "port", "8080", "-port=8080")
	f.BoolVar(&fc.tls, "tls", false, "-tls=true")

	return f
}

func (fc *StartServerCommand) Port() string {
	return fc.port
}

func (fc *StartServerCommand) Address() string {
	return fc.address
}

func (fc *StartServerCommand) Tls() bool {
	return fc.tls
}

// Name of the command
func (fc *StartServerCommand) Name() string {
	return "start-server"
}

// HelpString is the string shown as usage
func (fc *StartServerCommand) HelpString() string {
	return "Start the server using provided address and port"
}

// Run a command
func (fc *StartServerCommand) Run(ctx context.Context, args []string) error {
	if fc.Address() == "localhost" {
		fc.address = ""
	}

	hndl.StartServer(fc.Address(), fc.Port(), fc.Tls())

	return nil
}
