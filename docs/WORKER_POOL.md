# Worker Pool Implementation

This document explains the worker pool implementation for parallel task processing in Talis.

## Overview

The worker pool implementation allows Talis to process tasks (instance creation and termination) in parallel rather than sequentially. This improves overall system throughput and resource utilization.

## Architecture

The implementation consists of:

1. **Task Dispatcher**: A single goroutine responsible for fetching schedulable tasks from the database and placing them in a queue.
2. **Worker Pool**: Multiple worker goroutines that process tasks from the queue concurrently.
3. **Task Queue**: A buffered channel that holds tasks waiting to be processed.
4. **Concurrency Control**: Mutex locks to control access to shared resources like providers and provisioners.

## Core Components

- **WorkerPool struct**: Manages the worker pool, task queue, and shared resources
- **NewWorkerPool()**: Creates a new worker pool instance
- **LaunchWorkerPool()**: Starts the task dispatcher and worker goroutines
- **taskDispatcher()**: Fetches tasks and adds them to the queue
- **taskProcessor()**: Worker goroutines that process tasks from the queue

## Configuration

Configure the number of workers using the `WORKER_COUNT` environment variable:

```shell
# Set to the desired number of concurrent worker goroutines
WORKER_COUNT=10
```

If not specified, the system defaults to 5 workers (`DefaultWorkerCount`).

## Design Considerations

### Task Independence

The worker pool assumes that tasks are independent and can be processed in any order. The database query that fetches tasks orders them by:

1. Tasks without errors before tasks with errors
2. Oldest tasks first (by creation date)

### Resource Management

The implementation includes concurrency control for shared resources:
- Provider instances are cached and protected by a mutex
- Provisioner instances are cached and protected by a mutex

### Graceful Shutdown

The worker pool implements a graceful shutdown process:
1. When a shutdown signal is received, the dispatcher stops fetching new tasks
2. The task queue is closed after the dispatcher has completed
3. Workers finish processing their current tasks and then terminate

## Benefits

1. **Increased Throughput**: Multiple tasks can be processed concurrently, improving overall system throughput.
2. **Resource Efficiency**: System resources are used more efficiently by allowing IO-bound operations to proceed in parallel.
3. **Responsiveness**: The system can handle bursts of tasks more effectively.
4. **Scalability**: The number of workers can be adjusted based on the host system's capabilities.

## Future Improvements

Potential future improvements include:

1. **Resource-aware Scheduling**: Implement task priorities or resource allocation based on task types.
2. **Worker Health Metrics**: Track and expose metrics about worker pool performance.
3. **Dynamic Worker Scaling**: Automatically adjust the number of workers based on system load.
4. **Task Batching**: Group related tasks for more efficient processing. 