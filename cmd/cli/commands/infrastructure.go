package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// CreateRequest represents the JSON structure for creating infrastructure
type CreateRequest struct {
	Name        string                           `json:"name"`
	ProjectName string                           `json:"project_name"`
	WebhookURL  string                           `json:"webhook_url,omitempty"`
	Instances   []infrastructure.InstanceRequest `json:"instances"`
}

// DeleteRequest represents the JSON structure for deleting infrastructure
type DeleteRequest struct {
	ID          uint                             `json:"id"`
	Name        string                           `json:"name"`
	ProjectName string                           `json:"project_name"`
	Instances   []infrastructure.InstanceRequest `json:"instances"`
}

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
	Short: "Create new infrastructure",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get API client
		client := getAPIClient(cmd)

		// Parse input file
		jsonFile, _ := cmd.Flags().GetString("file")
		if jsonFile == "" {
			return fmt.Errorf("error: JSON file not provided")
		}
		if err := validateFilePath(jsonFile); err != nil {
			return fmt.Errorf("error validating file path: %w", err)
		}
		// #nosec G304 -- file path is validated before use
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("error reading JSON file: %w", err)
		}

		var req CreateRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("error parsing JSON file: %w", err)
		}

		// Validate that instances array is not empty
		if len(req.Instances) == 0 {
			return fmt.Errorf("error: no instances specified in the JSON file")
		}

		// Call API client
		ctx := context.Background()
		resp, err := client.CreateInfrastructure(ctx, req)
		if err != nil {
			return fmt.Errorf("error creating infrastructure: %w", err)
		}

		// Process response
		prettyJSON, _ := json.MarshalIndent(resp, "", "  ")
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(prettyJSON))

		// Generate delete file
		baseFileName := filepath.Base(jsonFile)
		deleteFileName := fmt.Sprintf("delete_%s", baseFileName)
		deleteFilePath := filepath.Join(filepath.Dir(jsonFile), deleteFileName)

		// Extract the ID from the response
		respMap, ok := resp.(map[string]interface{})
		if !ok {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Warning: Failed to parse response for delete file generation\n")
			return nil
		}

		// Create delete request
		deleteReq := DeleteRequest{
			Name:        req.Name,
			ProjectName: req.ProjectName,
			Instances:   req.Instances,
		}

		// Set the ID from the response
		if id, ok := respMap["id"].(float64); ok {
			deleteReq.ID = uint(id)
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Warning: Failed to extract ID from response\n")
			return nil
		}

		// Marshal the delete request to JSON
		deleteJSON, err := json.MarshalIndent(deleteReq, "", "    ")
		if err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Warning: Failed to generate delete file: %v\n", err)
			return nil
		}

		// Write the delete file
		if err := os.WriteFile(deleteFilePath, deleteJSON, 0600); err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Warning: Failed to write delete file: %v\n", err)
			return nil
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Delete file generated: %s (with job ID: %d)\n", deleteFilePath, deleteReq.ID)
		return nil
	},
}

var deleteInfraCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete existing infrastructure",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get API client
		client := getAPIClient(cmd)

		// Parse input file
		jsonFile, _ := cmd.Flags().GetString("file")
		if jsonFile == "" {
			return fmt.Errorf("error: JSON file not provided")
		}
		if err := validateFilePath(jsonFile); err != nil {
			return fmt.Errorf("error validating file path: %w", err)
		}
		// #nosec G304 -- file path is validated before use
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("error reading JSON file: %w", err)
		}

		var req DeleteRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("error parsing JSON file: %w", err)
		}

		// Call API client
		ctx := context.Background()
		resp, err := client.DeleteInfrastructure(ctx, req)
		if err != nil {
			return fmt.Errorf("error deleting infrastructure: %w", err)
		}

		// Process response
		prettyJSON, _ := json.MarshalIndent(resp, "", "  ")
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(prettyJSON))
		return nil
	},
}

// GetInfraCmd returns the infrastructure command
func GetInfraCmd() *cobra.Command {
	return infraCmd
}

// Helper function to validate file path
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
