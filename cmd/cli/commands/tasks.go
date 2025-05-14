package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
	"github.com/celestiaorg/talis/pkg/db/models"
)

// Task flag names
const (
	flagTaskID      = "id"
	flagProjectName = "project"
	flagTaskPage    = "page"
	flagInstanceID  = "instance-id"
	flagTaskAction  = "action"
	flagTaskLimit   = "limit"
	flagTaskOffset  = "offset"
)

// taskOutput represents the filtered output for a task
type taskOutput struct {
	ID      uint   `json:"id"`
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
	tasksCmd.AddCommand(listInstanceTasksCmd)

	// Add flags for get
	getTaskCmd.Flags().UintP(flagTaskID, "i", 0, "Task ID")
	_ = getTaskCmd.MarkFlagRequired(flagTaskID)

	// Add flags for list
	listTasksCmd.Flags().StringP(flagProjectName, "p", "", "Project name")
	listTasksCmd.Flags().IntP(flagTaskPage, "g", 1, "Page number for pagination") // Default is 1
	_ = listTasksCmd.MarkFlagRequired(flagProjectName)

	// Add flags for terminate
	terminateTaskCmd.Flags().UintP(flagTaskID, "i", 0, "Task ID")
	_ = terminateTaskCmd.MarkFlagRequired(flagTaskID)

	// Add flags for list-instance-tasks
	listInstanceTasksCmd.Flags().UintP(flagInstanceID, "I", 0, "Instance ID to list tasks for")
	listInstanceTasksCmd.Flags().StringP(flagTaskAction, "a", "", "Filter tasks by action (e.g., create_instances, terminate_instances)")
	listInstanceTasksCmd.Flags().Int(flagTaskLimit, 0, "Limit the number of tasks returned")
	listInstanceTasksCmd.Flags().Int(flagTaskOffset, 0, "Offset for paginating tasks")
	_ = listInstanceTasksCmd.MarkFlagRequired(flagInstanceID)
}

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage tasks",
}

var getTaskCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific task by its ID",
	RunE: func(cmd *cobra.Command, _ []string) error {
		taskID, err := cmd.Flags().GetUint(flagTaskID)
		if err != nil {
			return fmt.Errorf("error getting task ID flag: %w", err)
		}
		if taskID == 0 {
			return fmt.Errorf("task ID must be a positive number")
		}

		ownerID, err := getOwnerID(cmd)
		if err != nil {
			return fmt.Errorf("error getting owner_id: %w", err)
		}

		params := handlers.TaskGetParams{
			TaskID:  taskID,
			OwnerID: ownerID,
		}

		task, err := apiClient.GetTask(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error getting task: %w", err)
		}

		output := taskOutput{
			ID:      task.ID,
			Status:  string(task.Status),
			Action:  string(task.Action),
			Logs:    task.Logs,
			Error:   task.Error,
			Created: task.CreatedAt.Format("2006-01-02 15:04:05"),
		}

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

		ownerID, err := getOwnerID(cmd)
		if err != nil {
			return fmt.Errorf("error getting owner_id: %w", err)
		}

		params := handlers.TaskListParams{
			ProjectName: projectName,
			Page:        page,
			OwnerID:     ownerID,
		}

		tasks, err := apiClient.ListTasks(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error listing tasks: %w", err)
		}

		output := taskListOutput{
			Tasks: make([]taskOutput, len(tasks)),
		}
		for i, task := range tasks {
			output.Tasks[i] = taskOutput{
				ID:      task.ID,
				Status:  string(task.Status),
				Action:  string(task.Action),
				Logs:    task.Logs,
				Error:   task.Error,
				Created: task.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

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
	Short: "Terminate a running task by its ID",
	RunE: func(cmd *cobra.Command, _ []string) error {
		taskID, err := cmd.Flags().GetUint(flagTaskID)
		if err != nil {
			return fmt.Errorf("error getting task ID flag: %w", err)
		}
		if taskID == 0 {
			return fmt.Errorf("task ID must be a positive number")
		}

		ownerID, err := getOwnerID(cmd)
		if err != nil {
			return fmt.Errorf("error getting owner_id: %w", err)
		}

		params := handlers.TaskTerminateParams{
			TaskID:  taskID,
			OwnerID: ownerID,
		}

		if err := apiClient.TerminateTask(context.Background(), params); err != nil {
			return fmt.Errorf("error terminating task: %w", err)
		}

		fmt.Printf("Task ID %d termination request submitted successfully\n", taskID)
		return nil
	},
}

var listInstanceTasksCmd = &cobra.Command{
	Use:   "list-by-instance",
	Short: "List tasks for a specific instance ID, with optional action filter and pagination",
	RunE: func(cmd *cobra.Command, _ []string) error {
		instanceID, err := cmd.Flags().GetUint(flagInstanceID)
		if err != nil {
			return fmt.Errorf("error getting instance ID flag: %w", err)
		}
		if instanceID == 0 {
			return fmt.Errorf("instance ID must be a positive number")
		}

		actionFilter, _ := cmd.Flags().GetString(flagTaskAction)
		limit, _ := cmd.Flags().GetInt(flagTaskLimit)
		offset, _ := cmd.Flags().GetInt(flagTaskOffset)

		listOpts := &models.ListOptions{
			Limit:  limit,
			Offset: offset,
		}

		ownerID, err := getOwnerID(cmd)
		if err != nil {
			return fmt.Errorf("error getting owner_id: %w", err)
		}

		tasks, err := apiClient.ListTasksByInstanceID(context.Background(), ownerID, instanceID, actionFilter, listOpts)
		if err != nil {
			return fmt.Errorf("error listing tasks by instance ID: %w", err)
		}

		if len(tasks) == 0 {
			fmt.Printf("No tasks found for instance ID %d", instanceID)
			if actionFilter != "" {
				fmt.Printf(" with action '%s'", actionFilter)
			}
			fmt.Println(".")
			return nil
		}

		output := taskListOutput{
			Tasks: make([]taskOutput, len(tasks)),
		}
		for i, task := range tasks {
			output.Tasks[i] = taskOutput{
				ID:      task.ID,
				Status:  string(task.Status),
				Action:  string(task.Action),
				Logs:    task.Logs,
				Error:   task.Error,
				Created: task.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		prettyJSON, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("error formatting response: %w", err)
		}
		fmt.Println(string(prettyJSON))
		return nil
	},
}

// GetTasksCmd returns the tasks command
func GetTasksCmd() *cobra.Command {
	return tasksCmd
}
