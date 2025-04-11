// Package events provides event handling functionality
package events

import (
	"context"
	"sync"

	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

// EventType represents the type of infrastructure event
type EventType string

const (
	// EventInstancesCreated is emitted when instances are created
	EventInstancesCreated EventType = "instances_created"
	// EventInstancesDeleted is emitted when instances are deleted
	EventInstancesDeleted EventType = "instances_deleted"
	// EventInventoryRequested is emitted when inventory needs to be generated from DB
	EventInventoryRequested EventType = "inventory_requested"
	// EventChannelSize is the buffer size for the event channel
	EventChannelSize = 100
)

// Event represents an infrastructure event
type Event struct {
	Type      EventType               // The type of event
	JobID     string                  // The job ID
	JobName   string                  // The job name
	OwnerID   uint                    // The owner ID
	Instances []types.InstanceInfo    // The instances created
	Requests  []types.InstanceRequest // The requests that were made
}

// Handler is a function that handles an event
type Handler func(context.Context, Event) error

var (
	// handlers is a map of event types to their handlers
	handlers = make(map[EventType][]Handler)
	// handlersMu is a mutex for the handlers map
	handlersMu sync.RWMutex
	// eventChan is a channel for events
	eventChan = make(chan Event, EventChannelSize)
)

// Subscribe registers a handler for a specific event type
func Subscribe(eventType EventType, handler Handler) {
	handlersMu.Lock()
	defer handlersMu.Unlock()
	handlers[eventType] = append(handlers[eventType], handler)
	logger.Debugf("📝 Registered handler for event type: %s", eventType)
}

// Publish sends an event to be processed
func Publish(event Event) {
	eventChan <- event
	logger.Debugf("📢 Published event: %s (Job: %s)", event.Type, event.JobID)
}

// Start starts the event processing loop
func Start(ctx context.Context) {
	go processEvents(ctx)
	logger.Info("🎯 Started event processing loop")
}

// processEvents handles events in the background
func processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Info("🛑 Stopping event processing loop")
			return
		case event := <-eventChan:
			logger.Debugf("📥 Received event %s for job %s", event.Type, event.JobID)
			handlersMu.RLock()
			eventHandlers := handlers[event.Type]
			handlersMu.RUnlock()

			// Process event with all registered handlers
			for _, handler := range eventHandlers {
				go func(h Handler, e Event) {
					logger.Debugf("⚡ Processing event %s for job %s", e.Type, e.JobID)
					if err := h(ctx, e); err != nil {
						logger.Errorf("❌ Failed to handle event %s: %v", e.Type, err)
					} else {
						logger.Debugf("✅ Successfully processed event %s for job %s", e.Type, e.JobID)
					}
				}(handler, event)
			}
		}
	}
}
