package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

type worker struct {
	instanceService *Instance
	projectService  *Project
	taskService     *Task
}

func NewWorker(instanceService *Instance, projectService *Project, taskService *Task) *worker {
	return &worker{
		instanceService: instanceService,
		projectService:  projectService,
		taskService:     taskService,
	}
}

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

		for _, task := range tasks {
			// Check for shutdown signal here as well in case of long running tasks
			select {
			case <-ctx.Done():
				logger.Info("Worker received shutdown signal, stopping...")
				return
			default:
			}
			switch task.Action {
			case models.TaskActionCreateInstances:
				processCreateInstances(ctx, &task)
			case models.TaskActionTerminateInstances:
				processTerminateInstances(ctx, &task)
			default:
				logger.Errorf("Worker error fetching tasks: %v", err)
				// Wait before retrying to avoid spamming logs on persistent DB errors
				time.Sleep(backoff)
				continue
			}
		}

		// Wait before the next check
		time.Sleep(time.Second)
	}
}

func processCreateInstances(ctx context.Context, task *models.Task) error {
	// Unmarshal the payload
	var instancesRequest types.InstancesRequest
	if err := json.Unmarshal(task.Payload, &instancesRequest); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Create infrastructure client
	infraReq := &types.InstancesRequest{
		TaskName:  task.Name,
		Instances: instancesRequest.Instances,
		Action:    "create",
		Provider:  instancesRequest.Provider,
	}
	return nil
}

func processTerminateInstances(ctx context.Context, task *models.Task) error {
	return nil
}
