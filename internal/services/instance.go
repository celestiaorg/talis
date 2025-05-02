// Package services provides business logic implementation for the API
package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/pkg/types"
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
// It returns a list of task names created (or attempted) and an error if one occurred.
func (s *Instance) CreateInstance(ctx context.Context, instances []types.InstanceRequest) ([]string, error) {
	instancesToCreate := make([]*models.Instance, 0, len(instances))
	tasksToCreate := make([]*models.Task, 0, len(instances))
	taskNames := make([]string, 0, len(instances)) // Pre-allocate slice for task names

	for _, i := range instances {
		// Validate the instance request
		if err := i.Validate(); err != nil {
			return nil, fmt.Errorf("invalid instance request: %w", err)
		}

		for idx := 0; idx < i.NumberOfInstances; idx++ {
			// Create new instance request for task payload
			req := i
			// Update the name to be unique
			req.Name = fmt.Sprintf("%s-%d", i.Name, idx)

			// Marshal the request to JSON
			payload, err := json.Marshal(req)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal instance request: %w", err)
			}

			// Generate TaskName internally
			taskName := uuid.New().String()
			taskNames = append(taskNames, taskName) // Collect task name
			tasksToCreate = append(tasksToCreate, &models.Task{
				Name:      taskName,
				OwnerID:   i.OwnerID,
				ProjectID: i.ProjectID,
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
				ProjectID:     req.ProjectID,
				ProviderID:    req.Provider,
				Status:        models.InstanceStatusPending,
				Region:        req.Region,
				Size:          req.Size,
				Tags:          req.Tags,
				VolumeIDs:     []string{},
				VolumeDetails: models.VolumeDetails{}, // <<< Using pkg/models alias again
				PayloadStatus: initialPayloadStatus,
			})
		}
	}

	// Create the instances
	if err := s.repo.CreateBatch(ctx, instancesToCreate); err != nil {
		err = fmt.Errorf("failed to add instances to database: %w", err)
		// TODO: https://github.com/celestiaorg/talis/issues/246
		return nil, err
	}

	// Update the task payload with the instance ID
	for idx, instance := range instancesToCreate {
		// unmarshal the corresponding task payload
		var taskPayload types.InstanceRequest
		err := json.Unmarshal(tasksToCreate[idx].Payload, &taskPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal task payload: %w", err)
		}
		taskPayload.InstanceID = instance.ID

		// marshal the updated task payload
		updatedPayload, err := json.Marshal(taskPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal task payload: %w", err)
		}
		tasksToCreate[idx].Payload = updatedPayload
	}

	// Create the tasks
	if err := s.taskService.CreateBatch(ctx, tasksToCreate); err != nil {
		err = fmt.Errorf("failed to add tasks to database: %w", err)
		return nil, err
	}

	return taskNames, nil
}

// GetInstance retrieves an instance by ID
func (s *Instance) GetInstance(ctx context.Context, ownerID, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, ownerID, id)
}

// MarkAsTerminated marks an instance as terminated
func (s *Instance) MarkAsTerminated(ctx context.Context, ownerID, instanceID uint) error {
	return s.repo.Terminate(ctx, ownerID, instanceID)
}

// Terminate handles the termination of instances for a given project name and instance names.
func (s *Instance) Terminate(ctx context.Context, ownerID uint, projectID uint, instanceIDs []uint) error {
	// Verify we found all requested instances
	if len(instanceIDs) == 0 {
		logger.Infof("No active instances found with the specified ids for project '%d', request is a no-op", projectID)
		return nil
	}

	for _, instanceID := range instanceIDs {
		// Create a termination task for the instance
		taskName := uuid.New().String()
		taskPayload, err := json.Marshal(types.DeleteInstanceRequest{
			InstanceID: instanceID,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal task payload: %w", err)
		}
		err = s.taskService.Create(ctx, &models.Task{
			Name:      taskName,
			OwnerID:   ownerID,
			ProjectID: projectID,
			Status:    models.TaskStatusPending,
			Action:    models.TaskActionTerminateInstances,
			Payload:   taskPayload,
		})
		if err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}
	}
	return nil
}

// TerminateInstance handles the termination of a single instance
func (s *Instance) TerminateInstance(ctx context.Context, ownerID uint, projectID uint, instanceID uint) error {
	// Create a termination task for the instance
	taskName := uuid.New().String()
	taskPayload, err := json.Marshal(types.DeleteInstanceRequest{
		InstanceID: instanceID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	err = s.taskService.Create(ctx, &models.Task{
		Name:      taskName,
		OwnerID:   ownerID,
		ProjectID: projectID,
		Status:    models.TaskStatusPending,
		Action:    models.TaskActionTerminateInstances,
		Payload:   taskPayload,
	})
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

// addTaskLogs appends logs to the task record.
func (s *Instance) addTaskLogs(ctx context.Context, ownerID uint, task *models.Task, logs string) {
	if task == nil {
		logger.Warnf("Attempted to add logs to a nil task: %s", logs)
		return
	}
	currentTask, err := s.taskService.GetByID(ctx, ownerID, task.ID)
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

// GetByName retrieves an instance by name
func (s *Instance) GetByName(ctx context.Context, ownerID uint, name string) (*models.Instance, error) {
	return s.repo.GetByName(ctx, ownerID, name)
}

// GetByID retrieves an instance by ID
func (s *Instance) GetByID(ctx context.Context, ownerID uint, instanceID uint) (*models.Instance, error) {
	return s.repo.GetByID(ctx, ownerID, instanceID)
}

// GetByProjectIDAndInstanceNames retrieves instances by project ID and instance names
func (s *Instance) GetByProjectIDAndInstanceNames(ctx context.Context, ownerID uint, projectID uint, names []string) ([]models.Instance, error) {
	return s.repo.GetByProjectIDAndInstanceNames(ctx, ownerID, projectID, names)
}

// UpdateByName updates an instance by name
func (s *Instance) UpdateByName(ctx context.Context, ownerID uint, name string, instance *models.Instance) error {
	return s.repo.UpdateByName(ctx, ownerID, name, instance)
}

// UpdateByID updates an instance by ID
func (s *Instance) UpdateByID(ctx context.Context, ownerID uint, instanceID uint, instance *models.Instance) error {
	return s.repo.UpdateByID(ctx, ownerID, instanceID, instance)
}
