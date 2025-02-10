package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/tty47/talis/internal/types/infrastructure"
)

// HandleInfrastructure handles both creation and deletion of instances
func HandleInfrastructure(c *fiber.Ctx) error {
	var req infrastructure.InstanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Validate action
	if req.Action != "create" && req.Action != "delete" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("invalid action: %s", req.Action),
		})
	}

	var results []interface{}
	var provisioningErrors []string

	for _, instance := range req.Instances {
		infra, err := infrastructure.NewInfrastructure(req, instance)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to create infrastructure: %v", err),
			})
		}

		result, err := infra.Execute()
		if err != nil {
			if req.Action == "create" && result != nil {
				// Infrastructure created but provisioning failed
				provisioningErrors = append(provisioningErrors, err.Error())
				results = append(results, result)
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("failed to %s infrastructure: %v", req.Action, err),
				})
			}
		} else if result != nil {
			results = append(results, result)
		}
	}

	response := fiber.Map{
		"message": fmt.Sprintf("Infrastructure %s completed", req.Action),
	}

	if len(results) > 0 {
		response["results"] = results
	}

	if len(provisioningErrors) > 0 {
		response["provisioning_errors"] = provisioningErrors
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// DeleteInstance handles the deletion of instances through URL parameter
func DeleteInstance(c *fiber.Ctx) error {
	stackName := c.Params("stack")
	if stackName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Stack name is required",
		})
	}

	// Create a minimal request for deletion
	req := infrastructure.InstanceRequest{
		Name:        stackName,
		ProjectName: stackName,
		Action:      "delete",
		Instances: []infrastructure.Instance{
			{
				Provider: "digitalocean", // Default provider, could be made configurable
			},
		},
	}

	infra, err := infrastructure.NewInfrastructure(req, req.Instances[0])
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	result, err := infra.Execute()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Infrastructure deleted successfully",
		"result":  result,
	})
}
