package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/spf13/cobra"
)

func init() {
	infraCmd.AddCommand(createInfraCmd)
	infraCmd.AddCommand(deleteInfraCmd)

	// Add flags for create command
	createInfraCmd.Flags().StringP("file", "f", "", "JSON file containing infrastructure configuration")

	// Add flags for delete command
	deleteInfraCmd.Flags().StringP("file", "f", "", "JSON file containing infrastructure configuration")
}

var infraCmd = &cobra.Command{
	Use:   "infra",
	Short: "Manage infrastructure",
}

var createInfraCmd = &cobra.Command{
	Use:   "create",
	Short: "Create infrastructure",
	Long:  `Create infrastructure based on a JSON configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		jsonFile, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		// Validate file path
		if err := validateFilePath(jsonFile); err != nil {
			return err
		}

		// Read and parse JSON file
		// #nosec G304 -- file path is validated before use
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("error reading JSON file: %w", err)
		}

		var createReq infrastructure.InstancesRequest
		if err := json.Unmarshal(data, &createReq); err != nil {
			return fmt.Errorf("error parsing JSON file: %w", err)
		}

		// Validate that instances array is not empty
		if len(createReq.Instances) == 0 {
			return fmt.Errorf("error: no instances specified in the JSON file")
		}

		// Call API client
		ctx := context.Background()
		err = clientInstance.CreateInstance(ctx, createReq)
		if err != nil {
			return fmt.Errorf("error creating infrastructure: %w", err)
		}

		// Process response
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Infrastructure created successfully")

		// Get Job ID
		// TODO: need more efficient way to get job ID or info. Like get job by name.
		jobID, err := clientInstance.GetJobs(ctx, infrastructure.JobRequest{
			Name: createReq.JobName,
		})
		if err != nil {
			return fmt.Errorf("error creating job: %w", err)
		}
		// Generate delete file
		deleteFilePath := filepath.Join(filepath.Dir(jsonFile), fmt.Sprintf("delete_%s.json", strings.TrimSuffix(filepath.Base(jsonFile), filepath.Ext(jsonFile))))
		deleteReq := infrastructure.DeleteInstanceRequest{
			ID:           resp.ID,
			InstanceName: createReq.InstanceName,
			ProjectName:  createReq.ProjectName,
			Instances:    createReq.Instances,
		}

		deleteJSON, err := json.MarshalIndent(deleteReq, "", "  ")
		if err != nil {
			return fmt.Errorf("error generating delete file: %w", err)
		}

		// #nosec G304 -- file path is constructed from a validated input file
		if err := os.WriteFile(deleteFilePath, deleteJSON, 0600); err != nil {
			return fmt.Errorf("error writing delete file: %w", err)
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nDelete file generated: %s\n", deleteFilePath)

		return nil
	},
}

var deleteInfraCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete infrastructure",
	Long:  `Delete infrastructure based on a JSON configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		jsonFile, _ := cmd.Flags().GetString("file")

		// Validate file path
		if err := validateFilePath(jsonFile); err != nil {
			return err
		}

		// Read and parse JSON file
		// #nosec G304 -- file path is validated before use
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("error reading JSON file: %w", err)
		}

		var deleteReq infrastructure.DeleteInstanceRequest
		if err := json.Unmarshal(data, &deleteReq); err != nil {
			return fmt.Errorf("error parsing JSON file: %w", err)
		}

		// Call API client
		ctx := context.Background()
		// Use a dummy job ID since we're deleting infrastructure
		err = clientInstance.DeleteInstance(ctx, deleteReq)
		if err != nil {
			return fmt.Errorf("error deleting infrastructure: %w", err)
		}

		// Process response
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Infrastructure deleted successfully")

		return nil
	},
}

// GetInfraCmd returns the infrastructure command
func GetInfraCmd() *cobra.Command {
	return infraCmd
}

// Helper function to validate file path
//
//nolint:unused // This function is used in other files
func validateFilePath(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}

	// Check if path contains any directory traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains invalid characters")
	}

	return nil
}
