// Package commands implements the CLI commands for the application
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

		var req []types.InstanceRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("error parsing JSON file: %w", err)
		}

		// Validate that instances array is not empty
		if len(req) == 0 {
			return fmt.Errorf("no instances specified in the JSON file")
		}

		// Call the API client to create the infrastructure
		createdInstances, err := apiClient.CreateInstance(context.Background(), req)
		if err != nil {
			return fmt.Errorf("error creating infrastructure: %w", err)
		}

		if len(createdInstances) == 0 {
			// If TaskIDs were present, include them in the message.
			fmt.Println("Infrastructure creation request submitted, but no instance details were immediately returned. Task information is no longer part of this direct response.")
			return nil
		}

		fmt.Println("Infrastructure creation request submitted successfully.")

		// Generate delete file
		instanceIDs := make([]uint, 0, len(createdInstances))
		for _, inst := range createdInstances {
			if inst == nil {
				continue // skip invalid entries instead of panicking
			}
			if inst.ID != 0 {
				instanceIDs = append(instanceIDs, inst.ID)
			}
		}

		// Assume all instance requests in the input file share the same ProjectName and OwnerID
		// This was an implicit assumption in the previous name-based delete file generation as well.
		projectName := req[0].ProjectName
		ownerID := req[0].OwnerID

		deleteReq := types.DeleteInstancesRequest{
			OwnerID:     ownerID,
			ProjectName: projectName,
			InstanceIDs: instanceIDs,
		}

		deleteFileName := fmt.Sprintf("delete-%s-%d.json", projectName, time.Now().Unix())
		deleteFilePath := filepath.Join(filepath.Dir(jsonFile), deleteFileName) // Create in the same dir as input a.json

		deleteFileContent, err := json.MarshalIndent(deleteReq, "", "  ")
		if err != nil {
			// Log an error but don't fail the command execution as instance creation was successful
			fmt.Fprintf(os.Stderr, "Warning: successfully created instances but failed to generate delete file: %v\n", err)
			return nil
		}
		if err := os.WriteFile(deleteFilePath, deleteFileContent, 0644); err != nil { //nolint:gosec
			fmt.Fprintf(os.Stderr, "Warning: successfully created instances but failed to write delete file '%s': %v\n", deleteFilePath, err)
			return nil
		}

		fmt.Printf("Successfully created instances. A delete file has been generated: %s\n", deleteFilePath)
		fmt.Println("You can use this file with 'talis infra delete -f <filename>' to terminate these instances.")
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
