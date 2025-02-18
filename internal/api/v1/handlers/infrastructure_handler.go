package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/domain/job"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// Handler is the main handler for the API
type InfrastructureHandler struct {
	jobService job.Service // Service to manage jobs
}

// NewHandler creates a new Handler
func NewInfrastructureHandler(jobService job.Service) *InfrastructureHandler {
	return &InfrastructureHandler{
		jobService: jobService,
	}
}

// CreateInfrastructure creates a new infrastructure
func (h *InfrastructureHandler) CreateInfrastructure(c *fiber.Ctx) error {
	var req infrastructure.InstanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Set action based on HTTP method
	req.Action = "create"

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create job first
	j, err := h.jobService.CreateJob(c.Context(), req.WebhookURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Start async infrastructure creation
	go func() {
		fmt.Println("🚀 Starting async infrastructure creation...")

		// Update to initializing when starting Pulumi setup
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusInitializing, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to initializing: %v\n", err)
			return
		}

		infra, err := infrastructure.NewInfrastructure(req, req.Instances[0])
		if err != nil {
			fmt.Printf("❌ Failed to create infrastructure: %v\n", err)
			h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusFailed, nil, err.Error())
			return
		}

		// Update to provisioning when creating infrastructure
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusProvisioning, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to provisioning: %v\n", err)
			return
		}

		result, err := infra.Execute()
		if err != nil {
			// Check if error is due to resource not found
			if strings.Contains(err.Error(), "404") &&
				strings.Contains(err.Error(), "could not be found") {
				// Get Pulumi output result
				outputs, outputErr := infra.GetOutputs()
				if outputErr != nil {
					fmt.Printf("❌ Failed to get outputs: %v\n", outputErr)
					h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusFailed, nil, outputErr.Error())
					return
				}
				result = outputs
				fmt.Printf("⚠️ Warning: Some old resources were not found (already deleted)\n")
			} else {
				fmt.Printf("❌ Failed to execute infrastructure: %v\n", err)
				h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusFailed, nil, err.Error())
				return
			}
		}

		// Start Nix provisioning if creation was successful and provisioning is requested
		if req.Action == "create" && req.Instances[0].Provision {
			instances, ok := result.([]infrastructure.InstanceInfo)
			if !ok {
				fmt.Printf("❌ Invalid result type: %T\n", result)
				h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusFailed, nil,
					fmt.Sprintf("invalid result type: %T, expected []infrastructure.InstanceInfo", result))
				return
			}

			fmt.Printf("📝 Created instances: %+v\n", instances)

			// Update to configuring when setting up Nix
			if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusConfiguring, instances, ""); err != nil {
				fmt.Printf("❌ Failed to update job status to configuring: %v\n", err)
				return
			}

			if err := infra.RunProvisioning(instances); err != nil {
				fmt.Printf("❌ Failed to run provisioning: %v\n", err)
				h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusFailed, instances,
					fmt.Sprintf("infrastructure created but provisioning failed: %v", err))
				return
			}
		}

		// Update final status with result
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusCompleted, result, ""); err != nil {
			fmt.Printf("❌ Failed to update final job status: %v\n", err)
			return
		}

		fmt.Printf("✅ Infrastructure creation completed for job %s\n", j.ID)
	}()

	return c.Status(fiber.StatusCreated).JSON(j)
}

// DeleteInfrastructure deletes an infrastructure
func (h *InfrastructureHandler) DeleteInfrastructure(c *fiber.Ctx) error {
	var req infrastructure.DeleteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert DeleteRequest to InstanceRequest for backwards compatibility
	fullReq := infrastructure.InstanceRequest{
		Name:        req.Name,
		ProjectName: req.ProjectName,
		Action:      "delete",
		Instances:   make([]infrastructure.Instance, len(req.Instances)),
	}

	// Fill in only the required fields
	for i, inst := range req.Instances {
		fullReq.Instances[i] = infrastructure.Instance{
			Provider:          inst.Provider,
			NumberOfInstances: inst.NumberOfInstances,
			Region:            inst.Region,
			Size:              inst.Size,
		}
	}

	// Create job first
	j, err := h.jobService.CreateJob(c.Context(), req.WebhookURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Start async infrastructure deletion
	go func() {
		fmt.Println("🗑️ Starting async infrastructure deletion...")

		// Update to initializing when starting Pulumi setup
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusInitializing, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to initializing: %v\n", err)
			return
		}

		infra, err := infrastructure.NewInfrastructure(fullReq, fullReq.Instances[0])
		if err != nil {
			fmt.Printf("❌ Failed to create infrastructure client: %v\n", err)
			h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusFailed, nil, err.Error())
			return
		}

		// Update to deleting status
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusProvisioning, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to deleting: %v\n", err)
			return
		}

		result, err := infra.Execute()
		if err != nil {
			// Verificar si el error es por un recurso que no existe
			if strings.Contains(err.Error(), "404") &&
				strings.Contains(err.Error(), "could not be found") {
				fmt.Printf("⚠️ Warning: Resources were already deleted\n")
				result = map[string]string{
					"status": "deleted",
					"note":   "resources were already deleted",
				}
			} else {
				fmt.Printf("❌ Failed to delete infrastructure: %v\n", err)
				h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusFailed, nil, err.Error())
				return
			}
		}

		// Update final status with result
		if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, job.StatusDeleted, result, ""); err != nil {
			fmt.Printf("❌ Failed to update final job status: %v\n", err)
			return
		}

		fmt.Printf("✅ Infrastructure deletion completed for job %s\n", j.ID)
	}()

	return c.Status(fiber.StatusAccepted).JSON(j)
}

// GetInstance returns details of a specific instance
func (h *InfrastructureHandler) GetInstance(c *fiber.Ctx) error {
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
