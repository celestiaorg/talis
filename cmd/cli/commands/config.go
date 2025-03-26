package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/celestiaorg/talis/internal/api/v1/client"
	"github.com/celestiaorg/talis/internal/api/v1/routes"
)

// clientInstance is a singleton instance of the API client
var clientInstance client.Client

// getAPIClient returns the API client instance, creating it if necessary
func getAPIClient(cmd *cobra.Command) client.Client {
	if clientInstance != nil {
		return clientInstance
	}

	// Get configuration from flags or environment
	baseURL, _ := cmd.Flags().GetString("api-url")
	if baseURL == "" {
		baseURL = os.Getenv("TALIS_API_URL")
		if baseURL == "" {
			baseURL = routes.DefaultBaseURL // Default
		}
	}

	timeout, _ := cmd.Flags().GetDuration("timeout")
	if timeout == 0 {
		timeout = client.DefaultTimeout // Default
	}

	// Create client options
	opts := &client.ClientOptions{
		BaseURL: baseURL,
		Timeout: timeout,
	}

	// Create client
	var err error
	clientInstance, err = client.NewClient(opts)
	if err != nil {
		fmt.Printf("Error creating API client: %v\n", err)
		os.Exit(1)
	}

	return clientInstance
}

// AddClientFlags adds the API client configuration flags to the command
func AddClientFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("api-url", "", "API base URL (default: http://localhost:8080)")
	cmd.PersistentFlags().Duration("timeout", 30*time.Second, "API request timeout")
}

// GetAPIClient is the exported version of getAPIClient for testing
func GetAPIClient(cmd *cobra.Command) client.Client {
	return getAPIClient(cmd)
}
