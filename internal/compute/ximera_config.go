package compute

import (
	"fmt"
	"os"
	"strconv"

	computeTypes "github.com/celestiaorg/talis/internal/compute/types"
)

// InitXimeraConfig initializes the configuration from environment variables
func InitXimeraConfig() (*computeTypes.Configuration, error) {
	apiURL := os.Getenv("XIMERA_API_URL")
	apiToken := os.Getenv("XIMERA_API_TOKEN")
	userIDStr := os.Getenv("XIMERA_USER_ID")
	hypervisorGroupIDStr := os.Getenv("XIMERA_HYPERVISOR_GROUP_ID")

	if apiURL == "" {
		return nil, fmt.Errorf("XIMERA_API_URL is required in environment")
	}
	if apiToken == "" {
		return nil, fmt.Errorf("XIMERA_API_TOKEN is required in environment")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil && userIDStr != "" {
		return nil, fmt.Errorf("invalid XIMERA_USER_ID: %w", err)
	}
	hypervisorGroupID, err := strconv.Atoi(hypervisorGroupIDStr)
	if err != nil && hypervisorGroupIDStr != "" {
		return nil, fmt.Errorf("invalid XIMERA_HYPERVISOR_GROUP_ID: %w", err)
	}

	config := &computeTypes.Configuration{
		APIURL:       apiURL,
		APIToken:     apiToken,
		UserID:       userID,
		HypervisorID: hypervisorGroupID,
	}

	return config, nil
}
