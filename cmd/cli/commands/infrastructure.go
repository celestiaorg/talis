package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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
	Short: "Create new infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		var req infrastructure.InstanceRequest

		jsonFile, _ := cmd.Flags().GetString("file")
		if jsonFile == "" {
			fmt.Println("Error: JSON file not provided")
			os.Exit(1)
		}
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			fmt.Printf("Error reading JSON file: %v\n", err)
			os.Exit(1)
		}

		if err := json.Unmarshal(data, &req); err != nil {
			fmt.Printf("Error parsing JSON file: %v\n", err)
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
		defer resp.Body.Close()

		var result interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			os.Exit(1)
		}

		prettyJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(prettyJSON))
	},
}

var deleteInfraCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		var req infrastructure.DeleteRequest

		// Check if JSON file is provided
		jsonFile, _ := cmd.Flags().GetString("file")
		if jsonFile == "" {
			fmt.Println("Error: JSON file not provided")
			os.Exit(1)
		}
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			fmt.Printf("Error reading JSON file: %v\n", err)
			os.Exit(1)
		}

		if err := json.Unmarshal(data, &req); err != nil {
			fmt.Printf("Error parsing JSON file: %v\n", err)
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
		defer resp.Body.Close()

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
