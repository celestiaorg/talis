package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
)

// Flag names
const (
	flagName        = "name"
	flagProjectID   = "project_id"
	flagDescription = "description"
	flagConfig      = "config"
	flagPage        = "page"
)

// projectOutput represents the filtered output for a project
type projectOutput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// projectListOutput represents the filtered output for a list of projects
type projectListOutput struct {
	Projects []projectOutput `json:"projects"`
}

func init() {
	projectsCmd.AddCommand(createProjectCmd)
	projectsCmd.AddCommand(getProjectCmd)
	projectsCmd.AddCommand(listProjectsCmd)
	projectsCmd.AddCommand(deleteProjectCmd)
	projectsCmd.AddCommand(listProjectInstancesCmd)

	// Add flags for create
	createProjectCmd.Flags().StringP(flagName, "n", "", "Project name")
	createProjectCmd.Flags().StringP(flagDescription, "d", "", "Project description")
	createProjectCmd.Flags().StringP(flagConfig, "c", "", "Project configuration")
	if err := createProjectCmd.MarkFlagRequired(flagName); err != nil {
		panic(fmt.Errorf("failed to mark name flag as required for create project command: %w", err))
	}

	// Add flags for get
	getProjectCmd.Flags().StringP(flagName, "n", "", "Project name")
	if err := getProjectCmd.MarkFlagRequired(flagName); err != nil {
		panic(fmt.Errorf("failed to mark name flag as required for get project command: %w", err))
	}

	// Add flags for list
	listProjectsCmd.Flags().IntP(flagPage, "p", 1, "Page number for pagination")

	// Add flags for delete
	deleteProjectCmd.Flags().StringP(flagName, "n", "", "Project name")
	if err := deleteProjectCmd.MarkFlagRequired(flagName); err != nil {
		panic(fmt.Errorf("failed to mark name flag as required for delete project command: %w", err))
	}

	// Add flags for list instances
	listProjectInstancesCmd.Flags().StringP(flagName, "n", "", "Project name")
	listProjectInstancesCmd.Flags().IntP(flagPage, "p", 1, "Page number for pagination")
	if err := listProjectInstancesCmd.MarkFlagRequired(flagName); err != nil {
		panic(fmt.Errorf("failed to mark name flag as required for list project instances command: %w", err))
	}
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
}

var createProjectCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, err := cmd.Flags().GetString(flagName)
		if err != nil {
			return fmt.Errorf("error getting name flag: %w", err)
		}
		description, err := cmd.Flags().GetString(flagDescription)
		if err != nil {
			return fmt.Errorf("error getting description flag: %w", err)
		}
		config, err := cmd.Flags().GetString(flagConfig)
		if err != nil {
			return fmt.Errorf("error getting config flag: %w", err)
		}
		ownerID, err := getOwnerID(cmd)
		if err != nil {
			return fmt.Errorf("error getting owner_id: %w", err)
		}

		params := handlers.ProjectCreateParams{
			Name:        name,
			Description: description,
			Config:      config,
			OwnerID:     ownerID,
		}

		// Call the API client
		project, err := apiClient.CreateProject(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error creating project: %w", err)
		}

		// Filter the response to only include relevant fields
		output := projectOutput{
			Name:        project.Name,
			Description: project.Description,
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

var getProjectCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific project",
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, err := cmd.Flags().GetString(flagName)
		if err != nil {
			return fmt.Errorf("error getting name flag: %w", err)
		}
		ownerID, err := getOwnerID(cmd)
		if err != nil {
			return fmt.Errorf("error getting owner_id: %w", err)
		}

		params := handlers.ProjectGetParams{
			Name:    name,
			OwnerID: ownerID,
		}

		// Call the API client
		project, err := apiClient.GetProject(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error getting project: %w", err)
		}

		// Filter the response to only include relevant fields
		output := projectOutput{
			Name:        project.Name,
			Description: project.Description,
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

var listProjectsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, _ []string) error {
		page, err := cmd.Flags().GetInt(flagPage)
		if err != nil {
			return fmt.Errorf("error getting page flag: %w", err)
		}
		ownerID, err := getOwnerID(cmd)
		if err != nil {
			return fmt.Errorf("error getting owner_id: %w", err)
		}

		params := handlers.ProjectListParams{
			Page:    page,
			OwnerID: ownerID,
		}

		// Call the API client
		projects, err := apiClient.ListProjects(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error listing projects: %w", err)
		}

		// Filter the response to only include relevant fields
		output := projectListOutput{
			Projects: make([]projectOutput, len(projects)),
		}
		for i, project := range projects {
			output.Projects[i] = projectOutput{
				Name:        project.Name,
				Description: project.Description,
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

var deleteProjectCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a project",
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, err := cmd.Flags().GetString(flagName)
		if err != nil {
			return fmt.Errorf("error getting name flag: %w", err)
		}
		ownerID, err := getOwnerID(cmd)
		if err != nil {
			return fmt.Errorf("error getting owner_id: %w", err)
		}

		params := handlers.ProjectDeleteParams{
			Name:    name,
			OwnerID: ownerID,
		}

		// Call the API client
		if err := apiClient.DeleteProject(context.Background(), params); err != nil {
			return fmt.Errorf("error deleting project: %w", err)
		}

		fmt.Printf("Project '%s' deleted successfully\n", name)
		return nil
	},
}

var listProjectInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "List instances for a project",
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, err := cmd.Flags().GetString(flagName)
		if err != nil {
			return fmt.Errorf("error getting name flag: %w", err)
		}
		page, err := cmd.Flags().GetInt(flagPage)
		if err != nil {
			return fmt.Errorf("error getting page flag: %w", err)
		}
		ownerID, err := getOwnerID(cmd)
		if err != nil {
			return fmt.Errorf("error getting owner_id: %w", err)
		}

		// Lookup project ID by name and ownerID
		project, err := apiClient.GetProject(context.Background(), handlers.ProjectGetParams{
			Name:    name,
			OwnerID: ownerID,
		})
		if err != nil {
			return fmt.Errorf("error getting project: %w", err)
		}
		params := handlers.ProjectListInstancesParams{
			ProjectID: project.ID,
			Page:      page,
			OwnerID:   ownerID,
		}

		// Call the API client
		instances, err := apiClient.ListProjectInstances(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error listing project instances: %w", err)
		}

		// Create a simplified output structure
		type instanceOutput struct {
			Name     string `json:"name"`
			Status   string `json:"status"`
			PublicIP string `json:"public_ip,omitempty"`
			Region   string `json:"region"`
			Size     string `json:"size"`
		}

		output := struct {
			Instances []instanceOutput `json:"instances"`
		}{
			Instances: make([]instanceOutput, len(instances)),
		}

		for i, instance := range instances {
			output.Instances[i] = instanceOutput{
				Name:     instance.Name,
				Status:   instance.Status.String(),
				PublicIP: instance.PublicIP,
				Region:   instance.Region,
				Size:     instance.Size,
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

// GetProjectsCmd returns the projects command
func GetProjectsCmd() *cobra.Command {
	return projectsCmd
}
