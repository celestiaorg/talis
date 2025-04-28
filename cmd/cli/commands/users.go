package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
	"github.com/spf13/cobra"
)

func init() {
	userCmd.AddCommand(listUsersCmd)
	userCmd.AddCommand(createUserCmd)
	userCmd.AddCommand(deleteUserCmd)
	// add commands, sub commands and flags

	listUsersCmd.Flags().StringP("username", "u", "", "returns a user with given username")

	createUserCmd.Flags().StringP("username", "u", "", "username of the user to be created")
	_ = createUserCmd.MarkFlagRequired("username")

	deleteUserCmd.Flags().IntP("id", "i", 0, "ID of the user to be deleted")
	_ = deleteUserCmd.MarkFlagRequired("id")
}

var userCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
}

// listUsersCmd represents the command to list users
var listUsersCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	Long:  `List all users with optional filtering by username.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		username, _ := cmd.Flags().GetString("username")

		// Build query options
		opts := handlers.UserGetParams{}
		if username != "" {
			opts.Username = username
		}

		// Call the API Client
		response, err := apiClient.GetUsers(context.Background(), opts)
		if err != nil {
			return fmt.Errorf("error fetching users: %w", err)
		}

		// Pretty print the response
		prettyJSON, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("error formatting response: %w", err)
		}
		fmt.Println(string(prettyJSON))
		return nil
	},
}

// GetUsersCmd returns the infrastructure command
func GetUsersCmd() *cobra.Command {
	return userCmd
}

var createUserCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a user",
	Long:  "Create a user with the given username",
	RunE: func(cmd *cobra.Command, _ []string) error {
		username, _ := cmd.Flags().GetString("username")

		var req = handlers.CreateUserParams{Username: username}

		response, err := apiClient.CreateUser(context.Background(), req)
		if err != nil {
			return fmt.Errorf("error creating a user: %w", err)
		}

		// Pretty print the response
		prettyJSON, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("error formatting response: %w", err)
		}
		fmt.Println(string(prettyJSON))
		return nil
	},
}

var deleteUserCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	Long:  "Delete a user with a given ID",
	RunE: func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetUint("id")

		err := apiClient.DeleteUser(context.Background(), handlers.DeleteUserParams{ID: userID})
		if err != nil {
			return fmt.Errorf("error while deleting user : %w", err)
		}
		fmt.Println(string("User deleted successfully"))
		return nil
	},
}
