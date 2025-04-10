package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/celestiaorg/talis/internal/db/models"
)

// jobOutput represents the filtered output for a job
type jobOutput struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// jobListOutput represents the filtered output for a list of jobs
type jobListOutput struct {
	Jobs []jobOutput `json:"jobs"`
}

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
	RunE: func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetString("limit")
		status, _ := cmd.Flags().GetString("status")

		// Create list options
		opts := &models.ListOptions{}
		if limit != "" {
			limitInt, err := strconv.Atoi(limit)
			if err != nil {
				return fmt.Errorf("invalid limit value: %w", err)
			}
			opts.Limit = limitInt
		}
		if status != "" {
			jobStatus := models.JobStatus(status)
			opts.JobStatus = &jobStatus
		}

		// Call the API client
		response, err := apiClient.GetJobs(context.Background(), opts)
		if err != nil {
			return fmt.Errorf("error fetching jobs: %w", err)
		}

		// Filter the response to only include relevant fields
		output := jobListOutput{
			Jobs: make([]jobOutput, len(response.Jobs)),
		}
		for i, job := range response.Jobs {
			output.Jobs[i] = jobOutput{
				Name:   job.Name,
				Status: string(job.Status),
			}
		}

		// Pretty print the response
		prettyJSON, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("error formatting response: %w", err)
		}
		fmt.Println(string(prettyJSON))
		return nil
	},
}

var getJobCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific job",
	RunE: func(cmd *cobra.Command, _ []string) error {
		jobID, _ := cmd.Flags().GetString("id")

		// Call the API client
		job, err := apiClient.GetJob(context.Background(), jobID)
		if err != nil {
			return fmt.Errorf("error fetching job: %w", err)
		}

		// Filter the response to only include relevant fields
		output := jobOutput{
			Name:   job.Name,
			Status: string(job.Status),
		}

		// Pretty print the response
		prettyJSON, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("error formatting response: %w", err)
		}
		fmt.Println(string(prettyJSON))
		return nil
	},
}

// GetJobsCmd returns the jobs command
func GetJobsCmd() *cobra.Command {
	return jobsCmd
}
