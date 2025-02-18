package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/spf13/cobra"
)

func init() {
	infraCmd.AddCommand(createInfraCmd)
	infraCmd.AddCommand(deleteInfraCmd)

	// Add flags for create command
	createInfraCmd.Flags().StringP("name", "n", "", "Infrastructure name")
	createInfraCmd.Flags().StringP("project", "p", "", "Project name")
	createInfraCmd.Flags().StringP("provider", "", "digitalocean", "Provider (digitalocean/aws)")
	createInfraCmd.Flags().IntP("instances", "i", 1, "Number of instances")
	createInfraCmd.Flags().StringP("region", "r", "", "Region")
	createInfraCmd.Flags().StringP("size", "s", "", "Instance size")
	createInfraCmd.Flags().StringP("webhook", "w", "", "Webhook URL")
	createInfraCmd.Flags().BoolP("provision", "", true, "Enable Nix provisioning")
	createInfraCmd.Flags().StringP("file", "f", "", "JSON file containing infrastructure configuration")

	// Add flags for delete command
	deleteInfraCmd.Flags().StringP("file", "f", "", "JSON file containing infrastructure configuration")
	deleteInfraCmd.Flags().StringP("name", "n", "", "Infrastructure name")
	deleteInfraCmd.Flags().StringP("project", "p", "", "Project name")
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

		// Check if JSON file is provided
		if jsonFile, _ := cmd.Flags().GetString("file"); jsonFile != "" {
			data, err := ioutil.ReadFile(jsonFile)
			if err != nil {
				fmt.Printf("Error reading JSON file: %v\n", err)
				os.Exit(1)
			}

			if err := json.Unmarshal(data, &req); err != nil {
				fmt.Printf("Error parsing JSON file: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Use command line flags
			name, _ := cmd.Flags().GetString("name")
			project, _ := cmd.Flags().GetString("project")
			provider, _ := cmd.Flags().GetString("provider")
			instances, _ := cmd.Flags().GetInt("instances")
			region, _ := cmd.Flags().GetString("region")
			size, _ := cmd.Flags().GetString("size")
			webhook, _ := cmd.Flags().GetString("webhook")
			provision, _ := cmd.Flags().GetBool("provision")

			// Validate required flags when not using JSON file
			if name == "" || project == "" || region == "" || size == "" {
				fmt.Println("Error: required flags not provided")
				fmt.Println("Required: --name, --project, --region, --size")
				fmt.Println("Alternatively, provide a JSON file with --file")
				os.Exit(1)
			}

			req = infrastructure.InstanceRequest{
				Name:        name,
				ProjectName: project,
				WebhookURL:  webhook,
				Instances: []infrastructure.Instance{
					{
						Provider:          provider,
						NumberOfInstances: instances,
						Region:            region,
						Size:              size,
						Provision:         provision,
					},
				},
			}
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
		if jsonFile, _ := cmd.Flags().GetString("file"); jsonFile != "" {
			data, err := ioutil.ReadFile(jsonFile)
			if err != nil {
				fmt.Printf("Error reading JSON file: %v\n", err)
				os.Exit(1)
			}

			if err := json.Unmarshal(data, &req); err != nil {
				fmt.Printf("Error parsing JSON file: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Use command line flags
			name, _ := cmd.Flags().GetString("name")
			project, _ := cmd.Flags().GetString("project")

			// Validate required flags when not using JSON file
			if name == "" || project == "" {
				fmt.Println("Error: required flags not provided")
				fmt.Println("Required: --name, --project")
				fmt.Println("Alternatively, provide a JSON file with --file")
				os.Exit(1)
			}

			req = infrastructure.DeleteRequest{
				Name:        name,
				ProjectName: project,
			}
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
