// Package services provides business logic implementation for the API
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

// Instance provides business logic for instance operations
type Instance struct {
	repo           *repos.InstanceRepository
	taskService    *Task
	projectService *Project
}

// NewInstanceService creates a new instance service instance
func NewInstanceService(repo *repos.InstanceRepository, taskService *Task, projectService *Project) *Instance {
	return &Instance{
		repo:           repo,
		taskService:    taskService,
		projectService: projectService,
	}
}

// ListInstances retrieves a paginated list of instances
func (s *Instance) ListInstances(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Instance, error) {
	return s.repo.List(ctx, ownerID, opts)
}

// CreateInstance creates a new task to track instance creation and starts the process.
func (s *Instance) CreateInstance(ctx context.Context, instances []types.InstanceRequest) error {

	instancesToCreate := make([]*models.Instance, 0, len(instances))
	taskToCreate := make([]*models.Task, 0, len(instances))
	for _, i := range instances {
		// Validate the instance request
		if err := i.Validate(); err != nil {
			return fmt.Errorf("invalid instance request: %w", err)
		}

		// Get the project
		project, err := s.projectService.GetByName(ctx, i.OwnerID, i.ProjectName)
		if err != nil {
			return fmt.Errorf("failed to get project: %w", err)
		}

		// Sanity check that a user is not trying to create an instance for another user
		// TODO: in the future we should have authorized users for a project
		if i.OwnerID != project.OwnerID {
			return fmt.Errorf("instance owner_id does not match project owner_id")
		}

		for idx := 0; idx < i.NumberOfInstances; idx++ {
			// Create new instance request for task payload
			req := i
			req.Name = fmt.Sprintf("%s-%d", i.Name, idx)
			req.Action = "create"

			// Marshal the request to JSON
			payload, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal instance request: %w", err)
			}

			// Generate TaskName internally
			taskName := uuid.New().String()
			taskToCreate = append(taskToCreate, &models.Task{
				Name:      taskName,
				OwnerID:   i.OwnerID,
				ProjectID: project.ID,
				Status:    models.TaskStatusPending,
				Action:    models.TaskActionCreateInstances,
				Payload:   payload,
			})

			// Determine initial payload status
			initialPayloadStatus := models.PayloadStatusNone
			if i.PayloadPath != "" {
				initialPayloadStatus = models.PayloadStatusPendingCopy
			}

			instancesToCreate = append(instancesToCreate, &models.Instance{
				Name:          req.Name,
				OwnerID:       req.OwnerID,
				ProjectID:     project.ID,
				ProviderID:    req.Provider,
				Status:        models.InstanceStatusPending,
				Region:        req.Region,
				Size:          req.Size,
				Tags:          req.Tags,
				Volumes:       []string{},
				VolumeDetails: models.VolumeDetails{},
				PayloadStatus: initialPayloadStatus,
			})

		}
	}

	// Create the tasks
	if err := s.taskService.CreateBatch(ctx, taskToCreate); err != nil {
		err = fmt.Errorf("failed to add tasks to database: %w", err)
		return err
	}

	// Map the task IDs to the instances
	for idx, task := range taskToCreate {
		instancesToCreate[idx].LastTaskID = task.ID
	}

	// Create the instances
	if err := s.repo.CreateBatch(ctx, instancesToCreate); err != nil {
		err = fmt.Errorf("failed to add instances to database: %w", err)
		// TODO: Delete the tasks that were created
		return err
	}

	// TODO: remove these since the background worker would do this
	// s.addTaskLogs(ctx, ownerID, task, fmt.Sprintf("Created %d instances in database", len(instancesToCreate)))
	// s.updateTaskStatus(ctx, ownerID, task.ID, models.TaskStatusRunning)

	// Start provisioning in background
	// TODO: remove as this will be handled by the background worker
	go s.provisionInstances(ctx, instancesToCreate)

	return nil
}

// GetInstance retrieves an instance by ID
func (s *Instance) GetInstance(ctx context.Context, ownerID, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, ownerID, id)
}

// updateInstanceVolumes updates the volumes and volume details for an instance
func (s *Instance) updateInstanceVolumes(
	ctx context.Context,
	instanceName string,
	volumes []string,
	volumeDetails []types.VolumeDetails,
) error {
	// Get the instance first to get its ownerID
	instance, err := s.repo.GetByName(ctx, models.AdminID, instanceName)
	if err != nil {
		return fmt.Errorf("failed to get instance %s: %w", instanceName, err)
	}

	logger.Debugf("🔄 Converting volume details for instance %s", instanceName)
	logger.Debugf("📥 Input data:")
	logger.Debugf("  - Volumes: %v", volumes)
	logger.Debugf("  - Volume Details: %+v", volumeDetails)

	// Convert volume details to database model
	dbVolumeDetails := make(models.VolumeDetails, 0, len(volumeDetails))
	for _, vd := range volumeDetails {
		dbVolumeDetails = append(dbVolumeDetails, models.VolumeDetail{
			ID:         vd.ID,
			Name:       vd.Name,
			Region:     vd.Region,
			SizeGB:     vd.SizeGB,
			MountPoint: vd.MountPoint,
		})
	}

	logger.Debugf("📦 Preparing to update instance %s", instanceName)
	logger.Debugf("📝 Data to update:")
	logger.Debugf("  - Volumes: %#v", volumes)
	logger.Debugf("  - Volume Details: %#v", dbVolumeDetails)

	// Create update instance with only the fields we want to update
	updateInstance := &models.Instance{
		Volumes:       volumes,
		VolumeDetails: dbVolumeDetails,
	}

	// Update instance in database using a transaction to ensure atomicity
	err = s.repo.UpdateByName(ctx, instance.OwnerID, instanceName, updateInstance)
	if err != nil {
		logger.Errorf("❌ Failed to update instance %s volumes: %v", instanceName, err)
		return fmt.Errorf("failed to update instance %s volumes: %w", instanceName, err)
	}

	// Verify the update immediately
	updatedInstance, err := s.repo.GetByName(ctx, instance.OwnerID, instanceName)
	if err != nil {
		logger.Warnf("⚠️ Could not verify volume update for instance %s: %v", instanceName, err)
		return nil // Don't return error here as the update might have succeeded
	}

	logger.Debugf("✅ Verified volumes update for instance %s:", instanceName)
	logger.Debugf("📊 Database state after update:")
	logger.Debugf("  - Volumes: %#v", updatedInstance.Volumes)
	logger.Debugf("  - Volume Details: %#v", updatedInstance.VolumeDetails)

	// Verify data integrity
	if len(updatedInstance.Volumes) != len(volumes) {
		logger.Warnf("⚠️ Volume count mismatch - Expected: %d, Got: %d", len(volumes), len(updatedInstance.Volumes))
	}
	if len(updatedInstance.VolumeDetails) != len(dbVolumeDetails) {
		logger.Warnf("⚠️ Volume details count mismatch - Expected: %d, Got: %d", len(dbVolumeDetails), len(updatedInstance.VolumeDetails))
	}

	return nil
}

// provisionInstances provisions the instances. To not block the main thread, it should be called in a goroutine.
// NOTE: Currently this assumes that the instances are for the same provider.
func (s *Instance) provisionInstances(ctx context.Context, ownerID, taskID uint, instances []types.InstanceRequest) {
		// Get task details
		task, err := s.taskService.GetByID(ctx, ownerID, taskID)
		if err != nil {
			logger.Errorf("❌ Failed to get task details for taskID %d: %v", taskID, err)
			return
		}

		// Create infrastructure client
		infraReq := &types.InstancesRequest{
			TaskName:  task.Name,
			Instances: instances,
			Action:    "create",
			Provider:  instances[0].Provider,
		}

		infra, err := NewInfrastructure(infraReq)
		if err != nil {
			err = fmt.Errorf("❌ failed to create infrastructure client: %w", err)
			logger.Error(err)
			s.updateTaskError(ctx, ownerID, task, err)
			return
		}

		// Execute infrastructure creation
		result, err := infra.Execute()
		if err != nil {
			err = fmt.Errorf("❌ failed to create infrastructure: %w", err)
			logger.Error(err)
			s.updateTaskError(ctx, ownerID, task, err)
			return
		}

		// Type assert the result
		pInstances, ok := result.([]types.InstanceInfo)
		if !ok {
			err = fmt.Errorf("❌ Invalid result type: %T", result)
			logger.Error(err)
			s.updateTaskError(ctx, ownerID, task, err)
			return
		}

		logger.Debugf("📝 Created instances: %+v", pInstances)

		// Update instances with IP and status
		// Update instance information in database
		ownerID := instances[0].OwnerID
		for _, instance := range pInstances {
			logger.Debugf("🔄 Processing instance update for %s", instance.Name)
			logger.Debugf("  - Volumes: %v", instance.Volumes)
			logger.Debugf("  - Volume Details: %+v", instance.VolumeDetails)

			// Update volumes first if present
			if len(instance.Volumes) > 0 || len(instance.VolumeDetails) > 0 {
				logger.Debugf("🔄 Updating volumes for instance %s", instance.Name)
				if err := s.updateInstanceVolumes(ctx, instance.Name, instance.Volumes, instance.VolumeDetails); err != nil {
					err = fmt.Errorf("❌ Failed to update volumes for instance %s: %w", instance.Name, err)
					logger.Error(err)
					s.addTaskLogs(ctx, ownerID, task, err.Error())
					continue
				}
				logger.Debugf("✅ Successfully updated volumes for instance %s", instance.Name)
				s.addTaskLogs(ctx, ownerID, task, fmt.Sprintf("Updated volumes for instance %s", instance.Name))

				// Verify the update was successful
				updatedInstance, err := s.repo.GetByName(ctx, ownerID, instance.Name)
				if err != nil {
					errMsg := fmt.Sprintf("❌ Failed to verify volume update for instance %s: %v", instance.Name, err)
					logger.Error(errMsg)
					s.addTaskLogs(ctx, ownerID, task, errMsg)
					continue
				}

				logger.Debugf("📊 Current instance state after volume update:")
				logger.Debugf("  - Volumes: %v", updatedInstance.Volumes)
				logger.Debugf("  - Volume Details: %+v", updatedInstance.VolumeDetails)
				s.addTaskLogs(ctx, ownerID, task, fmt.Sprintf("Verified volume update for instance %s", instance.Name))
			}

			// Then update instance status and IP
			updateInstance, err := s.repo.GetByName(ctx, ownerID, instance.Name)
			if err != nil {
				errMsg := fmt.Sprintf("❌ Failed to get instance %s: %v", instance.Name, err)
				logger.Error(errMsg)
				s.addTaskLogs(ctx, ownerID, task, errMsg)
				continue
			}

			updateInstance.PublicIP = instance.PublicIP
			updateInstance.Status = models.InstanceStatusReady

			if err := s.repo.UpdateByID(ctx, ownerID, updateInstance.ID, updateInstance); err != nil {
				errMsg := fmt.Sprintf("❌ Failed to update instance %s: %v", instance.Name, err)
				logger.Error(errMsg)
				s.addTaskLogs(ctx, ownerID, task, errMsg)
				continue
			}
			logger.Debugf("✅ Updated instance %s with IP %s and status ready", instance.Name, instance.PublicIP)
		}

		// Start Ansible provisioning if requested
		if instances[0].Provision {
			s.addTaskLogs(ctx, ownerID, task, "Running Ansible provisioning")
			if err := infra.RunProvisioning(pInstances); err != nil {
				err = fmt.Errorf("❌ Failed to run provisioning: %w", err)
				logger.Error(err)
				s.updateTaskError(ctx, ownerID, task, err)
				return
			}
			s.addTaskLogs(ctx, ownerID, task, "Ansible provisioning completed")
			// Update payload status for instances with payloads
			for _, instance := range pInstances {
				if instance.PayloadPath == "" {
					continue
				}
				updateInstance := &models.Instance{
					PayloadStatus: models.PayloadStatusExecuted,
				}

				if err := s.repo.UpdateByName(ctx, ownerID, instance.Name, updateInstance); err != nil {
					logger.Errorf("❌ Failed to update payload status for instance %s: %v", instance.Name, err)
					continue
				}
				logger.Debugf("✅ Updated payload status to executed for instance %s", instance.Name)
			}
		}

		s.updateTaskStatus(ctx, ownerID, task.ID, models.TaskStatusCompleted)
		logger.Debugf("✅ Infrastructure creation completed for task %s", task.Name)
}

// Terminate handles the termination of instances for a given project name and instance names.
func (s *Instance) Terminate(ctx context.Context, ownerID uint, projectName string, instanceNames []string) (taskName string, err error) {
	// First verify the project exists and belongs to the owner
	project, err := s.projectService.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	// Get instances that belong to this project and match the provided names
	instances, err := s.repo.GetByProjectIDAndInstanceNames(ctx, ownerID, project.ID, instanceNames)
	if err != nil {
		return "", fmt.Errorf("failed to get instances: %w", err)
	}

	taskName = uuid.New().String()
	err = s.taskService.Create(ctx, &models.Task{
		Name:      taskName,
		OwnerID:   ownerID,
		ProjectID: project.ID,
		Status:    models.TaskStatusPending,
		Action:    models.TaskActionTerminateInstances,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	task, err := s.taskService.GetByName(ctx, ownerID, taskName)
	if err != nil {
		return "", fmt.Errorf("failed to get task: %w", err)
	}

	// Verify we found all requested instances
	if len(instances) == 0 {
		logger.Infof("No active instances found with the specified names for project '%s', request is a no-op", projectName)
		s.updateTaskStatus(ctx, ownerID, task.ID, models.TaskStatusCompleted)
		return taskName, nil
	}
	if len(instances) != len(instanceNames) {
		// Some instances were not found, log which ones
		foundNames := make(map[string]bool)
		for _, instance := range instances {
			foundNames[instance.Name] = true
		}
		var missingNames []string
		for _, name := range instanceNames {
			if !foundNames[name] {
				missingNames = append(missingNames, name)
			}
		}
		logMsg := fmt.Sprintf("Some instances were not found or are already deleted for project '%s': %v", projectName, missingNames)
		logger.Infof("%s", logMsg)
		s.addTaskLogs(ctx, ownerID, task, logMsg)
	}

	s.terminate(ctx, ownerID, task.ID, instances)
	return taskName, nil
}

// deleteRequest represents a single instance deletion request with tracking
type deleteRequest struct {
	instance     models.Instance
	infraRequest *types.InstancesRequest
	attempts     int
	lastError    error
	maxAttempts  int
}

// deletionResults tracks the overall results of the deletion operation
type deletionResults struct {
	successful []string         // names of successfully deleted instances
	failed     map[string]error // instance name -> last error for failed deletions
}

// terminate handles the infrastructure deletion process
func (s *Instance) terminate(ctx context.Context, ownerID, taskID uint, instances []models.Instance) {
	go func() {
		task, err := s.taskService.GetByID(ctx, ownerID, taskID)
		if err != nil {
			logger.Errorf("❌ Failed to get task details for taskID %d: %v", taskID, err)
			return
		}

		if len(instances) == 0 {
			err := fmt.Errorf("❌ No instances found to terminate")
			logger.Error(err)
			s.updateTaskError(ctx, ownerID, task, err)
			return
		}

		// Create queue of delete requests
		queue := make([]*deleteRequest, 0, len(instances))
		for _, instance := range instances {
			logMsg := fmt.Sprintf("🗑️ Attempting to terminate instance: %s", instance.Name)
			logger.Infof("%s", logMsg)
			s.addTaskLogs(ctx, ownerID, task, logMsg)

			// Create a new infrastructure request for each instance
			infraReq := &types.InstancesRequest{
				TaskName: task.Name,
				Instances: []types.InstanceRequest{
					{
						Name:     instance.Name,
						Provider: instance.ProviderID,
						Region:   instance.Region,
						Size:     instance.Size,
					},
				},
				Action: "delete",
			}

			// TODO: a hacky fix, but has to be addressed properly in another PR
			if infraReq.Provider == "" {
				infraReq.Provider = infraReq.Instances[0].Provider
			}

			queue = append(queue, &deleteRequest{
				instance:     instance,
				infraRequest: infraReq,
				maxAttempts:  10,
			})
		}

		results := &deletionResults{
			successful: make([]string, 0),
			failed:     make(map[string]error),
		}

		// Process queue until empty or all requests have failed
		defaultErrorSleep := 100 * time.Millisecond
		ownerID := instances[0].OwnerID
	REQUESTLOOP:
		for len(queue) > 0 {
			select {
			case <-ctx.Done():
				// Context cancelled, mark remaining requests as failed
				for _, req := range queue {
					results.failed[req.instance.Name] = fmt.Errorf("operation cancelled: %w", ctx.Err())
				}
				queue = nil // Clear the queue
				break REQUESTLOOP
			default:
			}

			request := queue[0]
			queue = queue[1:] // pop from front

			// Skip if max attempts reached
			if request.attempts >= request.maxAttempts {
				results.failed[request.instance.Name] = fmt.Errorf("max attempts reached (%d): %w", request.maxAttempts, request.lastError)
				continue
			}

			request.attempts++

			// Try to delete infrastructure
			infra, err := NewInfrastructure(request.infraRequest)
			if err != nil {
				request.lastError = fmt.Errorf("failed to create infrastructure client: %w", err)
				queue = append(queue, request) // add back to queue
				time.Sleep(defaultErrorSleep)
				continue
			}

			_, err = infra.Execute()
			// If error exists and it's not a "not found" error, add back to queue
			if err != nil && (!strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found")) {
				request.lastError = fmt.Errorf("failed to delete infrastructure: %w", err)
				queue = append(queue, request) // add back to queue
				time.Sleep(defaultErrorSleep)
				continue
			}

			// Try to update database
			if err := s.repo.Terminate(ctx, ownerID, request.instance.ID); err != nil {
				request.lastError = fmt.Errorf("failed to terminate in database: %w", err)
				queue = append(queue, request) // add back to queue
				time.Sleep(defaultErrorSleep)
				continue
			}

			// Successfully deleted both infrastructure and database record
			results.successful = append(results.successful, request.instance.Name)
		}

		// Log final results
		if len(results.successful) > 0 {
			logMsg := fmt.Sprintf("Successfully deleted instances: %v", results.successful)
			logger.Infof("%s", logMsg)
			s.addTaskLogs(ctx, ownerID, task, logMsg)
		}
		if len(results.failed) > 0 {
			logMsg := "Failed to delete instances:"
			logger.Infof("%s", logMsg)
			s.addTaskLogs(ctx, ownerID, task, logMsg)
			for name, err := range results.failed {
				logMsg := fmt.Sprintf("  %s: %v", name, err)
				logger.Infof("%s", logMsg)
				s.addTaskLogs(ctx, ownerID, task, logMsg)
			}
		}

		// Create deletion result for API response
		deletionResult := map[string]interface{}{
			"status":    "completed",
			"deleted":   results.successful,
			"failed":    results.failed,
			"completed": time.Now().UTC(),
		}
		task.Result, err = json.Marshal(deletionResult)
		if err != nil {
			logger.Errorf("failed to marshal deletion result: %v", err)
		}
		task.Status = models.TaskStatusCompleted

		// Update final status with result
		if err := s.taskService.Update(ctx, ownerID, task); err != nil {
			logger.Errorf("failed to update task: %v", err)
		}
	}()
}

func (s *Instance) updateTaskError(ctx context.Context, ownerID uint, task *models.Task, err error) {
	if err != nil {
		task.Error += fmt.Sprintf("\n%s", err.Error())
		task.Status = models.TaskStatusFailed
	}
	if err := s.taskService.Update(ctx, ownerID, task); err != nil {
		logger.Errorf("failed to update task: %v", err)
	}
}

func (s *Instance) addTaskLogs(ctx context.Context, ownerID uint, task *models.Task, logs string) {
	task.Logs += fmt.Sprintf("\n%s", logs)
	if err := s.taskService.Update(ctx, ownerID, task); err != nil {
		logger.Errorf("failed to update task: %v", err)
	}
}

func (s *Instance) updateTaskStatus(ctx context.Context, ownerID uint, taskID uint, status models.TaskStatus) {
	if err := s.taskService.UpdateStatus(ctx, ownerID, taskID, status); err != nil {
		logger.Errorf("failed to update task status: %v", err)
	}
}
