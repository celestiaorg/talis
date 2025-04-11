# Add Support for Instance Payload Execution

## Overview
Add the ability to send a payload (bash script) to instances during creation or as an update. The payload can either be executed immediately or copied to the instance for later execution. Payloads will be implemented using the RPC model and task executor pattern that is being adopted for the platform.

## Current Behavior
- Instances are created through the API with basic configuration (size, region, etc.)
- Ansible is used for base provisioning
- The system is transitioning to an RPC model with a task executor
- No support for custom payload execution

## Proposed Behavior
- Users can specify a local file path to a bash script payload in instance creation or update requests
- Users can specify if the payload should be executed immediately or just copied
- Payload execution has configurable timeouts (system default with optional override)
- Same payload can be deployed to multiple instances in a single request
- Payload size will be limited for reliability
- Payload operations will be executed asynchronously via the task executor

## Architectural Assumptions
> Note: These assumptions should be revisited once the RPC model and task executor implementation details are finalized.

1. **Task Executor Pattern**
   - A centralized task executor will poll the task database for pending tasks
   - Tasks have types, statuses, and dependencies (e.g., waiting for an instance to be ready)
   - Payload operations will be a new task type within this framework

2. **RPC Model Integration**
   - Payload operations will be exposed through the RPC-based API
   - Payload endpoints will follow the same patterns as other RPC endpoints
   - Existing authentication and authorization mechanisms will apply to payload operations

## Implementation Approach

1. **Data Model Updates**
   - Create a new Payload model that references instances
   - Update Instance model to include references to payloads
   - Create a new PayloadTask type in the task system
   - Create migration for database changes

2. **File Validation Layer**
   - Implement payload file validation (existence, size limit, permissions)
   - Add path security validation for remote_path
   - Extend validation logic for payload-related requests

3. **Task Executor Integration**
   - Implement payload task handler in the task executor
   - Add logic to check instance readiness before payload operations
   - Implement payload transfer and execution logic
   - Update task statuses based on operation progress

4. **Error Handling**
   - Create specific error types for payload failures
   - Implement cleanup for failed payload execution
   - Add detailed logging for payload operations

5. **RPC API Updates**
   - Add payload-related RPC methods to the Instance API
   - Support bulk payload operations to multiple instances
   - Implement status queries for payload operations

6. **Testing**
   - Add unit tests for validation logic
   - Create integration tests for payload operations
   - Extend API tests to cover payload scenarios

7. **Documentation**
   - Update API docs with new RPC methods
   - Add examples for payload usage
   - Document task dependencies and execution model

## Technical Details

### Data Model Changes

#### 1. Create New Payload Model
```go
// models/payload.go
type PayloadStatus string

const (
    PayloadStatusPending     PayloadStatus = "pending"
    PayloadStatusTransferred PayloadStatus = "transferred"
    PayloadStatusExecuting   PayloadStatus = "executing"
    PayloadStatusCompleted   PayloadStatus = "completed"
    PayloadStatusFailed      PayloadStatus = "failed"
)

type Payload struct {
    gorm.Model
    InstanceID    uint          `json:"instance_id" gorm:"not null;index"`
    RemotePath    string        `json:"remote_path" gorm:"not null"`
    Status        PayloadStatus `json:"status" gorm:"not null;default:'pending'"`
    ExecutionTime time.Duration `json:"execution_time,omitempty"`
    Error         string        `json:"error,omitempty"`
    Executed      bool          `json:"executed" gorm:"not null;default:false"`
    Output        string        `json:"output,omitempty" gorm:"type:text"`
    TaskID        uint          `json:"task_id" gorm:"index"` // Reference to the task handling this payload
}
```

#### 2. Update Task Model for Payload Tasks
```go
// Assuming current task model exists
// Add a new task type for payload operations
const (
    // ... existing task types ...
    TaskTypePayloadTransfer TaskType = "payload_transfer"
    TaskTypePayloadExecute  TaskType = "payload_execute"
)

// Task parameters specific to payload operations
type PayloadTaskParams struct {
    PayloadID      uint   `json:"payload_id"`
    InstanceID     uint   `json:"instance_id"`
    OriginalPath   string `json:"original_path"` // Local path on the server
    RemotePath     string `json:"remote_path"`   // Path on the instance
    ExecutePayload bool   `json:"execute_payload"`
    Timeout        int    `json:"timeout,omitempty"` // In seconds
}
```

#### 3. RPC Method Definitions
```go
// Define new RPC methods for payload operations
const (
    // ... existing RPC methods ...
    MethodDeployPayload       = "deploy_payload"
    MethodGetPayloadStatus    = "get_payload_status"
    MethodListInstancePayloads = "list_instance_payloads"
)

// Request for deploying a payload
type DeployPayloadRequest struct {
    InstanceIDs    []uint  `json:"instance_ids"`    // Instances to deploy to
    PayloadPath    string  `json:"payload_path"`    // Local path on the server
    RemotePath     string  `json:"remote_path"`     // Path on the instance
    ExecutePayload bool    `json:"execute_payload"` // Whether to execute after transfer
    Timeout        int     `json:"timeout,omitempty"` // In seconds
}

// Response for deploy payload operation
type DeployPayloadResponse struct {
    Success   bool       `json:"success"`
    PayloadIDs []uint    `json:"payload_ids"` // IDs of created payload records
    TaskIDs   []uint     `json:"task_ids"`    // IDs of created tasks
    Errors    []string   `json:"errors,omitempty"`
}
```

### Task Executor Implementation

1. **Payload Task Handling**
   - Task executor will poll for payload tasks
   - For each payload task:
     - Check if the target instance is ready
     - If not ready, reschedule the task for later
     - If ready, perform the payload operation (transfer/execute)
     - Update the payload status and task status accordingly

2. **Payload Transfer Logic**
   - Use Ansible (preferred) or direct SSH for file transfer
   - Validate connection to instance
   - Transfer file to the specified remote path
   - Update payload status

3. **Payload Execution Logic**
   - Execute the script on the instance if execution is requested
   - Apply timeout to prevent indefinite execution
   - Capture output and error messages
   - Update payload status with results

### API Flow

1. **Deployment Flow**
   - Client makes RPC call to `deploy_payload`
   - API validates request (file existence, size, etc.)
   - API creates Payload records in the database
   - API creates Task records for transfer/execution
   - API returns success with payload and task IDs
   - Task executor runs the tasks asynchronously

2. **Status Query Flow**
   - Client makes RPC call to `get_payload_status`
   - API looks up payload record(s)
   - API returns current status, output, and any errors

### Example Usage

#### Deploying a payload to existing instances
```json
// RPC request
{
  "jsonrpc": "2.0",
  "method": "deploy_payload",
  "params": {
    "instance_ids": [1, 2, 3],
    "payload_path": "/local/path/to/setup.sh",
    "remote_path": "/root/setup.sh",
    "execute_payload": true,
    "timeout": 300
  },
  "id": 1
}

// RPC response
{
  "jsonrpc": "2.0",
  "result": {
    "success": true,
    "payload_ids": [101, 102, 103],
    "task_ids": [201, 202, 203]
  },
  "id": 1
}
```

#### Checking payload status
```json
// RPC request
{
  "jsonrpc": "2.0",
  "method": "get_payload_status",
  "params": {
    "payload_id": 101
  },
  "id": 2
}

// RPC response
{
  "jsonrpc": "2.0",
  "result": {
    "id": 101,
    "instance_id": 1,
    "remote_path": "/root/setup.sh",
    "status": "completed",
    "execution_time": "5.321s",
    "executed": true,
    "output": "Setup completed successfully",
    "task_id": 201,
    "created_at": "2023-07-15T10:30:00Z",
    "updated_at": "2023-07-15T10:30:15Z"
  },
  "id": 2
}
```

## Testing Requirements

1. **Unit Tests**
   - Test payload file validation
   - Test size limits
   - Test path validation
   - Test RPC request validation
   - Test task creation logic

2. **Integration Tests**
   - Test task executor with payload tasks
   - Test payload transfer
   - Test payload execution
   - Test bulk operations
   - Test error scenarios
   - Test dependencies (waiting for instance readiness)

3. **API Tests**
   - Test RPC methods
   - Test bulk operations
   - Test invalid payload scenarios
   - Test size limit enforcement
   - Test error responses

## Documentation Updates Required

1. **API Documentation**
   - Document new RPC methods
   - Add examples for payload usage
   - Document size limits and restrictions
   - Document timeout behavior
   - Document task dependencies and execution model

2. **Error Documentation**
   - Document error response formats
   - Document payload-related error handling
   - Document task failure scenarios

## Acceptance Criteria

- [ ] Local payload file path can be specified in RPC requests
- [ ] Payload size is validated and limited
- [ ] Payload can be configured for immediate or delayed execution
- [ ] Payloads can be deployed to multiple instances in a single request
- [ ] Payload execution has configurable timeouts
- [ ] Task executor handles payload operations asynchronously
- [ ] Proper error handling and status updates are implemented
- [ ] File access errors are properly handled and reported
- [ ] All tests pass and coverage is maintained
- [ ] Documentation is updated and accurate

## Out of Scope

- Multi-file payload support
- Payload encryption
- Advanced execution environments
- Non-bash script support
- Payload scheduling for future execution
- Payload versioning

## Implementation Impact

- Integration with the new RPC model and task executor
- New task types and handlers in the task executor
- New RPC methods for the API
- Moderate database changes (new models + relationships)
- New file system access requirements for API service

## Questions/Concerns

- System default timeout for payload execution should be 5 minutes
- No basic payload validation needed (e.g., shebang line check)
- No default environment variables will be provided to the payload
- Task dependencies will use the task executor's dependency model (TBD)
- File permission issues on the API server are out of scope
