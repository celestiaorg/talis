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
func (s *Instance) CreateInstance(ctx context.Context, instances []types.InstanceRequest) error {
	instancesToCreate := make([]*models.Instance, 0, len(instances))
	tasksToCreate := make([]*models.Task, 0, len(instances))
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
			// Update the name to be unique
			req.Name = fmt.Sprintf("%s-%d", i.Name, idx)

			// Marshal the request to JSON
			payload, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal instance request: %w", err)
			}

			// Generate TaskName internally
			taskName := uuid.New().String()
			tasksToCreate = append(tasksToCreate, &models.Task{
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
				VolumeIDs:     []string{},
				VolumeDetails: models.VolumeDetails{},
				PayloadStatus: initialPayloadStatus,
			})
		}
	}

	// Create the instances
	if err := s.repo.CreateBatch(ctx, instancesToCreate); err != nil {
		err = fmt.Errorf("failed to add instances to database: %w", err)
		// TODO: https://github.com/celestiaorg/talis/issues/246
		return err
	}

	// Update the task payload with the instance ID
	for idx, instance := range instancesToCreate {
		// unmarshal the corresponding task payload
		var taskPayload types.InstanceRequest
		err := json.Unmarshal(tasksToCreate[idx].Payload, &taskPayload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal task payload: %w", err)
		}
		taskPayload.InstanceID = instance.ID

		// marshal the updated task payload
		updatedPayload, err := json.Marshal(taskPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal task payload: %w", err)
		}
		tasksToCreate[idx].Payload = updatedPayload
	}

	// Create the tasks
	if err := s.taskService.CreateBatch(ctx, tasksToCreate); err != nil {
		err = fmt.Errorf("failed to add tasks to database: %w", err)
		return err
	}

	return nil
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
func (s *Instance) Terminate(ctx context.Context, ownerID uint, projectName string, instanceNames []string) error {
	// First verify the project exists and belongs to the owner
	project, err := s.projectService.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get instances that belong to this project and match the provided names
	instances, err := s.repo.GetByProjectIDAndInstanceNames(ctx, ownerID, project.ID, instanceNames)
	if err != nil {
		return fmt.Errorf("failed to get instances: %w", err)
	}
	// Verify we found all requested instances
	if len(instances) == 0 {
		logger.Infof("No active instances found with the specified names for project '%s', request is a no-op", projectName)
		return nil
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
	}

	for _, instance := range instances {
		// Create a termination task for the instance
		taskName := uuid.New().String()
		taskPayload, err := json.Marshal(types.DeleteInstanceRequest{
			InstanceID: instance.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal task payload: %w", err)
		}
		err = s.taskService.Create(ctx, &models.Task{
			Name:      taskName,
			OwnerID:   ownerID,
			ProjectID: project.ID,
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
