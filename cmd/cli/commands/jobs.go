package commands

import (
	"context"
	"encoding/json"
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get API client
		client := getAPIClient(cmd)

		// Parse flags
		limitStr, _ := cmd.Flags().GetString("limit")
		status, _ := cmd.Flags().GetString("status")

		// Convert limit to int
		var limit int
		if limitStr != "" {
			if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
				return fmt.Errorf("error parsing limit: %w", err)
			}
		}

		// Call API client
		ctx := context.Background()
		resp, err := client.ListJobs(ctx, limit, status)
		if err != nil {
			return fmt.Errorf("error fetching jobs: %w", err)
		}

		// Process response
		prettyJSON, _ := json.MarshalIndent(resp, "", "  ")
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(prettyJSON))
		return nil
	},
}

var getJobCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific job",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get API client
		client := getAPIClient(cmd)

		// Parse flags
		jobID, _ := cmd.Flags().GetString("id")

		// Call API client
		ctx := context.Background()
		resp, err := client.GetJob(ctx, jobID)
		if err != nil {
			return fmt.Errorf("error fetching job: %w", err)
		}

		// Process response
		prettyJSON, _ := json.MarshalIndent(resp, "", "  ")
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(prettyJSON))
		return nil
	},
}

// GetJobsCmd returns the jobs command
func GetJobsCmd() *cobra.Command {
	return jobsCmd
}
