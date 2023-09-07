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

// Port gets the http server port
func (fc *StartServerCommand) Port() string {
	return fc.port
}

// Address gets the server address
func (fc *StartServerCommand) Address() string {
	return fc.address
}

// Tls gets the flag whether the server should run on TLS
func (fc *StartServerCommand) Tls() bool {
	return fc.tls
}

// Name gets the name of the command used in yacli package
func (fc *StartServerCommand) Name() string {
	return "start-server"
}

// HelpString gets the string shown as usage in cli
func (fc *StartServerCommand) HelpString() string {
	return "Start the server using provided address and port"
}

// Run the start server command
func (fc *StartServerCommand) Run(ctx context.Context, args []string) error {
	if fc.Address() == "localhost" {
		fc.address = ""
	}

	hndl.StartServer(fc.Address(), fc.Port(), fc.Tls())

	return nil
}
