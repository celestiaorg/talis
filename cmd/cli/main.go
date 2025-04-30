// Package main provides the entry point for the CLI application
package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/celestiaorg/talis/cmd/cli/commands"
)

// init runs before main and loads the .env file.
func init() {
	// Load .env file first, before any command initialization
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist, just print a notice if verbose or debug logging is enabled later
		// For now, we won't exit, allowing defaults or flags to take precedence.
		fmt.Printf("Notice: Error loading .env file: %v\n", err)
	}
}

func main() {
	// Environment variables are now loaded via init()

	if err := commands.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
