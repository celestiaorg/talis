# API Documentation

This document provides details on the available API endpoints, including RESTful and RPC-style endpoints.

**Base URL**: `http://localhost:8080` (default)
**API Prefix**: `/api/v1`

## Table of Contents

1.  [Health Check](#health-check)
2.  [Admin Endpoints](#admin-endpoints)
    *   [List All Instances (Admin)](#list-all-instances-admin)
    *   [Get All Instances Metadata (Admin)](#get-all-instances-metadata-admin)
3.  [Instance Endpoints](#instance-endpoints)
    *   [List Instances](#list-instances)
    *   [Get All Instances Metadata](#get-all-instances-metadata)
    *   [Get Public IPs of Instances](#get-public-ips-of-instances)
    *   [Create Instance(s)](#create-instances)
    *   [Get Instance Details](#get-instance-details)
    *   [Terminate Instances](#terminate-instances)
    *   [List Tasks for an Instance](#list-tasks-for-an-instance)
4.  [RPC Endpoint](#rpc-endpoint)
    *   [RPC Request Structure](#rpc-request-structure)
    *   [RPC Response Structure](#rpc-response-structure)
    *   [Project Methods](#project-methods)
        *   [`project.create`](#projectcreate)
        *   [`project.get`](#projectget)
        *   [`project.list`](#projectlist)
        *   [`project.delete`](#projectdelete)
        *   [`project.listInstances`](#projectlistinstances)
    *   [Task Methods](#task-methods)
        *   [`task.get`](#taskget)
        *   [`task.list`](#tasklist)
        *   [`task.terminate`](#taskterminate)
    *   [User Methods](#user-methods)
        *   [`user.create`](#usercreate)
        *   [`user.get`](#userget)
        *   [`user.get.id`](#usergetid)
        *   [`user.delete`](#userdelete)

---

## Health Check

### Health Check Endpoint

*   **Endpoint:** `GET /health`
*   **Description:** Checks the health status of the API.
*   **Request Body:** None
*   **Query Parameters:** None
*   **Example Request:**
    ```bash
    curl http://localhost:8080/health
    ```
*   **Example Response (200 OK):**
    ```json
    {
      "status": "healthy"
    }
    ```

---

## Admin Endpoints

These endpoints are typically for administrative purposes and might require special privileges.

### List All Instances (Admin)

*   **Endpoint:** `GET /api/v1/admin/instances`
*   **Route Name:** `AdminGetInstances`
*   **Handler:** `instanceHandler.ListInstances`
*   **Description:** Retrieves a list of all instances across all users/projects.
*   **Request Body:** None
*   **Query Parameters:**
    *   `limit` (int, optional, default: `DefaultPageSize` from `handlers`): Number of instances to return.
    *   `offset` (int, optional, default: 0): Offset for pagination.
    *   `include_deleted` (bool, optional, default: `false`): Whether to include deleted instances.
    *   `status` (string, optional): Filter instances by status (e.g., "running", "terminated"). See `models.InstanceStatus` for possible values.
*   **Example Request:**
    ```bash
    curl "http://localhost:8080/api/v1/admin/instances?limit=10&status=running"
    ```
*   **Example Response (200 OK):**
    ```json
    {
      "rows": [
        {
          "id": 1,
          "owner_id": 100,
          "project_name": "admin-project",
          "provider_instance_id": "prov-inst-id-1",
          "provider": "do",
          "name": "instance-01-admin",
          "public_ip": "192.0.2.1",
          "region": "nyc3",
          "size": "s-1vcpu-1gb",
          "image": "ubuntu-20-04-x64",
          "ssh_key_name": "admin-ssh-key",
          "status": "running",
          // ... other fields from models.Instance
        }
      ],
      "pagination": {
        "total": 1,
        "page": 1,
        "limit": 10,
        "offset": 0
      }
    }
    ```

### Get All Instances Metadata (Admin)

*   **Endpoint:** `GET /api/v1/admin/instances/all-metadata`
*   **Route Name:** `AdminGetInstancesMetadata`
*   **Handler:** `instanceHandler.GetAllMetadata`
*   **Description:** Retrieves metadata for all instances. Functionally similar to List All Instances (Admin) but might have a different internal purpose or representation in the future.
*   **Request Body:** None
*   **Query Parameters:** Same as [List All Instances (Admin)](#list-all-instances-admin).
*   **Example Request:**
    ```bash
    curl "http://localhost:8080/api/v1/admin/instances/all-metadata?include_deleted=true"
    ```
*   **Example Response (200 OK):** Same as [List All Instances (Admin)](#list-all-instances-admin).

---

## Instance Endpoints

### List Instances

*   **Endpoint:** `GET /api/v1/instances`
*   **Route Name:** `GetInstances`
*   **Handler:** `instanceHandler.ListInstances` (Note: The handler implies AdminID is used; this might be subject to OwnerID filtering in a real scenario based on TODOs in code).
*   **Description:** Retrieves a list of instances. (Assumed to be filterable by OwnerID in the future).
*   **Request Body:** None
*   **Query Parameters:** Same as [List All Instances (Admin)](#list-all-instances-admin).
*   **Example Request:**
    ```bash
    curl "http://localhost:8080/api/v1/instances?limit=5"
    ```
*   **Example Response (200 OK):** Similar structure to [List All Instances (Admin)](#list-all-instances-admin).

### Get All Instances Metadata

*   **Endpoint:** `GET /api/v1/instances/all-metadata`
*   **Route Name:** `GetMetadata`
*   **Handler:** `instanceHandler.GetAllMetadata`
*   **Description:** Retrieves metadata for instances. (Assumed to be filterable by OwnerID in the future).
*   **Request Body:** None
*   **Query Parameters:** Same as [List All Instances (Admin)](#list-all-instances-admin).
*   **Example Request:**
    ```bash
    curl "http://localhost:8080/api/v1/instances/all-metadata"
    ```
*   **Example Response (200 OK):** Similar structure to [List All Instances (Admin)](#list-all-instances-admin).

### Get Public IPs of Instances

*   **Endpoint:** `GET /api/v1/instances/public-ips`
*   **Route Name:** `GetPublicIPs`
*   **Handler:** `instanceHandler.GetPublicIPs`
*   **Description:** Retrieves a list of public IP addresses for instances.
*   **Request Body:** None
*   **Query Parameters:**
    *   `limit` (int, optional, default: `DefaultPageSize`): Number of items to return.
    *   `offset` (int, optional, default: 0): Offset for pagination.
    *   `include_deleted` (bool, optional, default: `false`): Whether to include IPs from deleted instances.
*   **Example Request:**
    ```bash
    curl "http://localhost:8080/api/v1/instances/public-ips"
    ```
*   **Example Response (200 OK):**
    ```json
    {
      "public_ips": [
        { "public_ip": "192.0.2.1" },
        { "public_ip": "192.0.2.2" }
      ],
      "pagination": {
        "total": 2,
        "page": 1,
        "limit": 10,
        "offset": 0
      }
    }
    ```

### Create Instance(s)

*   **Endpoint:** `POST /api/v1/instances`
*   **Route Name:** `CreateInstance`
*   **Handler:** `instanceHandler.CreateInstance`
*   **Description:** Creates one or more new instances based on the provided configurations.
*   **Request Body:** An array of `types.InstanceRequest` objects.
    ```json
    // types.InstanceRequest structure
    {
      "owner_id": 1, // Required: Owner ID of the instance
      "provider": "do", // Required: Cloud provider (e.g., "do", "aws", "gcp")
      "region": "nyc3", // Required: Region for instance creation
      "size": "s-1vcpu-1gb", // Required: Instance size/type
      "image": "ubuntu-20-04-x64", // Required: OS image
      "tags": ["web", "production"], // Optional: Tags
      "project_name": "my-web-app", // Required
      "ssh_key_name": "my-ssh-key", // Required: Name of the SSH key
      "number_of_instances": 1, // Required: Must be > 0
      "provision": true, // Optional: Whether to run Ansible provisioning
      "payload_path": "/abs/path/to/server/payload.sh", // Optional: Absolute path on API server to payload script
      "execute_payload": false, // Optional: Whether to execute the payload
      "volumes": [ // Required: At least one volume
        {
          "name": "data-volume",
          "size_gb": 50,
          "mount_point": "/mnt/data"
          // "region": "nyc3", // Optional: defaults to instance region
          // "filesystem": "ext4" // Optional
        }
      ],
      // "ssh_key_type": "rsa", // Optional: "rsa", "ed25519"
      // "ssh_key_path": "/custom/path/to/private_key" // Optional: Overrides default SSH key for Ansible
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '[
              {
                "owner_id": 1,
                "provider": "do",
                "region": "sfo3",
                "size": "s-2vcpu-4gb",
                "image": "debian-11-x64",
                "project_name": "batch-processing",
                "ssh_key_name": "batch-worker-key",
                "number_of_instances": 2,
                "provision": true,
                "volumes": [{ "name": "job-data", "size_gb": 100, "mount_point": "/data" }],
                "tags": ["batch", "worker"]
              }
            ]' \
         http://localhost:8080/api/v1/instances
    ```
*   **Example Response (201 Created):**
    ```json
    // types.Success(createdInstances)
    {
      "slug": "success",
      "data": [
        {
          "id": 10,
          "owner_id": 1,
          "project_name": "batch-processing",
          // ... other fields from models.Instance
          "status": "provisioning" // Or initial status
        },
        {
          "id": 11,
          "owner_id": 1,
          "project_name": "batch-processing",
          // ... other fields from models.Instance
          "status": "provisioning"
        }
      ]
    }
    ```

### Get Instance Details

*   **Endpoint:** `GET /api/v1/instances/:id`
*   **Route Name:** `GetInstance`
*   **Handler:** `instanceHandler.GetInstance`
*   **Description:** Retrieves details for a specific instance by its ID.
*   **Request Body:** None
*   **Path Parameters:**
    *   `id` (int, required): The ID of the instance.
*   **Example Request:**
    ```bash
    curl http://localhost:8080/api/v1/instances/10
    ```
*   **Example Response (200 OK):**
    ```json
    // models.Instance structure
    {
      "id": 10,
      "owner_id": 1,
      "project_name": "batch-processing",
      "provider_instance_id": "prov-inst-id-10",
      "provider": "do",
      "name": "instance-batch-01",
      "public_ip": "192.0.2.10",
      "region": "sfo3",
      "size": "s-2vcpu-4gb",
      "image": "debian-11-x64",
      "ssh_key_name": "batch-worker-key",
      "status": "running",
      "tags": ["batch", "worker"],
      // ... other fields
    }
    ```

### Terminate Instances

*   **Endpoint:** `DELETE /api/v1/instances`
*   **Route Name:** `TerminateInstances`
*   **Handler:** `instanceHandler.TerminateInstances`
*   **Description:** Terminates one or more specified instances.
*   **Request Body:** `types.DeleteInstancesRequest`
    ```json
    {
      "owner_id": 1, // Required
      "project_name": "batch-processing", // Required
      "instance_ids": [10, 11] // Required, array of instance IDs
    }
    ```
*   **Example Request:**
    ```bash
    curl -X DELETE -H "Content-Type: application/json" \
         -d '{
              "owner_id": 1,
              "project_name": "batch-processing",
              "instance_ids": [10, 11]
            }' \
         http://localhost:8080/api/v1/instances
    ```
*   **Example Response (200 OK):**
    ```json
    // types.Success(nil)
    {
      "slug": "success",
      "data": null
    }
    ```

### List Tasks for an Instance

*   **Endpoint:** `GET /api/v1/instances/:instance_id/tasks`
*   **Route Name:** `ListInstanceTasks`
*   **Handler:** `taskHandler.ListByInstanceID`
*   **Description:** Retrieves a list of tasks associated with a specific instance.
*   **Request Body:** None
*   **Path Parameters:**
    *   `instance_id` (int, required): The ID of the instance.
*   **Query Parameters:**
    *   `action` (string, optional): Filter tasks by action (e.g., "create_instances", "terminate_instances"). See `models.TaskAction`.
    *   `limit` (int, optional, default: `DefaultPageSize`): Number of tasks to return.
    *   `offset` (int, optional, default: 0): Offset for pagination.
*   **Example Request:**
    ```bash
    curl "http://localhost:8080/api/v1/instances/10/tasks?action=create_instances&limit=5"
    ```
*   **Example Response (200 OK):**
    ```json
    // types.Success(types.ListResponse[models.Task])
    {
      "slug": "success",
      "data": {
        "rows": [
          {
            "id": 1,
            "instance_id": 10,
            "project_name": "batch-processing",
            "action": "create_instances",
            "status": "completed",
            "payload": "{"instance_config": ...}",
            // ... other fields from models.Task
          }
        ],
        "pagination": {
          "total": 1,
          "page": 1,
          "limit": 5,
          "offset": 0
        }
      }
    }
    ```

---

## RPC Endpoint

The API provides a single RPC endpoint for various operations related to projects, tasks, and users.

*   **Endpoint:** `POST /api/v1/`
*   **Route Name:** `RPC`
*   **Handler:** `rpcHandler.HandleRPC`
*   **Description:** A general-purpose RPC endpoint. The specific operation is determined by the `method` field in the request body.

### RPC Request Structure

All RPC calls use the following JSON structure in the request body:

```json
{
  "method": "service.action", // e.g., "project.create", "task.list"
  "params": {
    // Parameters specific to the method
  },
  "id": "request-identifier-123" // Optional: A client-generated ID echoed in the response
}
```

### RPC Response Structure

RPC responses follow this structure:

```json
{
  "data": {
    // Result of the operation, structure varies by method
  },
  "error": { // Present only if an error occurred
    "code": 123, // Numeric error code
    "message": "Error description",
    "data": { /* Optional additional error details */ }
  },
  "id": "request-identifier-123", // Echoed from request if provided
  "success": true // or false if an error occurred
}
```

### Project Methods

Dispatched by `rpcHandler.handleProjectMethod` to `ProjectHandlers`.

#### `project.create`

*   **Description:** Creates a new project.
*   **Handler:** `ProjectHandlers.Create`
*   **Params (`handlers.ProjectCreateParams`):**
    ```json
    {
      "name": "new-project-alpha", // Required
      "description": "Alpha project description.", // Optional
      "config": "{"setting1": "value1"}", // Optional: JSON string or complex object for project config
      "owner_id": 1 // Required
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "project.create",
              "params": {
                "name": "my-new-project",
                "owner_id": 1,
                "description": "A project for testing."
              },
              "id": "proj-create-001"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": { // models.Project structure
        "id": 5,
        "owner_id": 1,
        "name": "my-new-project",
        "description": "A project for testing.",
        "config": null,
        // ... other project fields
      },
      "success": true,
      "id": "proj-create-001"
    }
    ```

#### `project.get`

*   **Description:** Retrieves details of a specific project by name and owner.
*   **Handler:** `ProjectHandlers.Get`
*   **Params (`handlers.ProjectGetParams`):**
    ```json
    {
      "name": "my-new-project", // Required
      "owner_id": 1 // Required
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "project.get",
              "params": {
                "name": "my-new-project",
                "owner_id": 1
              },
              "id": "proj-get-002"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": { // models.Project structure
        "id": 5,
        "owner_id": 1,
        "name": "my-new-project",
        // ...
      },
      "success": true,
      "id": "proj-get-002"
    }
    ```

#### `project.list`

*   **Description:** Lists projects for a given owner, with pagination.
*   **Handler:** `ProjectHandlers.List`
*   **Params (`handlers.ProjectListParams`):**
    ```json
    {
      "owner_id": 1, // Required
      "page": 1 // Optional, default: 1 (for pagination options)
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "project.list",
              "params": {
                "owner_id": 1,
                "page": 1
              },
              "id": "proj-list-003"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": { // types.ListResponse[models.Project]
        "rows": [
          {
            "id": 5,
            "owner_id": 1,
            "name": "my-new-project",
            // ...
          },
          {
            "id": 6,
            "owner_id": 1,
            "name": "another-project",
            // ...
          }
        ],
        "pagination": {
          "total": 2,
          "page": 1,
          "limit": 10, // DefaultPageSize
          "offset": 0
        }
      },
      "success": true,
      "id": "proj-list-003"
    }
    ```

#### `project.delete`

*   **Description:** Deletes a project by name and owner.
*   **Handler:** `ProjectHandlers.Delete`
*   **Params (`handlers.ProjectDeleteParams`):**
    ```json
    {
      "name": "my-new-project", // Required
      "owner_id": 1 // Required
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "project.delete",
              "params": {
                "name": "my-new-project",
                "owner_id": 1
              },
              "id": "proj-delete-004"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "success": true,
      "id": "proj-delete-004"
    }
    ```

#### `project.listInstances`

*   **Description:** Lists all instances associated with a specific project.
*   **Handler:** `ProjectHandlers.ListInstances`
*   **Params (`handlers.ProjectListInstancesParams`):**
    ```json
    {
      "name": "another-project", // Required: Project name
      "owner_id": 1, // Required
      "page": 1 // Optional, default: 1 (for pagination options)
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "project.listInstances",
              "params": {
                "name": "another-project",
                "owner_id": 1
              },
              "id": "proj-listinst-005"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": { // types.ListResponse[models.Instance]
        "rows": [
          {
            "id": 20,
            "owner_id": 1,
            "project_name": "another-project",
            "name": "instance-alpha",
            // ... other instance fields
          }
        ],
        "pagination": {
          "total": 1,
          "page": 1,
          "limit": 10, // DefaultPageSize
          "offset": 0
        }
      },
      "success": true,
      "id": "proj-listinst-005"
    }
    ```

### Task Methods

Dispatched by `rpcHandler.handleTaskMethod` to `TaskHandlers`.
Note: `ownerID` for task methods is currently hardcoded to `models.AdminID` in `rpcHandler.handleTaskMethod` and marked with a TODO to get from JWT. The `Task...Params` structs often include `OwnerID`, which might be used by the service layer.

#### `task.get`

*   **Description:** Retrieves details of a specific task by its ID.
*   **Handler:** `TaskHandlers.Get`
*   **Params (`handlers.TaskGetParams`):**
    ```json
    {
      "task_id": 15, // Required
      "owner_id": 1 // Required by params struct, but RPC handler uses AdminID for now.
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "task.get",
              "params": {
                "task_id": 15,
                "owner_id": 1
              },
              "id": "task-get-001"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": { // models.Task structure
        "id": 15,
        "project_name": "some-project",
        "action": "create_instances",
        "status": "completed",
        // ...
      },
      "success": true,
      "id": "task-get-001"
    }
    ```

#### `task.list`

*   **Description:** Lists tasks for a given project and owner, with pagination.
*   **Handler:** `TaskHandlers.List`
*   **Params (`handlers.TaskListParams`):**
    ```json
    {
      "projectName": "another-project", // Required
      "owner_id": 1, // Required by params struct
      "page": 1 // Optional, default: 1
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "task.list",
              "params": {
                "projectName": "another-project",
                "owner_id": 1
              },
              "id": "task-list-002"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": { // types.ListResponse[models.Task]
        "rows": [
          {
            "id": 22,
            "project_name": "another-project",
            "action": "terminate_instances",
            // ...
          }
        ],
        "pagination": {
          "total": 1,
          "page": 1,
          "limit": 10, // DefaultPageSize
          "offset": 0
        }
      },
      "success": true,
      "id": "task-list-002"
    }
    ```

#### `task.terminate`

*   **Description:** Terminates a running task. (Internally, this updates the task status to `models.TaskStatusTerminated`).
*   **Handler:** `TaskHandlers.Terminate`
*   **Params (`handlers.TaskTerminateParams`):**
    ```json
    {
      "task_id": 25, // Required
      "owner_id": 1 // Required by params struct
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "task.terminate",
              "params": {
                "task_id": 25,
                "owner_id": 1
              },
              "id": "task-terminate-003"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "success": true,
      "id": "task-terminate-003"
    }
    ```

### User Methods

Dispatched by `rpcHandler.handleUserMethod` to `UserHandlers`.

#### `user.create`

*   **Description:** Creates a new user.
*   **Handler:** `UserHandlers.CreateUser`
*   **Params (`handlers.CreateUserParams`):**
    ```json
    {
      "username": "newuser", // Required
      "email": "newuser@example.com", // Optional, validated if provided
      "role": "user", // Optional, see models.UserRole for values (e.g., "user", "admin")
      "public_ssh_key": "ssh-rsa AAAA..." // Optional
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "user.create",
              "params": {
                "username": "johndoe",
                "email": "johndoe@example.com",
                "role": "user"
              },
              "id": "user-create-001"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": { // types.CreateUserResponse
        "user_id": 12
      },
      "success": true,
      "id": "user-create-001"
    }
    ```

#### `user.get`

*   **Description:** Retrieves users. If `username` is provided, gets a specific user. Otherwise, lists all users with pagination.
*   **Handler:** `UserHandlers.GetUsers`
*   **Params (`handlers.UserGetParams`):**
    ```json
    {
      "username": "johndoe", // Optional: If provided, fetches this specific user
      "page": 1 // Optional: For listing all users, default: 1
    }
    ```
*   **Example Request (Get Specific User):**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "user.get",
              "params": {
                "username": "johndoe"
              },
              "id": "user-get-specific-002"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Specific User):**
    ```json
    {
      "data": { // types.UserResponse with single User
        "user": {
          "id": 12,
          "username": "johndoe",
          "email": "johndoe@example.com",
          "role": "user",
          // ...
        }
      },
      "success": true,
      "id": "user-get-specific-002"
    }
    ```
*   **Example Request (List All Users):**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "user.get",
              "params": {
                "page": 1
              },
              "id": "user-get-list-003"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (List Users):**
    ```json
    {
      "data": { // types.UserResponse with Users array and Pagination
        "users": [
          { "id": 1, "username": "admin", /* ... */ },
          { "id": 12, "username": "johndoe", /* ... */ }
        ],
        "pagination": {
          "total": 2,
          "page": 1,
          "limit": 10, // DefaultPageSize
          "offset": 0
        }
      },
      "success": true,
      "id": "user-get-list-003"
    }
    ```

#### `user.get.id`

*   **Description:** Retrieves a specific user by their ID.
*   **Handler:** `UserHandlers.GetUserByID`
*   **Params (`handlers.UserGetByIDParams`):**
    ```json
    {
      "id": 12 // Required
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "user.get.id",
              "params": {
                "id": 12
              },
              "id": "user-get-id-004"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": { // types.UserResponse with single User
        "user": {
          "id": 12,
          "username": "johndoe",
          // ...
        }
      },
      "success": true,
      "id": "user-get-id-004"
    }
    ```

#### `user.delete`

*   **Description:** Deletes a user by their ID.
*   **Handler:** `UserHandlers.DeleteUser`
*   **Params (`handlers.DeleteUserParams`):**
    ```json
    {
      "id": 12 // Required
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" \
         -d '{
              "method": "user.delete",
              "params": {
                "id": 12
              },
              "id": "user-delete-005"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "success": true,
      "id": "user-delete-005"
    }
    ```

</rewritten_file> 