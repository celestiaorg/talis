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

	task, err := s.taskService.GetByName(ctx, ownerID, taskName)
	if err != nil {
		return "", fmt.Errorf("failed to get task: %w", err)
	}

	instancesToCreate := make([]*models.Instance, 0, len(instances))
	// Create a map to store the final generated instance name to its intended DB owner ID
	instanceNameToOwnerID := make(map[string]uint)

	for _, i := range instances {
		// Sanity check the ownerID fields.
		// TODO: this is a little verbose, maybe we can clean it up?
		bothZero := i.OwnerID == 0 && ownerID == 0
		bothNonZero := i.OwnerID != 0 && ownerID != 0
		// At least one of the ownerID fields is required
		if bothZero {
			return "", fmt.Errorf("instance owner_id is required")
		}
		// Sanity check that a user is not trying to create an instance for another user
		if bothNonZero && i.OwnerID != ownerID {
			return "", fmt.Errorf("instance owner_id does not match project owner_id")
		}

		// Determine the OwnerID to be stored in the database for this instance
		dbOwnerID := ownerID // Default to the authenticated user's ID
		if i.OwnerID != 0 {
			dbOwnerID = i.OwnerID // Override with the ID specified in the request item
		}

		baseName := i.Name
		if baseName == "" {
			baseName = fmt.Sprintf("instance-%s", uuid.New().String())
		}

		// Create multiple instances if requested
		numInstances := i.NumberOfInstances
		if numInstances < 1 {
			numInstances = 1
		}

		for idx := 0; idx < numInstances; idx++ {
			instanceName := baseName
			if numInstances > 1 {
				instanceName = fmt.Sprintf("%s-%d", baseName, idx)
			}

			// Determine initial payload status
			initialPayloadStatus := models.PayloadStatusNone
			if i.PayloadPath != "" {
				initialPayloadStatus = models.PayloadStatusPendingCopy
			}

			// Store the mapping from the final instance name to its DB owner ID
			instanceNameToOwnerID[instanceName] = dbOwnerID

			instancesToCreate = append(instancesToCreate, &models.Instance{
				Name:          instanceName,
				OwnerID:       dbOwnerID, // Use the determined DB owner ID
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

	// Start provisioning in background, passing only the name-to-owner map
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
	volumeDetails []types.VolumeDetails,
) error {
	// Get the instance first using ownerID and instanceID
	instance, err := s.repo.Get(ctx, ownerID, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance %d for owner %d: %w", instanceID, ownerID, err)
	}
	instanceName := instance.Name // Keep instanceName for logging

	logger.Debugf("🔄 Converting volume details for instance %s (ID: %d)", instanceName, instanceID)
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

	logger.Debugf("📦 Preparing to update instance %s (ID: %d)", instanceName, instanceID)
	logger.Debugf("📝 Data to update:")
	logger.Debugf("  - Volumes: %#v", volumes)
	logger.Debugf("  - Volume Details: %#v", dbVolumeDetails)

	// Create update instance with only the fields we want to update
	updateInstance := &models.Instance{
		Volumes:       volumes,
		VolumeDetails: dbVolumeDetails,
	}

	// Update instance in database using a transaction to ensure atomicity
	err = s.repo.UpdateByID(ctx, ownerID, instanceID, updateInstance)
	if err != nil {
		logger.Errorf("❌ Failed to update instance %s (ID: %d) volumes: %v", instanceName, instanceID, err)
		return fmt.Errorf("failed to update instance %s (ID: %d) volumes: %w", instanceName, instanceID, err)
	}

	// Verify the update immediately
	updatedInstance, err := s.repo.Get(ctx, ownerID, instanceID)
	if err != nil {
		logger.Warnf("⚠️ Could not verify volume update for instance %s (ID: %d): %v", instanceName, instanceID, err)
		return nil // Don't return error here as the update might have succeeded
	}

	logger.Debugf("✅ Verified volumes update for instance %s (ID: %d):", instanceName, instanceID)
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

// provisionInstances provisions the job asynchronously
func (s *Instance) provisionInstances(ctx context.Context, ownerID, taskID uint, instances []types.InstanceRequest, instanceNameToOwnerID map[string]uint) {
	go func() {
		// Get task details (use the initial ownerID for task operations)
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
		for _, pInstance := range pInstances {
			logger.Debugf("🔄 Processing instance update for %s", pInstance.Name)
			logger.Debugf("  - Provisioned Info: %+v", pInstance)

			// Find the corresponding DB instance record (should be in Pending state for this task)
			// Use the correct owner ID from the map passed from CreateInstance
			dbOwnerID, ok := instanceNameToOwnerID[pInstance.Name]
			if !ok {
				errMsg := fmt.Sprintf("❌ Internal error: Owner ID not found in map for instance name %s", pInstance.Name)
				logger.Error(errMsg)
				s.addTaskLogs(ctx, ownerID, task, errMsg) // Log to the original task owner
				continue
			}

			var foundCorrectInstance bool           // Flag to track if we identified the correct instance
			var identifiedInstance *models.Instance // Store the identified instance here

			// Attempt 1: GetByName
			getInstance, err := s.repo.GetByName(ctx, dbOwnerID, pInstance.Name)
			if err == nil {
				// GetByName succeeded, check if it's the correct one
				if getInstance.LastTaskID == taskID && getInstance.Status == models.InstanceStatusPending {
					identifiedInstance = getInstance
					foundCorrectInstance = true
					logger.Debugf("Instance %s (ID: %d) identified via GetByName for Task %d", pInstance.Name, identifiedInstance.ID, taskID)
				} else {
					// GetByName returned the wrong instance (different task or status)
					warnMsg := fmt.Sprintf("⚠️ GetByName returned instance %s (ID: %d) with wrong state (TaskID: %d/%d, Status: %d/%d). Attempting fallback List.",
						pInstance.Name, getInstance.ID, getInstance.LastTaskID, taskID, getInstance.Status, models.InstanceStatusPending)
					logger.Warn(warnMsg)
					s.addTaskLogs(ctx, ownerID, task, warnMsg)
				}
			} else {
				// GetByName failed. Log the error.
				errMsg := fmt.Sprintf("ℹ️ GetByName failed for instance %s (Owner: %d): %v. Will attempt fallback List.", pInstance.Name, dbOwnerID, err)
				logger.Warn(errMsg) // Log as warning because fallback might succeed
				s.addTaskLogs(ctx, ownerID, task, errMsg)
			}
			// Regardless of GetByName outcome, if we haven't found the correct instance yet, try fallback.

			// Attempt 2: Fallback List (if GetByName didn't find the correct one)
			if !foundCorrectInstance {
				logger.Debugf("Attempting fallback List search for instance %s (Owner: %d, Task: %d)", pInstance.Name, dbOwnerID, taskID)
				// Fallback: List all instances for the owner and find the one matching name, task, and status
				allInstances, err := s.repo.List(ctx, dbOwnerID, nil) // Assuming nil fetches all
				if err != nil {
					errMsg := fmt.Sprintf("❌ Fallback List search failed for owner %d: %v", dbOwnerID, err)
					logger.Error(errMsg)
					s.addTaskLogs(ctx, ownerID, task, errMsg)
					// Cannot proceed without list result
				} else {
					// Iterate and find the match
					for i := range allInstances {
						inst := &allInstances[i]
						if inst.Name == pInstance.Name && inst.LastTaskID == taskID && inst.Status == models.InstanceStatusPending {
							identifiedInstance = inst // Assign the pointer
							foundCorrectInstance = true
							logger.Infof("✅ Fallback List search successful for instance %s (ID: %d, Task: %d)", pInstance.Name, identifiedInstance.ID, taskID)
							break
						}
					}
				}
			} // End of fallback attempt

			// Final Check: If we didn't find the correct instance by either method, log error and skip.
			if !foundCorrectInstance {
				errMsg := fmt.Sprintf("❌ Failed to identify unique pending DB instance for provisioned resource %s (Owner: %d, Task: %d). Skipping update.",
					pInstance.Name, dbOwnerID, taskID)
				logger.Error(errMsg)
				s.addTaskLogs(ctx, ownerID, task, errMsg)
				continue // Skip this pInstance
			}

			// If we reach here, identifiedInstance points to the correct, pending instance.
			dbInstance := identifiedInstance // Use the identified instance for subsequent operations
			logger.Debugf("  - Processing DB Instance ID: %d, Owner: %d, Status: %d, TaskID: %d", dbInstance.ID, dbOwnerID, dbInstance.Status, dbInstance.LastTaskID)

			// Update volumes first if present
			if len(pInstance.Volumes) > 0 || len(pInstance.VolumeDetails) > 0 {
				logger.Debugf("🔄 Updating volumes for instance %s (ID: %d, Owner: %d)", pInstance.Name, dbInstance.ID, dbOwnerID)
				// Pass the correct dbOwnerID and dbInstance.ID to updateInstanceVolumes
				if err := s.updateInstanceVolumes(ctx, dbOwnerID, dbInstance.ID, pInstance.Volumes, pInstance.VolumeDetails); err != nil {
					err = fmt.Errorf("❌ Failed to update volumes for instance %s (ID: %d, Owner: %d): %w", pInstance.Name, dbInstance.ID, dbOwnerID, err)
					logger.Error(err)
					s.addTaskLogs(ctx, ownerID, task, err.Error()) // Log error to original task owner
					continue
				}
				logger.Debugf("✅ Successfully updated volumes for instance %s (ID: %d, Owner: %d)", pInstance.Name, dbInstance.ID, dbOwnerID)
				s.addTaskLogs(ctx, ownerID, task, fmt.Sprintf("Updated volumes for instance %s (ID: %d, Owner: %d)", pInstance.Name, dbInstance.ID, dbOwnerID))

				// Verify the update was successful
				// Use the correct dbOwnerID
				updatedInstance, err := s.repo.Get(ctx, dbOwnerID, dbInstance.ID)
				if err != nil {
					errMsg := fmt.Sprintf("❌ Failed to verify volume update for instance %s (ID: %d, Owner: %d): %v", pInstance.Name, dbInstance.ID, dbOwnerID, err)
					logger.Error(errMsg)
					s.addTaskLogs(ctx, ownerID, task, errMsg)
					continue
				}

				logger.Debugf("📊 Current instance state after volume update:")
				logger.Debugf("  - Volumes: %v", updatedInstance.Volumes)
				logger.Debugf("  - Volume Details: %+v", updatedInstance.VolumeDetails)
				s.addTaskLogs(ctx, ownerID, task, fmt.Sprintf("Verified volume update for instance %s (ID: %d, Owner: %d)", pInstance.Name, dbInstance.ID, dbOwnerID))
			}

			// Then update instance status and IP
			// Re-fetch the instance AFTER potential volume updates to ensure we have the latest state
			instanceToUpdate, err := s.repo.Get(ctx, dbOwnerID, dbInstance.ID)
			if err != nil {
				errMsg := fmt.Sprintf("❌ Failed to re-fetch instance %s (ID: %d, Owner: %d) before final update: %v", dbInstance.Name, dbInstance.ID, dbOwnerID, err)
				logger.Error(errMsg)
				s.addTaskLogs(ctx, ownerID, task, errMsg)
				continue
			}

			updateInstance := instanceToUpdate // Start with the re-fetched instance
			updateInstance.PublicIP = pInstance.PublicIP
			updateInstance.Status = models.InstanceStatusReady

			// Update using the specific ID and correct dbOwnerID
			if err := s.repo.UpdateByID(ctx, dbOwnerID, updateInstance.ID, updateInstance); err != nil {
				errMsg := fmt.Sprintf("❌ Failed to update instance %s (ID: %d, Owner: %d) IP/Status: %v", updateInstance.Name, updateInstance.ID, dbOwnerID, err)
				logger.Error(errMsg)
				s.addTaskLogs(ctx, ownerID, task, errMsg)
				continue
			}
			logger.Debugf("✅ Updated instance %s (ID: %d, Owner: %d) with IP %s and status ready", updateInstance.Name, updateInstance.ID, dbOwnerID, updateInstance.PublicIP)
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
			for _, pInstance := range pInstances {
				if pInstance.PayloadPath == "" {
					continue
				}
				updateInstancePayload := &models.Instance{
					PayloadStatus: models.PayloadStatusExecuted,
				}

				// Find the correct DB instance ID and Owner ID again
				dbOwnerID, ok := instanceNameToOwnerID[pInstance.Name]
				if !ok {
					logger.Errorf("❌ Internal error: Owner ID not found in map for payload status update on instance %s", pInstance.Name)
					continue
				}

				var foundPayloadInstance bool
				var identifiedPayloadInstance *models.Instance

				// Attempt 1: GetByName for payload update
				getInstancePayload, err := s.repo.GetByName(ctx, dbOwnerID, pInstance.Name)
				if err == nil {
					if getInstancePayload.LastTaskID == taskID {
						identifiedPayloadInstance = getInstancePayload
						foundPayloadInstance = true
						logger.Debugf("Payload instance %s (ID: %d) identified via GetByName for Task %d", pInstance.Name, identifiedPayloadInstance.ID, taskID)
					} else {
						warnMsg := fmt.Sprintf("⚠️ GetByName returned wrong task (%d, expected %d) for payload update on %s. Attempting fallback List.",
							getInstancePayload.LastTaskID, taskID, pInstance.Name)
						logger.Warn(warnMsg)
						s.addTaskLogs(ctx, ownerID, task, warnMsg)
					}
				} else {
					// GetByName failed for payload update. Log the error.
					errMsg := fmt.Sprintf("ℹ️ GetByName failed for payload update on %s: %v. Will attempt fallback List.", pInstance.Name, err)
					logger.Warn(errMsg)
					s.addTaskLogs(ctx, ownerID, task, errMsg)
				}

				// Attempt 2: Fallback List for payload update
				if !foundPayloadInstance {
					logger.Debugf("Attempting fallback List for payload update on instance %s (Owner: %d, Task: %d)", pInstance.Name, dbOwnerID, taskID)
					allInstancesPayload, err := s.repo.List(ctx, dbOwnerID, nil)
					if err != nil {
						errMsg := fmt.Sprintf("❌ Fallback List failed for payload update owner %d: %v", dbOwnerID, err)
						logger.Error(errMsg)
						s.addTaskLogs(ctx, ownerID, task, errMsg)
					} else {
						for i := range allInstancesPayload {
							inst := &allInstancesPayload[i]
							if inst.Name == pInstance.Name && inst.LastTaskID == taskID {
								identifiedPayloadInstance = inst // Assign the pointer
								foundPayloadInstance = true
								logger.Infof("✅ Fallback List successful for payload update on instance %s (ID: %d, Task: %d)", pInstance.Name, identifiedPayloadInstance.ID, taskID)
								break
							}
						}
					}
				}

				// Final Check: If we didn't find the correct instance for payload update, skip.
				if !foundPayloadInstance {
					errMsg := fmt.Sprintf("❌ Failed to identify unique DB instance for payload update on %s (Owner: %d, Task: %d). Skipping.",
						pInstance.Name, dbOwnerID, taskID)
					logger.Error(errMsg)
					s.addTaskLogs(ctx, ownerID, task, errMsg)
					continue
				}

				// If we reach here, identifiedPayloadInstance points to the correct instance for this task
				dbInstanceToUpdatePayload := identifiedPayloadInstance
				if err := s.repo.UpdateByID(ctx, dbOwnerID, dbInstanceToUpdatePayload.ID, updateInstancePayload); err != nil {
					logger.Errorf("❌ Failed to update payload status for instance %s (ID: %d, Owner: %d): %v", pInstance.Name, dbInstanceToUpdatePayload.ID, dbOwnerID, err)
					continue
				}
				logger.Debugf("✅ Updated payload status to executed for instance %s (ID: %d, Owner: %d)", pInstance.Name, dbInstanceToUpdatePayload.ID, dbOwnerID)
			}
		}

		s.updateTaskStatus(ctx, ownerID, task.ID, models.TaskStatusCompleted)
		logger.Debugf("✅ Infrastructure creation completed for task %s", task.Name)
	}()
}

// Terminate handles the termination of instances for a given project name and instance names.
func (s *Instance) Terminate(ctx context.Context, ownerID uint, projectName string, instanceNames []string) (taskName string, err error) {
	// First verify the project exists and belongs to the owner making the request
	project, err := s.projectService.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	instancesToTerminate := make([]*models.Instance, 0)
	notFoundOrInvalidNames := make([]string, 0)

	// Fetch all instances for the project owner once, to avoid multiple List calls if possible
	// Note: If List doesn't support filtering by ProjectID, this fetches all for the owner.
	// This could be inefficient for owners with many instances across many projects.
	allOwnerInstances, listErr := s.repo.List(ctx, project.OwnerID, nil) // Assuming nil fetches all
	if listErr != nil {
		// If we can't list instances, we cannot proceed reliably.
		logger.Errorf("❌ Failed to list instances for owner %d during termination check: %v", project.OwnerID, listErr)
		// Create a task to report this failure
		taskName = uuid.New().String()
		createTaskErr := s.taskService.Create(ctx, &models.Task{
			Name:    taskName,
			OwnerID: ownerID, ProjectID: project.ID, Status: models.TaskStatusFailed,
			Action: models.TaskActionTerminateInstances, Error: fmt.Sprintf("failed to list instances for project owner %d: %v", project.OwnerID, listErr),
		})
		if createTaskErr != nil {
			logger.Errorf("❌ Additionally failed to create failure task: %v", createTaskErr)
		}
		return taskName, fmt.Errorf("failed to list instances for project owner %d: %w", project.OwnerID, listErr)
	}

	// Now iterate through requested names and check against the fetched list
	for _, name := range instanceNames {
		foundInProject := false
		for i := range allOwnerInstances { // Iterate through the fetched instances
			inst := &allOwnerInstances[i]

			// Check if this instance matches the requested name AND project ID
			if inst.Name == name && inst.ProjectID == project.ID {
				foundInProject = true
				logger.Debugf("Checking instance found in list: Name=%s, InstanceID=%d, InstanceProjectID=%d, InstanceStatus=%d against ProjectID=%d", inst.Name, inst.ID, inst.ProjectID, inst.Status, project.ID)

				// Check if it's already terminated
				if inst.Status != models.InstanceStatusTerminated {
					instancesToTerminate = append(instancesToTerminate, inst)
				} else {
					// Instance found in project but already terminated
					notFoundOrInvalidNames = append(notFoundOrInvalidNames, fmt.Sprintf("%s (already terminated)", name))
				}
				break // Found the specific instance for this name and project, stop inner loop
			}
		}

		// If after checking all instances, we didn't find one matching the name AND project ID
		if !foundInProject {
			notFoundOrInvalidNames = append(notFoundOrInvalidNames, fmt.Sprintf("%s (not found in project %s)", name, projectName))
		}
	}

	// Create the termination task
	taskName = uuid.New().String()
	createTaskErr := s.taskService.Create(ctx, &models.Task{
		Name:      taskName,
		OwnerID:   ownerID,
		ProjectID: project.ID,
		Status:    models.TaskStatusPending,
		Action:    models.TaskActionTerminateInstances,
	})
	if createTaskErr != nil {
		// If task creation fails, we can't proceed.
		return "", fmt.Errorf("failed to create termination task: %w", createTaskErr)
	}

	task, getTaskErr := s.taskService.GetByName(ctx, ownerID, taskName)
	if getTaskErr != nil {
		// Even if we can't get the task right away, the termination might still proceed.
		// Log the error but return the taskName.
		logger.Errorf("❌ Failed to get termination task %s immediately after creation: %v", taskName, getTaskErr)
		// Let the background process handle task updates.
		// Decide if we should return error here or let background handle it.
		// For now, returning taskName seems reasonable.
	}

	// Log skipped instances
	if len(notFoundOrInvalidNames) > 0 {
		logMsg := fmt.Sprintf("Skipped termination for the following names (not found, wrong project, or already terminated): %v", notFoundOrInvalidNames)
		logger.Infof(logMsg)
		if task != nil { // Add logs only if task was successfully retrieved
			s.addTaskLogs(ctx, ownerID, task, logMsg)
		}
	}

	// Check if there are any valid instances to terminate
	if len(instancesToTerminate) == 0 {
		finalLogMsg := "No valid instances found to terminate for this request."
		logger.Infof(finalLogMsg)
		if task != nil {
			s.addTaskLogs(ctx, ownerID, task, finalLogMsg)
			s.updateTaskStatus(ctx, ownerID, task.ID, models.TaskStatusCompleted) // Mark task as completed (no-op)
		} else {
			// If task fetch failed, we can't update it here.
			logger.Warnf("Cannot update task %s status to completed as task fetch failed.", taskName)
		}
		return taskName, nil // Successful request, even if nothing was terminated
	}

	// Proceed with termination for the found valid instances
	s.terminate(ctx, ownerID, task.ID, instancesToTerminate)
	return taskName, nil // Termination initiated successfully
}

// deleteRequest represents a single instance deletion request with tracking
// Change instance type to pointer
type deleteRequest struct {
	instance     *models.Instance
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
// Change instances parameter to slice of pointers
func (s *Instance) terminate(ctx context.Context, ownerID, taskID uint, instances []*models.Instance) {
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
			logMsg := fmt.Sprintf("🗑️ Attempting to terminate instance: %s (ID: %d)", instance.Name, instance.ID)
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

		// Determine the correct OwnerID for termination operations
		// All instances in the list should belong to the same project owner
		terminateOwnerID := ownerID // Default to caller ID (used for task updates)
		if len(instances) > 0 {
			terminateOwnerID = instances[0].OwnerID // Use the actual owner of the instances for infra/DB ops
		}

		results := &deletionResults{
			successful: make([]string, 0),
			failed:     make(map[string]error),
		}

		// Process queue until empty or all requests have failed
		defaultErrorSleep := 100 * time.Millisecond
		// ownerID variable is unused inside loop, removed redeclaration
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
			// Use terminateOwnerID (the actual instance owner)
			if err := s.repo.Terminate(ctx, terminateOwnerID, request.instance.ID); err != nil {
				request.lastError = fmt.Errorf("failed to terminate instance %d in database: %w", request.instance.ID, err)
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
	// Fetch the task again before updating to ensure we have the latest version
	currentTask, err := s.taskService.GetByID(ctx, ownerID, task.ID)
	if err != nil {
		logger.Errorf("failed to get task %d before adding logs: %v", task.ID, err)
		// Attempt to update with the potentially stale task object anyway
		task.Logs += fmt.Sprintf("\n%s", logs)
		if updateErr := s.taskService.Update(ctx, ownerID, task); updateErr != nil {
			logger.Errorf("failed to update task %d after failing to fetch: %v", task.ID, updateErr)
		}
		return
	}

	currentTask.Logs += fmt.Sprintf("\n%s", logs)
	if err := s.taskService.Update(ctx, ownerID, currentTask); err != nil {
		logger.Errorf("failed to update task %d with new logs: %v", currentTask.ID, err)
	}
}

func (s *Instance) updateTaskStatus(ctx context.Context, ownerID uint, taskID uint, status models.TaskStatus) {
	if err := s.taskService.UpdateStatus(ctx, ownerID, taskID, status); err != nil {
		logger.Errorf("failed to update task status: %v", err)
	}
}
