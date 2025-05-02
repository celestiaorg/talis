// Package commands implements the CLI commands for the application
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/celestiaorg/talis/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	infraCmd.AddCommand(createInfraCmd)
	infraCmd.AddCommand(deleteInfraCmd)

	// Add flags for create command
	createInfraCmd.Flags().StringP("file", "f", "", "JSON file containing infrastructure configuration")
	_ = createInfraCmd.MarkFlagRequired("file")

	// Add flags for delete command
	deleteInfraCmd.Flags().StringP("file", "f", "", "JSON file containing infrastructure configuration")
	_ = deleteInfraCmd.MarkFlagRequired("file")
}

var infraCmd = &cobra.Command{
	Use:   "infra",
	Short: "Manage infrastructure",
}

var createInfraCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new infrastructure",
	RunE: func(cmd *cobra.Command, _ []string) error {
		jsonFile, _ := cmd.Flags().GetString("file")
		if err := validateFilePath(jsonFile); err != nil {
			return fmt.Errorf("error validating file path: %w", err)
		}

		// Read and parse the JSON file
		data, err := os.ReadFile(jsonFile) //nolint:gosec
		if err != nil {
			return fmt.Errorf("error reading JSON file: %w", err)
		}

		var req []types.InstanceRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("error parsing JSON file: %w", err)
		}

		// Validate that instances array is not empty
		if len(req) == 0 {
			return fmt.Errorf("no instances specified in the JSON file")
		}

		// Call the API client to create the infrastructure
		if err := apiClient.CreateInstance(context.Background(), req); err != nil {
			return fmt.Errorf("error creating infrastructure: %w", err)
		}

		fmt.Println("Infrastructure creation request submitted successfully.")
		return nil
	},
}

var deleteInfraCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete infrastructure",
	RunE: func(cmd *cobra.Command, _ []string) error {
		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("error getting file path: %w", err)
		}

		// Read and parse the JSON file
		data, err := os.ReadFile(filepath.Clean(filePath))
		if err != nil {
			return fmt.Errorf("error reading JSON file: %w", err)
		}

		var req types.DeleteInstancesRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("error parsing JSON file: %w", err)
		}

		// Call the API client
		err = apiClient.DeleteInstances(context.Background(), req)
		if err != nil {
			return fmt.Errorf("error deleting infrastructure: %w", err)
		}

		fmt.Println("Infrastructure deletion started.")
		return nil
	},
}

// GetInfraCmd returns the infrastructure command
func GetInfraCmd() *cobra.Command {
	return infraCmd
}

// validateFilePath checks if the file path is valid and exists
func validateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("JSON file not provided")
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}
