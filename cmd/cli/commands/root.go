package commands

import (
	"github.com/spf13/cobra"

	"github.com/celestiaorg/talis/pkg/api/v1/client"
)

var (
	// apiClient is the shared API client instance
	apiClient client.Client
)

// initClient initializes the API client with default options
func initClient() error {
	var err error
	apiClient, err = client.NewClient(client.DefaultOptions())
	if err != nil {
		return err
	}
	return nil
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "talis",
	Short: "Talis CLI tool",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		return initClient()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}
