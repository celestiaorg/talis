package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/celestiaorg/talis/cmd/cli/commands"
)

var rootCmd = &cobra.Command{
	Use:   "talis",
	Short: "Talis CLI - A command line interface for Talis API",
	Long: `Talis CLI is a command line tool for managing infrastructure and jobs through the Talis API.
Complete documentation is available at https://github.com/celestiaorg/talis`,
}

func init() {
	// Add client flags to the root command
	commands.AddClientFlags(rootCmd)

	// Add all subcommands to root command
	rootCmd.AddCommand(commands.GetInfraCmd())
	rootCmd.AddCommand(commands.GetJobsCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
