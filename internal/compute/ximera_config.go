package compute

import (
	"fmt"
	"os"
	"strconv"

	computeTypes "github.com/celestiaorg/talis/internal/compute/types"
)

// InitXimeraConfig initializes the configuration from environment variables
func InitXimeraConfig() (*computeTypes.XimeraConfiguration, error) {
	apiURL := os.Getenv("XIMERA_API_URL")
	apiToken := os.Getenv("XIMERA_API_TOKEN")
	userIDStr := os.Getenv("XIMERA_USER_ID")
	hypervisorGroupIDStr := os.Getenv("XIMERA_HYPERVISOR_GROUP_ID")
	packageIDStr := os.Getenv("XIMERA_PACKAGE_ID")

	if apiURL == "" {
		return nil, fmt.Errorf("XIMERA_API_URL is required in environment")
	}
	if apiToken == "" {
		return nil, fmt.Errorf("XIMERA_API_TOKEN is required in environment")
	}

	userID := 0
	if userIDStr != "" {
		var err error
		userID, err = strconv.Atoi(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid XIMERA_USER_ID: %w", err)
		}
		if userID <= 0 {
			return nil, fmt.Errorf("XIMERA_USER_ID must be a positive integer")
		}
	}

	hypervisorGroupID := 0
	if hypervisorGroupIDStr != "" {
		var err error
		hypervisorGroupID, err = strconv.Atoi(hypervisorGroupIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid XIMERA_HYPERVISOR_GROUP_ID: %w", err)
		}
		if hypervisorGroupID <= 0 {
			return nil, fmt.Errorf("XIMERA_HYPERVISOR_GROUP_ID must be a positive integer")
		}
	}

	// Default package ID is 1 if not specified
	packageID := 1
	if packageIDStr != "" {
		var err error
		packageID, err = strconv.Atoi(packageIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid XIMERA_PACKAGE_ID: %w", err)
		}
		if packageID <= 0 {
			return nil, fmt.Errorf("XIMERA_PACKAGE_ID must be a positive integer")
		}
	}

	config := &computeTypes.XimeraConfiguration{
		APIURL:       apiURL,
		APIToken:     apiToken,
		UserID:       userID,
		HypervisorID: hypervisorGroupID,
		PackageID:    packageID,
	}

	return config, nil
}
