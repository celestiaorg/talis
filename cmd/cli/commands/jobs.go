package commands

import (
	"encoding/json"
	"fmt"
	"net/http"

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
		limit, _ := cmd.Flags().GetString("limit")
		status, _ := cmd.Flags().GetString("status")

		url := "http://localhost:8080/api/v1/jobs"
		if limit != "" || status != "" {
			url += "?"
			if limit != "" {
				url += fmt.Sprintf("limit=%s", limit)
			}
			if status != "" {
				if limit != "" {
					url += "&"
				}
				url += fmt.Sprintf("status=%s", status)
			}
		}

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error fetching jobs: %v\n", err)
			return
		}
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				fmt.Printf("Error closing response body: %v\n", cerr)
			}
		}()

		var result interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			return
		}

		prettyJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(prettyJSON))
	},
}

var getJobCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific job",
	Run: func(cmd *cobra.Command, args []string) {
		jobID, _ := cmd.Flags().GetString("id")

		resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/v1/jobs/%s", jobID))
		if err != nil {
			fmt.Printf("Error fetching job: %v\n", err)
			return
		}
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				fmt.Printf("Error closing response body: %v\n", cerr)
			}
		}()

		var result interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			return
		}

		prettyJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(prettyJSON))
	},
}

// GetJobsCmd returns the jobs command
func GetJobsCmd() *cobra.Command {
	return jobsCmd
}
