package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
)

// Task flag names
const (
	flagTaskName    = "name" // Reusing "name" constant from projects if applicable, or define separately
	flagProjectName = "project"
	flagTaskPage    = "page"
)

// taskOutput represents the filtered output for a task
type taskOutput struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Action  string `json:"action"`
	Logs    string `json:"logs,omitempty"`
	Error   string `json:"error,omitempty"`
	Created string `json:"created_at"`
}

// taskListOutput represents the filtered output for a list of tasks
type taskListOutput struct {
	Tasks []taskOutput `json:"tasks"`
}

func init() {
	tasksCmd.AddCommand(getTaskCmd)
	tasksCmd.AddCommand(listTasksCmd)
	tasksCmd.AddCommand(terminateTaskCmd)

	// Add flags for get
	getTaskCmd.Flags().StringP(flagTaskName, "n", "", "Task name")
	if err := getTaskCmd.MarkFlagRequired(flagTaskName); err != nil {
		panic(fmt.Errorf("failed to mark name flag as required for get task command: %w", err))
	}

	// Add flags for list
	listTasksCmd.Flags().StringP(flagProjectName, "p", "", "Project name")
	listTasksCmd.Flags().IntP(flagTaskPage, "g", 1, "Page number for pagination") // Default is 1
	if err := listTasksCmd.MarkFlagRequired(flagProjectName); err != nil {
		panic(fmt.Errorf("failed to mark project flag as required for list tasks command: %w", err))
	}

	// Add flags for terminate
	terminateTaskCmd.Flags().StringP(flagTaskName, "n", "", "Task name")
	if err := terminateTaskCmd.MarkFlagRequired(flagTaskName); err != nil {
		panic(fmt.Errorf("failed to mark name flag as required for terminate task command: %w", err))
	}
}

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage tasks",
}

var getTaskCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific task",
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, err := cmd.Flags().GetString(flagTaskName)
		if err != nil {
			return fmt.Errorf("error getting name flag: %w", err)
		}

		params := handlers.TaskGetParams{
			TaskName: name,
		}

		// Call the API client
		task, err := apiClient.GetTask(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error getting task: %w", err)
		}

		// Filter the response to only include relevant fields
		output := taskOutput{
			Name:    task.Name,
			Status:  string(task.Status),
			Action:  string(task.Action),
			Logs:    task.Logs,
			Error:   task.Error,
			Created: task.CreatedAt.Format("2006-01-02 15:04:05"),
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

var listTasksCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks for a project",
	RunE: func(cmd *cobra.Command, _ []string) error {
		projectName, err := cmd.Flags().GetString(flagProjectName)
		if err != nil {
			return fmt.Errorf("error getting project flag: %w", err)
		}

		page, err := cmd.Flags().GetInt(flagTaskPage)
		if err != nil {
			return fmt.Errorf("error getting page flag: %w", err)
		}

		params := handlers.TaskListParams{
			ProjectName: projectName,
			Page:        page,
		}

		// Call the API client
		tasks, err := apiClient.ListTasks(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error listing tasks: %w", err)
		}

		// Filter the response to only include relevant fields
		output := taskListOutput{
			Tasks: make([]taskOutput, len(tasks)),
		}
		for i, task := range tasks {
			output.Tasks[i] = taskOutput{
				Name:    task.Name,
				Status:  string(task.Status),
				Action:  string(task.Action),
				Error:   task.Error,
				Created: task.CreatedAt.Format("2006-01-02 15:04:05"),
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

var terminateTaskCmd = &cobra.Command{
	Use:   "terminate",
	Short: "Terminate a running task",
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, err := cmd.Flags().GetString(flagTaskName)
		if err != nil {
			return fmt.Errorf("error getting name flag: %w", err)
		}

		params := handlers.TaskTerminateParams{
			TaskName: name,
		}

		// Call the API client
		if err := apiClient.TerminateTask(context.Background(), params); err != nil {
			return fmt.Errorf("error terminating task: %w", err)
		}

		fmt.Printf("Task '%s' termination request submitted successfully\n", name)
		return nil
	},
}

// GetTasksCmd returns the tasks command
func GetTasksCmd() *cobra.Command {
	return tasksCmd
}
