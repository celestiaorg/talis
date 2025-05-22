// Package services provides business logic implementation for the API
package services

import (
	"context"
	"encoding/json"
	"fmt"

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

// CreateInstance creates a new instance and a new task to track the instance creation in the DB.
// It returns the created instances and an error if one occurred.
func (s *Instance) CreateInstance(ctx context.Context, instances []types.InstanceRequest) ([]*models.Instance, error) {
	instancesToCreate := make([]*models.Instance, 0, len(instances))
	tasksToCreate := make([]*models.Task, 0, len(instances))

	for _, i := range instances {
		// Validate the instance request
		if err := i.Validate(); err != nil {
			return nil, fmt.Errorf("invalid instance request: %w", err)
		}

		// Get the project
		project, err := s.projectService.GetByName(ctx, i.OwnerID, i.ProjectName)
		if err != nil {
			return nil, fmt.Errorf("failed to get project: %w", err)
		}

		// Sanity check that a user is not trying to create an instance for another user
		// TODO: in the future we should have authorized users for a project
		if i.OwnerID != project.OwnerID {
			return nil, fmt.Errorf("instance owner_id does not match project owner_id")
		}

		for idx := 0; idx < i.NumberOfInstances; idx++ {
			// Create new instance request for task payload
			req := i

			// Construct the actual instance name
			instanceName := req.Name
			if req.Name != "" && req.NumberOfInstances > 1 {
				// Add the instance index to create unique names for multiple instances
				instanceName = fmt.Sprintf("%s-%d", req.Name, idx+1)
				// Add the instance index to the request so the provider
				// can use it to generate the correct name
				req.InstanceIndex = idx
			}

			// Marshal the request to JSON
			payload, err := json.Marshal(req)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal instance request: %w", err)
			}

			tasksToCreate = append(tasksToCreate, &models.Task{
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
				OwnerID:       req.OwnerID,
				ProjectID:     project.ID,
				Name:          instanceName,
				ProviderID:    req.Provider,
				Status:        models.InstanceStatusPending,
				Region:        req.Region,
				Size:          req.Size,
				Tags:          req.Tags,
				VolumeIDs:     []string{},
				VolumeDetails: models.VolumeDetails{},
				PayloadStatus: initialPayloadStatus,
			})
		}
	}

	// Create the instances
	createdInstances, err := s.repo.CreateBatch(ctx, instancesToCreate)
	if err != nil {
		err = fmt.Errorf("failed to add instances to database: %w", err)
		return nil, err
	}

	// Update the task payload with the instance ID
	// This loop assumes a 1:1 mapping between createdInstances and tasksToCreate based on order.
	// This should be safe given how they are populated in parallel.
	for idx, instance := range createdInstances {
		if idx < len(tasksToCreate) { // Boundary check
			// Set the InstanceID on the task model itself
			tasksToCreate[idx].InstanceID = instance.ID

			var taskPayload types.InstanceRequest
			err := json.Unmarshal(tasksToCreate[idx].Payload, &taskPayload)
			if err != nil {
				return createdInstances, fmt.Errorf("failed to unmarshal task payload for instance %d: %w", instance.ID, err)
			}
			taskPayload.InstanceID = instance.ID

			updatedPayload, err := json.Marshal(taskPayload)
			if err != nil {
				return createdInstances, fmt.Errorf("failed to marshal updated task payload for instance %d: %w", instance.ID, err)
			}
			tasksToCreate[idx].Payload = updatedPayload
		}
	}

	// Create the tasks
	if err := s.taskService.CreateBatch(ctx, tasksToCreate); err != nil {
		err = fmt.Errorf("failed to add tasks to database: %w", err)
		return nil, err
	}

	return createdInstances, nil
}

// GetInstance retrieves an instance by ID
func (s *Instance) GetInstance(ctx context.Context, ownerID, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, ownerID, id)
}

// MarkAsTerminated marks an instance as terminated
func (s *Instance) MarkAsTerminated(ctx context.Context, ownerID, instanceID uint) error {
	return s.repo.Terminate(ctx, ownerID, instanceID)
}

// Terminate handles the termination of instances for a given project name and instance IDs.
func (s *Instance) Terminate(ctx context.Context, ownerID uint, projectName string, instanceIDs []uint) error {
	// First verify the project exists and belongs to the owner
	project, err := s.projectService.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return fmt.Errorf("failed to get project '%s': %w", projectName, err)
	}

	if len(instanceIDs) == 0 {
		logger.Infof("No instance IDs provided for termination in project '%s'. Request is a no-op.", projectName)
		return nil
	}

	instancesToTerminate, err := s.repo.GetByProjectIDAndInstanceIDs(ctx, ownerID, project.ID, instanceIDs)
	if err != nil {
		return fmt.Errorf("failed to get instances for project '%s' by IDs: %w", projectName, err)
	}

	if len(instancesToTerminate) == 0 {
		logger.Infof("No active instances found matching the provided IDs for project '%s'. Request is a no-op.", projectName)
		return nil
	}

	// Optional: Log if not all requested IDs were found.
	// To do this accurately, we'd need to compare the found IDs against the requested IDs.
	// For now, we can log a general message if counts don't match.
	if len(instancesToTerminate) < len(uniqueRequestedIDs(instanceIDs)) { // Helper to count unique requested IDs
		logger.Warnf(
			"Project '%s': Not all requested instance IDs were found or matched the project. Requested %d unique IDs, found %d matching instances.",
			projectName,
			len(uniqueRequestedIDs(instanceIDs)),
			len(instancesToTerminate),
		)
	}

	for _, instance := range instancesToTerminate { // instance is models.Instance
		// Create a termination task for the instance
		taskPayload, err := json.Marshal(types.DeleteInstanceRequest{
			InstanceID: instance.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal task payload for instance ID %d: %w", instance.ID, err)
		}
		err = s.taskService.Create(ctx, &models.Task{
			OwnerID:    ownerID,
			ProjectID:  project.ID,
			InstanceID: instance.ID,
			Status:     models.TaskStatusPending,
			Action:     models.TaskActionTerminateInstances,
			Payload:    taskPayload,
		})
		if err != nil {
			return fmt.Errorf("failed to create termination task for instance ID %d: %w", instance.ID, err)
		}
	}
	return nil
}

// helper function to count unique IDs in a slice
func uniqueRequestedIDs(ids []uint) []uint {
	seen := make(map[uint]bool)
	result := []uint{}
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

// addTaskLogs appends logs to the task record.
func (s *Instance) addTaskLogs(ctx context.Context, ownerID uint, task *models.Task, logs string) {
	if task == nil {
		logger.Warnf("Attempted to add logs to a nil task: %s", logs)
		return
	}
	currentTask, err := s.taskService.Get(ctx, ownerID, task.ID)
	if err != nil {
		logger.Errorf("failed to get task %d before adding logs: %v", task.ID, err)
		// Attempt to update with potentially stale task object anyway
		task.Logs += fmt.Sprintf("\n%s", logs)
		if updateErr := s.taskService.Update(ctx, ownerID, task); updateErr != nil {
			logger.Errorf("failed to update task %d with new logs: %v", task.ID, updateErr)
		}
		return
	}

	currentTask.Logs += fmt.Sprintf("\n%s", logs)
	if err := s.taskService.Update(ctx, ownerID, currentTask); err != nil {
		logger.Errorf("failed to update task %d with new logs: %v", task.ID, err)
	}
}

// Get retrieves an instance by ID
func (s *Instance) Get(ctx context.Context, ownerID uint, instanceID uint) (*models.Instance, error) {
	return s.repo.Get(ctx, ownerID, instanceID)
}

// Update updates an instance by ID
func (s *Instance) Update(ctx context.Context, ownerID uint, instanceID uint, instance *models.Instance) error {
	return s.repo.Update(ctx, ownerID, instanceID, instance)
}
