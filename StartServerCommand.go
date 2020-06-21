package main

import (
	"context"
	"flag"
	"fmt"
)

// FooCommand is foo command
type StartServerCommand struct {
	port string
	address string
}

// Flags returns the flag sets
func (fc *StartServerCommand) Flags() *flag.FlagSet {
	f := &flag.FlagSet{}

	f.StringVar(&fc.address, "address", "localhost", "-address=localhost")
	f.StringVar(&fc.port, "port", "8080", "-port=8080")

	return f
}

func (fc *StartServerCommand) getPort() string {
	return fc.port
}

func (fc *StartServerCommand) getAddress() string {
	return fc.address
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
	fmt.Println(fc.getAddress())
	fmt.Println(fc.getPort())
	return nil
}
