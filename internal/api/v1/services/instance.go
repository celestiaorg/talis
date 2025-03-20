package services

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// Instance defines the interface for instance operations
type Instance interface {
	ListInstances(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error)
	CreateInstance(ctx context.Context, name, projectName, webhookURL string, instances []infrastructure.InstanceRequest) (*models.Job, error)
	CreateInstancesForJob(ctx context.Context, job *models.Job, instances []infrastructure.InstanceRequest) error
	DeleteInstance(ctx context.Context, jobID uint, name, projectName string, instances []infrastructure.InstanceRequest) (*models.Job, error)
	GetInstance(ctx context.Context, id uint) (*models.Instance, error)
	GetInstancesByJobID(ctx context.Context, jobID uint) ([]models.Instance, error)
}

// Job defines the interface for job operations
type Job interface {
	CreateJob(ctx context.Context, job *models.Job) (*models.Job, error)
	UpdateJobStatus(ctx context.Context, id uint, status models.JobStatus, result interface{}, errMsg string) error
	GetByProjectName(ctx context.Context, projectName string) (*models.Job, error)
	GetJobStatus(ctx context.Context, ownerID uint, id uint) (models.JobStatus, error)
	ListJobs(ctx context.Context, status models.JobStatus, ownerID uint, opts *models.ListOptions) ([]models.Job, error)
}

// InstanceService provides business logic for instance operations
type InstanceService struct {
	repo       *repos.InstanceRepository
	jobService Job
}

// NewInstanceService creates a new instance service instance
func NewInstanceService(repo *repos.InstanceRepository, jobService Job) *InstanceService {
	return &InstanceService{
		repo:       repo,
		jobService: jobService,
	}
}

// ListInstances retrieves a paginated list of instances
func (s *InstanceService) ListInstances(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	return s.repo.List(ctx, opts)
}

// CreateInstance creates a new instance
func (s *InstanceService) CreateInstance(
	ctx context.Context,
	instanceName,
	projectName,
	webhookURL string,
	instances []infrastructure.InstanceRequest,
) (*models.Job, error) {
	log.Printf("📝 Starting instance creation process for project %s with name %s", projectName, instanceName)

	// Validate instances array is not empty
	if len(instances) == 0 {
		log.Printf("❌ No instances provided in request")
		return nil, fmt.Errorf("no instances provided in request")
	}

	// Create job request
	jobName := fmt.Sprintf("%s-%s-creation", projectName, instanceName)
	jobReq := &infrastructure.JobRequest{
		JobName:     jobName,
		ProjectName: projectName,
		Provider:    instances[0].Provider,
		Instances:   instances,
		Action:      "create",
	}

	log.Printf("🔍 Validating job request: %+v", jobReq)
	if err := jobReq.Validate(); err != nil {
		log.Printf("❌ Job request validation failed: %v", err)
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create job first
	job := &models.Job{
		Name:        jobName,
		ProjectName: projectName,
		Status:      models.JobStatusPending,
		WebhookURL:  webhookURL,
		OwnerID:     0, // Default owner ID for now
	}

	log.Printf("📝 Creating job in database: %+v", job)
	job, err := s.jobService.CreateJob(ctx, job)
	if err != nil {
		log.Printf("❌ Failed to create job in database: %v", err)
		return nil, fmt.Errorf("failed to create job: %w", err)
	}
	log.Printf("✅ Job created successfully with ID: %d", job.ID)

	// Create pending instances in the database
	instanceConfig := instances[0] // Safe because we validate there's at least one instance in the request
	log.Printf("📝 Will create %d instances with config: %+v", instanceConfig.NumberOfInstances, instanceConfig)

	for i := 0; i < instanceConfig.NumberOfInstances; i++ {
		// Create default tags if none provided
		var tags []string
		if len(instanceConfig.Tags) > 0 {
			tags = instanceConfig.Tags
		} else {
			tags = []string{fmt.Sprintf("%s-do-instance", job.Name)}
		}

		// Create instance in database with pending status
		instance := &models.Instance{
			JobID:      job.ID,
			ProviderID: models.ProviderID(jobReq.Provider),
			Name:       fmt.Sprintf("%s-%d", instanceName, i), // Match DO's naming convention
			Region:     instanceConfig.Region,
			Size:       instanceConfig.Size,
			Image:      instanceConfig.Image,
			Tags:       tags,
			Status:     models.InstanceStatusPending,
		}

		log.Printf("📝 Creating instance %d/%d in database: %+v", i+1, instanceConfig.NumberOfInstances, instance)

		if err := s.repo.Create(ctx, instance); err != nil {
			log.Printf("❌ Failed to create instance in database:\nInstance data: %+v\nError details: %v\nError type: %T",
				instance, err, err)

			// Log more details about the error
			log.Printf("🔍 Detailed error information:")
			log.Printf("  - Error type: %T", err)
			log.Printf("  - Error message: %v", err)
			log.Printf("  - Instance data: %+v", instance)

			// If this is the first instance and it fails, we should fail the entire operation
			if i == 0 {
				log.Printf("❌ First instance creation failed, aborting operation")
				// Try to update job status to failed
				updateErr := s.jobService.UpdateJobStatus(ctx, job.ID, models.JobStatusFailed, nil,
					fmt.Sprintf("Failed to create first instance: %v", err))
				if updateErr != nil {
					log.Printf("❌ Additionally failed to update job status: %v", updateErr)
				}
				return nil, fmt.Errorf("failed to create first instance in database: %w", err)
			}

			// For subsequent instances, continue but log the error
			continue
		}
		log.Printf("✅ Successfully created instance %s in database with ID %d", instance.Name, instance.ID)
	}

	log.Printf("🚀 Starting async infrastructure creation for job %d", job.ID)
	// Start async infrastructure creation
	go func() {
		if err := s.handleInfrastructureCreation(context.Background(), job, jobReq, instances); err != nil {
			log.Printf("❌ Failed to handle infrastructure creation: %v", err)
			s.updateJobStatusWithError(context.Background(), job.ID, models.JobStatusFailed, nil, err.Error())
		}
	}()

	log.Printf("✅ Instance creation process initiated successfully for job %d", job.ID)
	return job, nil
}

// DeleteInstance deletes an instance
func (s *InstanceService) DeleteInstance(
	ctx context.Context,
	jobID uint,
	instanceName,
	projectName string,
	instances []infrastructure.InstanceRequest,
) (*models.Job, error) {
	// Create job for deletion
	job := &models.Job{
		Model: gorm.Model{
			ID: jobID,
		},
		Name:        instanceName,
		ProjectName: projectName,
		Status:      models.JobStatusPending,
	}

	// Create infrastructure request
	jobReq := &infrastructure.JobRequest{
		JobName:      job.Name,
		InstanceName: instanceName,
		ProjectName:  job.ProjectName,
		Provider:     instances[0].Provider,
		Instances:    instances,
		Action:       "delete",
	}

	// Start async infrastructure deletion
	go func() {
		if err := s.handleInfrastructureDeletion(context.Background(), job, jobReq); err != nil {
			fmt.Printf("❌ Failed to handle infrastructure deletion: %v\n", err)
			s.updateJobStatusWithError(context.Background(), job.ID, models.JobStatusFailed, nil, err.Error())
		}
	}()

	return job, nil
}

// GetInstance retrieves an instance by ID
func (s *InstanceService) GetInstance(ctx context.Context, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, id)
}

// handleInfrastructureCreation handles the infrastructure creation process
func (s *InstanceService) handleInfrastructureCreation(
	ctx context.Context,
	job *models.Job,
	jobReq *infrastructure.JobRequest,
	instances []infrastructure.InstanceRequest,
) error {
	log.Printf("🚀 Starting async infrastructure creation for job %d...", job.ID)

	// Update to initializing when starting setup
	log.Printf("📝 Updating job status to initializing...")
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusInitializing, nil, "")

	log.Printf("🔧 Creating infrastructure client...")
	infra, err := infrastructure.NewInfrastructure(jobReq)
	if err != nil {
		log.Printf("❌ Failed to create infrastructure client: %v", err)
		return fmt.Errorf("failed to create infrastructure client: %w", err)
	}

	// Update to provisioning when creating infrastructure
	log.Printf("📝 Updating job status to provisioning...")
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusProvisioning, nil, "")

	log.Printf("🚀 Executing infrastructure creation...")
	result, err := infra.Execute()
	if err != nil {
		// Check if error is due to resource not found
		if strings.Contains(err.Error(), "404") && strings.Contains(err.Error(), "could not be found") {
			log.Printf("⚠️ Warning: Some old resources were not found (already deleted)")
		} else {
			log.Printf("❌ Failed to execute infrastructure: %v", err)
			return fmt.Errorf("failed to execute infrastructure: %w", err)
		}
	}

	// Get instance information from the result
	instanceInfos, ok := result.([]infrastructure.InstanceInfo)
	if !ok {
		log.Printf("❌ Invalid result type: expected []infrastructure.InstanceInfo, got %T", result)
		return fmt.Errorf("invalid result type: expected []infrastructure.InstanceInfo, got %T", result)
	}

	log.Printf("📝 Got instance information from provider: %+v", instanceInfos)

	// Get existing instances from database
	log.Printf("🔍 Getting existing instances from database for job %d...", job.ID)
	dbInstances, err := s.repo.GetByJobID(ctx, job.ID)
	if err != nil {
		log.Printf("❌ Failed to get instances from database: %v", err)
		return fmt.Errorf("failed to get instances from database: %w", err)
	}

	// Update instances in database with their new information
	log.Printf("📝 Updating instances in database with provider information...")
	for _, info := range instanceInfos {
		// Find the instance to update by checking if DO's name starts with our base name
		var instanceToUpdate *models.Instance
		for i := range dbInstances {
			// The DO instance name should start with our database instance name
			if strings.HasPrefix(info.Name, dbInstances[i].Name) {
				instanceToUpdate = &dbInstances[i]
				break
			}
		}

		if instanceToUpdate != nil {
			// Update existing instance with DO's information
			instanceToUpdate.Name = info.Name // Update with actual DO name
			instanceToUpdate.PublicIP = info.IP
			instanceToUpdate.Status = models.InstanceStatusReady
			if err := s.repo.Update(ctx, instanceToUpdate.ID, instanceToUpdate); err != nil {
				log.Printf("❌ Failed to update instance in database: %v", err)
				continue
			}
			log.Printf("✅ Updated instance %s in database with IP %s", instanceToUpdate.Name, info.IP)
		} else {
			log.Printf("⚠️ Warning: No matching instance found in database for DO instance %s", info.Name)
			// Log all database instances for debugging
			log.Printf("🔍 Database instances available:")
			for _, dbInstance := range dbInstances {
				log.Printf("  - %s (ID: %d)", dbInstance.Name, dbInstance.ID)
			}
		}
	}

	// Handle provisioning if requested
	if instances[0].Provision {
		log.Printf("🔧 Starting provisioning process...")
		if err := s.handleProvisioning(ctx, job, infra, instanceInfos); err != nil {
			log.Printf("❌ Provisioning failed: %v", err)
			return err
		}
	}

	// Update final status with result
	log.Printf("📝 Updating job status to completed...")
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusCompleted, instanceInfos, "")

	log.Printf("✅ Infrastructure creation completed for job ID %d", job.ID)
	return nil
}

// handleProvisioning handles the provisioning process for instances
func (s *InstanceService) handleProvisioning(
	ctx context.Context,
	job *models.Job,
	infra *infrastructure.Infrastructure,
	result interface{},
) error {
	instances, ok := result.([]infrastructure.InstanceInfo)
	if !ok {
		log.Printf("❌ Invalid result type for provisioning: %T", result)
		return fmt.Errorf("invalid result type: %T, expected []infrastructure.InstanceInfo", result)
	}

	log.Printf("📝 Starting provisioning for instances: %+v", instances)

	// Update to configuring when setting up Ansible
	log.Printf("📝 Updating job status to configuring...")
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusConfiguring, instances, "")

	// Update instance status to provisioning
	log.Printf("📝 Updating instance statuses to provisioning...")
	dbInstances, err := s.repo.GetByJobID(ctx, job.ID)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to get instances from database: %v", err)
	} else {
		for i := range dbInstances {
			if err := s.repo.UpdateStatus(ctx, dbInstances[i].ID, models.InstanceStatusProvisioning); err != nil {
				log.Printf("⚠️ Warning: Failed to update instance %s status: %v", dbInstances[i].Name, err)
			} else {
				log.Printf("✅ Updated instance %s status to provisioning", dbInstances[i].Name)
			}
		}
	}

	log.Printf("🔧 Running provisioning process...")
	if err := infra.RunProvisioning(instances); err != nil {
		log.Printf("❌ Provisioning failed: %v", err)
		return fmt.Errorf("infrastructure created but provisioning failed: %w", err)
	}

	// Update instance status to ready after successful provisioning
	log.Printf("📝 Updating instance statuses to ready...")
	dbInstances, err = s.repo.GetByJobID(ctx, job.ID)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to get instances from database: %v", err)
	} else {
		for i := range dbInstances {
			if err := s.repo.UpdateStatus(ctx, dbInstances[i].ID, models.InstanceStatusReady); err != nil {
				log.Printf("⚠️ Warning: Failed to update instance %s status: %v", dbInstances[i].Name, err)
			} else {
				log.Printf("✅ Updated instance %s status to ready", dbInstances[i].Name)
			}
		}
	}

	log.Printf("✅ Provisioning completed successfully")
	return nil
}

// handleInfrastructureDeletion handles the infrastructure deletion process
func (s *InstanceService) handleInfrastructureDeletion(
	ctx context.Context,
	job *models.Job,
	infraReq *infrastructure.JobRequest,
) error {
	log.Printf("🗑️ Starting async deletion for job %d", job.ID)

	// Update to initializing when starting setup
	log.Printf("📝 Updating job status to initializing...")
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusInitializing, nil, "")

	// Get instances from database for this job, ordered by creation time
	log.Printf("🔍 Getting instances from database for job %d...", job.ID)
	instances, err := s.repo.GetByJobIDOrdered(ctx, job.ID)
	if err != nil {
		log.Printf("❌ Failed to get instances from database: %v", err)
		return fmt.Errorf("failed to get instances: %w", err)
	}

	if len(instances) == 0 {
		log.Printf("❌ No instances found for job %d", job.ID)
		return fmt.Errorf("no instances found for job %d", job.ID)
	}

	// TODO: this section of code could be refactored to be unit tested
	// Check if we have specific instance names to delete
	var instancesToDelete []models.Instance
	var specificNamesToDelete []string
	var numberOfInstancesToDelete int

	// Collect specific instance names to delete and track the total number of instances to delete
	for _, instanceReq := range infraReq.Instances {
		if instanceReq.Name != "" {
			specificNamesToDelete = append(specificNamesToDelete, instanceReq.Name)
		}
		numberOfInstancesToDelete += instanceReq.NumberOfInstances
	}

	// Check if we have specific instance names to delete
	if len(specificNamesToDelete) > 0 {
		for _, instance := range instances {
			if slices.Contains(specificNamesToDelete, instance.Name) {
				instancesToDelete = append(instancesToDelete, instance)
			}
		}
	} else {
		// If no specific instance names are provided, delete the oldest instances
		instancesToDelete = instances[:numberOfInstancesToDelete]
	}

	if numberOfInstancesToDelete <= 0 {
		log.Printf("❌ Invalid number of instances to delete: %d", numberOfInstancesToDelete)
		return fmt.Errorf("invalid number of instances to delete: %d", numberOfInstancesToDelete)
	}

	// Limit the number of instances to delete to the available ones
	if numberOfInstancesToDelete > len(instances) {
		log.Printf("⚠️ Warning: Requested to delete %d instances but only %d are available",
			numberOfInstancesToDelete, len(instances))
	}

	// Prepare deletion result
	deletionResult := map[string]interface{}{
		"status":  "deleting",
		"deleted": []string{},
	}

	log.Printf("ℹ️ Will delete %d oldest instances out of %d total instances",
		numberOfInstancesToDelete, len(instances))

	// Try to delete each selected instance
	for _, instance := range instancesToDelete {
		log.Printf("🗑️ Attempting to delete instance: %s", instance.Name)

		// Create a new infrastructure request for each instance
		instanceInfraReq := &infrastructure.JobRequest{
			InstanceName: instance.Name,
			ProjectName:  infraReq.ProjectName,
			Provider:     infraReq.Provider,
			Instances: []infrastructure.InstanceRequest{{
				Provider: string(instance.ProviderID),
				Region:   instance.Region,
				Size:     instance.Size,
			}},
			Action: "delete",
		}

		// Create infrastructure client for this specific instance
		infra, err := infrastructure.NewInfrastructure(instanceInfraReq)
		if err != nil {
			fmt.Printf("❌ Failed to create infrastructure client for instance %s: %v\n", instance.Name, err)
			continue
		}

		// Execute the deletion for this specific instance
		_, err = infra.Execute()
		if err != nil {
			if !(strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found")) {
				fmt.Printf("❌ Error deleting instance %s: %v\n", instance.Name, err)
				continue
			}
			fmt.Printf("⚠️ Warning: Instance %s was already deleted\n", instance.Name)
		}

		// Instance doesn't exist in DO or deletion was successful, safe to mark as deleted
		if err := s.repo.Delete(ctx, instance.ID); err != nil {
			fmt.Printf("❌ Failed to mark instance %s as deleted in database: %v\n", instance.Name, err)
			continue
		}

		fmt.Printf("✅ Marked instance %s as deleted in database\n", instance.Name)
		if deleted, ok := deletionResult["deleted"].([]string); ok {
			deletionResult["deleted"] = append(deleted, instance.Name)
		}
	}

	deletionResult["status"] = "completed"

	// Update final status with result
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusCompleted, deletionResult, "")

	fmt.Printf("✅ Infrastructure deletion completed for job %d\n", job.ID)
	return nil
}

// Update job status with error handling
func (s *InstanceService) updateJobStatusWithError(
	ctx context.Context,
	jobID uint,
	status models.JobStatus,
	result interface{},
	errMsg string,
) {
	if err := s.jobService.UpdateJobStatus(ctx, jobID, status, result, errMsg); err != nil {
		log.Printf("Failed to update job status: %v", err)
	}
}

// GetInstancesByJobID retrieves all instances for a specific job
func (s *InstanceService) GetInstancesByJobID(ctx context.Context, jobID uint) ([]models.Instance, error) {
	fmt.Printf("📥 Getting instances for job ID %d from database...\n", jobID)

	// Get instances for the specific job
	instances, err := s.repo.GetByJobID(ctx, jobID)
	if err != nil {
		fmt.Printf("❌ Error getting instances for job %d: %v\n", jobID, err)
		return nil, fmt.Errorf("failed to get instances for job %d: %w", jobID, err)
	}

	fmt.Printf("✅ Retrieved %d instances for job %d from database\n", len(instances), jobID)
	return instances, nil
}

// CreateInstancesForJob creates instances for an existing job
func (s *InstanceService) CreateInstancesForJob(
	ctx context.Context,
	job *models.Job,
	instances []infrastructure.InstanceRequest,
) error {
	log.Printf("📝 Starting instance creation process for job %d", job.ID)

	// Validate instances array is not empty
	if len(instances) == 0 {
		log.Printf("❌ No instances provided in request")
		return fmt.Errorf("no instances provided in request")
	}

	// Create job request
	jobReq := &infrastructure.JobRequest{
		JobName:      job.Name,
		InstanceName: job.Name,
		ProjectName:  job.ProjectName,
		Provider:     instances[0].Provider,
		Instances:    instances,
		Action:       "create",
	}

	log.Printf("🔍 Validating job request: %+v", jobReq)
	if err := jobReq.Validate(); err != nil {
		log.Printf("❌ Job request validation failed: %v", err)
		return fmt.Errorf("invalid request: %w", err)
	}

	// Create pending instances in the database
	instanceConfig := instances[0] // Safe because we validate there's at least one instance in the request
	log.Printf("📝 Will create %d instances with config: %+v", instanceConfig.NumberOfInstances, instanceConfig)

	for i := 0; i < instanceConfig.NumberOfInstances; i++ {
		// Create default tags if none provided
		var tags []string
		if len(instanceConfig.Tags) > 0 {
			tags = instanceConfig.Tags
		} else {
			tags = []string{fmt.Sprintf("%s-do-instance", job.Name)}
		}

		// Create instance in database with pending status
		instance := &models.Instance{
			JobID:      job.ID,
			ProviderID: models.ProviderID(jobReq.Provider),
			Name:       fmt.Sprintf("%s-%d", job.Name, i), // Match DO's naming convention
			Region:     instanceConfig.Region,
			Size:       instanceConfig.Size,
			Image:      instanceConfig.Image,
			Tags:       tags,
			Status:     models.InstanceStatusPending,
		}

		log.Printf("📝 Creating instance %d/%d in database: %+v", i+1, instanceConfig.NumberOfInstances, instance)

		if err := s.repo.Create(ctx, instance); err != nil {
			log.Printf("❌ Failed to create instance in database:\nInstance data: %+v\nError details: %v\nError type: %T",
				instance, err, err)

			// Log more details about the error
			log.Printf("🔍 Detailed error information:")
			log.Printf("  - Error type: %T", err)
			log.Printf("  - Error message: %v", err)
			log.Printf("  - Instance data: %+v", instance)

			// If this is the first instance and it fails, we should fail the entire operation
			if i == 0 {
				log.Printf("❌ First instance creation failed, aborting operation")
				// Try to update job status to failed
				updateErr := s.jobService.UpdateJobStatus(ctx, job.ID, models.JobStatusFailed, nil,
					fmt.Sprintf("Failed to create first instance: %v", err))
				if updateErr != nil {
					log.Printf("❌ Additionally failed to update job status: %v", updateErr)
				}
				return fmt.Errorf("failed to create first instance in database: %w", err)
			}

			// For subsequent instances, continue but log the error
			continue
		}
		log.Printf("✅ Successfully created instance %s in database with ID %d", instance.Name, instance.ID)
	}

	log.Printf("🚀 Starting async infrastructure creation for job %d", job.ID)
	// Start async infrastructure creation
	go func() {
		if err := s.handleInfrastructureCreation(context.Background(), job, jobReq, instances); err != nil {
			log.Printf("❌ Failed to handle infrastructure creation: %v", err)
			s.updateJobStatusWithError(context.Background(), job.ID, models.JobStatusFailed, nil, err.Error())
		}
	}()

	log.Printf("✅ Instance creation process initiated successfully for job %d", job.ID)
	return nil
}
