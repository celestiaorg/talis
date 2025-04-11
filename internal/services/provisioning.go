// Package services provides business logic implementation for the API
package services

import (
	"context"
	"fmt"

	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/events"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

// Provisioning represents the provisioning service that handles
// configuration management of infrastructure resources.
type Provisioning struct {
	provisioner     compute.Provisioner
	instanceService *Instance
	jobService      *Job
}

// NewProvisioningService creates a new provisioning service instance and registers it as an event handler
func NewProvisioningService(instanceService *Instance, jobService *Job) *Provisioning {
	p := &Provisioning{
		instanceService: instanceService,
		jobService:      jobService,
	}

	// Subscribe to instance creation events
	events.Subscribe(events.EventInstancesCreated, p.handleInstancesCreated)
	// Subscribe to inventory generation requests
	events.Subscribe(events.EventInventoryRequested, p.handleInventoryRequested)

	return p
}

// getInstancesFromDB gets instances from the database for a given job
func (p *Provisioning) getInstancesFromDB(ctx context.Context, jobName string, ownerID uint) ([]types.InstanceInfo, error) {
	// Get job ID from job name
	jobIDUint, err := p.jobService.GetJobIDByName(ctx, ownerID, jobName)
	if err != nil {
		return nil, fmt.Errorf("failed to get job ID for job %s: %w", jobName, err)
	}

	// Get instances from database using the job ID
	instances, err := p.instanceService.GetInstancesByJobID(ctx, ownerID, jobIDUint)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances from database for job %s: %w", jobName, err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances found for job %s", jobName)
	}

	// Convert DB instances to InstanceInfo
	instanceInfos := make([]types.InstanceInfo, len(instances))
	for i, instance := range instances {
		// Convert VolumeDetails to the correct type
		volumeDetails := make([]types.VolumeDetails, 0, len(instance.VolumeDetails))
		for _, vd := range instance.VolumeDetails {
			volumeDetails = append(volumeDetails, types.VolumeDetails{
				ID:         vd.ID,
				Name:       vd.Name,
				Region:     vd.Region,
				SizeGB:     vd.SizeGB,
				MountPoint: vd.MountPoint,
			})
		}

		instanceInfos[i] = types.InstanceInfo{
			Name:          instance.Name,
			PublicIP:      instance.PublicIP,
			Provider:      instance.ProviderID,
			Region:        instance.Region,
			Size:          instance.Size,
			Tags:          instance.Tags,
			Volumes:       instance.Volumes,
			VolumeDetails: volumeDetails,
		}
	}

	return instanceInfos, nil
}

// handleInstancesCreated handles the instances created event
func (p *Provisioning) handleInstancesCreated(ctx context.Context, event events.Event) error {
	// Check if any instance requires provisioning
	needsProvisioning := false
	for _, req := range event.Requests {
		if req.Provision {
			needsProvisioning = true
			break
		}
	}

	if !needsProvisioning {
		logger.Info("⏭️ Skipping provisioning as requested")
		return nil
	}

	logger.InfoWithFields("🔄 Getting instances from database for provisioning", map[string]interface{}{
		"job_name": event.JobName,
		"owner_id": event.OwnerID,
	})

	// Always get instances from database to ensure we have the most up-to-date information
	instances, err := p.getInstancesFromDB(ctx, event.JobName, event.OwnerID)
	if err != nil {
		return fmt.Errorf("failed to get instances from database: %w", err)
	}

	logger.Info("⚙️ Starting provisioning...")

	// Initialize provisioner with job ID from event
	p.provisioner = compute.NewProvisioner(event.JobID)

	// Configure the provisioner with instances from DB
	if err := p.provisioner.Configure(ctx, instances); err != nil {
		return fmt.Errorf("failed to configure instances: %w", err)
	}

	return nil
}

// handleInventoryRequested handles requests to generate inventory from DB
func (p *Provisioning) handleInventoryRequested(ctx context.Context, event events.Event) error {
	logger.InfoWithFields("🔄 Generating inventory from database", map[string]interface{}{
		"job_name": event.JobName,
		"owner_id": event.OwnerID,
	})

	// Get instances from database
	instances, err := p.getInstancesFromDB(ctx, event.JobName, event.OwnerID)
	if err != nil {
		return fmt.Errorf("failed to get instances from database: %w", err)
	}

	// Configure the provisioner with instances from DB
	if err := p.provisioner.Configure(ctx, instances); err != nil {
		return fmt.Errorf("failed to configure instances: %w", err)
	}

	return nil
}
