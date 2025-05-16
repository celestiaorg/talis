package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/celestiaorg/talis/pkg/api/v1/client"
)

// flagOwnerID is the flag for the owner ID
const (
	// flagOwnerID is the flag for the owner ID
	flagOwnerID = "owner-id"
	// flagAPIURL is the flag for the API URL
	flagAPIURL = "api-url"
)

var (
	// apiClient is the shared API client instance
	apiClient client.Client
	// apiURLFlag holds the value for the API URL flag
	apiURLFlag string
)

// getAPIBaseURL determines the API base URL from flag, environment variable, or default value
func getAPIBaseURL() string {
	if apiURLFlag != "" {
		return apiURLFlag
	}
	if env := os.Getenv("TALIS_API_URL"); env != "" {
		return env
	}
	return client.DefaultOptions().BaseURL
}

// initClient initializes the API client with the appropriate base URL
func initClient() error {
	opts := client.DefaultOptions()
	opts.BaseURL = getAPIBaseURL()

	var err error
	apiClient, err = client.NewClient(opts)
	return err
}

func init() {
	RootCmd.PersistentFlags().StringP(flagOwnerID, "o", "", "Owner ID for resources")
	RootCmd.PersistentFlags().StringVar(&apiURLFlag, flagAPIURL, "", "Base URL for the Talis API (overrides TALIS_API_URL)")
	RootCmd.AddCommand(GetInfraCmd())
	RootCmd.AddCommand(GetUsersCmd())
	RootCmd.AddCommand(GetTasksCmd())
	RootCmd.AddCommand(GetProjectsCmd())
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "talis",
	Short: "Talis CLI - A command line interface for Talis API",
	Long: `Talis CLI is a command line tool for managing infrastructure and jobs through the Talis API.
	Complete documentation is available at https://github.com/celestiaorg/talis`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		return initClient()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}

// getOwnerID retrieves the owner ID from the command's persistent flags.
func getOwnerID(cmd *cobra.Command) (uint, error) {
	// Try to get the flag; cmd.Flag() searches the current command and its parents.
	flag := cmd.Flag(flagOwnerID)
	if flag == nil {
		// This means the flag wasn't defined anywhere in the command hierarchy.
		return 0, fmt.Errorf("flag '%s' is not defined", flagOwnerID)
	}

	ownerID := flag.Value.String()

	// Check if the ownerID flag was actually provided and is not empty
	if ownerID == "" {
		// Check if the flag was actually changed by the user (i.e., explicitly set, even if empty)
		// If it wasn't changed, it means the user didn't provide the flag.
		if !flag.Changed {
			return 0, fmt.Errorf("required flag(s) \"%s\" not set", flagOwnerID)
		}
		// If it was changed but is empty, it's an invalid value (though technically possible depending on flag type)
		// For our uint case, an empty string is invalid.
		return 0, fmt.Errorf("owner-id cannot be empty")
	}

	// Convert ownerID to uint
	ownerIDUint, err := strconv.ParseUint(ownerID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid owner-id format: %w", err)
	}

	return uint(ownerIDUint), nil
}
