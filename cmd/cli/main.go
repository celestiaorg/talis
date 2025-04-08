// Package main provides the entry point for the CLI application
package main

import (
	"fmt"
	"os"

	"github.com/celestiaorg/talis/cmd/cli/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
