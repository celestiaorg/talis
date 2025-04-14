# Events Package

This package implements an event-driven system for decoupling components in the Talis infrastructure management system. It provides a simple yet powerful event bus that enables asynchronous communication between different parts of the application.

## Overview

The events package is a core component that enables:
- Asynchronous communication between services
- Decoupling of infrastructure creation and provisioning
- Event-based workflow management
- Scalable and extensible architecture

## Event Types

Currently supported event types:
- `EventInstancesCreated`: Emitted when new instances are created
- `EventInstancesDeleted`: Emitted when instances are deleted

## Usage

### Starting the Event System

```go
ctx := context.Background()
events.Start(ctx)
```

### Publishing Events

```go
events.Publish(events.Event{
    Type:      events.EventInstancesCreated,
    JobID:     "job-123",
    JobName:   "create-instances",
    OwnerID:   1,
    Instances: instances,
    Requests:  requests,
})
```

### Subscribing to Events

```go
handler := func(ctx context.Context, event Event) error {
    // Handle the event
    return nil
}

events.Subscribe(events.EventInstancesCreated, handler)
```

## Best Practices

1. **Error Handling**
   - Always return errors from handlers
   - Handle errors appropriately in the event processing loop
   - Use context for cancellation

2. **Event Design**
   - Keep events focused and specific
   - Include all necessary data in the event
   - Avoid circular references

3. **Handler Implementation**
   - Keep handlers lightweight
   - Avoid blocking operations
   - Use timeouts where appropriate
   - Handle concurrent access safely

4. **Context Usage**
   - Pass context through to all async operations
   - Use context cancellation for cleanup
   - Set appropriate timeouts

## Example: Complete Handler

```go
func handleInstanceCreation(ctx context.Context, event Event) error {
    // Validate event data
    if len(event.Instances) == 0 {
        return errors.New("no instances in event")
    }

    // Process with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    // Handle the event
    for _, instance := range event.Instances {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Process instance
        }
    }

    return nil
}
```

## Testing

The package includes comprehensive tests covering:
- Basic publish/subscribe functionality
- Multiple handler scenarios
- Context cancellation
- Different event types
- Concurrent event processing

Run tests with:
```bash
go test -v ./internal/events
```

## Benefits

1. **Loose Coupling**
   - Services communicate through events without direct dependencies
   - Easy to add new functionality without modifying existing code
   - Simple to extend with new event types

2. **Improved Error Handling**
   - Failures in one component don't affect others
   - Easy to implement retries and recovery
   - Better error isolation

3. **Scalability**
   - Events can be processed asynchronously
   - Multiple handlers can process events concurrently
   - Easy to distribute across services

4. **Testability**
   - Events and handlers can be tested in isolation
   - Easy to mock and verify event flow
   - Clear component boundaries

## Architecture Integration

```
+------------+     +------------+     +------------+
|            |     |            |     |            |
| services   +---->+  events    +---->+provisioner |
|            |     |            |     |            |
+------------+     +------------+     +------------+
```

The events package serves as the central communication hub, enabling:
- Decoupled service communication
- Async workflow processing
- Plugin-like extensibility
- Clear separation of concerns
