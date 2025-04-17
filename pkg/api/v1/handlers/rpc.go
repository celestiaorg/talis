// Package handlers provides HTTP request handling
package handlers

import (
	"encoding/json"

	"github.com/celestiaorg/talis/internal/db/models"

	fiber "github.com/gofiber/fiber/v2"
)

// RPCRequest defines the structure for RPC-style API requests
type RPCRequest struct {
	// Method is the operation to perform (e.g., "project.create", "task.list")
	Method string `json:"method"`

	// Params contains the operation parameters
	Params interface{} `json:"params"`

	// ID is an optional request identifier that will be echoed back in the response
	ID string `json:"id,omitempty"`
}

// RPCResponse defines the structure for RPC-style API responses
type RPCResponse struct {
	// Data contains the operation result
	Data interface{} `json:"data,omitempty"`

	// Error contains error information if the operation failed
	Error *RPCError `json:"error,omitempty"`

	// ID echoes back the request ID if provided
	ID string `json:"id,omitempty"`

	// Success indicates if the operation was successful
	Success bool `json:"success"`
}

// RPCError defines the structure for RPC errors
type RPCError struct {
	// Code is a numeric error code
	Code int `json:"code"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Data contains additional error details (optional)
	Data interface{} `json:"data,omitempty"`
}

// RPCHandler handles RPC-style API requests for projects and tasks
type RPCHandler struct {
	ProjectHandlers *ProjectHandlers
	TaskHandlers    *TaskHandlers
}

// HandleRPC handles all RPC requests for various resource types
func (h *RPCHandler) HandleRPC(c *fiber.Ctx) error {
	var req RPCRequest
	if err := c.BodyParser(&req); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, "Invalid request format", err.Error(), req.ID)
	}

	// Check if method is provided
	if req.Method == "" {
		return respondWithRPCError(c, fiber.StatusBadRequest, "Method is required", nil, req.ID)
	}

	// Route to appropriate handler based on method prefix
	switch {
	case IsProjectMethod(req.Method):
		return h.handleProjectMethod(c, req)
	case IsTaskMethod(req.Method):
		return h.handleTaskMethod(c, req)
	default:
		return respondWithRPCError(c, fiber.StatusBadRequest, "Unknown method", nil, req.ID)
	}
}

// handleProjectMethod routes project methods to their respective handlers
func (h *RPCHandler) handleProjectMethod(c *fiber.Ctx, req RPCRequest) error {
	if h.ProjectHandlers == nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, "Project handlers not configured", nil, req.ID)
	}

	ownerID := models.AdminID // TODO: get owner id from the JWT token

	switch req.Method {
	case ProjectCreate:
		return h.ProjectHandlers.Create(c, ownerID, req)
	case ProjectGet:
		return h.ProjectHandlers.Get(c, ownerID, req)
	case ProjectList:
		return h.ProjectHandlers.List(c, ownerID, req)
	case ProjectDelete:
		return h.ProjectHandlers.Delete(c, ownerID, req)
	case ProjectListInstances:
		return h.ProjectHandlers.ListInstances(c, ownerID, req)
	default:
		return respondWithRPCError(c, fiber.StatusBadRequest, "Unknown project method", nil, req.ID)
	}
}

// handleTaskMethod routes task methods to their respective handlers
func (h *RPCHandler) handleTaskMethod(c *fiber.Ctx, req RPCRequest) error {
	if h.TaskHandlers == nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, "Task handlers not configured", nil, req.ID)
	}

	ownerID := models.AdminID // TODO: get owner id from the JWT token

	switch req.Method {
	case TaskGet:
		return h.TaskHandlers.Get(c, ownerID, req)
	case TaskList:
		return h.TaskHandlers.List(c, ownerID, req)
	case TaskTerminate:
		return h.TaskHandlers.Terminate(c, ownerID, req)
	default:
		return respondWithRPCError(c, fiber.StatusBadRequest, "Unknown task method", nil, req.ID)
	}
}

// parseParams is a helper function to parse RPC parameters into a specific struct type
func parseParams[T any](req RPCRequest) (T, error) {
	var params T

	// Convert params to JSON
	paramsJSON, err := json.Marshal(req.Params)
	if err != nil {
		return params, err
	}

	// Unmarshal to target type
	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return params, err
	}

	return params, nil
}

// Helper to create a standardized RPC error response
func respondWithRPCError(c *fiber.Ctx, httpCode int, message string, data interface{}, id string) error {
	return c.Status(httpCode).JSON(RPCResponse{
		Error: &RPCError{
			Code:    httpCode,
			Message: message,
			Data:    data,
		},
		Success: false,
		ID:      id,
	})
}
