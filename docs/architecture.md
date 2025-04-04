# Talis Architecture

## Overview

Talis is a service designed to manage and orchestrate distributed systems deployments, with a focus on blockchain networks and testing environments. This document outlines the high-level architecture of Talis, describing its core components, data models, API design, and internal service interactions.

## Core Concepts

Talis is built around four primary entities that work together to provide a comprehensive deployment and management solution:

1. **Owners**: The users of the Talis system who initiate and manage deployments
2. **Projects**: Scopes of work that represent a specific deployment goal (e.g., test environment, long-running chain, testnet)
3. **Tasks**: Discrete actions that Talis executes to achieve the desired state
4. **Instances**: The actual deployed resources (servers, nodes, etc.) managed by Talis

## Data Models

### Owner
- Represents a user of the Talis system
- Manages projects and has permissions to perform actions
- Can create and monitor multiple projects

### Project
- Represents a scope of work (test, chain, testnet)
- Contains configuration for the desired deployment
- Associated with multiple instances and tasks
- Owned by a specific owner

### Task
- Represents a discrete action to be executed
- Contains all necessary information for execution
- Has a clear lifecycle (pending, running, completed, failed)
- Associated with a project and potentially specific instances

### Instance
- Represents a deployed resource
- Contains metadata about the deployment (IP, status, configuration)
- Associated with a specific project
- May have multiple tasks associated with it

## API Design

The API is designed with SDK users in mind, following these principles:

1. **Request-Body Focused**
   - Prioritizes POST requests with structured request bodies
   - Enables strong type validation and clear contract definition
   - Makes SDK integration straightforward and maintainable

2. **Synchronous Database Operations**
   - API calls block until database operations complete
   - Successful response guarantees successful database update
   - Provides clear consistency guarantees to clients

3. **Asynchronous Task Execution**
   - Long-running operations are handled asynchronously
   - Tasks are created and tracked in the database
   - Clients can poll for task status and completion

## Data Flow

### Request Processing
1. User submits request through API
2. Request is validated
3. Database is updated synchronously
   - Project/Instance/Task records created/updated
   - All related entities are updated in a transaction
4. API returns success response
5. Asynchronous processing begins

### Task Execution
1. TaskExecutor routine runs as a background process
2. Continuously monitors Task database for new entries
3. Picks up pending tasks for execution
4. Updates task status during execution
5. Marks tasks as completed or failed
6. Updates related entities with results

### Example Flow: Project Creation with Instances
1. User submits project creation request with 100 instances
2. API validates request
3. Database transaction begins
   - Project record created
   - Instance records created
   - Associated tasks created
4. Transaction commits
5. API returns success
6. TaskExecutor begins processing instance creation tasks
7. Instances are created asynchronously
8. Task and instance statuses are updated as work completes

## Implementation Status

Note: This architecture represents the target design of Talis. The actual implementation is a work in progress, and some components may be partially implemented or planned for future development.
