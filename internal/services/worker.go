package services

import (
	"context"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/logger"
)

// LaunchWorker launches a goroutine that will initialize the worker and execute tasks
func LaunchWorker(ctx context.Context, wg *sync.WaitGroup, taskService *Task) {
	defer wg.Done()
	const taskLimit = 10
	const backoff = time.Second

	logger.Info("Worker started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Worker received shutdown signal, stopping...")
			return
		default:
		}

		// Fetch schedulable tasks
		tasks, err := taskService.GetSchedulableTasks(ctx, taskLimit)
		if err != nil {
			logger.Errorf("Worker error fetching tasks: %v", err)
			// Wait before retrying to avoid spamming logs on persistent DB errors
			time.Sleep(backoff)
			continue
		}

		if len(tasks) == 0 {
			logger.Debug("Worker: No tasks to process")
			// Wait before retrying to give time for tasks to be created
			time.Sleep(backoff)
			continue
		}
		// Log fetched tasks (just IDs for brevity)
		taskIDs := make([]uint, len(tasks))
		for i, task := range tasks {
			taskIDs[i] = task.ID
		}
		logger.Infof("Worker fetched %d tasks: %v", len(tasks), taskIDs)

		// TODO: Implement actual task processing logic here
		// For now, we just log and discard them.

		// Wait before the next check
		time.Sleep(time.Second)
	}
}
