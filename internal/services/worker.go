package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

// DefaultBackoff is the default backoff time for the worker
const DefaultBackoff = time.Second

const (
	// DefaultWorkerCount is the default number of workers in the pool
	DefaultWorkerCount = 100

	// DefaultHighPriorityRatio is the default ratio of workers assigned to high priority tasks
	DefaultHighPriorityRatio = 0.7

	// QueueSize is the size of the task queue
	QueueSize = 100
)

// WorkerPool is a struct that contains the worker pool's dependencies
type WorkerPool struct {
	// Services
	instanceService *Instance
	projectService  *Project
	taskService     *Task
	userService     *User

	// Providers & Provisioners
	providers map[models.ProviderID]compute.Provider
	// Create a provisioner for each provider
	provisioners map[models.ProviderID]compute.Provisioner
	computeMU    sync.RWMutex

	// Config
	backoff           time.Duration
	workerCount       int
	highPriorityRatio float64

	// Task queues
	highPriorityQueue chan *models.Task
	lowPriorityQueue  chan *models.Task
}

// NewWorkerPool creates a new WorkerPool
func NewWorkerPool(instanceService *Instance, projectService *Project, taskService *Task, userService *User, backoff time.Duration) *WorkerPool {
	return &WorkerPool{
		instanceService:   instanceService,
		projectService:    projectService,
		taskService:       taskService,
		userService:       userService,
		providers:         make(map[models.ProviderID]compute.Provider),
		provisioners:      make(map[models.ProviderID]compute.Provisioner),
		backoff:           backoff,
		workerCount:       DefaultWorkerCount,
		highPriorityRatio: DefaultHighPriorityRatio,
		highPriorityQueue: make(chan *models.Task, QueueSize),
		lowPriorityQueue:  make(chan *models.Task, QueueSize),
	}
}

// WithWorkerCount sets the number of workers in the pool
func (w *WorkerPool) WithWorkerCount(count int) *WorkerPool {
	if count > 0 {
		w.workerCount = count
	}
	return w
}

// WithHighPriorityRatio sets the ratio of workers assigned to high priority tasks
func (w *WorkerPool) WithHighPriorityRatio(ratio float64) *WorkerPool {
	if ratio > 0 && ratio <= 1.0 {
		w.highPriorityRatio = ratio
	}
	return w
}

// LaunchWorkerPool launches a task dispatcher and worker pool to process tasks
func (w *WorkerPool) LaunchWorkerPool(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	const taskLimit = 10

	// Recover stale tasks from previous runs
	w.recoverStaleTasks(ctx)

	// Create a wait group for the workers
	var workersWg sync.WaitGroup

	// Launch task dispatchers for each priority level
	dispatcherCtx, cancelDispatcher := context.WithCancel(ctx)
	workersWg.Add(2) // One for each priority dispatcher
	go w.taskDispatcher(dispatcherCtx, &workersWg, models.TaskPriorityHigh, taskLimit)
	go w.taskDispatcher(dispatcherCtx, &workersWg, models.TaskPriorityLow, taskLimit)

	// Calculate worker distribution
	highPriorityWorkers := int(float64(w.workerCount) * w.highPriorityRatio)
	lowPriorityWorkers := w.workerCount - highPriorityWorkers

	// Launch high priority workers
	for i := 0; i < highPriorityWorkers; i++ {
		workerID := i + 1
		workersWg.Add(1)
		go w.highPriorityTaskProcessor(ctx, &workersWg, workerID)
	}

	// Launch low priority workers
	for i := 0; i < lowPriorityWorkers; i++ {
		workerID := highPriorityWorkers + i + 1
		workersWg.Add(1)
		go w.lowPriorityTaskProcessor(ctx, &workersWg, workerID)
	}

	logger.Infof("Worker pool started with %d workers (%d high priority, %d low priority)",
		w.workerCount, highPriorityWorkers, lowPriorityWorkers)

	// Wait for context cancellation
	<-ctx.Done()
	logger.Info("Worker pool received shutdown signal, stopping dispatchers...")
	cancelDispatcher()

	// Close task queues after dispatchers are done
	workersWg.Wait()
	close(w.highPriorityQueue)
	close(w.lowPriorityQueue)

	logger.Info("Worker pool shutdown complete")
}

// recoverStaleTasks finds and resets tasks that were in progress when the system crashed
func (w *WorkerPool) recoverStaleTasks(ctx context.Context) {
	count, err := w.taskService.RecoverStaleTasks(ctx)
	if err != nil {
		logger.Errorf("Failed to recover stale tasks: %v", err)
		return
	}

	if count > 0 {
		logger.Infof("Recovered %d stale tasks that were in progress during previous run", count)
	}
}

// taskDispatcher fetches tasks from database and puts them in the appropriate queue
func (w *WorkerPool) taskDispatcher(ctx context.Context, wg *sync.WaitGroup, priority models.TaskPriority, taskLimit int) {
	defer wg.Done()

	// Initialize ticker with 1 second interval to reduce DB polling on startup
	t := time.NewTicker(time.Second)
	defer t.Stop()

	// Determine which queue to use based on priority
	var queue chan *models.Task
	priorityName := priority.String()

	if priority == models.TaskPriorityHigh {
		queue = w.highPriorityQueue
	} else {
		queue = w.lowPriorityQueue
	}

	logger.Infof("%s priority task dispatcher started", priorityName)

	for {
		select {
		case <-ctx.Done():
			logger.Infof("%s priority task dispatcher received shutdown signal, stopping...", priorityName)
			return
		case <-t.C:
		}

		// Fetch schedulable tasks for this priority level
		tasks, err := w.taskService.GetSchedulableTasks(ctx, priority, taskLimit)
		if err != nil {
			logger.Errorf("%s priority task dispatcher error fetching tasks: %v", priorityName, err)
			// Wait before retrying to avoid spamming logs on persistent DB errors
			t.Reset(w.backoff)
			continue
		}

		if len(tasks) == 0 {
			logger.Debugf("%s priority task dispatcher: No tasks to process", priorityName)
			// Wait before retrying to give time for tasks to be created
			t.Reset(w.backoff)
			continue
		}

		// Distribute tasks to workers through the queue
		for i := range tasks {
			select {
			case <-ctx.Done():
				logger.Infof("%s priority task dispatcher received shutdown signal, stopping...", priorityName)
				return
			case queue <- &tasks[i]:
				logger.Debugf("%s priority task dispatcher: Queued task %d for processing", priorityName, tasks[i].ID)
			}
		}
	}
}

// highPriorityTaskProcessor processes tasks from the high priority queue
func (w *WorkerPool) highPriorityTaskProcessor(ctx context.Context, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	logger.Infof("High priority worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			logger.Infof("High priority worker %d received shutdown signal, stopping...", workerID)
			return
		case task, ok := <-w.highPriorityQueue:
			if !ok {
				// Queue is closed
				logger.Infof("High priority worker %d shutting down, task queue is closed", workerID)
				return
			}

			// Process the task
			w.processTask(ctx, workerID, task)
		}
	}
}

// lowPriorityTaskProcessor processes tasks from the low priority queue
func (w *WorkerPool) lowPriorityTaskProcessor(ctx context.Context, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	logger.Infof("Low priority worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			logger.Infof("Low priority worker %d received shutdown signal, stopping...", workerID)
			return
		case task, ok := <-w.lowPriorityQueue:
			if !ok {
				// Queue is closed
				logger.Infof("Low priority worker %d shutting down, task queue is closed", workerID)
				return
			}

			// Process the task
			w.processTask(ctx, workerID, task)
		}
	}
}

// processTask handles the actual task processing with locking
func (w *WorkerPool) processTask(ctx context.Context, workerID int, task *models.Task) {
	priorityName := task.Priority.String()
	logger.Debugf("%s priority worker %d attempting to process task %d", priorityName, workerID, task.ID)

	// Atomically increment attempts count before trying to acquire the lock
	// This ensures we track all processing attempts, even if lock acquisition fails
	err := w.taskService.IncrementAttempts(ctx, task.ID)
	if err != nil {
		logger.Errorf("%s priority worker %d failed to increment attempts for task %d: %v",
			priorityName, workerID, task.ID, err)
		// Continue processing despite the error, as this is not critical
	} else {
		// Update the local task object to reflect the increment
		task.Attempts++
	}

	// Try to acquire a lock on the task
	err = w.taskService.AcquireTaskLock(ctx, task.ID)
	if err != nil {
		if errors.Is(err, ErrTaskLockNotAcquired) {
			logger.Debugf("%s priority worker %d could not acquire lock for task %d, skipping",
				priorityName, workerID, task.ID)
		} else {
			logger.Errorf("%s priority worker %d failed to acquire lock for task %d: %v",
				priorityName, workerID, task.ID, err)
		}
		return
	}

	logger.Debugf("%s priority worker %d processing task %d", priorityName, workerID, task.ID)

	// Process the task based on its action
	var processErr error
	switch task.Action {
	case models.TaskActionCreateInstances:
		processErr = w.processCreateInstanceTask(ctx, task)
		if processErr != nil {
			logMsg := fmt.Sprintf("❌ %s priority worker %d failed to process create instance task %d: %v",
				priorityName, workerID, task.ID, processErr)
			logger.Error(logMsg)
			task.Logs += fmt.Sprintf("\n%s", logMsg)
			err = w.taskService.UpdateFailed(ctx, task, processErr.Error(), logMsg)
			if err != nil {
				logger.Errorf("%s priority worker %d: Failed to update task: %v", priorityName, workerID, err)
			}
		} else {
			err = w.taskService.UpdateStatus(ctx, task.OwnerID, task.ID, models.TaskStatusCompleted)
			if err != nil {
				logger.Errorf("%s priority worker %d: Failed to update task status: %v", priorityName, workerID, err)
			}
		}
	case models.TaskActionTerminateInstances:
		processErr = w.processTerminateInstanceTask(ctx, task)
		if processErr != nil {
			logMsg := fmt.Sprintf("❌ %s priority worker %d failed to process terminate instance task %d: %v",
				priorityName, workerID, task.ID, processErr)
			logger.Error(logMsg)
			task.Logs += fmt.Sprintf("\n%s", logMsg)
			err = w.taskService.UpdateFailed(ctx, task, processErr.Error(), logMsg)
			if err != nil {
				logger.Errorf("%s priority worker %d: Failed to update task: %v", priorityName, workerID, err)
			}
		} else {
			err = w.taskService.UpdateStatus(ctx, task.OwnerID, task.ID, models.TaskStatusCompleted)
			if err != nil {
				logger.Errorf("%s priority worker %d: Failed to update task status: %v", priorityName, workerID, err)
			}
		}
	default:
		logger.Errorf("%s priority worker %d: Unknown task action %s for task %d",
			priorityName, workerID, task.Action, task.ID)
	}

	// Release the lock regardless of success or failure
	if err := w.taskService.ReleaseTaskLock(ctx, task.ID); err != nil {
		logger.Errorf("%s priority worker %d failed to release lock for task %d: %v",
			priorityName, workerID, task.ID, err)
	}
}

// processCreateInstanceTask processes a create instance task. It will handle the instance creation, provisioning, and status updates for the instance.
func (w *WorkerPool) processCreateInstanceTask(ctx context.Context, task *models.Task) error {
	err := w.taskService.UpdateStatus(ctx, task.OwnerID, task.ID, models.TaskStatusRunning)
	if err != nil {
		return fmt.Errorf("worker: failed to update task status: %w", err)
	}
	logger.Debugf("Creating instance for task %d", task.ID)

	// Unmarshal the task payload
	var instanceReq types.InstanceRequest
	err = json.Unmarshal(task.Payload, &instanceReq)
	if err != nil {
		return fmt.Errorf("worker: failed to unmarshal task payload for task %d: %w", task.ID, err)
	}

	// Check the instance status
	instance, err := w.instanceService.Get(ctx, instanceReq.OwnerID, instanceReq.InstanceID)
	if err != nil {
		return fmt.Errorf("worker: failed to get instance: %w", err)
	}
	if instance == nil {
		return fmt.Errorf("worker: instance %d not found", instanceReq.InstanceID)
	}

	switch instance.Status {
	case models.InstanceStatusPending:
		logger.Debugf("Instance ID %d is in status %s, creating", instance.ID, instance.Status)
		w.instanceService.addTaskLogs(ctx, instanceReq.OwnerID, task, fmt.Sprintf("Creating instance ID %d", instance.ID))
		// Create the instance
		// NOTE: need to understand if server creation via the hypervisor is atomic or if we need to understand how to pick up where we left off

		// Get the compute provider or create a new one
		provider, err := w.getProvider(instanceReq.Provider)
		if err != nil {
			return fmt.Errorf("worker: failed to get compute provider for provider %s: %w", instanceReq.Provider, err)
		}

		// Create the instance
		// NOTE: since the instance request type is now being updated during the create instance process we might need to update the task payload to include the updates. This is more of a concern if we want to support resuming from a failed task.
		err = provider.CreateInstance(ctx, &instanceReq)
		if err != nil {
			return fmt.Errorf("worker: failed to create instance: %w", err)
		}

		// Update instance DB with IP and volume info
		// Convert volume details to database model
		dbVolumeDetails := make(models.VolumeDetails, 0, len(instanceReq.VolumeDetails))
		for _, vd := range instanceReq.VolumeDetails {
			dbVolumeDetails = append(dbVolumeDetails, models.VolumeDetail{
				ID:         vd.ID,
				Name:       vd.Name,
				Region:     vd.Region,
				SizeGB:     vd.SizeGB,
				MountPoint: vd.MountPoint,
			})
		}
		instance.PublicIP = instanceReq.PublicIP
		instance.VolumeIDs = instanceReq.VolumeIDs
		instance.VolumeDetails = dbVolumeDetails
		instance.Status = models.InstanceStatusCreated
		instance.ProviderInstanceID = instanceReq.ProviderInstanceID
		err = w.instanceService.Update(ctx, instanceReq.OwnerID, instance.ID, instance)
		if err != nil {
			return fmt.Errorf("worker: failed to update instance: %w", err)
		}

		// Fall through to the next case and step in the process
		fallthrough
	case models.InstanceStatusCreated:
		logger.Debugf("Instance ID %d is in status %s, determine if provisioning is needed", instance.ID, instance.Status)
		w.instanceService.addTaskLogs(ctx, instanceReq.OwnerID, task, fmt.Sprintf("Instance ID %d created, determining if provisioning is needed", instance.ID))

		// Check if the instance needs to be provisioned
		if !instanceReq.Provision {
			logger.Debugf("Provisioning is not needed for instance ID %d, updating status to ready", instance.ID)
			// Instance is ready, update and return
			instance.Status = models.InstanceStatusReady
			err = w.instanceService.Update(ctx, instanceReq.OwnerID, instance.ID, instance)
			if err != nil {
				return fmt.Errorf("worker: failed to update instance ID %d: %w", instance.ID, err)
			}
			logger.Debugf("✅ Instance ID %d is ready", instance.ID)
			w.instanceService.addTaskLogs(ctx, instanceReq.OwnerID, task, fmt.Sprintf("Instance ID %d is ready", instance.ID))
			return nil
		}

		// Instance needs to be provisioned, update the status to provisioning
		logger.Debugf("Provisioning is needed for instance ID %d, updating status to provisioning", instance.ID)
		instance.Status = models.InstanceStatusProvisioning
		err = w.instanceService.Update(ctx, instanceReq.OwnerID, instance.ID, instance)
		if err != nil {
			return fmt.Errorf("worker: failed to update instance ID %d: %w", instance.ID, err)
		}

		fallthrough
	case models.InstanceStatusProvisioning:
		logger.Debugf("Instance ID %d is in status %s, provisioning", instance.ID, instance.Status)
		w.instanceService.addTaskLogs(ctx, instanceReq.OwnerID, task, fmt.Sprintf("Provisioning instance ID %d", instance.ID))

		// Get provisioning tasks
		provisioner, err := w.getProvisioner(instanceReq.Provider)
		if err != nil {
			return fmt.Errorf("worker: failed to get provisioner for provider %s: %w", instanceReq.Provider, err)
		}

		// Verify that the instance IP is available
		if instance.PublicIP == "" {
			return fmt.Errorf("worker: instance ID %d has no public IP, can't provision", instance.ID)
		}

		// TODO: Validate inputs

		// create a hosts file with the instance IP to provision.
		inventoryPath, err := provisioner.CreateInventory(&instanceReq, getAnsibleSSHKeyPath(instanceReq))
		if err != nil {
			return fmt.Errorf("worker: failed to create inventory file for instance ID %d: %w", instance.ID, err)
		}

		if inventoryPath == "" {
			// Should not happen if hosts were provided and valid, but handle defensively
			logger.Warnf("Worker: Inventory path empty for instance ID %d, skipping playbook run", instance.ID)
		} else {
			// tags depending on the provisioner
			tags := []string{}
			if instanceReq.Provider == models.ProviderXimera {
				tags = []string{"setup"}
			}
			if instanceReq.Provider == models.ProviderDO {
				tags = []string{"setup", "volumes"}
			}

			if err := provisioner.RunAnsiblePlaybook(inventoryPath, tags); err != nil {
				return fmt.Errorf("Worker: Failed to run ansible playbook for instance ID %d: %w", instance.ID, err)
			}
			// Optionally remove inventory file after successful run
			// if err := os.Remove(inventoryPath); err != nil {
			// 	 logger.Warnf("Worker: Failed to remove inventory file %s: %v", inventoryPath, err)
			// }
		}

		// Update status to Ready
		instance.Status = models.InstanceStatusReady
		err = w.instanceService.Update(ctx, instanceReq.OwnerID, instance.ID, instance)
		if err != nil {
			return fmt.Errorf("worker: failed to update instance ID %d to ready: %w", instance.ID, err)
		}
		logger.Debugf("✅ Instance ID %d successfully provisioned, marking as ready", instance.ID)
		w.instanceService.addTaskLogs(ctx, instanceReq.OwnerID, task, fmt.Sprintf("Instance ID %d successfully provisioned and is ready", instance.ID))

	case models.InstanceStatusReady, models.InstanceStatusTerminated:
		// Instance is already in a final state for this task
		logger.Debugf("Instance ID %d is already ready or terminated, nothing to do for create task.", instance.ID)
		w.instanceService.addTaskLogs(ctx, instanceReq.OwnerID, task, fmt.Sprintf("Instance ID %d already in final state (%s)", instance.ID, instance.Status))
		return nil
	default:
		return fmt.Errorf("worker: instance ID %d is in an unknown state %s", instance.ID, instance.Status)
	}
	return nil
}

// processTerminateInstanceTask processes a terminate instance task. It will handle the infrastructure deletion and status updates for the instance.
func (w *WorkerPool) processTerminateInstanceTask(ctx context.Context, task *models.Task) error {
	err := w.taskService.UpdateStatus(ctx, task.OwnerID, task.ID, models.TaskStatusRunning)
	if err != nil {
		return fmt.Errorf("worker: failed to update task status: %w", err)
	}
	logger.Debugf("Terminating instance for task %d", task.ID)

	// Unmarshal the task payload
	var deleteReq types.DeleteInstanceRequest
	err = json.Unmarshal(task.Payload, &deleteReq)
	if err != nil {
		return fmt.Errorf("worker: failed to unmarshal task payload for task %d: %w", task.ID, err)
	}

	// Get the instance
	instance, err := w.instanceService.Get(ctx, task.OwnerID, deleteReq.InstanceID)
	if err != nil {
		return fmt.Errorf("worker: failed to get instance: %w", err)
	}
	if instance == nil {
		return fmt.Errorf("worker: instance %d not found", deleteReq.InstanceID)
	}

	// Confirm the instance is not already terminated
	if instance.Status == models.InstanceStatusTerminated {
		logger.Debugf("Instance ID %d is already terminated, skipping", instance.ID)
		return nil
	}

	// Get the compute provider or create a new one
	provider, err := w.getProvider(instance.ProviderID)
	if err != nil {
		return fmt.Errorf("worker: failed to get compute provider for provider %s: %w", instance.ProviderID, err)
	}

	// Delete the instance
	logger.Infof("🗑️ Deleting %v droplet ID: %d in region %v", instance.ProviderID, instance.ProviderInstanceID, instance.Region)
	err = provider.DeleteInstance(ctx, instance.ProviderInstanceID)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			logger.Warnf("⚠️ Warning: Instance %v was already deleted", instance.ProviderInstanceID)
			// Make sure the instance is marked as terminated in the database
			return w.instanceService.MarkAsTerminated(ctx, instance.OwnerID, instance.ID)
		}
		return fmt.Errorf("failed to delete instance %v: %w", instance.ProviderInstanceID, err)
	}
	logger.Debugf("✅ Successfully deleted instance: %v", instance.ProviderInstanceID)

	// Update database
	if err := w.instanceService.MarkAsTerminated(ctx, instance.OwnerID, instance.ID); err != nil {
		return fmt.Errorf("failed to terminate instance ID %d in database: %w", instance.ID, err)
	}
	return nil
}

// getProvider returns the compute provider for the given instance
func (w *WorkerPool) getProvider(providerID models.ProviderID) (compute.Provider, error) {
	w.computeMU.RLock()

	provider, ok := w.providers[providerID]
	if !ok {
		w.computeMU.RUnlock()
		w.computeMU.Lock()
		var err error
		provider, err = compute.NewComputeProvider(providerID)
		if err != nil {
			w.computeMU.Unlock()
			return nil, fmt.Errorf("worker: failed to create compute provider for provider %s: %w", providerID, err)
		}
		w.providers[providerID] = provider
		w.computeMU.Unlock()
		w.computeMU.RLock()
	}
	w.computeMU.RUnlock()

	return provider, nil
}

// getProvisioner returns the provisioner for the given instance
func (w *WorkerPool) getProvisioner(providerID models.ProviderID) (compute.Provisioner, error) {
	w.computeMU.RLock()

	provisioner, ok := w.provisioners[providerID]
	if !ok {
		w.computeMU.RUnlock()
		w.computeMU.Lock()
		jobID := fmt.Sprintf("job-%s", time.Now().Format("20060102-150405"))
		provisioner = compute.NewProvisioner(jobID)
		w.provisioners[providerID] = provisioner
		w.computeMU.Unlock()
		w.computeMU.RLock()
	}
	w.computeMU.RUnlock()

	return provisioner, nil
}
