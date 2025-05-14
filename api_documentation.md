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
*   **Authentication:** Required. Pass the API key in the `apikey` header.
*   **Request Body:** None
*   **Query Parameters:** None
*   **Example Request:**
    ```bash
    curl -H "apikey: YOUR_API_KEY" http://localhost:8080/health
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
*   **Authentication:** Required. Pass the API key in the `apikey` header.
*   **Request Body:** None
*   **Query Parameters:**
    *   `limit` (int, optional, default: `DefaultPageSize` from `handlers`): Number of instances to return.
    *   `offset` (int, optional, default: 0): Offset for pagination.
    *   `include_deleted` (bool, optional, default: `false`): Whether to include deleted instances.
    *   `status` (string, optional): Filter instances by status. Valid values: `unknown`, `pending`, `created`, `provisioning`, `ready`, `terminated`.
*   **Example Request:**
    ```bash
    curl -H "apikey: YOUR_API_KEY" "http://localhost:8080/api/v1/admin/instances?limit=10&status=ready"
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
*   **Authentication:** Required. Pass the API key in the `apikey` header.
*   **Request Body:** None
*   **Query Parameters:** Same as [List All Instances (Admin)](#list-all-instances-admin).
*   **Example Request:**
    ```bash
    curl -H "apikey: YOUR_API_KEY" "http://localhost:8080/api/v1/admin/instances/all-metadata?include_deleted=true&status=pending"
    ```
*   **Example Response (200 OK):** Same as [List All Instances (Admin)](#list-all-instances-admin).

---

## Instance Endpoints

### List Instances

*   **Endpoint:** `GET /api/v1/instances`
*   **Route Name:** `GetInstances`
*   **Handler:** `instanceHandler.ListInstances` (Note: The handler implies AdminID is used; this might be subject to OwnerID filtering in a real scenario based on TODOs in code).
*   **Description:** Retrieves a list of instances. (Assumed to be filterable by OwnerID in the future).
*   **Authentication:** Required. Pass the API key in the `apikey` header.
*   **Request Body:** None
*   **Query Parameters:** Same as [List All Instances (Admin)](#list-all-instances-admin).
*   **Example Request:**
    ```bash
    curl -H "apikey: YOUR_API_KEY" "http://localhost:8080/api/v1/instances?limit=5&status=created"
    ```
*   **Example Response (200 OK):** Similar structure to [List All Instances (Admin)](#list-all-instances-admin).

### Get All Instances Metadata

*   **Endpoint:** `GET /api/v1/instances/all-metadata`
*   **Route Name:** `GetMetadata`
*   **Handler:** `instanceHandler.GetAllMetadata`
*   **Description:** Retrieves metadata for instances. (Assumed to be filterable by OwnerID in the future).
*   **Authentication:** Required. Pass the API key in the `apikey` header.
*   **Request Body:** None
*   **Query Parameters:** Same as [List All Instances (Admin)](#list-all-instances-admin).
*   **Example Request:**
    ```bash
    curl -H "apikey: YOUR_API_KEY" "http://localhost:8080/api/v1/instances/all-metadata?status=terminated"
    ```
*   **Example Response (200 OK):** Similar structure to [List All Instances (Admin)](#list-all-instances-admin).

### Get Public IPs of Instances

*   **Endpoint:** `GET /api/v1/instances/public-ips`
*   **Route Name:** `GetPublicIPs`
*   **Handler:** `instanceHandler.GetPublicIPs`
*   **Description:** Retrieves a list of public IP addresses for instances.
*   **Authentication:** Required. Pass the API key in the `apikey` header.
*   **Request Body:** None
*   **Query Parameters:**
    *   `limit` (int, optional, default: `DefaultPageSize`): Number of items to return.
    *   `offset` (int, optional, default: 0): Offset for pagination.
    *   `include_deleted` (bool, optional, default: `false`): Whether to include IPs from deleted instances.
*   **Example Request:**
    ```bash
    curl -H "apikey: YOUR_API_KEY" "http://localhost:8080/api/v1/instances/public-ips"
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
*   **Description:** Creates one or more new instances based on the provided configurations. Instance names are auto-generated by the system.
*   **Authentication:** Required. Pass the API key in the `apikey` header.
*   **Request Body:** An array of `types.InstanceRequest` objects. The `name` field for the instance itself should be omitted as it is auto-generated.
    ```json
    // types.InstanceRequest structure (example)
    {
      // "name": "my-instance-01", // OMIT - Instance name is auto-generated
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
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
         -d '[
              {
                // "name": "batch-worker-01", // OMIT - Instance name is auto-generated
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
    {
      "slug": "success",
      "error": "",
      "data": [
        {
          "ID": 10,
          "owner_id": 1,
          "project_id": 5,
          // "name": "autogenerated-name-xyz", // Name is auto-generated and might not be in response, or its field name could vary.
          // ... other fields from models.Instance, excluding user-defined name
          "status": "pending"
        },
        {
          "ID": 11,
          "owner_id": 1,
          "project_id": 5,
          // ... other fields from models.Instance
          "status": "pending"
        }
      ]
    }
    ```

### Get Instance Details

*   **Endpoint:** `GET /api/v1/instances/:id`
*   **Route Name:** `GetInstance`
*   **Handler:** `instanceHandler.GetInstance`
*   **Description:** Retrieves details for a specific instance by its ID.
*   **Authentication:** Required. Pass the API key in the `apikey` header.
*   **Request Body:** None
*   **Path Parameters:**
    *   `id` (int, required): The ID of the instance.
*   **Example Request:**
    ```bash
    curl -H "apikey: YOUR_API_KEY" http://localhost:8080/api/v1/instances/10
    ```
*   **Example Response (200 OK):**
    ```json
    // models.Instance structure
    {
      "id": 10,
      "owner_id": 1,
      "project_name": "batch-processing", // Note: In current DB model this is project_id (int)
      "provider_instance_id": "prov-inst-id-10",
      "provider": "do",
      "name": "instance-batch-01-0", // Name may include a numeric suffix (e.g., -0, -1)
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

*   **Endpoint**: `DELETE /api/v1/instances`
*   **Method**: `DELETE`
*   **Description**: Terminates one or more specified instances within a given project.
*   **Requires API Key**: Yes
*   **Request Body**: JSON object with the following fields:
    *   `owner_id` (integer, required): The ID of the owner.
    *   `project_name` (string, required): The name of the project to which the instances belong.
    *   `instance_ids` (array of integers, required): A list of instance IDs to terminate.

*   **Example Request**:
    ```bash
    curl -X DELETE \
      http://163.172.162.109:8000/talis/api/v1/instances \
      -H "apikey: YOUR_API_KEY" \
      -H "Content-Type: application/json" \
      -d '{
            "owner_id": 1,
            "project_name": "project-id-retest",
            "instance_ids": [1]
          }'
    ```
*   **Success Response (200 OK)**:
    ```json
    {
        "message": "Tasks created successfully for terminating instances",
        "task_ids": ["xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"]
    }
    ```
    (Task IDs will be actual UUIDs)
*   **Error Responses**:
    *   `400 Bad Request`: If `project_name` or `instance_ids` are missing or invalid (e.g., "project name and instance IDs are required").
    *   `401 Unauthorized`: If the API key is missing or invalid.
    *   `404 Not Found`: If the project or any of the specified instances do not exist.
    *   `500 Internal Server Error`: If there's an issue on the server side during the termination process.
*   **Notes**:
    *   This operation is asynchronous. It creates tasks to terminate the instances. You can use the task IDs to track the status of the termination.
    *   The server now correctly expects `instance_ids` as per the type definitions.

### List Tasks for an Instance

*   **Endpoint:** `GET /api/v1/instances/:instance_id/tasks`
*   **Route Name:** `ListInstanceTasks`
*   **Handler:** `taskHandler.ListByInstanceID` (from `routes.go`)
*   **Description:** Retrieves a list of tasks associated with a specific instance.
*   **Authentication:** Required. Pass the API key in the `apikey` header.
*   **Request Body:** None
*   **Path Parameters:**
    *   `instance_id` (int, required): The ID of the instance.
*   **Query Parameters:** (Assumed standard pagination might apply, e.g., `limit`, `offset` - needs verification)
    *   `limit` (int, optional): Number of tasks to return.
    *   `offset` (int, optional): Offset for pagination.
*   **Example Request:**
    ```bash
    curl -H "apikey: YOUR_API_KEY" http://localhost:8080/api/v1/instances/10/tasks?limit=5
    ```
*   **Example Response (200 OK):** (Hypothetical - endpoint returned 404 in tests)
    ```json
    {
      "rows": [
        {
          "id": 101,
          "owner_id": 1,
          "project_id": 5,
          "instance_id": 10,
          "status": "completed",
          "action": "create_instance",
          // ... other task fields ...
          "created_at": "2023-10-26T10:00:00Z"
        },
        {
          "id": 105,
          "owner_id": 1,
          "project_id": 5,
          "instance_id": 10,
          "status": "pending",
          "action": "terminate_instance",
          // ... other task fields ...
          "created_at": "2023-10-27T11:00:00Z"
        }
      ],
      "pagination": {
        "total": 2,
        "page": 1,
        "limit": 5,
        "offset": 0
      }
    }
    ```
*   **Current Status (Testing Note):** As of the last test against `http://163.172.162.109:8000/talis/`, this endpoint returned a `404 Not Found` with `{"error":"Cannot GET /api/v1/instances/:instance_id/tasks"}`. This suggests a routing or deployment issue on the server. It's also possible this is related to the "record not found" error observed with the RPC `task.list` method if the underlying task retrieval logic has issues distinguishing "no tasks found" from "parent record (instance/project) not found".

---

## RPC Endpoint

The API provides a single RPC endpoint for various operations related to projects, tasks, and users.

*   **Endpoint:** `POST /api/v1/`
*   **Route Name:** `RPC`
*   **Handler:** `rpcHandler.HandleRPC`
*   **Description:** A general-purpose RPC endpoint. The specific operation is determined by the `method` field in the request body.
*   **Authentication:** Required. Pass the API key in the `apikey` header for all RPC calls.

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
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
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
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
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
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
*   **Params (`handlers.ProjectGetParams`):**
    ```json
    {
      "name": "my-new-project", // Required
      "owner_id": 1 // Required
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
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
*   **Example Response (Not Found - 404):**
    ```json
    {
      "error": {
        "code": 404,
        "message": "Project not found",
        "data": "record not found"
      },
      "id": "proj-get-002",
      "success": false
    }
    ```

#### `project.list`

*   **Description:** Lists projects for a given owner, with pagination.
*   **Handler:** `ProjectHandlers.List`
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
*   **Params (`handlers.ProjectListParams`):
    *   `owner_id` (uint, required)
    *   `limit` (int, optional, default: 50 in `handlers.getPaginationOptions`)
    *   `offset` (int, optional, default: 0)
    *   `page` (int, optional, default: 1) - Note: `limit` & `offset` are preferred for direct service layer calls.
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
         -d '{
              "method": "project.list",
              "params": {
                "owner_id": 1,
                "page": 1
              },
              "id": "proj-list-001"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": {
        "rows": [
          {
            "ID": 1,
            "name": "cli-test-project",
            // ... other project fields
          }
          // ... other projects
        ],
        "pagination": {
          "total": 1, // Actual total number of projects for the owner
          "page": 1,
          "limit": 50, // Note: Server might return its default limit (e.g., 50) despite request
          "offset": 0
        }
      },
      "success": true,
      "id": "proj-list-001"
    }
    ```
*   **Testing Note:** This RPC method was tested successfully. The server returned a default pagination limit different from the one requested, which is noted.

#### `project.delete`

*   **Description:** Deletes a project by name and owner.
*   **Handler:** `ProjectHandlers.Delete`
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
*   **Params (`handlers.ProjectDeleteParams`):**
    ```json
    {
      "name": "my-new-project", // Required
      "owner_id": 1 // Required
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
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
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
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
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
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

Note: Task methods allow you to interact with asynchronous operations created by the system. These are useful for monitoring the status of long-running operations like instance creation or termination.

#### `task.get`

*   **Description:** Retrieves details of a specific task by ID.
*   **Handler:** `TaskHandlers.Get`
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
*   **Params (`handlers.TaskGetParams`):**
    ```json
    {
      "id": "task-uuid-123", // Required: UUID of the task
      "owner_id": 1 // Required: Owner ID associated with the task
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
         -d '{
              "method": "task.get",
              "params": {
                "id": "550e8400-e29b-41d4-a716-446655440000",
                "owner_id": 1
              },
              "id": "task-get-001"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "owner_id": 1,
        "project_id": 5,
        "instance_id": 20,
        "status": "completed",
        "action": "create_instance",
        "message": "Instance creation completed successfully",
        "error": null,
        "created_at": "2023-10-29T15:30:00Z",
        "updated_at": "2023-10-29T15:35:12Z"
      },
      "success": true,
      "id": "task-get-001"
    }
    ```
*   **Example Response (Not Found):**
    ```json
    {
      "error": {
        "code": 404,
        "message": "Task not found",
        "data": "record not found"
      },
      "id": "task-get-001",
      "success": false
    }
    ```

#### `task.list`

*   **Description:** Lists tasks for a given project, with pagination.
*   **Handler:** `TaskHandlers.List`
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
*   **Params (`handlers.TaskListParams`):**
    ```json
    {
      "owner_id": 1, // Required: Owner ID
      "project_name": "my-project", // Required: Project name
      "limit": 10, // Optional: Number of tasks per page
      "page": 1, // Optional: Page number
      "status": "pending" // Optional: Filter by task status (e.g., "pending", "completed", "failed")
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
         -d '{
              "method": "task.list",
              "params": {
                "owner_id": 1,
                "project_name": "my-project",
                "page": 1,
                "limit": 5
              },
              "id": "task-list-001"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": {
        "rows": [
          {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "owner_id": 1,
            "project_id": 5,
            "instance_id": 20,
            "status": "completed",
            "action": "create_instance",
            "message": "Instance creation completed successfully",
            "error": null,
            "created_at": "2023-10-29T15:30:00Z",
            "updated_at": "2023-10-29T15:35:12Z"
          },
          {
            "id": "550e8400-e29b-41d4-a716-446655440001",
            "owner_id": 1,
            "project_id": 5,
            "instance_id": 21,
            "status": "pending",
            "action": "terminate_instance",
            "message": "Termination in progress",
            "error": null,
            "created_at": "2023-10-29T16:20:00Z",
            "updated_at": "2023-10-29T16:20:00Z"
          }
        ],
        "pagination": {
          "total": 2,
          "page": 1,
          "limit": 5,
          "offset": 0
        }
      },
      "success": true,
      "id": "task-list-001"
    }
    ```
*   **Example Response (Project Not Found):**
    ```json
    {
      "error": {
        "code": 404,
        "message": "Project not found",
        "data": "record not found"
      },
      "id": "task-list-001",
      "success": false
    }
    ```

#### `task.terminate`

*   **Description:** Terminates a running task by ID. This is useful for canceling long-running operations if supported.
*   **Handler:** `TaskHandlers.Terminate`
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
*   **Params (`handlers.TaskTerminateParams`):**
    ```json
    {
      "id": "task-uuid-123", // Required: UUID of the task to terminate
      "owner_id": 1 // Required: Owner ID associated with the task
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
         -d '{
              "method": "task.terminate",
              "params": {
                "id": "550e8400-e29b-41d4-a716-446655440001",
                "owner_id": 1
              },
              "id": "task-term-001"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "status": "terminated",
        "message": "Task terminated by user request"
      },
      "success": true,
      "id": "task-term-001"
    }
    ```
*   **Example Response (Task Not Found):**
    ```json
    {
      "error": {
        "code": 404,
        "message": "Task not found",
        "data": "record not found"
      },
      "id": "task-term-001",
      "success": false
    }
    ```
*   **Example Response (Cannot Terminate):**
    ```json
    {
      "error": {
        "code": 400,
        "message": "Cannot terminate task",
        "data": "Task is already completed or terminated"
      },
      "id": "task-term-001",
      "success": false
    }
    ```

### User Methods

Dispatched by `rpcHandler.handleUserMethod` to `UserHandlers`.

#### `user.create`

*   **Description:** Creates a new user.
*   **Handler:** `UserHandlers.Create`
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
*   **Params (`handlers.UserCreateParams`):**
    ```json
    {
      "name": "new-user-alpha", // Required
      "email": "user@example.com", // Required
      "password": "password123", // Required
      "role": "admin" // Optional: Default is "user"
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
         -d '{
              "method": "user.create",
              "params": {
                "name": "new-user-alpha",
                "email": "user@example.com",
                "password": "password123"
              },
              "id": "user-create-001"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": {
        "id": 100,
        "name": "new-user-alpha",
        "email": "user@example.com",
        "role": "admin"
      },
      "success": true,
      "id": "user-create-001"
    }
    ```

#### `user.get`

*   **Description:** Retrieves details of a specific user by ID.
*   **Handler:** `UserHandlers.Get`
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
*   **Params (`handlers.UserGetParams`):**
    ```json
    {
      "id": 100 // Required: User ID
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
         -d '{
              "method": "user.get",
              "params": {
                "id": 100
              },
              "id": "user-get-001"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": {
        "id": 100,
        "name": "new-user-alpha",
        "email": "user@example.com",
        "role": "admin"
      },
      "success": true,
      "id": "user-get-001"
    }
    ```

#### `user.get.id`

*   **Description:** Retrieves the ID of a specific user by name.
*   **Handler:** `UserHandlers.GetID`
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
*   **Params (`handlers.UserGetIDParams`):**
    ```json
    {
      "name": "new-user-alpha" // Required: User name
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
         -d '{
              "method": "user.get.id",
              "params": {
                "name": "new-user-alpha"
              },
              "id": "user-getid-001"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "data": {
        "id": 100
      },
      "success": true,
      "id": "user-getid-001"
    }
    ```

#### `user.delete`

*   **Description:** Deletes a user by ID.
*   **Handler:** `UserHandlers.Delete`
*   **Authentication (for RPC endpoint):** Required. Pass the API key in the `apikey` header.
*   **Params (`handlers.UserDeleteParams`):**
    ```json
    {
      "id": 100 // Required: User ID
    }
    ```
*   **Example Request:**
    ```bash
    curl -X POST -H "Content-Type: application/json" -H "apikey: YOUR_API_KEY" \
         -d '{
              "method": "user.delete",
              "params": {
                "id": 100
              },
              "id": "user-delete-001"
            }' \
         http://localhost:8080/api/v1/
    ```
*   **Example Response (Success):**
    ```json
    {
      "success": true,
      "id": "user-delete-001"
    }
    ```