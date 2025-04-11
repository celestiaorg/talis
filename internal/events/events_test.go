package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/celestiaorg/talis/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestEventSystem(t *testing.T) {
	t.Run("Subscribe and Publish", func(t *testing.T) {
		// Reset handlers for clean test environment
		handlers = make(map[EventType][]Handler)
		eventChan = make(chan Event, EventChannelSize)

		var wg sync.WaitGroup
		wg.Add(1)

		var receivedEvent Event
		testHandler := func(ctx context.Context, event Event) error {
			receivedEvent = event
			wg.Done()
			return nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start event processing
		Start(ctx)

		// Subscribe to test event
		Subscribe(EventInstancesCreated, testHandler)

		// Create test event
		testEvent := Event{
			Type:    EventInstancesCreated,
			JobID:   "test-job-123",
			JobName: "test-job",
			OwnerID: 1,
			Instances: []types.InstanceInfo{
				{
					Name:     "test-instance",
					PublicIP: "1.2.3.4",
				},
			},
		}

		// Publish event
		Publish(testEvent)

		// Wait for handler to process event with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success case
		case <-time.After(2 * time.Second):
			t.Fatal("Test timed out waiting for event handler")
		}

		// Verify received event matches published event
		assert.Equal(t, testEvent.Type, receivedEvent.Type)
		assert.Equal(t, testEvent.JobID, receivedEvent.JobID)
		assert.Equal(t, testEvent.JobName, receivedEvent.JobName)
		assert.Equal(t, testEvent.OwnerID, receivedEvent.OwnerID)
		assert.Equal(t, testEvent.Instances[0].Name, receivedEvent.Instances[0].Name)
		assert.Equal(t, testEvent.Instances[0].PublicIP, receivedEvent.Instances[0].PublicIP)
	})

	t.Run("Multiple Handlers", func(t *testing.T) {
		// Reset handlers for clean test environment
		handlers = make(map[EventType][]Handler)
		eventChan = make(chan Event, EventChannelSize)

		var wg sync.WaitGroup
		wg.Add(2) // Expecting two handlers to be called

		handlerCalls := make(map[string]bool)
		var mu sync.Mutex

		handler1 := func(ctx context.Context, event Event) error {
			mu.Lock()
			handlerCalls["handler1"] = true
			mu.Unlock()
			wg.Done()
			return nil
		}

		handler2 := func(ctx context.Context, event Event) error {
			mu.Lock()
			handlerCalls["handler2"] = true
			mu.Unlock()
			wg.Done()
			return nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start event processing
		Start(ctx)

		// Subscribe both handlers
		Subscribe(EventInstancesCreated, handler1)
		Subscribe(EventInstancesCreated, handler2)

		// Publish test event
		Publish(Event{
			Type:    EventInstancesCreated,
			JobID:   "test-job-456",
			JobName: "test-job",
		})

		// Wait for handlers with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success case
		case <-time.After(2 * time.Second):
			t.Fatal("Test timed out waiting for event handlers")
		}

		// Verify both handlers were called
		mu.Lock()
		assert.True(t, handlerCalls["handler1"], "Handler 1 should have been called")
		assert.True(t, handlerCalls["handler2"], "Handler 2 should have been called")
		mu.Unlock()
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		// Reset handlers for clean test environment
		handlers = make(map[EventType][]Handler)
		eventChan = make(chan Event, EventChannelSize)

		ctx, cancel := context.WithCancel(context.Background())

		// Start event processing
		Start(ctx)

		// Subscribe a handler that should not be called
		Subscribe(EventInstancesCreated, func(ctx context.Context, event Event) error {
			t.Error("Handler should not be called after context cancellation")
			return nil
		})

		// Cancel context immediately
		cancel()

		// Give some time for the goroutine to process the cancellation
		time.Sleep(100 * time.Millisecond)

		// Try to publish an event after cancellation
		// This should not block or panic
		Publish(Event{
			Type:    EventInstancesCreated,
			JobID:   "test-job-789",
			JobName: "test-job",
		})

		// Wait a bit to ensure no handlers are called
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Different Event Types", func(t *testing.T) {
		// Reset handlers for clean test environment
		handlers = make(map[EventType][]Handler)
		eventChan = make(chan Event, EventChannelSize)

		var wg sync.WaitGroup
		wg.Add(2)

		receivedEvents := make(map[EventType]bool)
		var mu sync.Mutex

		createdHandler := func(ctx context.Context, event Event) error {
			mu.Lock()
			receivedEvents[EventInstancesCreated] = true
			mu.Unlock()
			wg.Done()
			return nil
		}

		deletedHandler := func(ctx context.Context, event Event) error {
			mu.Lock()
			receivedEvents[EventInstancesDeleted] = true
			mu.Unlock()
			wg.Done()
			return nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start event processing
		Start(ctx)

		// Subscribe to different event types
		Subscribe(EventInstancesCreated, createdHandler)
		Subscribe(EventInstancesDeleted, deletedHandler)

		// Publish both types of events
		Publish(Event{Type: EventInstancesCreated, JobID: "job1"})
		Publish(Event{Type: EventInstancesDeleted, JobID: "job2"})

		// Wait for handlers with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success case
		case <-time.After(2 * time.Second):
			t.Fatal("Test timed out waiting for event handlers")
		}

		// Verify both event types were handled
		mu.Lock()
		assert.True(t, receivedEvents[EventInstancesCreated], "Created event should have been handled")
		assert.True(t, receivedEvents[EventInstancesDeleted], "Deleted event should have been handled")
		mu.Unlock()
	})
}
