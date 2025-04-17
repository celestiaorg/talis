// Package commands implements the CLI commands for the application
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/celestiaorg/talis/internal/types"
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

		var req types.InstancesRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("error parsing JSON file: %w", err)
		}

		// Validate that instances array is not empty
		if len(req.Instances) == 0 {
			return fmt.Errorf("no instances specified in the JSON file")
		}

		// Call the API client to create the infrastructure
		if err := apiClient.CreateInstance(context.Background(), req); err != nil {
			return fmt.Errorf("error creating infrastructure: %w", err)
		}

		fmt.Println("Infrastructure creation request submitted successfully")

		// Create the delete request
		deleteReq := types.DeleteInstanceRequest{
			ProjectName: req.ProjectName,
			InstanceNames: func() []string {
				names := make([]string, 0)
				for _, instance := range req.Instances {
					if instance.Name != "" {
						names = append(names, instance.Name)
					} else {
						// If no specific name, use the base name pattern
						for i := 0; i < instance.NumberOfInstances; i++ {
							names = append(names, fmt.Sprintf("%s-%d", req.InstanceName, i))
						}
					}
				}
				return names
			}(),
		}

		// Generate the delete file name based on the create file name
		baseFileName := filepath.Base(jsonFile)
		deleteFileName := fmt.Sprintf("delete_%s", baseFileName)
		deleteFilePath := filepath.Join(filepath.Dir(jsonFile), deleteFileName)

		// Marshal the delete request to JSON
		deleteJSON, err := json.MarshalIndent(deleteReq, "", "  ")
		if err != nil {
			return fmt.Errorf("error generating delete file: %w", err)
		}

		// Write the delete file
		if err := os.WriteFile(deleteFilePath, deleteJSON, 0600); err != nil {
			return fmt.Errorf("error writing delete file: %w", err)
		}

		fmt.Printf("Delete file generated: %s (with project name: %s)\n", deleteFilePath, deleteReq.ProjectName)
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

		var req types.DeleteInstanceRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("error parsing JSON file: %w", err)
		}

		// Call the API client
		resp, err := apiClient.DeleteInstance(context.Background(), req)
		if err != nil {
			return fmt.Errorf("error deleting infrastructure: %w", err)
		}

		fmt.Printf("Infrastructure deletion task %s started. Use 'talis tasks get -n %s' to check the status\n", resp.TaskName, resp.TaskName)
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
