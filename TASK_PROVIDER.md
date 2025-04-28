# VirtFusion Provider Integration Task

## Overview
Implement a provider integration for VirtFusion hypervisor management platform. This provider will allow creating and managing virtual machines through VirtFusion's API.

## Requirements

### Environment Configuration
- [x] Add the following environment variables:
  ```
  VIRTFUSION_API_TOKEN=your_api_token
  VIRTFUSION_HOST=your_virtfusion_host  # e.g., https://api.virtfusion.com
  ```

### API Integration Requirements

#### Core Functionality
1. Connection Testing
   - Implement connection test using `/connect` endpoint
   - Validate API credentials and host availability

2. Hypervisor Management
   - List available hypervisors (`GET /compute/hypervisors`)
   - Get specific hypervisor details (`GET /compute/hypervisors/{hypervisorId}`)

3. Server Operations
   - Create servers (`POST /servers`)
   - List servers (`GET /servers`)
   - Get server details (`GET /servers/{serverId}`)
   - Delete servers (`DELETE /servers/{serverId}`)
   - Build server after creation (`POST /servers/{serverId}/build`)
   - Suspend/Unsuspend servers
     - `POST /servers/{serverId}/suspend`
     - `POST /servers/{serverId}/unsuspend`

### API Endpoints and Request Details

#### Connection Testing
```http
GET /connect
Headers:
  - Authorization: Bearer {VIRTFUSION_API_TOKEN}
  - Content-Type: application/json
```

#### Hypervisor Management

1. List Hypervisors
```http
GET /compute/hypervisors
Headers:
  - Authorization: Bearer {VIRTFUSION_API_TOKEN}
  - Content-Type: application/json
Response:
{
  "hypervisors": [
    {
      "id": 1,
      "name": "hypervisor-1",
      "status": "active",
      "ipAddress": "1.2.3.4",
      "location": "nyc",
      "totalMemory": 32768,
      "usedMemory": 16384,
      "totalDisk": 1000,
      "usedDisk": 500
    }
  ]
}
```

2. Get Hypervisor Details
```http
GET /compute/hypervisors/{hypervisorId}
Headers:
  - Authorization: Bearer {VIRTFUSION_API_TOKEN}
  - Content-Type: application/json
```

#### Server Management

1. Create Server
```http
POST /servers
Headers:
  - Authorization: Bearer {VIRTFUSION_API_TOKEN}
  - Content-Type: application/json
Request Body:
{
  "name": "server-name",
  "memory": 2048,
  "disk": 50,
  "cpu": 2,
  "template": "ubuntu-22.04",
  "hypervisorId": 1,
  "sshKeys": ["ssh-rsa AAAA..."],
  "userData": "#!/bin/bash\n..."
}
```

2. Build Server
```http
POST /servers/{serverId}/build
Headers:
  - Authorization: Bearer {VIRTFUSION_API_TOKEN}
  - Content-Type: application/json
Request Body:
{
  "template": "ubuntu-22.04",
  "sshKeys": ["ssh-rsa AAAA..."],
  "userData": "#!/bin/bash\n..."
}
```

3. Get Server Details
```http
GET /servers/{serverId}
Headers:
  - Authorization: Bearer {VIRTFUSION_API_TOKEN}
  - Content-Type: application/json
```

4. Delete Server
```http
DELETE /servers/{serverId}
Headers:
  - Authorization: Bearer {VIRTFUSION_API_TOKEN}
```

5. Suspend/Unsuspend Server
```http
POST /servers/{serverId}/suspend
POST /servers/{serverId}/unsuspend
Headers:
  - Authorization: Bearer {VIRTFUSION_API_TOKEN}
```

### Request Format

The VirtFusion provider will use the following request format:

```json
{
    "project_name": "project-name",
    "instances": [
        {
            "name": "instance-name",
            "provider": "virtfusion",
            "number_of_instances": 1,
            "provision": false,
            "hypervisor_id": 1,
            "resources": {
                "memory": 1024,           // MB
                "disk": 25,               // GB
                "cpu": 1                  // Cores
            },
            "template": "ubuntu-22.04",
            "ssh_key_name": "key-name",
            "network": {
                "bandwidth": 1000,        // Mbps
                "public_ip": true
            },
            "volumes": [
                {
                    "name": "volume-name",
                    "size_gb": 20,
                    "mount_point": "/mnt/data"
                }
            ]
        }
    ]
}
```

### Error Handling
- Implement proper error handling for all API calls
- Create specific error types for VirtFusion-related errors
- Handle rate limiting and API timeouts
- Implement retry mechanisms for transient failures

### Testing
1. Unit Tests
   - Test provider class methods
   - Test request/response mapping
   - Test error handling

2. Integration Tests
   - Test actual API interactions
   - Test instance lifecycle (create, get, delete)
   - Test error scenarios

## Acceptance Criteria
1. Successfully create instances on VirtFusion using the specified request format
2. Proper error handling and reporting
3. Complete test coverage
4. Documentation for VirtFusion features and limitations
5. Environment configuration guide
6. Working examples in the documentation

## Future Enhancements
1. Support for backup management (`PUT /servers/{serverId}/backups/plan/{planId}`)
2. Support for owner changes (`PUT /servers/{serverId}/owner/{newOwnerId}`)
3. Support for VNC management (`POST /servers/{serverId}/vnc`)
4. Traffic statistics integration (`GET /servers/{serverId}/traffic`)
5. CPU throttling capabilities (`PUT /servers/{serverId}/modify/cpuThrottle`)
6. Add support for custom network configurations
7. Implement automatic failover between hypervisors
8. Add support for volume snapshots and backups
9. Implement resource usage monitoring and alerts
10. Add support for custom firewall rules

## Completed Tasks
1. ✅ Basic provider setup and configuration
2. ✅ Core API integration with VirtFusion
3. ✅ Instance management implementation
4. ✅ Testing suite implementation
5. ✅ Documentation and examples
6. ✅ Environment configuration
7. ✅ Error handling and recovery mechanisms
8. ✅ SSH key management
9. ✅ Volume management
10. ✅ Instance provisioning workflow

## Next Steps
1. Deploy to staging environment for final testing
2. Conduct performance testing under load
3. Set up monitoring and alerting
4. Create user documentation
5. Plan rollout strategy

## Dependencies
- Access to VirtFusion API
- API documentation
- Test environment credentials

## Implementation Tasks

### 1. Basic Setup
- [x] Update provider constants in `internal/db/models/provider.go`
- [x] Create VirtFusion types in `internal/compute/types.go`
- [x] Create configuration struct for VirtFusion in `internal/config/virtfusion.go`
- [x] Add environment variable validation for VirtFusion credentials
- [ ] Verification:
  - [ ] Run `make test` to verify config tests pass
  - [ ] Run `make run` to verify environment variables are properly loaded

### 2. Core Provider Implementation
- [x] Update `internal/compute/provider.go`:
  - [x] Implement `VirtFusionProvider` struct
  - [x] Implement HTTP client with authentication
  - [x] Implement base API request methods
  - [x] Add retry mechanisms and error handling
- [x] Implement Core Services:
  - [x] Implement `HypervisorService`
  - [x] Implement `ServerService`
  - [x] Add response parsing and error handling
  - [x] Add logging for operations
- [ ] Verification:
  - [ ] Run `make test` to verify all service tests pass
  - [ ] Run `make run` to verify client initialization and connection test

### 3. Instance Management
- [x] Implement instance creation flow:
  - [x] Add hypervisor selection logic
  - [x] Add resource mapping
  - [x] Add template mapping
  - [x] Implement volume handling
  - [x] Add SSH key management
- [x] Implement instance lifecycle:
  - [x] Add status monitoring
  - [x] Add build process handling
  - [x] Add cleanup procedures
  - [x] Add error recovery
- [x] Verification:
  - [x] Run `make test` to verify instance management tests
  - [x] Run `make run` to test full instance lifecycle with real API

### 4. Testing
- [x] Create mock client for testing:
  - [x] Add mock responses
  - [x] Add error scenarios
  - [x] Add state management
- [x] Add unit tests:
  - [x] Test provider methods
  - [x] Test service methods
  - [x] Test error handling
  - [x] Test request/response mapping
- [x] Add integration tests:
  - [x] Test full instance lifecycle
  - [x] Test error recovery
  - [x] Test concurrent operations
- [x] Verification:
  - [x] Run `make test` to verify all test suites pass
  - [x] Run test coverage report and ensure >80% coverage
  - [x] Run integration tests in CI environment

### 5. Documentation
- [x] Add code documentation
- [x] Update README with VirtFusion setup
- [x] Add example configurations
- [x] Add troubleshooting guide
- [x] Verification:
  - [x] Run `go doc` to verify documentation is properly formatted
  - [x] Test example configurations with `make run`
  - [x] Verify all setup steps in a clean environment

## Network Profile Issues and Testing

### Current Issues
1. Network Profile Error
   - Error: "Invalid network profile. The specified profile is not in use"
   - Root Cause: The network profile ID in the request doesn't match the hypervisor's network configuration
   - Current Implementation: Using hypervisor's primary network ID, but still failing

2. Request Format Issues
   - Need to support network profile specification in the request JSON
   - Current format doesn't include network profile configuration

### Proposed Solutions

1. Update Request Format
```json
{
    "project_name": "project-name",
    "instances": [
        {
            "name": "instance-name",
            "provider": "virtfusion",
            "number_of_instances": 1,
            "provision": false,
            "hypervisor_id": 1,
            "resources": {
                "memory": 1024,
                "disk": 25,
                "cpu": 1
            },
            "network": {
                "profile_id": 2,        // Added: Specific network profile ID
                "bandwidth": 1000,
                "public_ip": true
            },
            "template": "ubuntu-22.04",
            "ssh_key_name": "key-name",
            "volumes": [
                {
                    "name": "volume-name",
                    "size_gb": 20,
                    "mount_point": "/mnt/data"
                }
            ]
        }
    ]
}
```

2. Network Profile Selection Logic
   - First try: Use network profile from request if specified
   - Second try: Use hypervisor's primary network if available
   - Third try: Use hypervisor's default network
   - Fail if no valid network profile is found

### Testing Procedure

1. Start Local Environment
```bash
make run
```

2. Test with CLI
```bash
# Basic test with default network
./bin/talis-cli infra create -f create.json

# Test with specific network profile
./bin/talis-cli infra create -f create-with-network.json
```

3. Verification Steps
   - Check server creation response
   - Verify network configuration in created server
   - Test connectivity once server is active
   - Verify network bandwidth matches request

4. Error Cases to Test
   - Invalid network profile ID
   - Missing network configuration
   - Network profile not available on selected hypervisor

### Implementation Tasks

1. [ ] Update types and request validation
   - [ ] Add network profile fields to request structs
   - [ ] Update validation logic
   - [ ] Add network profile mapping

2. [ ] Enhance Provider Implementation
   - [ ] Update network profile selection logic
   - [ ] Add network validation against hypervisor
   - [ ] Improve error messages for network issues

3. [ ] Testing
   - [ ] Add unit tests for network profile selection
   - [ ] Add integration tests with various network configs
   - [ ] Document test cases and expected results

4. [ ] Documentation
   - [ ] Update API documentation with network fields
   - [ ] Add network configuration examples
   - [ ] Document error scenarios and solutions

### Example Test Files

1. `create.json` (Basic test)
```json
{
    "project_name": "talis-test",
    "instances": [
        {
            "name": "test-server",
            "provider": "virtfusion",
            "hypervisor_id": 2,
            "hypervisor_group": "Prague",
            "resources": {
                "memory": 1024,
                "disk": 25,
                "cpu": 1
            },
            "network": {
                "profile_id": 2,
                "bandwidth": 1000
            },
            "template": "ubuntu-22.04",
            "ssh_key_name": "default"
        }
    ]
}
```

2. `create-with-network.json` (Full network config)
```json
{
    "project_name": "talis-test",
    "instances": [
        {
            "name": "test-server",
            "provider": "virtfusion",
            "hypervisor_id": 2,
            "hypervisor_group": "Prague",
            "resources": {
                "memory": 1024,
                "disk": 25,
                "cpu": 1
            },
            "network": {
                "profile_id": 2,
                "bandwidth": 1000,
                "public_ip": true,
                "firewall_rulesets": [2, 3]
            },
            "template": "ubuntu-22.04",
            "ssh_key_name": "default"
        }
    ]
}
```

### Verification Checklist
- [ ] Server creation succeeds with default network
- [ ] Server creation succeeds with specific network profile
- [ ] Network bandwidth is correctly applied
- [ ] Firewall rules are properly configured
- [ ] Public IP is assigned when requested
- [ ] Network errors are properly handled and reported
- [ ] Network configuration persists after server reboot
