package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	jobsCmd.AddCommand(listJobsCmd)
	jobsCmd.AddCommand(getJobCmd)

	// Add flags
	listJobsCmd.Flags().StringP("limit", "l", "", "Limit the number of jobs returned")
	listJobsCmd.Flags().StringP("status", "s", "", "Filter jobs by status")

	getJobCmd.Flags().StringP("id", "i", "", "Job ID to fetch")
	_ = getJobCmd.MarkFlagRequired("id")
}

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage jobs",
}

var listJobsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all jobs",
	Run: func(cmd *cobra.Command, args []string) {
		// Get API client
		client := getAPIClient(cmd)

		// Parse flags
		limitStr, _ := cmd.Flags().GetString("limit")
		status, _ := cmd.Flags().GetString("status")

		// Convert limit to int
		var limit int
		if limitStr != "" {
			if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
				fmt.Printf("Error parsing limit: %v\n", err)
				os.Exit(1)
			}
		}

		// Call API client
		ctx := context.Background()
		resp, err := client.ListJobs(ctx, limit, status)
		if err != nil {
			fmt.Printf("Error fetching jobs: %v\n", err)
			os.Exit(1)
		}

		// Process response
		prettyJSON, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(prettyJSON))
	},
}

var getJobCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific job",
	Run: func(cmd *cobra.Command, args []string) {
		// Get API client
		client := getAPIClient(cmd)

		// Parse flags
		jobID, _ := cmd.Flags().GetString("id")

		// Call API client
		ctx := context.Background()
		resp, err := client.GetJob(ctx, jobID)
		if err != nil {
			fmt.Printf("Error fetching job: %v\n", err)
			os.Exit(1)
		}

		// Process response
		prettyJSON, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(prettyJSON))
	},
}

// GetJobsCmd returns the jobs command
func GetJobsCmd() *cobra.Command {
	return jobsCmd
}
