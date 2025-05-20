# Talis Go API Client Usage

This document provides comprehensive guidance and examples on how to use the Talis Go API client (`pkg/api/v1/client`).

## Table of Contents

- [Getting Started](#getting-started)
  - [Installation](#installation)
  - [Initialization](#initialization)
  - [Client Options](#client-options)
- [Core Concepts](#core-concepts)
  - [Authentication](#authentication)
  - [Context Usage](#context-usage)
  - [Error Handling](#error-handling)
- [API Operations](#api-operations)
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
    - [List Tasks by Instance ID](#list-tasks-by-instance-id)
    - [Terminate Task](#terminate-task)
    - [Update Task Status](#update-task-status)
- [Best Practices](#best-practices)
  - [Timeouts and Cancellation](#timeouts-and-cancellation)
  - [Pagination](#pagination)
  - [Resource Cleanup](#resource-cleanup)
  - [Error Handling Strategies](#error-handling-strategies)
- [Troubleshooting](#troubleshooting)
  - [Common Errors](#common-errors)
  - [Debugging Tips](#debugging-tips)
  - [Support Resources](#support-resources)

## Getting Started

### Installation

The Talis API client is part of the Talis project. If you're developing within the Talis codebase, you already have access to it. If you're building an external application that needs to interact with a Talis API server, you'll need to import the client package:

```go
import "github.com/celestiaorg/talis/pkg/api/v1/client"
```

### Initialization

To use the API client, you first need to create an instance of it. The `client.NewClient` function is used for this, which accepts an `client.Options` struct or `nil` for default options.

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
    // Method 1: Using default options
    apiClient, err := talisClient.NewClient(nil)
    if err != nil {
        log.Fatalf("Error creating API client: %v", err)
    }
    
    // Method 2: Using custom options
    opts := &talisClient.Options{
        BaseURL: "http://localhost:8080", // Replace with your Talis API server URL
        APIKey:  "your-api-key",          // Optional: if your API requires an API key
        Timeout: 30 * time.Second,
    }

    apiClient, err = talisClient.NewClient(opts)
    if err != nil {
        log.Fatalf("Error creating API client: %v", err)
    }

    // You can now use apiClient to interact with the Talis API
    // Example: Perform a health check
    health, err := apiClient.HealthCheck(context.Background())
    if err != nil {
        log.Fatalf("Health check failed: %v", err)
    }
    fmt.Printf("API Health: %v\n", health)
}
```

### Client Options

The `client.Options` struct allows you to customize the client's behavior:

- `BaseURL` (string): The base URL of the Talis API server (e.g., `http://localhost:8080`). This should include the protocol (http:// or https://) and host, but not the endpoint paths.
- `APIKey` (string): An optional API key for authentication. If provided, this key will be included in all API requests.
- `Timeout` (time.Duration): The default request timeout. This value is used when no deadline is set in the context. If not specified, a default timeout of 30 seconds is used.

## Core Concepts

### Authentication

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

You can also update the API key after client creation:

```go
apiClient.SetAPIKey("new-api-key")
```

### Context Usage

All client methods accept a `context.Context` as their first argument. This allows you to manage request cancellation, deadlines, and pass request-scoped values.

```go
// Create a context with a timeout
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

// Create a context with cancellation
ctx, cancel = context.WithCancel(context.Background())
defer cancel()

// In another goroutine or based on some condition:
// cancel() // This will cancel any in-flight requests using this context
```

### Error Handling

Client methods return an `error` as their last return value. Always check this error to ensure the operation was successful. Errors can be standard Go errors or `*fiber.Error` for HTTP-specific errors, which include the status code.

```go
user, err := apiClient.GetUserByID(ctx, handlers.UserGetByIDParams{ID: 1})
if err != nil {
    if fe, ok := err.(*fiber.Error); ok {
        // Handle HTTP-specific errors
        switch fe.Code {
        case 404:
            log.Printf("User not found: %s", fe.Message)
        case 401, 403:
            log.Printf("Authentication/authorization error: %s", fe.Message)
        default:
            log.Printf("API error (HTTP %d): %s", fe.Code, fe.Message)
        }
    } else {
        // Handle other errors (network, context deadline, etc.)
        log.Printf("Generic error getting user: %v", err)
    }
    return
}
// process user
```

## API Operations

### Health Check

You can check the health of the Talis API server:

```go
health, err := apiClient.HealthCheck(context.Background())
if err != nil {
    log.Fatalf("Health check failed: %v", err)
}
fmt.Printf("API Health Status: %v\n", health)
// A successful response typically includes entries like {"status": "ok"}
```

### User Management

#### Create User

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
fmt.Printf("Created User ID: %d, Username: %s\n", userResponse.ID, userResponse.Username)
```

#### Get User by ID

To retrieve a user by their ID:

```go
userIDToGet := uint(1) // Example User ID
user, err := apiClient.GetUserByID(context.Background(), handlers.UserGetByIDParams{ID: userIDToGet})
if err != nil {
    log.Fatalf("Error getting user by ID %d: %v", userIDToGet, err)
}
fmt.Printf("Fetched User: %+v\n", user)
```

#### List Users

To list users, potentially with filters:

```go
// Example: List all users (params might support pagination/filtering)
listUserParams := handlers.UserGetParams{
    Limit:  10, // Optional: limit the number of results
    Offset: 0,  // Optional: pagination offset
}
userResponse, err := apiClient.GetUsers(context.Background(), listUserParams)
if err != nil {
    log.Fatalf("Error listing users: %v", err)
}
fmt.Printf("Found %d users. Users: %+v\n", userResponse.Total, userResponse.Users)
for _, u := range userResponse.Users {
    fmt.Printf(" - User ID: %d, Username: %s\n", u.ID, u.Username)
}
```

#### Delete User

To delete a user:

```go
userIDToDelete := uint(1) // Example User ID
err := apiClient.DeleteUser(context.Background(), handlers.DeleteUserParams{ID: userIDToDelete})
if err != nil {
    log.Fatalf("Error deleting user ID %d: %v", userIDToDelete, err)
}
fmt.Printf("User ID %d deleted successfully.\n", userIDToDelete)
```

### Project Management

#### Create Project

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
fmt.Printf("Created Project ID: %d, Name: %s\n", project.ID, project.Name)
```

#### Get Project

To retrieve a project by its name and owner ID:

```go
projectName := "my-awesome-project"
ownerID := uint(1) 
project, err := apiClient.GetProject(context.Background(), handlers.ProjectGetParams{Name: projectName, OwnerID: ownerID})
if err != nil {
    log.Fatalf("Error getting project '%s': %v", projectName, err)
}
fmt.Printf("Fetched Project: %+v\n", project)
```

#### List Projects

To list projects, potentially filtered by owner ID or other parameters:

```go
listProjectParams := handlers.ProjectListParams{
    OwnerID: ownerID, // Optional: filter by owner
    // Add other filters as needed
}
projects, err := apiClient.ListProjects(context.Background(), listProjectParams)
if err != nil {
    log.Fatalf("Error listing projects: %v", err)
}
fmt.Printf("Found %d projects:\n", len(projects))
for _, p := range projects {
    fmt.Printf(" - Project ID: %d, Name: %s, OwnerID: %d\n", p.ID, p.Name, p.OwnerID)
}
```

#### Delete Project

To delete a project:

```go
projectName := "my-awesome-project"
ownerID := uint(1) 
err := apiClient.DeleteProject(context.Background(), handlers.ProjectDeleteParams{Name: projectName, OwnerID: ownerID})
if err != nil {
    log.Fatalf("Error deleting project '%s': %v", projectName, err)
}
fmt.Printf("Project '%s' deleted successfully.\n", projectName)
```

#### List Project Instances

To list all instances associated with a specific project:

```go
projectName := "my-awesome-project"
ownerID := uint(1) 
projectInstancesParams := handlers.ProjectListInstancesParams{
    Name:    projectName,
    OwnerID: ownerID,
    // You can also use ListOptions here for pagination, etc.
    ListOptions: &models.ListOptions{
        Limit:  10,
        Offset: 0,
    },
}
instances, err := apiClient.ListProjectInstances(context.Background(), projectInstancesParams)
if err != nil {
    log.Fatalf("Error listing instances for project '%s': %v", projectName, err)
}
fmt.Printf("Found %d instances for project '%s':\n", len(instances), projectName)
for _, inst := range instances {
    fmt.Printf(" - Instance ID: %d, Name: %s, Status: %s\n", inst.ID, inst.Name, inst.Status)
}
```

### Instance Management

#### Create Instance(s)

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

createdInstances, err := apiClient.CreateInstance(context.Background(), instanceRequests)
if err != nil {
    log.Fatalf("Error creating instance(s): %v", err)
}
fmt.Printf("Created %d instances:\n", len(createdInstances))
for _, inst := range createdInstances {
    fmt.Printf(" - Instance ID: %d, Name: %s\n", inst.ID, inst.Name)
}
```

**Note on `types.InstanceRequest` fields:**
- `OwnerID`: The ID of the user owning the project and instances.
- `ProjectName`: The name of the project these instances belong to.
- `Provider`: Cloud provider ID (e.g., `models.ProviderDO` for DigitalOcean).
- `SSHKeyName`: The name of the SSH key as recognized by the cloud provider (this key must be pre-uploaded to the provider).
- `NumberOfInstances`: If greater than 1, multiple instances based on this configuration will be created, typically named `name-0`, `name-1`, etc.
- `Provision`: Boolean indicating whether to run Ansible playbooks after instance creation.

#### Get Instance

To retrieve details of a specific instance by its Talis ID (not the provider's instance ID):

```go
instanceTalisID := "123" // This is the ID from the Talis database
instance, err := apiClient.GetInstance(context.Background(), instanceTalisID)
if err != nil {
    log.Fatalf("Error getting instance %s: %v", instanceTalisID, err)
}
fmt.Printf("Fetched Instance: %+v\n", instance)
fmt.Printf("Instance Details:\n")
fmt.Printf(" - ID: %d\n", instance.ID)
fmt.Printf(" - Name: %s\n", instance.Name)
fmt.Printf(" - Provider: %s\n", instance.ProviderID)
fmt.Printf(" - Region: %s\n", instance.Region)
fmt.Printf(" - Size: %s\n", instance.Size)
fmt.Printf(" - Status: %s\n", instance.Status)
fmt.Printf(" - Public IP: %s\n", instance.PublicIP)
```

#### List Instances

To list instances, with optional filtering using `models.ListOptions`:

```go
// Create ListOptions with various filters
listOpts := &models.ListOptions{
    Limit:          10,
    Offset:         0,
    IncludeDeleted: false, // Set to true to include terminated instances
}

// Optionally filter by instance status
readyStatus := models.InstanceStatusReady
listOpts.InstanceStatus = &readyStatus

// Get instances
instances, err := apiClient.GetInstances(context.Background(), listOpts)
if err != nil {
    log.Fatalf("Error listing instances: %v", err)
}
fmt.Printf("Found %d instances:\n", len(instances))
for _, inst := range instances {
    fmt.Printf(" - ID: %d, Name: %s, Provider: %s, IP: %s, Status: %s\n",
        inst.ID, inst.Name, inst.ProviderID, inst.PublicIP, inst.Status)
}
```

#### List Instance Metadata

Similar to `ListInstances` but returns a more concise set of data, focused on metadata. This is more efficient when you don't need all instance details.

```go
listOpts := &models.ListOptions{Limit: 5}
metadata, err := apiClient.GetInstancesMetadata(context.Background(), listOpts)
if err != nil {
    log.Fatalf("Error listing instance metadata: %v", err)
}
fmt.Printf("Instance Metadata (%d results):\n", len(metadata))
for _, meta := range metadata {
    fmt.Printf(" - ID: %d, Name: %s, Region: %s, Size: %s\n", meta.ID, meta.Name, meta.Region, meta.Size)
}
```

#### List Instance Public IPs

To get a list of public IP addresses for instances, with optional filtering:

```go
listOpts := &models.ListOptions{} // No specific filters, get all
publicIPsResponse, err := apiClient.GetInstancesPublicIPs(context.Background(), listOpts)
if err != nil {
    log.Fatalf("Error getting instance public IPs: %v", err)
}
fmt.Println("Public IPs:")
for instanceID, ipInfo := range publicIPsResponse.PublicIPs {
    fmt.Printf(" - Instance ID: %d, Name: %s, IP: %s\n", 
        instanceID, ipInfo.InstanceName, ipInfo.PublicIP)
}
```

#### Delete Instances

To delete one or more instances. This operation is typically based on project name and a list of instance IDs.

```go
deleteRequest := types.DeleteInstancesRequest{
    ProjectName: "my-awesome-project",
    InstanceIDs: []uint{123, 124}, // IDs of instances to delete
    OwnerID:     ownerID,          // Required for authorization
}
err := apiClient.DeleteInstances(context.Background(), deleteRequest)
if err != nil {
    log.Fatalf("Error deleting instances: %v", err)
}
fmt.Println("Instance deletion request submitted successfully.")

// Note: Instance deletion is asynchronous. You can check the status
// by listing instances with IncludeDeleted=true and checking their status.
```

### Task Management

Tasks represent asynchronous operations within Talis (e.g., instance provisioning).

#### Get Task

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
fmt.Printf("Fetched Task: %+v\n", task)
fmt.Printf("Task Details:\n")
fmt.Printf(" - ID: %d\n", task.ID)
fmt.Printf(" - Action: %s\n", task.Action)
fmt.Printf(" - Status: %s\n", task.Status)
fmt.Printf(" - Created At: %s\n", task.CreatedAt.Format(time.RFC3339))
fmt.Printf(" - Updated At: %s\n", task.UpdatedAt.Format(time.RFC3339))
```

#### List Tasks

To list tasks, potentially with filters:

```go
listTaskParams := handlers.TaskListParams{
    OwnerID: 1, // Example Owner ID
    Limit:   10,
    Offset:  0,
    // ProjectID: &projectID, // Optional filter by project
    // Status: string(models.TaskStatusCompleted), // Optional filter by status
}
tasks, err := apiClient.ListTasks(context.Background(), listTaskParams)
if err != nil {
    log.Fatalf("Error listing tasks: %v", err)
}
fmt.Printf("Found %d tasks:\n", len(tasks))
for _, t := range tasks {
    fmt.Printf(" - Task ID: %d, Action: %s, Status: %s\n", t.ID, t.Action, t.Status)
}
```

#### List Tasks by Instance ID

To list tasks associated with a specific instance:

```go
ownerID := uint(1)
instanceID := uint(123)
actionFilter := string(models.TaskActionCreateInstances) // Optional: filter by action type
listOpts := &models.ListOptions{
    Limit:  10,
    Offset: 0,
}

tasks, err := apiClient.ListTasksByInstanceID(context.Background(), ownerID, instanceID, actionFilter, listOpts)
if err != nil {
    log.Fatalf("Error listing tasks for instance %d: %v", instanceID, err)
}
fmt.Printf("Found %d tasks for instance %d:\n", len(tasks), instanceID)
for _, t := range tasks {
    fmt.Printf(" - Task ID: %d, Action: %s, Status: %s\n", t.ID, t.Action, t.Status)
}
```

#### Terminate Task

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

#### Update Task Status

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

## Best Practices

### Timeouts and Cancellation

Always use appropriate timeouts for your API requests to prevent hanging operations:

```go
// Set a reasonable timeout for the operation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel() // Always defer cancel to prevent context leaks

// Use the context in your API call
result, err := apiClient.SomeOperation(ctx, params)
```

For long-running operations, consider implementing a cancellation mechanism:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// In a separate goroutine or based on user input:
go func() {
    time.Sleep(5 * time.Second)
    fmt.Println("Cancelling operation...")
    cancel()
}()

// This operation will be cancelled after 5 seconds
result, err := apiClient.LongRunningOperation(ctx, params)
if err != nil {
    if errors.Is(err, context.Canceled) {
        fmt.Println("Operation was cancelled")
    } else {
        fmt.Printf("Operation failed: %v\n", err)
    }
}
```

### Pagination

When listing resources, use pagination to avoid loading too much data at once:

```go
// Define how many items to fetch per page
const pageSize = 10
var offset uint = 0
var allItems []*models.SomeResource

// Loop until we've fetched all items
for {
    listOpts := &models.ListOptions{
        Limit:  pageSize,
        Offset: offset,
    }
    
    items, err := apiClient.ListSomeResource(ctx, listOpts)
    if err != nil {
        log.Fatalf("Error listing resources: %v", err)
    }
    
    // Add the items to our collection
    allItems = append(allItems, items...)
    
    // If we got fewer items than the page size, we've reached the end
    if len(items) < pageSize {
        break
    }
    
    // Update the offset for the next page
    offset += uint(len(items))
}

fmt.Printf("Fetched a total of %d items\n", len(allItems))
```

### Resource Cleanup

Always clean up resources when they're no longer needed:

```go
// Create some instances
instances, err := apiClient.CreateInstance(ctx, instanceRequests)
if err != nil {
    log.Fatalf("Failed to create instances: %v", err)
}

// Use defer to ensure cleanup happens even if there's an error later
defer func() {
    fmt.Println("Cleaning up instances...")
    deleteRequest := types.DeleteInstancesRequest{
        ProjectName: projectName,
        InstanceIDs: extractInstanceIDs(instances),
        OwnerID:     ownerID,
    }
    if err := apiClient.DeleteInstances(context.Background(), deleteRequest); err != nil {
        log.Printf("Warning: Failed to clean up instances: %v", err)
    }
}()

// Continue with your operations...
```

### Error Handling Strategies

Implement robust error handling for different types of errors:

```go
result, err := apiClient.SomeOperation(ctx, params)
if err != nil {
    // Check for specific error types
    if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout
        log.Printf("Operation timed out: %v", err)
        return
    }
    
    if errors.Is(err, context.Canceled) {
        // Handle cancellation
        log.Printf("Operation was canceled: %v", err)
        return
    }
    
    // Check for HTTP-specific errors
    if fe, ok := err.(*fiber.Error); ok {
        switch fe.Code {
        case 400:
            log.Printf("Bad request: %s", fe.Message)
        case 401:
            log.Printf("Authentication required: %s", fe.Message)
        case 403:
            log.Printf("Permission denied: %s", fe.Message)
        case 404:
            log.Printf("Resource not found: %s", fe.Message)
        case 429:
            log.Printf("Rate limit exceeded: %s", fe.Message)
            // Implement backoff/retry logic here
        default:
            log.Printf("API error (HTTP %d): %s", fe.Code, fe.Message)
        }
        return
    }
    
    // Handle other errors
    log.Printf("Unexpected error: %v", err)
    return
}

// Process the successful result
```

## Troubleshooting

### Common Errors

#### Connection Issues

If you're experiencing connection issues:

1. Verify the `BaseURL` is correct and the API server is running
2. Check network connectivity (firewalls, VPNs, etc.)
3. Ensure the API server is accessible from your environment

```go
// Test basic connectivity with a health check
health, err := apiClient.HealthCheck(ctx)
if err != nil {
    log.Printf("Connectivity issue: %v", err)
    // Implement further diagnostics...
}
```

#### Authentication Errors

If you're seeing 401 (Unauthorized) or 403 (Forbidden) errors:

1. Verify your API key is correct
2. Check that the API key has the necessary permissions
3. Ensure the API key hasn't expired

```go
// Update your API key if needed
apiClient.SetAPIKey("new-api-key")
```

#### Resource Not Found

If you're seeing 404 (Not Found) errors:

1. Double-check resource IDs, names, and other identifiers
2. Verify the resource exists (e.g., with a list operation)
3. Check if the resource might have been deleted

#### Rate Limiting

If you're being rate limited (429 Too Many Requests):

1. Implement exponential backoff and retry logic
2. Reduce the frequency of your API calls
3. Contact the API administrator if you need higher limits

### Debugging Tips

#### Enable Verbose Logging

Consider implementing a debug mode with verbose logging:

```go
// Set up a logger with appropriate level
logger := log.New(os.Stdout, "API CLIENT: ", log.LstdFlags)

// Log before and after API calls
logger.Printf("Calling GetInstances with opts: %+v", listOpts)
instances, err := apiClient.GetInstances(ctx, listOpts)
logger.Printf("GetInstances result: %d instances, err: %v", len(instances), err)
```

#### Inspect Request/Response

For deeper debugging, you might need to inspect the actual HTTP requests and responses. Consider using a proxy tool like [Charles](https://www.charlesproxy.com/) or [Fiddler](https://www.telerik.com/fiddler) to intercept and inspect the traffic.

### Support Resources

If you're still experiencing issues:

1. Check the Talis API documentation for updates or known issues
2. Review the client source code for additional insights
3. File an issue in the Talis repository with detailed reproduction steps
4. Contact the Talis team for support
