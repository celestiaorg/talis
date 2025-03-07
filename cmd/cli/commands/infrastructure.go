package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		var req CreateRequest

		jsonFile, _ := cmd.Flags().GetString("file")
		if jsonFile == "" {
			fmt.Println("Error: JSON file not provided")
			os.Exit(1)
		}
		if err := validateFilePath(jsonFile); err != nil {
			fmt.Printf("Error validating file path: %v\n", err)
			os.Exit(1)
		}
		// #nosec G304 -- file path is validated before use
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			fmt.Printf("Error reading JSON file: %v\n", err)
			os.Exit(1)
		}

		if err := json.Unmarshal(data, &req); err != nil {
			fmt.Printf("Error parsing JSON file: %v\n", err)
			os.Exit(1)
		}

		// Validate that instances array is not empty
		if len(req.Instances) == 0 {
			fmt.Println("Error: No instances specified in the JSON file")
			os.Exit(1)
		}

		jsonData, err := json.Marshal(req)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			os.Exit(1)
		}

		resp, err := http.Post("http://localhost:8080/api/v1/instances",
			"application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating infrastructure: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				err = fmt.Errorf("error closing response body: %w", cerr)
			}
		}()

		// Check if the response status code is an error
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error creating infrastructure. Status code: %d, Response: %s\n", resp.StatusCode, string(body))
			os.Exit(1)
		}

		// Decode the response body into a map
		var result map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		decoder.UseNumber() // Use json.Number instead of float64 for numbers
		if err := decoder.Decode(&result); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			os.Exit(1)
		}

		// Print the response in a pretty format
		prettyJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(prettyJSON))

		// The request was successful, generate a delete.json file
		// Create a delete request based on the create request
		deleteReq := DeleteRequest{
			Name:        req.Name,
			ProjectName: req.ProjectName,
			Instances:   req.Instances,
		}

		// Extract the job ID from the response
		idFound := false
		if id, ok := result["ID"]; ok {
			if num, ok := id.(json.Number); ok {
				if i, err := num.Int64(); err == nil {
					deleteReq.ID = uint(i)
					idFound = true
				}
			}
		}
		if !idFound {
			fmt.Println("Warning: Could not extract job ID from response. Using ID: 0")
			deleteReq.ID = 0
		}

		// Generate the delete file name based on the create file name
		baseFileName := filepath.Base(jsonFile)
		deleteFileName := fmt.Sprintf("delete_%s", baseFileName)

		// If the create file is in a different directory, preserve that path
		deleteFilePath := filepath.Join(filepath.Dir(jsonFile), deleteFileName)

		// Marshal the delete request to JSON
		deleteJSON, err := json.MarshalIndent(deleteReq, "", "    ")
		if err != nil {
			fmt.Printf("Warning: Failed to generate delete file: %v\n", err)
			return
		}

		// Write the delete file
		if err := os.WriteFile(deleteFilePath, deleteJSON, 0600); err != nil {
			fmt.Printf("Warning: Failed to write delete file: %v\n", err)
			return
		}

		fmt.Printf("Delete file generated: %s (with job ID: %d)\n", deleteFilePath, deleteReq.ID)
	},
}

var deleteInfraCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		var req DeleteRequest

		// Check if JSON file is provided
		jsonFile, _ := cmd.Flags().GetString("file")
		if jsonFile == "" {
			fmt.Println("Error: JSON file not provided")
			os.Exit(1)
		}
		if err := validateFilePath(jsonFile); err != nil {
			fmt.Printf("Error validating file path: %v\n", err)
			os.Exit(1)
		}
		// #nosec G304 -- file path is validated before use
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			fmt.Printf("Error reading JSON file: %v\n", err)
			os.Exit(1)
		}

		if err := json.Unmarshal(data, &req); err != nil {
			fmt.Printf("Error parsing JSON file: %v\n", err)
			os.Exit(1)
		}

		// Validate that instances array is not empty
		if len(req.Instances) == 0 {
			fmt.Println("Error: No instances specified in the JSON file")
			os.Exit(1)
		}

		jsonData, err := json.Marshal(req)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			os.Exit(1)
		}

		// Create a new request
		httpReq, err := http.NewRequest("DELETE", "http://localhost:8080/api/v1/instances",
			bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			os.Exit(1)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			fmt.Printf("Error deleting infrastructure: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				err = fmt.Errorf("error closing response body: %w", cerr)
			}
		}()

		var result interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			os.Exit(1)
		}

		prettyJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(prettyJSON))
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
