package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// InstanceHandler handles HTTP requests for instance operations
type InstanceHandler struct {
	service    *services.InstanceService
	jobService *services.JobService
}

// NewInstanceHandler creates a new instance handler instance
func NewInstanceHandler(service *services.InstanceService, jobService *services.JobService) *InstanceHandler {
	return &InstanceHandler{
		service:    service,
		jobService: jobService,
	}
}

// CreateInstance handles the request to create a new instance
func (h *InstanceHandler) CreateInstance(c *fiber.Ctx) error {
	var req struct {
		Name        string                           `json:"name"`
		ProjectName string                           `json:"project_name"`
		WebhookURL  string                           `json:"webhook_url"`
		Instances   []infrastructure.InstanceRequest `json:"instances"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert to Request and validate
	jobReq := &infrastructure.JobRequest{
		Name:        req.Name,
		ProjectName: req.ProjectName,
		Provider:    req.Instances[0].Provider,
		Instances:   req.Instances,
		Action:      "create",
	}

	if err := jobReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create job first
	job := &models.Job{
		Name:        req.Name,
		ProjectName: req.ProjectName,
		Status:      models.JobStatusPending,
		WebhookURL:  req.WebhookURL,
	}

	j, err := h.jobService.CreateJob(context.Background(), job)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Start async infrastructure creation
	go func() {
		fmt.Println("🚀 Starting async infrastructure creation...")

		// Update to initializing when starting setup
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusInitializing, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to initializing: %v\n", err)
			return
		}

		infra, err := infrastructure.NewInfrastructure(jobReq)
		if err != nil {
			fmt.Printf("❌ Failed to create infrastructure client: %v\n", err)
			h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil, err.Error())
			return
		}

		// Update to provisioning when creating infrastructure
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusProvisioning, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to provisioning: %v\n", err)
			return
		}

		result, err := infra.Execute()
		if err != nil {
			// Check if error is due to resource not found
			if strings.Contains(err.Error(), "404") &&
				strings.Contains(err.Error(), "could not be found") {
				fmt.Printf("⚠️ Warning: Some old resources were not found (already deleted)\n")
			} else {
				fmt.Printf("❌ Failed to execute infrastructure: %v\n", err)
				h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil, err.Error())
				return
			}
		}

		// Start Nix provisioning if creation was successful and provisioning is requested
		if req.Instances[0].Provision {
			instances, ok := result.([]infrastructure.InstanceInfo)
			if !ok {
				fmt.Printf("❌ Invalid result type: %T\n", result)
				h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil,
					fmt.Sprintf("invalid result type: %T, expected []infrastructure.InstanceInfo", result))
				return
			}

			fmt.Printf("📝 Created instances: %+v\n", instances)

			// Update to configuring when setting up Nix
			if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusConfiguring, instances, ""); err != nil {
				fmt.Printf("❌ Failed to update job status to configuring: %v\n", err)
				return
			}

			if err := infra.RunProvisioning(instances); err != nil {
				fmt.Printf("❌ Failed to run provisioning: %v\n", err)
				h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, instances, fmt.Sprintf("infrastructure created but provisioning failed: %v", err))
				return
			}
		}

		// Update final status with result
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusCompleted, result, ""); err != nil {
			fmt.Printf("❌ Failed to update final job status: %v\n", err)
			return
		}

		fmt.Printf("✅ Infrastructure creation completed for job %s\n", j.ID)
	}()

	return c.Status(fiber.StatusAccepted).JSON(j)
}

// DeleteInstance handles the request to delete an instance
func (h *InstanceHandler) DeleteInstance(c *fiber.Ctx) error {
	instanceID := c.Params("id")
	if instanceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "instance id is required",
		})
	}

	// Create job for deletion
	job := &models.Job{
		Name:        fmt.Sprintf("delete-%s", instanceID),
		ProjectName: "talis", // Default project for now
		Status:      models.JobStatusPending,
	}

	j, err := h.jobService.CreateJob(context.Background(), job)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create infrastructure request
	infraReq := &infrastructure.JobRequest{
		Name:        instanceID,
		ProjectName: "talis",
		Provider:    "digitalocean",
		Action:      "delete",
		Instances: []infrastructure.InstanceRequest{
			{
				Provider:          "digitalocean",
				NumberOfInstances: 1,
				Region:            "nyc3",
			},
		},
	}

	// Start async infrastructure deletion
	go func() {
		fmt.Printf("🗑️ Starting async deletion of instance: %s\n", instanceID)

		// Update to initializing when starting setup
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusInitializing, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to initializing: %v\n", err)
			return
		}

		infra, err := infrastructure.NewInfrastructure(infraReq)
		if err != nil {
			fmt.Printf("❌ Failed to create infrastructure client: %v\n", err)
			h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil, err.Error())
			return
		}

		// Update to deleting status
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusDeleting, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to deleting: %v\n", err)
			return
		}

		result, err := infra.Execute()
		if err != nil {
			// Check if the error is due to resources already being deleted
			if strings.Contains(err.Error(), "404") ||
				strings.Contains(err.Error(), "not found") {
				fmt.Printf("⚠️ Warning: Resources were already deleted\n")
				result = map[string]string{
					"status": "deleted",
					"note":   "resources were already deleted",
				}
				err = nil // Clear the error since this is an acceptable condition
			} else {
				fmt.Printf("❌ Failed to delete infrastructure: %v\n", err)
				h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil, err.Error())
				return
			}
		}

		// Update final status with result
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusCompleted, result, ""); err != nil {
			fmt.Printf("❌ Failed to update final job status: %v\n", err)
			return
		}

		fmt.Printf("✅ Infrastructure deletion completed for job %d\n", j.ID)
	}()

	return c.Status(fiber.StatusAccepted).JSON(j)
}

// GetInstance returns details of a specific instance
func (h *InstanceHandler) GetInstance(c *fiber.Ctx) error {
	instanceID := c.Params("id")
	if instanceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "instance id is required",
		})
	}

	// TODO: Implement instance retrieval
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "get instance not implemented yet",
	})
}

// convertToInstances converts DeleteInstance to InstanceRequest
func convertToInstances(deleteInstances []infrastructure.DeleteInstance) []infrastructure.InstanceRequest {
	instances := make([]infrastructure.InstanceRequest, len(deleteInstances))
	for i, di := range deleteInstances {
		instances[i] = infrastructure.InstanceRequest{
			Provider:          di.Provider,
			NumberOfInstances: di.NumberOfInstances,
			Region:            di.Region,
			Size:              di.Size,
			Image:             di.Image,
			Tags:              di.Tags,
			SSHKeyName:        di.SSHKeyName,
		}
	}
	return instances
}
