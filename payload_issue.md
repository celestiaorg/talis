# Add Support for Instance Payload Execution

## Overview
Add the ability to send a payload (bash script) to instances during creation or as an update. The payload can either be executed immediately as part of instance creation or copied to the instance for later execution.

## Current Behavior
- Instances are created through the API with basic configuration (size, region, etc.)
- Ansible is used for base provisioning
- No support for custom payload execution

## Proposed Behavior
- Users can specify a local file path to a bash script payload in instance creation requests
- Users can specify if the payload should be executed immediately or just copied
- Instance creation with immediate execution will only complete after payload execution
- Payload size will be limited for reliability

## Technical Details

### API Changes

#### 1. Update Instance Request Model
Add to `types.InstanceRequest`:
```go
type InstanceRequest struct {
    // ... existing fields ...
    PayloadPath      string `json:"payload_path"`      // Local path to the bash script
    ExecutePayload   bool   `json:"execute_payload"`   // Whether to execute immediately
    RemotePath       string `json:"remote_path"`       // Path where payload should be saved on instance
}
```

#### 2. Database Changes
Add to `models.Instance`:
```go
type Instance struct {
    // ... existing fields ...
    PayloadStatus    string `json:"payload_status"`    // Status of payload execution if applicable
}
```

### Implementation Requirements

1. **Payload Validation**
   - Add payload size validation (suggest 1MB limit)
   - Validate file exists at provided path
   - Validate remote_path (prevent security issues with path)
   - Validate file is a readable text file

2. **Infrastructure Service Changes**
   - Extend `Infrastructure.Execute()` to handle payload operations
   - Add payload transfer functionality using SSH
   - Implement payload execution if `execute_payload` is true
   - Add proper error handling for payload operations

3. **Instance Status Updates**
   - Add payload-related status updates
   - Handle failed payload execution scenarios
   - Update instance status only after payload execution if specified

4. **Error Handling**
   - Add specific error types for payload-related failures
   - Implement proper cleanup if payload execution fails
   - Add logging for payload operations

### Example Usage

```json
{
  "instances": [{
    "provider": "do",
    "region": "nyc1",
    "size": "s-1vcpu-1gb",
    "name": "test-instance",
    "payload_path": "/local/path/to/setup.sh",
    "execute_payload": true,
    "remote_path": "/root/setup.sh"
  }]
}
```

## Testing Requirements

1. **Unit Tests**
   - Test payload file validation
   - Test size limits
   - Test path validation
   - Test status updates

2. **Integration Tests**
   - Test payload transfer
   - Test immediate execution
   - Test delayed execution
   - Test error scenarios
   - Test with various file sizes
   - Test with invalid file paths

3. **API Tests**
   - Test payload in instance creation
   - Test invalid payload scenarios
   - Test size limit enforcement
   - Test file access errors

## Documentation Updates Required

1. **API Documentation**
   - Document new payload-related fields
   - Add examples for payload usage
   - Document size limits and restrictions
   - Document file path requirements

2. **Error Documentation**
   - Document new error types
   - Document payload-related error handling
   - Document file access errors

## Acceptance Criteria

- [ ] Local payload file path can be specified in instance creation request
- [ ] Payload size is validated and limited
- [ ] Payload can be configured for immediate or delayed execution
- [ ] Instance creation with immediate execution waits for payload completion
- [ ] Proper error handling and status updates are implemented
- [ ] File access errors are properly handled and reported
- [ ] All tests pass and coverage is maintained
- [ ] Documentation is updated and accurate

## Out of Scope

- Multi-file payload support
- Payload encryption
- Payload storage/history
- Advanced execution environments
- Non-bash script support

## Implementation Impact

- Low impact on existing instance creation flow
- Minimal database changes required
- No changes to existing instance management endpoints
- New file system access requirements for API service

## Questions/Concerns

- Should there be a timeout for payload execution?
- Should we add basic payload validation (e.g., shebang line check)?
- Should we provide any default environment variables to the payload?
- How should we handle file permission issues on the API server? 