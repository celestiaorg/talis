# Talis Go API Client Usage

This document provides guidance and examples on how to use the Talis Go API client (`pkg/api/v1/client`).

## Table of Contents

- [Initialization](#initialization)
- [Authentication](#authentication)
- [Context](#context)
- [Error Handling](#error-handling)
- [Health Check](#health-check)
- [User Management](#user-management)
  - [Create User](#create-user)
  - [Get User by ID](#get-user-by-id)
  - [List Users](#list-users)
  - [Delete User](#delete-user)
- [Project Management](#project-management)
  - [Create Project](#create-project)
  - [Get Project](#get-project)
  - [List Projects](#list-projects)
  - [Delete Project](#delete-project)
  - [List Project Instances](#list-project-instances)
- [Instance Management](#instance-management)
  - [Create Instance(s)](#create-instances)
  - [Get Instance](#get-instance)
  - [List Instances](#list-instances)
  - [List Instance Metadata](#list-instance-metadata)
  - [List Instance Public IPs](#list-instance-public-ips)
  - [Delete Instances](#delete-instances)
- [Task Management](#task-management)
  - [Get Task](#get-task)
  - [List Tasks](#list-tasks)
  - [Terminate Task](#terminate-task)
  - [Update Task Status](#update-task-status)

## Initialization

To use the API client, you first need to create an instance of it. The `client.NewClient` function is used for this, which accepts an `client.Options` struct.

```go
import (
	"context"
	"fmt"
	"log"
	"time"

	talisClient "github.com/celestiaorg/talis/pkg/api/v1/client"
	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
	"github.com/celestiaorg/talis/pkg/db/models"
	"github.com/celestiaorg/talis/pkg/types"
)

func main() {
	opts := &talisClient.Options{
		BaseURL: "http://localhost:8080", // Replace with your Talis API server URL
		APIKey:  "your-api-key",          // Optional: if your API requires an API key
		Timeout: 30 * time.Second,
	}

	apiClient, err := talisClient.NewClient(opts)
	if err != nil {
		log.Fatalf("Error creating API client: %v", err)
	}

	// You can now use apiClient to interact with the Talis API
	// Example: Perform a health check
	health, err := apiClient.HealthCheck(context.Background())
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Printf("API Health: %v\\n", health)
}
```

### Client Options

- `BaseURL` (string): The base URL of the Talis API server (e.g., `http://localhost:8080`).
- `APIKey` (string): At the moment it is an optional API key.
- `Timeout` (time.Duration): The default timeout for API requests.

## Authentication

If the Talis API server is configured to use API key authentication, provide the `APIKey` in the `client.Options` during initialization. The client will automatically include this key in the headers of its requests.

```go
opts := &talisClient.Options{
	BaseURL: "http://localhost:8080",
	APIKey:  "your-secret-api-key", // Set your API key here
	Timeout: 30 * time.Second,
}
apiClient, err := talisClient.NewClient(opts)
// ...
```

## Context

All client methods accept a `context.Context` as their first argument. This allows you to manage request cancellation, deadlines, and pass request-scoped values.

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Example: Listing projects with a timeout
projects, err := apiClient.ListProjects(ctx, handlers.ProjectListParams{})
if err != nil {
	// Handle error, potentially a context deadline exceeded
	log.Printf("Error listing projects: %v", err)
	return
}
// ...
```

## Error Handling

Client methods return an `error` as their last return value. Always check this error to ensure the operation was successful. Errors can be standard Go errors or `*fiber.Error` for HTTP-specific errors, which include the status code.

```go
user, err := apiClient.GetUserByID(ctx, handlers.UserGetByIDParams{ID: 1})
if err != nil {
	if fe, ok := err.(*fiber.Error); ok {
		log.Printf("API error (HTTP %d): %s", fe.Code, fe.Message)
	} else {
		log.Printf("Generic error getting user: %v", err)
	}
	return
}
// process user
```

## Health Check

You can check the health of the Talis API server:

```go
health, err := apiClient.HealthCheck(context.Background())
if err != nil {
	log.Fatalf("Health check failed: %v", err)
}
fmt.Printf("API Health Status: %v\\n", health)
```

## User Management

### Create User

To create a new user:

```go
createUserParams := handlers.CreateUserParams{
	Username: "newuser",
	Email:    "newuser@example.com",
	// Role is not implemented yet
}
userResponse, err := apiClient.CreateUser(context.Background(), createUserParams)
if err != nil {
	log.Fatalf("Error creating user: %v", err)
}
fmt.Printf("Created User ID: %d, Username: %s\\n", userResponse.ID, userResponse.Username)
```

### Get User by ID

To retrieve a user by their ID:

```go
userIDToGet := uint(1) // Example User ID
user, err := apiClient.GetUserByID(context.Background(), handlers.UserGetByIDParams{ID: userIDToGet})
if err != nil {
	log.Fatalf("Error getting user by ID %d: %v", userIDToGet, err)
}
fmt.Printf("Fetched User: %+v\\n", user)
```

### List Users

To list users, potentially with filters:

```go
// Example: List all users (params might support pagination/filtering)
listUserParams := handlers.UserGetParams{
	// Limit: 10,
	// Offset: 0,
}
userResponse, err := apiClient.GetUsers(context.Background(), listUserParams)
if err != nil {
	log.Fatalf("Error listing users: %v", err)
}
fmt.Printf("Found %d users. Users: %+v\\n", userResponse.Total, userResponse.Users)
for _, u := range userResponse.Users {
	fmt.Printf(" - User ID: %d, Username: %s\\n", u.ID, u.Username)
}
```

### Delete User

To delete a user:

```go
userIDToDelete := uint(1) // Example User ID
err := apiClient.DeleteUser(context.Background(), handlers.DeleteUserParams{ID: userIDToDelete})
if err != nil {
	log.Fatalf("Error deleting user ID %d: %v", userIDToDelete, err)
}
fmt.Printf("User ID %d deleted successfully.\\n", userIDToDelete)
```

## Project Management

### Create Project

To create a new project:

```go
// Assuming you have a user ID, e.g., from a previous CreateUser call or a known admin ID
ownerID := uint(1) 

createProjectParams := handlers.ProjectCreateParams{
	Name:        "my-awesome-project",
	Description: "This is a test project for Talis.",
	OwnerID:     ownerID,
}
project, err := apiClient.CreateProject(context.Background(), createProjectParams)
if err != nil {
	log.Fatalf("Error creating project: %v", err)
}
fmt.Printf("Created Project ID: %d, Name: %s\\n", project.ID, project.Name)
```

### Get Project

To retrieve a project by its name and owner ID:

```go
projectName := "my-awesome-project"
ownerID := uint(1) 
project, err := apiClient.GetProject(context.Background(), handlers.ProjectGetParams{Name: projectName, OwnerID: ownerID})
if err != nil {
	log.Fatalf("Error getting project '%s': %v", projectName, err)
}
fmt.Printf("Fetched Project: %+v\\n", project)
```

### List Projects

To list projects, potentially filtered by owner ID or other parameters:

```go
listProjectParams := handlers.ProjectListParams{
	// OwnerID: &ownerID, // Optional: filter by owner
}
projects, err := apiClient.ListProjects(context.Background(), listProjectParams)
if err != nil {
	log.Fatalf("Error listing projects: %v", err)
}
fmt.Printf("Found %d projects:\\n", len(projects))
for _, p := range projects {
	fmt.Printf(" - Project ID: %d, Name: %s, OwnerID: %d\\n", p.ID, p.Name, p.OwnerID)
}
```

### Delete Project

To delete a project:

```go
projectName := "my-awesome-project"
ownerID := uint(1) 
err := apiClient.DeleteProject(context.Background(), handlers.ProjectDeleteParams{Name: projectName, OwnerID: ownerID})
if err != nil {
	log.Fatalf("Error deleting project '%s': %v", projectName, err)
}
fmt.Printf("Project '%s' deleted successfully.\\n", projectName)
```

### List Project Instances

To list all instances associated with a specific project:

```go
projectName := "my-awesome-project"
ownerID := uint(1) 
projectInstancesParams := handlers.ProjectListInstancesParams{
	Name:    projectName,
	OwnerID: ownerID,
	// You can also use ListOptions here for pagination, etc.
	// ListOptions: &models.ListOptions{Limit: 10},
}
instances, err := apiClient.ListProjectInstances(context.Background(), projectInstancesParams)
if err != nil {
	log.Fatalf("Error listing instances for project '%s': %v", projectName, err)
}
fmt.Printf("Found %d instances for project '%s':\\n", len(instances), projectName)
for _, inst := range instances {
	fmt.Printf(" - Instance ID: %d, Name: %s, Status: %s\\n", inst.ID, inst.Name, inst.Status)
}
```

## Instance Management

### Create Instance(s)

To create one or more instances. This typically requires associating them with a project and an owner.

```go
// Ensure you have a valid project name and owner ID
projectName := "my-awesome-project"
ownerID := uint(1)      // The ID of the user who owns this project/instance
sshKeyName := "your-ssh-key-name" // The name of the SSH key registered with the cloud provider

instanceRequests := []types.InstanceRequest{
	{
		Name:              "webserver-01",
		OwnerID:           ownerID,
		ProjectName:       projectName,
		Provider:          models.ProviderDO, // e.g., DigitalOcean. Refer to models.ProviderID for options
		Region:            "nyc3",
		Size:              "s-1vcpu-1gb", // Provider-specific size slug
		Image:             "ubuntu-22-04-x64", // Provider-specific image slug
		SSHKeyName:        sshKeyName,
		NumberOfInstances: 1, // Creates one instance with this config
		Provision:         true, // Whether to run Ansible provisioning
		Tags:              []string{"web", "production"},
		Volumes: []types.VolumeConfig{
			{
				Name:       "data-volume",
				SizeGB:     50,
				MountPoint: "/mnt/data",
			},
		},
	},
	// Add more types.InstanceRequest structs here to create multiple instances in one call
}

err := apiClient.CreateInstance(context.Background(), instanceRequests)
if err != nil {
	log.Fatalf("Error creating instance(s): %v", err)
}
fmt.Println("Instance creation request submitted successfully.")
```
**Note on `types.InstanceRequest` fields:**
- `OwnerID`: The ID of the user owning the project and instances.
- `ProjectName`: The name of the project these instances belong to.
- `Provider`: Cloud provider ID (e.g., `models.ProviderDO` for DigitalOcean).
- `SSHKeyName`: The name of the SSH key as recognized by the cloud provider (this key must be pre-uploaded to the provider).
- `NumberOfInstances`: If greater than 1, multiple instances based on this configuration will be created, typically named `name-0`, `name-1`, etc.
- `Provision`: Boolean indicating whether to run Ansible playbooks after instance creation.

### Get Instance

To retrieve details of a specific instance by its Talis ID (not the provider's instance ID):

```go
instanceTalisID := "123" // This is the ID from the Talis database
instance, err := apiClient.GetInstance(context.Background(), instanceTalisID)
if err != nil {
	log.Fatalf("Error getting instance %s: %v", instanceTalisID, err)
}
fmt.Printf("Fetched Instance: %+v\\n", instance)
```

### List Instances

To list instances, with optional filtering using `models.ListOptions`:

```go
listOpts := &models.ListOptions{
	Limit:          10,
	Offset:         0,
	// StatusFilter: string(models.InstanceStatusReady), // Filter by specific status
}
instances, err := apiClient.GetInstances(context.Background(), listOpts)
if err != nil {
	log.Fatalf("Error listing instances: %v", err)
}
fmt.Printf("Found %d instances:\\n", len(instances))
for _, inst := range instances {
	fmt.Printf(" - ID: %d, Name: %s, Provider: %s, IP: %s, Status: %s\\n",
		inst.ID, inst.Name, inst.ProviderID, inst.PublicIP, inst.Status)
}
```

### List Instance Metadata

Similar to `ListInstances` but might return a more concise set of data, focused on metadata.

```go
listOpts := &models.ListOptions{Limit: 5}
metadata, err := apiClient.GetInstancesMetadata(context.Background(), listOpts)
if err != nil {
	log.Fatalf("Error listing instance metadata: %v", err)
}
fmt.Printf("Instance Metadata (%d results):\\n", len(metadata))
for _, meta := range metadata {
	fmt.Printf(" - ID: %d, Name: %s, Region: %s, Size: %s\\n", meta.ID, meta.Name, meta.Region, meta.Size)
}

```

### List Instance Public IPs

To get a list of public IP addresses for instances, with optional filtering:

```go
listOpts := &models.ListOptions{} // No specific filters, get all
publicIPsResponse, err := apiClient.GetInstancesPublicIPs(context.Background(), listOpts)
if err != nil {
	log.Fatalf("Error getting instance public IPs: %v", err)
}
fmt.Println("Public IPs:")
for _, ipInfo := range publicIPsResponse.PublicIPs {
	fmt.Printf(" - Instance Name: %s, IP: %s\\n", ipInfo.InstanceName, ipInfo.PublicIP)
}
```

### Delete Instances

To delete one or more instances. This operation is typically based on project name and a list of instance names.

```go
deleteRequest := types.DeleteInstancesRequest{
	ProjectName:   "my-awesome-project",
	InstanceNames: []string{"webserver-01", "webserver-02"}, // Names of instances to delete
	// OwnerID might be required depending on API implementation for authorization
}
err := apiClient.DeleteInstances(context.Background(), deleteRequest)
if err != nil {
	log.Fatalf("Error deleting instances: %v", err)
}
fmt.Println("Instance deletion request submitted successfully.")
```

## Task Management

Tasks represent asynchronous operations within Talis (e.g., instance provisioning).

### Get Task

To retrieve details of a specific task by its parameters (e.g., ID and Owner ID):

```go
taskParams := handlers.TaskGetParams{
	ID:      1, // Example Task ID
	OwnerID: 1, // Example Owner ID
}
task, err := apiClient.GetTask(context.Background(), taskParams)
if err != nil {
	log.Fatalf("Error getting task: %v", err)
}
fmt.Printf("Fetched Task: %+v\\n", task)
```

### List Tasks

To list tasks, potentially with filters:

```go
listTaskParams := handlers.TaskListParams{
	OwnerID: 1, // Example Owner ID
	// Limit: 10,
	// Offset: 0,
	// ProjectID: &projectID, // Optional filter by project
	// Status: string(models.TaskStatusCompleted), // Optional filter by status
}
tasks, err := apiClient.ListTasks(context.Background(), listTaskParams)
if err != nil {
	log.Fatalf("Error listing tasks: %v", err)
}
fmt.Printf("Found %d tasks:\\n", len(tasks))
for _, t := range tasks {
	fmt.Printf(" - Task ID: %d, Type: %s, Status: %s\\n", t.ID, t.Type, t.Status)
}
```

### Terminate Task

To request termination of an ongoing task:

```go
terminateParams := handlers.TaskTerminateParams{
	ID:      1, // Example Task ID
	OwnerID: 1, // Example Owner ID
}
err := apiClient.TerminateTask(context.Background(), terminateParams)
if err != nil {
	log.Fatalf("Error terminating task: %v", err)
}
fmt.Println("Task termination request submitted.")
```

### Update Task Status

This is likely an internal or admin-only endpoint for manually updating a task's status.

```go
updateStatusParams := handlers.TaskUpdateStatusParams{
	ID:      1, // Example Task ID
	OwnerID: 1, // Example Owner ID
	Status:  string(models.TaskStatusFailed), // New status, e.g., "failed"
	Message: "Manually marked as failed due to external issue.",
}
err := apiClient.UpdateTaskStatus(context.Background(), updateStatusParams)
if err != nil {
	log.Fatalf("Error updating task status: %v", err)
}
fmt.Println("Task status update request submitted.")
``` 