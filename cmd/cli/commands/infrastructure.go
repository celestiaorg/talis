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
	Run: func(cmd *cobra.Command, args []string) {
		// Get API client
		client := getAPIClient(cmd)

		// Parse input file
		jsonFile, _ := cmd.Flags().GetString("file")
		if jsonFile == "" {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error: JSON file not provided")
			os.Exit(1)
		}
		if err := validateFilePath(jsonFile); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error validating file path: %v\n", err)
			os.Exit(1)
		}
		// #nosec G304 -- file path is validated before use
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error reading JSON file: %v\n", err)
			os.Exit(1)
		}

		var req CreateRequest
		if err := json.Unmarshal(data, &req); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error parsing JSON file: %v\n", err)
			os.Exit(1)
		}

		// Validate that instances array is not empty
		if len(req.Instances) == 0 {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error: No instances specified in the JSON file")
			os.Exit(1)
		}

		// Call API client
		ctx := context.Background()
		resp, err := client.CreateInfrastructure(ctx, req)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error creating infrastructure: %v\n", err)
			os.Exit(1)
		}

		// Process response
		prettyJSON, _ := json.MarshalIndent(resp, "", "  ")
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(prettyJSON))

		// Generate delete file
		baseFileName := filepath.Base(jsonFile)
		deleteFileName := fmt.Sprintf("delete_%s", baseFileName)
		deleteFilePath := filepath.Join(filepath.Dir(jsonFile), deleteFileName)

		// Create a delete request based on the create request
		deleteReq := DeleteRequest{
			Name:        req.Name,
			ProjectName: req.ProjectName,
			Instances:   req.Instances,
		}

		// Extract the job ID from the response
		if respMap, ok := resp.(map[string]interface{}); ok {
			if id, ok := respMap["id"]; ok {
				if idFloat, ok := id.(float64); ok {
					deleteReq.ID = uint(idFloat)
				}
			}
		} else if respPtr, ok := resp.(*map[string]interface{}); ok {
			if id, ok := (*respPtr)["id"]; ok {
				if idFloat, ok := id.(float64); ok {
					deleteReq.ID = uint(idFloat)
				}
			}
		}

		// Marshal the delete request to JSON
		deleteJSON, err := json.MarshalIndent(deleteReq, "", "    ")
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Failed to generate delete file: %v\n", err)
			return
		}

		// Write the delete file
		if err := os.WriteFile(deleteFilePath, deleteJSON, 0600); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Failed to write delete file: %v\n", err)
			return
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Delete file generated: %s (with job ID: %d)\n", deleteFilePath, deleteReq.ID)
	},
}

var deleteInfraCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		// Get API client
		client := getAPIClient(cmd)

		// Parse input file
		jsonFile, _ := cmd.Flags().GetString("file")
		if jsonFile == "" {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error: JSON file not provided")
			os.Exit(1)
		}
		if err := validateFilePath(jsonFile); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error validating file path: %v\n", err)
			os.Exit(1)
		}
		// #nosec G304 -- file path is validated before use
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error reading JSON file: %v\n", err)
			os.Exit(1)
		}

		var req DeleteRequest
		if err := json.Unmarshal(data, &req); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error parsing JSON file: %v\n", err)
			os.Exit(1)
		}

		// Validate that instances array is not empty
		if len(req.Instances) == 0 {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error: No instances specified in the JSON file")
			os.Exit(1)
		}

		// Call API client
		ctx := context.Background()
		resp, err := client.DeleteInfrastructure(ctx, req)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error deleting infrastructure: %v\n", err)
			os.Exit(1)
		}

		// Process response
		prettyJSON, _ := json.MarshalIndent(resp, "", "  ")
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(prettyJSON))
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
