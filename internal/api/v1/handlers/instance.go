package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// InstanceHandler handles HTTP requests for instance operations
type InstanceHandler struct {
	service *services.InstanceService
}

// NewInstanceHandler creates a new instance handler instance
func NewInstanceHandler(service *services.InstanceService) *InstanceHandler {
	return &InstanceHandler{
		service: service,
	}
}

// DeleteInstance handles the request to delete an instance
func (h *InstanceHandler) DeleteInstance(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "delete instance not implemented yet",
	})

	// var req infrastructure.DeleteRequest
	// if err := c.BodyParser(&req); err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"error": err.Error(),
	// 	})
	// }

	// // Create job first
	// j, err := h.jobService.CreateJob(context.Background(), req.Name, req.WebhookURL)
	// if err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"error": err.Error(),
	// 	})
	// }

	// // Convert InstanceRequest to Request
	// infraReq := &infrastructure.JobRequest{
	// 	Name:        req.Name,
	// 	ProjectName: req.ProjectName,
	// 	Provider:    req.Instances[0].Provider,
	// 	Instances:   convertToInstances(req.Instances),
	// 	Action:      "delete",
	// }

	// // Start async infrastructure deletion
	// go func() {
	// 	fmt.Println("🗑️ Starting async infrastructure deletion...")

	// 	// Update to initializing when starting Pulumi setup
	// 	if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusInitializing, nil, ""); err != nil {
	// 		fmt.Printf("❌ Failed to update job status to initializing: %v\n", err)
	// 		return
	// 	}

	// 	infra, err := infrastructure.NewInfrastructure(infraReq)
	// 	if err != nil {
	// 		fmt.Printf("❌ Failed to create infrastructure client: %v\n", err)
	// 		h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil, err.Error())
	// 		return
	// 	}

	// 	// Update to deleting status
	// 	if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusProvisioning, nil, ""); err != nil {
	// 		fmt.Printf("❌ Failed to update job status to deleting: %v\n", err)
	// 		return
	// 	}

	// 	result, err := infra.Execute()
	// 	if err != nil {
	// 		// Verificar si el error es por un recurso que no existe
	// 		if strings.Contains(err.Error(), "404") &&
	// 			strings.Contains(err.Error(), "could not be found") {
	// 			fmt.Printf("⚠️ Warning: Resources were already deleted\n")
	// 			result = map[string]string{
	// 				"status": "deleted",
	// 				"note":   "resources were already deleted",
	// 			}
	// 		} else {
	// 			fmt.Printf("❌ Failed to delete infrastructure: %v\n", err)
	// 			h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil, err.Error())
	// 			return
	// 		}
	// 	}

	// 	// Update final status with result
	// 	if err := h.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusDeleted, result, ""); err != nil {
	// 		fmt.Printf("❌ Failed to update final job status: %v\n", err)
	// 		return
	// 	}

	// 	fmt.Printf("✅ Infrastructure deletion completed for job %s\n", j.ID)
	// }()

	// return c.Status(fiber.StatusAccepted).JSON(j)
}

// GetInstance returns details of a specific instance
func (h *InstanceHandler) GetInstance(c *fiber.Ctx) error {
	instanceID := c.Params("id")
	if instanceID == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput("instance id is required"))
	}

	// TODO: Implement instance retrieval
	return c.Status(fiber.StatusNotImplemented).
		JSON(errGeneral("get instance not implemented yet"))
}

// convertToInstances converts DeleteInstance to InstanceRequest
//
//nolint:unused // Will be used in future implementation
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
