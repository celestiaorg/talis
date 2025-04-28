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
func (s *Instance) CreateInstance(ctx context.Context, ownerID uint, projectName string, instances []types.InstanceRequest) (string, error) {
	project, err := s.projectService.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	logger.Debugf("🔄 Creating instances for project %s", projectName)
	logger.Debugf("🔄 Instances: %+v", instances)

	// validate the instances array is not empty
	if len(instances) == 0 {
		return "", fmt.Errorf("at least one instance is required")
	}

	// Generate TaskName internally
	taskName := uuid.New().String()
	err = s.taskService.Create(ctx, &models.Task{
		Name:      taskName,
		OwnerID:   ownerID,
		ProjectID: project.ID,
		Status:    models.TaskStatusPending,
		Action:    models.TaskActionCreateInstances,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	// Re-fetch task to get the ID and use it for logging/error updates
	task, err := s.taskService.GetByName(ctx, ownerID, taskName)
	if err != nil {
		// Log error, but maybe creation can continue? Depends on whether task ID is strictly needed later.
		logger.Errorf("Failed to get task %s immediately after creation: %v", taskName, err)
		// Surface the failure on the task itself so users see the problem.
		s.updateTaskError(ctx, ownerID, &models.Task{ // safe: we know Name & OwnerID
			Name:    taskName,
			OwnerID: ownerID,
		}, fmt.Errorf("failed to fetch task after creation: %w", err))
		return "", fmt.Errorf("failed to get task %s after creation: %w", taskName, err)
	}

	instancesToCreate := make([]*models.Instance, 0, len(instances))
	instanceNameToOwnerID := make(map[string]uint)

	for _, i := range instances {
		// Determine the OwnerID to be stored in the database for this instance
		dbOwnerID := ownerID // Default to the authenticated user's ID
		if i.OwnerID != 0 {
			if ownerID != 0 && i.OwnerID != ownerID {
				s.updateTaskError(ctx, ownerID, task, fmt.Errorf("instance owner_id %d does not match project owner_id %d", i.OwnerID, ownerID))
				return taskName, fmt.Errorf("instance owner_id %d does not match project owner_id %d", i.OwnerID, ownerID)
			}
			dbOwnerID = i.OwnerID // Override with the ID specified in the request item
		} else if ownerID == 0 {
			// Should not happen if request validation requires OwnerID OR JWT
			s.updateTaskError(ctx, ownerID, task, fmt.Errorf("instance owner_id is required when request has no authenticated user"))
			return taskName, fmt.Errorf("instance owner_id is required when request has no authenticated user")
		}

		baseName := i.Name
		if baseName == "" {
			baseName = fmt.Sprintf("instance-%s", uuid.New().String()[:8]) // Shorter UUID
		}

		numInstances := i.NumberOfInstances
		if numInstances < 1 {
			numInstances = 1
		}

		for idx := 0; idx < numInstances; idx++ {
			instanceName := baseName
			if numInstances > 1 {
				instanceName = fmt.Sprintf("%s-%d", baseName, idx)
			}

			initialPayloadStatus := models.PayloadStatusNone
			if i.PayloadPath != "" {
				initialPayloadStatus = models.PayloadStatusPendingCopy
			}

			instanceNameToOwnerID[instanceName] = dbOwnerID

			instancesToCreate = append(instancesToCreate, &models.Instance{
				Name:          instanceName,
				OwnerID:       dbOwnerID,
				ProjectID:     project.ID,
				LastTaskID:    task.ID,
				ProviderID:    i.Provider,
				Status:        models.InstanceStatusPending,
				Region:        i.Region,
				Size:          i.Size,
				Tags:          i.Tags,
				Volumes:       []string{},
				VolumeDetails: models.VolumeDetails{},
				PayloadStatus: initialPayloadStatus,
			})
		}
	}

	if err := s.repo.CreateBatch(ctx, instancesToCreate); err != nil {
		err = fmt.Errorf("failed to add instances to database: %w", err)
		s.updateTaskError(ctx, ownerID, task, err)
		return taskName, err
	}

	s.addTaskLogs(ctx, ownerID, task, fmt.Sprintf("Created %d instances in database", len(instancesToCreate)))
	s.updateTaskStatus(ctx, ownerID, task.ID, models.TaskStatusRunning)

	// Start provisioning in background
	go s.provisionInstances(ctx, ownerID, task.ID, instances, instanceNameToOwnerID)

	return taskName, nil
}

// GetInstance retrieves an instance by ID
func (s *Instance) GetInstance(ctx context.Context, ownerID, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, ownerID, id)
}

// updateInstanceVolumes updates the volumes and volume details for an instance
func (s *Instance) updateInstanceVolumes(
	ctx context.Context,
	ownerID uint,
	instanceID uint,
	volumes []string,
) error {
	instance, err := s.repo.Get(ctx, ownerID, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance %d for owner %d: %w", instanceID, ownerID, err)
	}
	instanceName := instance.Name

	logger.Debugf("🔄 Converting volume details for instance %s (ID: %d)", instanceName, instanceID)
	logger.Debugf("📥 Input data:")
	logger.Debugf("  - Volumes: %v", volumes)

	logger.Debugf("📦 Preparing to update instance %s", instanceName)
	logger.Debugf("📝 Data to update:")
	logger.Debugf("  - Volumes: %#v", volumes)

	// Create update instance with only the fields we want to update
	updateData := &models.Instance{
		Volumes: volumes,
	}

	logger.Debugf("📦 Preparing to update volumes for instance %s (ID: %d): Volumes=%#v", instanceName, instanceID, volumes)
	err = s.repo.UpdateByID(ctx, ownerID, instanceID, updateData)
	if err != nil {
		err = fmt.Errorf("failed to update instance %s (ID: %d) volumes: %w", instanceName, instanceID, err)
		logger.Errorf("❌ Failed to update instance %s (ID: %d) volumes: %v", instanceName, instanceID, err)
		return err
	}

	return nil
}

// findPendingInstanceForTask attempts to find the unique database instance record
// corresponding to a given instance name for a specific task, ensuring it's in Pending state.
// It uses GetByName first and falls back to listing all instances for the owner if necessary.
func (s *Instance) findPendingInstanceForTask(
	ctx context.Context,
	task *models.Task,
	dbOwnerID uint,
	instanceName string,
) (*models.Instance, error) {
	var identifiedInstance *models.Instance
	foundCorrectInstance := false
	callerOwnerID := task.OwnerID // Owner who initiated the task

	// Attempt 1: GetByName
	getInstance, err := s.repo.GetByName(ctx, dbOwnerID, instanceName)
	if err == nil {
		if getInstance.LastTaskID == task.ID && getInstance.Status == models.InstanceStatusPending {
			identifiedInstance = getInstance
			foundCorrectInstance = true
			logger.Debugf("Instance %s (ID: %d) identified via GetByName for Task %d", instanceName, identifiedInstance.ID, task.ID)
		} else {
			warnMsg := fmt.Sprintf("⚠️ GetByName returned instance %s (ID: %d) with wrong state (TaskID: %d/%d, Status: %d/%d). Attempting fallback List.",
				instanceName, getInstance.ID, getInstance.LastTaskID, task.ID, getInstance.Status, models.InstanceStatusPending)
			logger.Warnf("⚠️ GetByName returned instance %s (ID: %d) with wrong state (TaskID: %d/%d, Status: %d/%d). Attempting fallback List.", instanceName, getInstance.ID, getInstance.LastTaskID, task.ID, getInstance.Status, models.InstanceStatusPending)
			s.addTaskLogs(ctx, callerOwnerID, task, warnMsg)
		}
	} else {
		errMsg := fmt.Sprintf("ℹ️ GetByName failed for instance %s (Owner: %d): %v. Will attempt fallback List.", instanceName, dbOwnerID, err)
		logger.Warnf("ℹ️ GetByName failed for instance %s (Owner: %d): %v. Will attempt fallback List.", instanceName, dbOwnerID, err)
		s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
	}

	// Attempt 2: Fallback List
	if !foundCorrectInstance {
		logger.Debugf("Attempting fallback List search for instance %s (Owner: %d, Task: %d)", instanceName, dbOwnerID, task.ID)
		allInstances, listErr := s.repo.List(ctx, dbOwnerID, nil)
		if listErr != nil {
			errMsg := fmt.Sprintf("❌ Fallback List search failed for owner %d: %v", dbOwnerID, listErr)
			logger.Errorf("❌ Fallback List search failed for owner %d: %v", dbOwnerID, listErr)
			s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
			return nil, fmt.Errorf("fallback List search failed for owner %d: %w", dbOwnerID, listErr)
		}
		for i := range allInstances {
			inst := &allInstances[i]
			if inst.Name == instanceName && inst.LastTaskID == task.ID && inst.Status == models.InstanceStatusPending {
				identifiedInstance = inst
				foundCorrectInstance = true
				logger.Infof("✅ Fallback List search successful for instance %s (ID: %d, Task: %d)", instanceName, identifiedInstance.ID, task.ID)
				break
			}
		}
	}

	if !foundCorrectInstance {
		errMsg := fmt.Sprintf("❌ Failed to identify unique pending DB instance for provisioned resource %s (Owner: %d, Task: %d).",
			instanceName, dbOwnerID, task.ID)
		logger.Errorf("❌ Failed to identify unique pending DB instance for provisioned resource %s (Owner: %d, Task: %d).", instanceName, dbOwnerID, task.ID)
		s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
		return nil, fmt.Errorf("failed to identify unique pending DB instance for provisioned resource %s (Owner: %d, Task: %d)", instanceName, dbOwnerID, task.ID)
	}

	return identifiedInstance, nil
}

// provisionInstances handles the background process of provisioning infrastructure,
// updating instance details (volumes, IP, status), and running optional provisioning.
func (s *Instance) provisionInstances(
	ctx context.Context,
	callerOwnerID, taskID uint,
	instancesReq []types.InstanceRequest,
	instanceNameToOwnerID map[string]uint,
) {
	// Fetch the task using the caller's ID
	task, err := s.taskService.GetByID(ctx, callerOwnerID, taskID)
	if err != nil {
		logger.Errorf("❌ Provisioning: Failed to get task details for taskID %d: %v", taskID, err)
		// Cannot update task status without task object
		return
	}

	// Create infrastructure client (assuming all instances in request use the same provider)
	if len(instancesReq) == 0 {
		s.updateTaskError(ctx, callerOwnerID, task, fmt.Errorf("no instance requests provided"))
		return
	}
	infraReq := &types.InstancesRequest{
		TaskName:        task.Name,
		Instances:       instancesReq,
		Action:          "create",
		Provider:        instancesReq[0].Provider,
		HypervisorID:    instancesReq[0].HypervisorID,
		HypervisorGroup: instancesReq[0].HypervisorGroup,
	}
	infra, err := NewInfrastructure(infraReq)
	if err != nil {
		err = fmt.Errorf("failed to create infrastructure client: %w", err)
		s.updateTaskError(ctx, callerOwnerID, task, err)
		return
	}

	// Execute infrastructure creation
	result, err := infra.Execute(ctx, instancesReq)
	if err != nil {
		err = fmt.Errorf("failed to create infrastructure: %w", err)
		s.updateTaskError(ctx, callerOwnerID, task, err)
		return
	}
	pInstances, ok := result.([]types.InstanceInfo)
	if !ok {
		err = fmt.Errorf("invalid result type from infrastructure execution: %T", result)
		s.updateTaskError(ctx, callerOwnerID, task, err)
		return
	}
	logger.Debugf("📝 Infrastructure creation successful, received %d instance info objects.", len(pInstances))

	// --- Update DB for each provisioned instance ---
	anyUpdateFailed := false
	for _, pInstance := range pInstances {
		logger.Debugf("🔄 Processing update for provisioned instance %s", pInstance.Name)

		dbOwnerID, ok := instanceNameToOwnerID[pInstance.Name]
		if !ok {
			errMsg := fmt.Sprintf("❌ Internal error: Owner ID not found in map for instance name %s. Skipping.", pInstance.Name)
			logger.Error(errMsg)
			s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
			anyUpdateFailed = true
			continue
		}

		// Find the correct DB record corresponding to this provisioned instance and task
		dbInstance, err := s.findPendingInstanceForTask(ctx, task, dbOwnerID, pInstance.Name)
		if err != nil {
			// Error already logged by helper
			anyUpdateFailed = true
			continue
		}
		logger.Debugf("  - Matched DB Instance ID: %d for %s", dbInstance.ID, pInstance.Name)

		// Update Volumes if provided by infrastructure result
		if len(pInstance.Volumes) > 0 {
			if err := s.updateInstanceVolumes(ctx, dbOwnerID, dbInstance.ID, pInstance.Volumes); err != nil {
				errMsg := fmt.Sprintf("❌ Failed to update volumes for instance %s (ID: %d): %v", pInstance.Name, dbInstance.ID, err)
				logger.Errorf("❌ Failed to update volumes for instance %s (ID: %d): %v", pInstance.Name, dbInstance.ID, err)
				s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
				anyUpdateFailed = true
				// Consider whether to continue updating IP/Status if volume update fails
				// For now, let's continue
			} else {
				logMsg := fmt.Sprintf("Updated volumes for instance %s (ID: %d)", pInstance.Name, dbInstance.ID)
				logger.Debugf("✅ Updated volumes for instance %s (ID: %d)", pInstance.Name, dbInstance.ID)
				s.addTaskLogs(ctx, callerOwnerID, task, logMsg)
			}
		}

		// Update IP and Status
		// Re-fetch required after potential volume update to avoid stale data
		instanceToUpdate, err := s.repo.Get(ctx, dbOwnerID, dbInstance.ID)
		if err != nil {
			errMsg := fmt.Sprintf("❌ Failed to re-fetch instance %s (ID: %d) before IP/Status update: %v", pInstance.Name, dbInstance.ID, err)
			logger.Errorf("❌ Failed to re-fetch instance %s (ID: %d) before IP/Status update: %v", pInstance.Name, dbInstance.ID, err)
			s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
			anyUpdateFailed = true
			continue
		}

		updateData := &models.Instance{
			PublicIP: pInstance.PublicIP,
			Status:   models.InstanceStatusReady,
		}
		if err := s.repo.UpdateByID(ctx, dbOwnerID, instanceToUpdate.ID, updateData); err != nil {
			errMsg := fmt.Sprintf("❌ Failed to update instance %s (ID: %d) IP/Status: %v", pInstance.Name, instanceToUpdate.ID, err)
			logger.Errorf("❌ Failed to update instance %s (ID: %d) IP/Status: %v", pInstance.Name, instanceToUpdate.ID, err)
			s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
			anyUpdateFailed = true
			continue
		}
		logMsg := fmt.Sprintf("Updated instance %s (ID: %d) with IP %s and status Ready", pInstance.Name, instanceToUpdate.ID, pInstance.PublicIP)
		logger.Debugf("✅ Updated instance %s (ID: %d) with IP %s and status Ready", pInstance.Name, instanceToUpdate.ID, pInstance.PublicIP)
		s.addTaskLogs(ctx, callerOwnerID, task, logMsg)
	}

	// --- Optional Ansible Provisioning ---
	if !anyUpdateFailed && len(instancesReq) > 0 && instancesReq[0].Provision {
		s.runAnsibleProvisioning(ctx, task, infra, pInstances, instanceNameToOwnerID)
	} else if anyUpdateFailed {
		logger.Warnf("Skipping Ansible provisioning for task %d due to previous update errors.", taskID)
		s.addTaskLogs(ctx, callerOwnerID, task, "Skipping Ansible provisioning due to previous errors.")
	} // No else needed if Provision flag was false

	// --- Final Task Update ---
	if anyUpdateFailed {
		s.updateTaskError(ctx, callerOwnerID, task, fmt.Errorf("one or more instance updates failed during provisioning"))
	} else {
		s.updateTaskStatus(ctx, callerOwnerID, task.ID, models.TaskStatusCompleted)
		logger.Debugf("✅ Infrastructure creation and updates completed for task %s", task.Name)
	}
}

// runAnsibleProvisioning executes the Ansible provisioning step and updates payload status.
func (s *Instance) runAnsibleProvisioning(
	ctx context.Context,
	task *models.Task,
	infra *Infrastructure,
	pInstances []types.InstanceInfo,
	instanceNameToOwnerID map[string]uint,
) {
	callerOwnerID := task.OwnerID
	s.addTaskLogs(ctx, callerOwnerID, task, "Running Ansible provisioning")

	if err := infra.RunProvisioning(pInstances); err != nil {
		err = fmt.Errorf("failed to run provisioning: %w", err)
		s.updateTaskError(ctx, callerOwnerID, task, err) // This also sets task status to Failed
		return
	}
	s.addTaskLogs(ctx, callerOwnerID, task, "Ansible provisioning completed")

	// Update payload status for instances that had a payload path
	for _, pInstance := range pInstances {
		if pInstance.PayloadPath == "" {
			continue
		}

		dbOwnerID, ok := instanceNameToOwnerID[pInstance.Name]
		if !ok {
			errMsg := fmt.Sprintf("❌ Internal error: Owner ID not found in map for payload status update on instance %s. Skipping.", pInstance.Name)
			logger.Errorf("❌ Internal error: Owner ID not found in map for payload status update on instance %s. Skipping.", pInstance.Name)
			s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
			continue
		}

		// We need to find the instance ID again. Reuse the finder logic,
		// but we don't need to check for Pending status here, just the TaskID.
		// TODO: Refactor finding logic further to avoid repetition.
		payloadInstance, err := s.findInstanceForTask(ctx, task, dbOwnerID, pInstance.Name)
		if err != nil {
			// Error already logged by helper
			continue
		}

		updateData := &models.Instance{PayloadStatus: models.PayloadStatusExecuted}
		if err := s.repo.UpdateByID(ctx, dbOwnerID, payloadInstance.ID, updateData); err != nil {
			errMsg := fmt.Sprintf("❌ Failed to update payload status for instance %s (ID: %d): %v", pInstance.Name, payloadInstance.ID, err)
			logger.Errorf("❌ Failed to update payload status for instance %s (ID: %d): %v", pInstance.Name, payloadInstance.ID, err)
			s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
			// Consider if this failure should mark the whole task as failed?
		} else {
			logMsg := fmt.Sprintf("Updated payload status to executed for instance %s (ID: %d)", pInstance.Name, payloadInstance.ID)
			logger.Debugf("✅ Updated payload status to executed for instance %s (ID: %d)", pInstance.Name, payloadInstance.ID)
			s.addTaskLogs(ctx, callerOwnerID, task, logMsg)
		}
	}
}

// findInstanceForTask is similar to findPendingInstanceForTask but doesn't require Pending status.
// It prioritizes finding the instance associated with the specific task ID.
// TODO: Combine common logic with findPendingInstanceForTask.
func (s *Instance) findInstanceForTask(
	ctx context.Context,
	task *models.Task,
	dbOwnerID uint,
	instanceName string,
) (*models.Instance, error) {
	callerOwnerID := task.OwnerID
	foundCorrectInstance := false // Renamed from identifiedInstance to avoid shadowing later
	var correctInstance *models.Instance

	// Attempt 1: GetByName
	getInstance, err := s.repo.GetByName(ctx, dbOwnerID, instanceName)
	if err != nil {
		// Log error and prepare for fallback
		errMsg := fmt.Sprintf("ℹ️ GetByName failed for instance %s (Owner: %d) during payload check: %v. Fallback List.", instanceName, dbOwnerID, err)
		logger.Warnf("ℹ️ GetByName failed for instance %s (Owner: %d) during payload check: %v. Fallback List.", instanceName, dbOwnerID, err)
		s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
		// Go directly to fallback check below
	} else {
		// GetByName succeeded, check if it's the correct task
		if getInstance.LastTaskID == task.ID {
			logger.Debugf("Instance %s (ID: %d) identified via GetByName for Task %d (Payload Check)", instanceName, getInstance.ID, task.ID)
			correctInstance = getInstance // Found the correct one
			foundCorrectInstance = true
			// No need for fallback if found here
		} else {
			// GetByName found instance from wrong task, log and prepare for fallback
			warnMsg := fmt.Sprintf("⚠️ GetByName returned instance %s (ID: %d) from wrong task (%d, expected %d) during payload check. Fallback List.",
				instanceName, getInstance.ID, getInstance.LastTaskID, task.ID)
			logger.Warnf("⚠️ GetByName returned instance %s (ID: %d) from wrong task (%d, expected %d) during payload check. Fallback List.", instanceName, getInstance.ID, getInstance.LastTaskID, task.ID)
			s.addTaskLogs(ctx, callerOwnerID, task, warnMsg)
			// Proceed to fallback check below
		}
	}

	// Attempt 2: Fallback List (only if not found by GetByName correctly)
	if !foundCorrectInstance {
		logger.Debugf("Attempting fallback List search for instance %s (Owner: %d, Task: %d) for payload check", instanceName, dbOwnerID, task.ID)
		allInstances, listErr := s.repo.List(ctx, dbOwnerID, nil)
		if listErr != nil {
			errMsg := fmt.Sprintf("❌ Fallback List search failed for owner %d during payload check: %v", dbOwnerID, listErr)
			logger.Errorf("❌ Fallback List search failed for owner %d during payload check: %v", dbOwnerID, listErr)
			s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
			return nil, fmt.Errorf("fallback List search failed for owner %d during payload check: %w", dbOwnerID, listErr)
		}
		// Removed else block, iterate directly
		for i := range allInstances {
			inst := &allInstances[i]
			if inst.Name == instanceName && inst.LastTaskID == task.ID {
				logger.Infof("✅ Fallback List search successful for instance %s (ID: %d, Task: %d) for payload check", instanceName, inst.ID, task.ID)
				correctInstance = inst // Found it via fallback
				foundCorrectInstance = true
				break
			}
		}
	}

	// Final check after both attempts
	if !foundCorrectInstance {
		errMsg := fmt.Sprintf("❌ Failed to identify DB instance for payload update %s (Owner: %d, Task: %d).", instanceName, dbOwnerID, task.ID)
		logger.Errorf("❌ Failed to identify DB instance for payload update %s (Owner: %d, Task: %d).", instanceName, dbOwnerID, task.ID)
		s.addTaskLogs(ctx, callerOwnerID, task, errMsg)
		return nil, fmt.Errorf("failed to identify DB instance for payload update %s (Owner: %d, Task: %d)", instanceName, dbOwnerID, task.ID)
	}

	return correctInstance, nil // Return the instance found by either method
}

// findActiveInstancesForTermination finds instances matching the requested names for a project,
// excluding those already terminated. It returns a list of valid instances to terminate
// and a list of names that were skipped (with reasons).
func (s *Instance) findActiveInstancesForTermination(
	ctx context.Context,
	project *models.Project,
	instanceNames []string,
) ([]*models.Instance, []string, error) {
	instancesToTerminate := make([]*models.Instance, 0)
	notFoundOrInvalidNames := make([]string, 0)

	allOwnerInstances, listErr := s.repo.List(ctx, project.OwnerID, nil)
	if listErr != nil {
		logger.Errorf("❌ Failed to list instances for owner %d during termination check: %v", project.OwnerID, listErr)
		// Return error so the caller can create a failed task
		return nil, nil, fmt.Errorf("failed to list instances for project owner %d: %w", project.OwnerID, listErr)
	}

	requestedNames := make(map[string]bool)
	for _, name := range instanceNames {
		requestedNames[name] = true
	}
	processedNames := make(map[string]bool)

	for i := range allOwnerInstances {
		inst := &allOwnerInstances[i]

		if requestedNames[inst.Name] && inst.ProjectID == project.ID {
			processedNames[inst.Name] = true
			logger.Debugf("Checking instance found in list: Name=%s, InstanceID=%d, InstanceProjectID=%d, InstanceStatus=%d against ProjectID=%d", inst.Name, inst.ID, inst.ProjectID, inst.Status, project.ID)

			if inst.Status != models.InstanceStatusTerminated {
				instancesToTerminate = append(instancesToTerminate, inst)
			} else {
				notFoundOrInvalidNames = append(notFoundOrInvalidNames, fmt.Sprintf("%s (already terminated)", inst.Name))
			}
		}
	}

	for name := range requestedNames {
		if !processedNames[name] {
			notFoundOrInvalidNames = append(notFoundOrInvalidNames, fmt.Sprintf("%s (not found in project %s)", name, project.Name))
		}
	}

	return instancesToTerminate, notFoundOrInvalidNames, nil // Return lists and nil error
}

// handleTerminationInBackground contains the logic previously in Terminate, executed async.
func (s *Instance) handleTerminationInBackground(ctx context.Context, callerOwnerID, taskID uint, project *models.Project, instanceNames []string) {
	// Fetch the task we just created (or should exist)
	task, err := s.taskService.GetByID(ctx, callerOwnerID, taskID)
	if err != nil {
		// If we can't even get the task, something is wrong, cannot update status/logs.
		logger.Errorf("❌ handleTerminationInBackground: Failed to get task %d: %v", taskID, err)
		return
	}

	// Find valid instances to terminate
	instancesToTerminate, notFoundOrInvalidNames, listErr := s.findActiveInstancesForTermination(ctx, project, instanceNames)
	if listErr != nil {
		// Update the task to failed status if listing failed
		s.updateTaskError(ctx, callerOwnerID, task, listErr)
		return
	}

	// Log skipped instances
	if len(notFoundOrInvalidNames) > 0 {
		logMsg := fmt.Sprintf("Skipped termination for the following names: %v", notFoundOrInvalidNames)
		logger.Infof("Skipped termination for the following names: %v", notFoundOrInvalidNames)
		if task != nil {
			s.addTaskLogs(ctx, callerOwnerID, task, logMsg)
		}
	}

	// Check if there are any valid instances to terminate
	if len(instancesToTerminate) == 0 {
		finalLogMsg := "No valid instances found to terminate for this request."
		logger.Infof("No valid instances found to terminate for this request.")
		if task != nil {
			s.addTaskLogs(ctx, callerOwnerID, task, finalLogMsg)
			s.updateTaskStatus(ctx, callerOwnerID, task.ID, models.TaskStatusCompleted) // Mark task as completed (no-op)
		}
		return
	}

	// Proceed with termination for the found valid instances
	// s.terminate itself runs the core logic in another goroutine.
	s.terminate(ctx, callerOwnerID, task.ID, instancesToTerminate)
	// Note: The s.terminate function handles updating the task status upon completion/failure.
}

// Terminate handles the termination of instances for a given project name and instance names.
func (s *Instance) Terminate(ctx context.Context, ownerID uint, projectName string, instanceNames []string) (taskName string, err error) {
	project, err := s.projectService.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	// Create the termination task immediately with Pending status
	taskName = uuid.New().String()
	task := &models.Task{
		Name:      taskName,
		OwnerID:   ownerID,
		ProjectID: project.ID,
		Status:    models.TaskStatusPending, // Start as Pending
		Action:    models.TaskActionTerminateInstances,
	}
	createTaskErr := s.taskService.Create(ctx, task)
	if createTaskErr != nil {
		logger.Errorf("❌ Failed to create termination task: %v", createTaskErr)
		return "", fmt.Errorf("failed to create termination task: %w", createTaskErr)
	}

	// Fetch the created task ID immediately (assuming Create doesn't return it)
	// This is needed to pass to the background handler.
	createdTask, getTaskErr := s.taskService.GetByName(ctx, ownerID, taskName)
	if getTaskErr != nil {
		// If we can't get the task ID right after creating it, we can't reliably
		// update it later. Mark the just-created task as failed.
		logger.Errorf("❌ Failed to get termination task %s immediately after creation: %v. Marking task as failed.", taskName, getTaskErr)
		s.updateTaskStatus(ctx, ownerID, task.ID, models.TaskStatusFailed) // Attempt to mark as failed
		s.addTaskLogs(ctx, ownerID, task, fmt.Sprintf("Failed to fetch task details immediately after creation: %v", getTaskErr))
		return taskName, fmt.Errorf("failed to get created task %s: %w", taskName, getTaskErr)
	}

	// Launch the main termination logic in the background
	go s.handleTerminationInBackground(ctx, ownerID, createdTask.ID, project, instanceNames)

	// Return the task name immediately
	logger.Infof("✅ Termination task %s created for project %s, processing in background.", taskName, projectName)
	return taskName, nil
}

// deleteRequest represents a single instance deletion request with tracking
type deleteRequest struct {
	instance     *models.Instance
	infraRequest *types.InstancesRequest
	attempts     int
	lastError    error
	maxAttempts  int
}

// deletionResults tracks the overall results of the deletion operation
type deletionResults struct {
	successful []string
	failed     map[string]error
}

// terminate handles the infrastructure deletion process
func (s *Instance) terminate(ctx context.Context, callerOwnerID, taskID uint, instances []*models.Instance) {
	// Fetch task using caller's ID
	task, err := s.taskService.GetByID(ctx, callerOwnerID, taskID)
	if err != nil {
		logger.Errorf("❌ Terminate: Failed to get task details for taskID %d: %v", taskID, err)
		return
	}

	if len(instances) == 0 {
		// This check might be redundant due to caller checks, but safe to keep
		s.updateTaskError(ctx, callerOwnerID, task, fmt.Errorf("no instances passed to terminate goroutine"))
		return
	}

	// Determine the actual OwnerID for infra/DB operations
	instanceOwnerID := instances[0].OwnerID

	// Create queue of delete requests
	queue := make([]*deleteRequest, 0, len(instances))
	for _, instance := range instances {
		logMsg := fmt.Sprintf("🗑️ Adding instance %s (ID: %d) to termination queue.", instance.Name, instance.ID)
		logger.Infof("🗑️ Adding instance %s (ID: %d) to termination queue.", instance.Name, instance.ID)
		s.addTaskLogs(ctx, callerOwnerID, task, logMsg)

		infraReq := &types.InstancesRequest{
			TaskName: task.Name,
			Instances: []types.InstanceRequest{{
				Name:     instance.Name,
				Provider: instance.ProviderID,
				Region:   instance.Region,
				Size:     instance.Size,
			}},
			Action:   "delete",
			Provider: instance.ProviderID,
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

	defaultErrorSleep := 100 * time.Millisecond

	// Process queue with retries
REQUESTLOOP:
	for len(queue) > 0 {
		select {
		case <-ctx.Done():
			for _, req := range queue {
				results.failed[req.instance.Name] = fmt.Errorf("operation cancelled: %w", ctx.Err())
			}
			queue = nil
			break REQUESTLOOP
		default:
		}

		request := queue[0]
		queue = queue[1:]

		if request.attempts >= request.maxAttempts {
			results.failed[request.instance.Name] = fmt.Errorf("max attempts reached (%d): %w", request.maxAttempts, request.lastError) // Use %w to wrap error
			continue
		}
		request.attempts++

		// Try infra deletion
		infra, err := NewInfrastructure(request.infraRequest)
		if err != nil {
			request.lastError = fmt.Errorf("failed to create infrastructure client: %w", err)
			queue = append(queue, request)
			time.Sleep(defaultErrorSleep)
			continue
		}
		_, err = infra.Execute(ctx, request.infraRequest.Instances)
		if err != nil && (!strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found")) {
			request.lastError = fmt.Errorf("failed to delete infrastructure: %w", err)
			queue = append(queue, request)
			time.Sleep(defaultErrorSleep)
			continue
		}

		// Try DB termination
		if err := s.repo.Terminate(ctx, instanceOwnerID, request.instance.ID); err != nil {
			request.lastError = fmt.Errorf("failed to terminate instance %d in database: %w", request.instance.ID, err)
			queue = append(queue, request)
			time.Sleep(defaultErrorSleep)
			continue
		}

		// Success for this instance
		results.successful = append(results.successful, request.instance.Name)
		s.addTaskLogs(ctx, callerOwnerID, task, fmt.Sprintf("Instance %s (ID: %d) terminated successfully.", request.instance.Name, request.instance.ID))
	}

	// Log final results
	if len(results.successful) > 0 {
		logMsg := fmt.Sprintf("Successfully deleted instances: %v", results.successful)
		logger.Infof("Successfully deleted instances: %v", results.successful)
		s.addTaskLogs(ctx, callerOwnerID, task, logMsg)
	}
	if len(results.failed) > 0 {
		logMsg := "Failed to delete instances after multiple attempts:"
		logger.Error(logMsg)
		s.addTaskLogs(ctx, callerOwnerID, task, logMsg)
		for name, err := range results.failed {
			logMsg := fmt.Sprintf("  - %s: %v", name, err)
			logger.Error(logMsg)
			s.addTaskLogs(ctx, callerOwnerID, task, logMsg)
		}
	}

	// Create deletion result for API response
	deletionResult := map[string]interface{}{
		"status":    "completed",
		"deleted":   results.successful,
		"failed":    results.failed,
		"completed": time.Now().UTC(),
	}
	resultJSON, err := json.Marshal(deletionResult)
	if err != nil {
		logger.Errorf("failed to marshal deletion result: %v", err)
		// Don't overwrite potential failure status if marshalling fails
		if len(results.failed) == 0 {
			s.updateTaskError(ctx, callerOwnerID, task, fmt.Errorf("failed to marshal result: %w", err))
		}
		return // Exit after attempting to mark task as failed
	}

	// Update final task status and result
	task.Result = resultJSON
	if len(results.failed) > 0 {
		task.Status = models.TaskStatusFailed
		task.Error += "\nOne or more instances failed to terminate after retries."
	} else {
		task.Status = models.TaskStatusCompleted
	}

	if err := s.taskService.Update(ctx, callerOwnerID, task); err != nil {
		logger.Errorf("failed to update final task status/result: %v", err)
	}
}

// updateTaskError updates task status to Failed and appends the error message.
func (s *Instance) updateTaskError(ctx context.Context, ownerID uint, task *models.Task, err error) {
	if task == nil || err == nil {
		return
	}
	currentTask, fetchErr := s.taskService.GetByID(ctx, ownerID, task.ID)
	if fetchErr != nil {
		logger.Errorf("failed to get task %d before updating error: %v", task.ID, fetchErr)
		// Attempt to update with potentially stale task object anyway
		task.Error += fmt.Sprintf("\n%s", err.Error())
		task.Status = models.TaskStatusFailed
		if updateErr := s.taskService.Update(ctx, ownerID, task); updateErr != nil {
			logger.Errorf("failed to update task %d (stale) with error: %v", task.ID, updateErr)
		}
		return
	}

	currentTask.Error += fmt.Sprintf("\n%s", err.Error())
	currentTask.Status = models.TaskStatusFailed
	if err := s.taskService.Update(ctx, ownerID, currentTask); err != nil {
		logger.Errorf("failed to update task %d with error: %v", currentTask.ID, err)
	}
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

// updateTaskStatus updates the status of a task.
func (s *Instance) updateTaskStatus(ctx context.Context, ownerID uint, taskID uint, status models.TaskStatus) {
	if err := s.taskService.UpdateStatus(ctx, ownerID, taskID, status); err != nil {
		logger.Errorf("failed to update task %d status to %v: %v", taskID, status, err)
	}
}
