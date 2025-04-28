package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/pkg/models"
	"github.com/celestiaorg/talis/pkg/types"
)

// DefaultBackoff is the default backoff time for the worker
const DefaultBackoff = time.Second

// Worker is a struct that contains the worker's dependencies
type Worker struct {
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
	backoff time.Duration
}

// NewWorker creates a new Worker
func NewWorker(instanceService *Instance, projectService *Project, taskService *Task, userService *User, backoff time.Duration) *Worker {
	return &Worker{
		instanceService: instanceService,
		projectService:  projectService,
		taskService:     taskService,
		userService:     userService,
		providers:       make(map[models.ProviderID]compute.Provider),
		provisioners:    make(map[models.ProviderID]compute.Provisioner),
		backoff:         backoff,
	}
}

// LaunchWorker launches a goroutine that will initialize the worker and execute tasks
func (w *Worker) LaunchWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	const taskLimit = 10

	// NOTE: tickers need a non-zero duration, this will just cause a small delay before the worker starts
	t := time.NewTicker(time.Millisecond)

	logger.Info("Worker started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Worker received shutdown signal, stopping...")
			return
		case <-t.C:
		}

		// Fetch schedulable tasks. Initially we exclude the delete upload tasks as we want to prioritize the instance related tasks
		tasks, err := w.taskService.GetSchedulableTasks(ctx, taskLimit, models.TaskActionDeleteUpload)
		if err != nil {
			logger.Errorf("Worker error fetching tasks: %v", err)
			// Wait before retrying to avoid spamming logs on persistent DB errors
			t.Reset(w.backoff)
			continue
		}

		// If no tasks are found, we need to look for deleted uploads tasks
		if len(tasks) == 0 {
			logger.Debug("Worker: No tasks found on first fetch, looking for deleted uploads tasks")
			tasks, err = w.taskService.GetSchedulableTasks(ctx, taskLimit)
			if err != nil {
				logger.Errorf("Worker error fetching tasks: %v", err)
				// Wait before retrying to avoid spamming logs on persistent DB errors
				t.Reset(w.backoff)
				continue
			}
			if len(tasks) == 0 {
				logger.Debug("Worker: No tasks found on second fetch, waiting for next tick")
				t.Reset(w.backoff)
				continue
			}
		}

		// Process tasks
		for i := range tasks {
			// Check if the context has been cancelled to avoid processing tasks after shutdown
			select {
			case <-ctx.Done():
				logger.Info("Worker received shutdown signal, stopping...")
				return
			default:
			}
			tasks[i].Attempts++
			switch tasks[i].Action {
			case models.TaskActionCreateInstances:
				err := w.processCreateInstanceTask(ctx, &tasks[i])
				if err != nil {
					w.handleFailure(ctx, &tasks[i], fmt.Sprintf("❌ Failed to process create instance task %d: %v", tasks[i].ID, err), err)
					// No time reset needed here as we are just continuing through the tasks
					continue
				}
				err = w.taskService.UpdateStatus(ctx, tasks[i].OwnerID, tasks[i].ID, models.TaskStatusCompleted)
				if err != nil {
					logger.Errorf("Worker: Failed to update task status: %v", err)
				}
			case models.TaskActionTerminateInstances:
				err := w.processTerminateInstanceTask(ctx, &tasks[i])
				if err != nil {
					w.handleFailure(ctx, &tasks[i], fmt.Sprintf("❌ Failed to process terminate instance task %d: %v", tasks[i].ID, err), err)
					// No time reset needed here as we are just continuing through the tasks
					continue
				}
				err = w.taskService.UpdateStatus(ctx, tasks[i].OwnerID, tasks[i].ID, models.TaskStatusCompleted)
				if err != nil {
					logger.Errorf("Worker: Failed to update task status: %v", err)
				}
			case models.TaskActionDeleteUpload:
				attempted, err := w.processDeleteUploadTask(ctx, &tasks[i])
				if err != nil {
					w.handleFailure(ctx, &tasks[i], fmt.Sprintf("❌ Failed to process delete upload task %d: %v", tasks[i].ID, err), err)
					// No time reset needed here as we are just continuing through the tasks
					continue
				}
				if !attempted {
					// Reset the ticker to prevent rapid cycling through deletion tasks that are not yet ready to be deleted
					// NOTE: this reset only impacts the next attempt to pull new tasks, it will not impact iteration through the current tasks
					// NOTE: the task attempts are incremented at the beginning of the for loop but since we haven't updated the task in the DB, besides the status, the attempt won't be persisted
					t.Reset(w.backoff)
					continue
				}
				err = w.taskService.UpdateStatus(ctx, tasks[i].OwnerID, tasks[i].ID, models.TaskStatusCompleted)
				if err != nil {
					logger.Errorf("Worker: Failed to update task status: %v", err)
				}
			default:
				logger.Errorf("Worker: Unknown task action %s for task %d", tasks[i].Action, tasks[i].ID)
			}
		}
	}
}

// handleFailure handles a failure for a task. It will log the error and update the task status to failed.
func (w *Worker) handleFailure(ctx context.Context, task *models.Task, logMsg string, err error) {
	logger.Error(logMsg)
	task.Logs += fmt.Sprintf("\n%s", logMsg)
	err = w.taskService.UpdateFailed(ctx, task, err.Error(), logMsg)
	if err != nil {
		logger.Errorf("Worker: Failed to update task: %v", err)
	}
}

// processCreateInstanceTask processes a create instance task. It will handle the instance creation, provisioning, and status updates for the instance.
func (w *Worker) processCreateInstanceTask(ctx context.Context, task *models.Task) error {
	err := w.taskService.UpdateStatus(ctx, task.OwnerID, task.ID, models.TaskStatusRunning)
	if err != nil {
		return fmt.Errorf("Worker: Failed to update task status: %w", err)
	}
	logger.Debugf("Creating instance for task %d", task.ID)

	// Unmarshal the task payload
	var instanceReq types.InstanceRequest
	err = json.Unmarshal(task.Payload, &instanceReq)
	if err != nil {
		return fmt.Errorf("Worker: Failed to unmarshal task payload for task %d: %w", task.ID, err)
	}

	// Check the instance status
	instance, err := w.instanceService.GetByID(ctx, instanceReq.OwnerID, instanceReq.InstanceID)
	if err != nil {
		return fmt.Errorf("Worker: Failed to get instance: %w", err)
	}
	if instance == nil {
		return fmt.Errorf("Worker: Instance %d not found", instanceReq.InstanceID)
	}

	switch instance.Status {
	case models.InstanceStatusPending:
		logger.Debugf("Instance %s is in status %s, creating", instanceReq.Name, instance.Status)
		w.instanceService.addTaskLogs(ctx, instanceReq.OwnerID, task, "Creating instance")
		// Create the instance
		// NOTE: need to understand if server creation via the hypervisor is atomic or if we need to understand how to pick up where we left off

		// Get the compute provider or create a new one
		provider, err := w.getProvider(instanceReq.Provider)
		if err != nil {
			return fmt.Errorf("Worker: Failed to get compute provider for provider %s: %w", instanceReq.Provider, err)
		}

		// Create the instance
		// NOTE: since the instance request type is now being updated during the create instance process we might need to update the task payload to include the updates. This is more of a concern if we want to support resuming from a failed task.
		err = provider.CreateInstance(ctx, &instanceReq)
		if err != nil {
			return fmt.Errorf("Worker: Failed to create instance: %w", err)
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
		err = w.instanceService.UpdateByID(ctx, instanceReq.OwnerID, instance.ID, instance)
		if err != nil {
			return fmt.Errorf("Worker: Failed to update instance: %w", err)
		}

		// Fall through to the next case and step in the process
		fallthrough
	case models.InstanceStatusCreated:
		logger.Debugf("Instance %s is in status %s, determine if provisioning is needed", instanceReq.Name, instance.Status)
		w.instanceService.addTaskLogs(ctx, instanceReq.OwnerID, task, "Determining if provisioning is needed")

		// Check if the instance needs to be provisioned
		if !instanceReq.Provision {
			logger.Debugf("Provisioning is not needed for instance %s, updating status to ready", instanceReq.Name)
			// Instance is ready, update and return
			instance.Status = models.InstanceStatusReady
			err = w.instanceService.UpdateByID(ctx, instanceReq.OwnerID, instance.ID, instance)
			if err != nil {
				return fmt.Errorf("Worker: Failed to update instance: %w", err)
			}
			logger.Debugf("✅ Instance %s is ready", instanceReq.Name)
			return nil
		}

		// Instance needs to be provisioned, update the status to provisioning
		logger.Debugf("Provisioning is needed for instance %s, updating status to provisioning", instanceReq.Name)
		instance.Status = models.InstanceStatusProvisioning
		err = w.instanceService.UpdateByID(ctx, instanceReq.OwnerID, instance.ID, instance)
		if err != nil {
			return fmt.Errorf("Worker: Failed to update instance: %w", err)
		}

		fallthrough
	case models.InstanceStatusProvisioning:
		logger.Debugf("Instance %s is in status %s, provisioning", instanceReq.Name, instance.Status)
		w.instanceService.addTaskLogs(ctx, instanceReq.OwnerID, task, "Running Ansible provisioning")
		// Run the Ansible playbook
		// Ansible playbooks can be rerun so we shouldn't need to worry about where we left off

		// Get the provisioner, or create a new one if it doesn't exist
		provisioner := w.getProvisioner(instanceReq.Provider)
		if provisioner == nil {
			return fmt.Errorf("Worker: Failed to get provisioner for provider %v", instanceReq.Provider)
		}

		// --- Step 1: Ensure all hosts are ready for SSH connections ---
		// TODO: This currently uses a default/shared key path assumption. Improve if needed.
		sshKeyPath := getAnsibleSSHKeyPath(instanceReq)
		hosts := make([]string, 1)
		hosts[0] = instanceReq.PublicIP
		if err := provisioner.ConfigureHosts(ctx, hosts, sshKeyPath); err != nil {
			// ConfigureHosts already logs details
			return fmt.Errorf("failed to ensure SSH readiness for all hosts: %w", err)
		}

		// --- Step 2: Create the inventory file using InstanceRequest ---
		// CreateInventory now handles extracting info and determining the key path internally
		// TODO: should inventory path be stored?
		inventoryPath, err := provisioner.CreateInventory(&instanceReq, sshKeyPath)
		if err != nil {
			return fmt.Errorf("failed to create Ansible inventory: %w", err)
		}

		// --- Step 3: Run the Ansible playbook ---
		fmt.Println("📝 Running Ansible playbook...")
		if err := provisioner.RunAnsiblePlaybook(inventoryPath); err != nil {
			return fmt.Errorf("failed to run Ansible playbook: %w", err)
		}

		fmt.Println("✅ Ansible playbook completed successfully")
		w.instanceService.addTaskLogs(ctx, instanceReq.OwnerID, task, "Ansible provisioning completed")

		// Update the instance status to ready and continue
		instance.Status = models.InstanceStatusReady
		if instanceReq.PayloadPath != "" {
			instance.PayloadStatus = models.PayloadStatusExecuted
			logger.Debugf("✅ Updated payload status to executed for instance %s", instanceReq.Name)
		}
		err = w.instanceService.UpdateByID(ctx, instanceReq.OwnerID, instance.ID, instance)
		if err != nil {
			return fmt.Errorf("Worker: Failed to update instance: %w", err)
		}

		logger.Debugf("✅ Instance %s is ready", instanceReq.Name)

		return nil
	case models.InstanceStatusReady:
		// Assume the task was not updated properly, just log a warning and return nil so that the task is updated as successful
		logger.Warnf("Worker: Instance %s is in status %s, assuming task was not updated properly", instance.Name, instance.Status)
		return nil
	case models.InstanceStatusTerminated:
		return fmt.Errorf("Worker: Instance %s is in status %s, cannot create", instance.Name, instance.Status)
	default:
		return fmt.Errorf("Worker: Unknown instance status %s for instance %s", instance.Status, instance.Name)
	}
}

// processTerminateInstanceTask processes a terminate instance task. It will handle the infrastructure deletion and status updates for the instance.
func (w *Worker) processTerminateInstanceTask(ctx context.Context, task *models.Task) error {
	err := w.taskService.UpdateStatus(ctx, task.OwnerID, task.ID, models.TaskStatusRunning)
	if err != nil {
		return fmt.Errorf("Worker: Failed to update task status: %w", err)
	}
	logger.Debugf("Terminating instance for task %d", task.ID)

	// Unmarshal the task payload
	var deleteReq types.DeleteInstanceRequest
	err = json.Unmarshal(task.Payload, &deleteReq)
	if err != nil {
		return fmt.Errorf("Worker: Failed to unmarshal task payload for task %d: %w", task.ID, err)
	}

	// Get the instance
	instance, err := w.instanceService.GetByID(ctx, task.OwnerID, deleteReq.InstanceID)
	if err != nil {
		return fmt.Errorf("Worker: Failed to get instance: %w", err)
	}
	if instance == nil {
		return fmt.Errorf("Worker: Instance %d not found", deleteReq.InstanceID)
	}

	// Confirm the instance is not already terminated
	if instance.Status == models.InstanceStatusTerminated {
		logger.Debugf("Instance %s is already terminated, skipping", instance.Name)
		return nil
	}

	// Get the compute provider or create a new one
	provider, err := w.getProvider(instance.ProviderID)
	if err != nil {
		return fmt.Errorf("Worker: Failed to get compute provider for provider %s: %w", instance.ProviderID, err)
	}

	// Delete the instance
	logger.Infof("🗑️ Deleting %v droplet: %v in region %v", instance.ProviderID, instance.Name, instance.Region)
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
		return fmt.Errorf("failed to terminate instance %s in database: %w", instance.Name, err)
	}
	return nil
}

// processDeleteUploadTask processes a delete upload task.
func (w *Worker) processDeleteUploadTask(ctx context.Context, task *models.Task) (bool, error) {
	// Unmarshal the task payload
	var payload types.UploadDeletionPayload
	err := json.Unmarshal(task.Payload, &payload)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal task payload for task %d: %w", task.ID, err)
	}

	// Check if the deletion timestamp has passed
	if !time.Now().After(payload.DeletionTimestamp) {
		logger.Debugf("Deletion timestamp has not passed for upload %s, skipping", payload.UploadPath)
		return false, nil
	}

	err = w.taskService.UpdateStatus(ctx, task.OwnerID, task.ID, models.TaskStatusRunning)
	if err != nil {
		return false, fmt.Errorf("failed to update task status: %w", err)
	}
	logger.Debugf("Deleting upload for task %d", task.ID)

	// Delete the upload
	logger.Debugf("Deleting upload %s", payload.UploadPath)
	err = os.RemoveAll(payload.UploadPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warnf("⚠️ Warning: Upload file %s was already deleted", payload.UploadPath)
			return true, nil
		}
		return false, fmt.Errorf("failed to delete upload file %s: %w", payload.UploadPath, err)
	}
	logger.Debugf("✅ Successfully deleted upload file: %s", payload.UploadPath)
	return true, nil
}

// getProvider returns the compute provider for the given instance
func (w *Worker) getProvider(providerID models.ProviderID) (compute.Provider, error) {
	w.computeMU.RLock()

	provider, ok := w.providers[providerID]
	if !ok {
		w.computeMU.RUnlock()
		w.computeMU.Lock()
		var err error
		provider, err = compute.NewComputeProvider(providerID)
		if err != nil {
			w.computeMU.Unlock()
			return nil, fmt.Errorf("Worker: Failed to create compute provider for provider %s: %w", providerID, err)
		}
		w.providers[providerID] = provider
		w.computeMU.Unlock()
		w.computeMU.RLock()
	}
	w.computeMU.RUnlock()

	return provider, nil
}

// getProvisioner returns the provisioner for the given instance
func (w *Worker) getProvisioner(providerID models.ProviderID) compute.Provisioner {
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

	return provisioner
}
