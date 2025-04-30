package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/celestiaorg/talis/pkg/api/v1/client"
	"github.com/celestiaorg/talis/pkg/api/v1/routes"
)

// flag names and environment variable names
const (
	// Flags
	flagOwnerID       = "owner-id"
	flagServerAddress = "server-address" // Renamed from flagAPIURL

	// Environment Variables
	envServerAddress = "TALIS_API_URL" // Use the merged env var name
)

var (
	// apiClient is the shared API client instance
	apiClient client.Client
	// serverAddress holds the target API server address. Flag parsing sets this.
	serverAddress string
)

// initClient initializes the API client with the appropriate base URL
func initClient() error {
	opts := client.DefaultOptions()
	opts.BaseURL = serverAddress // Use the finalized serverAddress

	var err error
	apiClient, err = client.NewClient(opts)
	return err
}

func init() {
	// Set a basic default for the flag. PersistentPreRunE will handle env var override.
	// Use flagServerAddress and the TALIS_API_URL env var name in the description.
	RootCmd.PersistentFlags().StringVarP(&serverAddress, flagServerAddress, "s", routes.DefaultBaseURL, "Address of the Talis API server (env: TALIS_API_URL)")

	RootCmd.PersistentFlags().StringP(flagOwnerID, "o", "", "Owner ID for resources")

	// Removed duplicate flag registration for apiURLFlag

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
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		// Check if the server address flag was explicitly set by the user.
		if !cmd.Flags().Changed(flagServerAddress) {
			// If not set via flag, check the environment variable *after* godotenv.Load() has run.
			// Use the correct env var name: envServerAddress (which is TALIS_API_URL)
			envAddr := os.Getenv(envServerAddress)
			if envAddr != "" {
				serverAddress = envAddr // Override the default value with the env var
			}
		}

		// Now serverAddress has the correct precedence: Flag > Env Var > Default
		fmt.Println("Talis Server address:", serverAddress) // Debug print

		// Validate the final server address
		if serverAddress == "" {
			return fmt.Errorf("server address cannot be empty")
		}
		// Initialization now happens here, using the correct serverAddress
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
