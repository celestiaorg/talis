// Package handlers provides HTTP request handling
package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

// SSHKeyHandlers provides HTTP handlers for SSH key operations
type SSHKeyHandlers struct {
	SSHKeyService SSHKeyService
}

// SSHKeyService defines the interface for SSH key-related operations
// This is an interface to allow for easy mocking in tests
type SSHKeyService interface {
	CreateSSHKey(ctx context.Context, key *models.SSHKey) error
	ListKeys(ctx context.Context, ownerID uint) ([]*models.SSHKey, error)
	DeleteKey(ctx context.Context, ownerID uint, name string) error
}

// SSHKeyCreateParams defines the parameters for creating an SSH key
type SSHKeyCreateParams struct {
	Name      string `json:"name" validate:"required"`
	PublicKey string `json:"public_key" validate:"required"`
	OwnerID   uint   `json:"owner_id" validate:"required"`
}

// SSHKeyListParams defines the parameters for listing SSH keys
type SSHKeyListParams struct {
	OwnerID uint `json:"owner_id" validate:"required"`
}

// SSHKeyDeleteParams defines the parameters for deleting an SSH key
type SSHKeyDeleteParams struct {
	Name    string `json:"name" validate:"required"`
	OwnerID uint   `json:"owner_id" validate:"required"`
}

// ValidateSSHPublicKey checks if the provided string is a valid SSH public key
// by verifying it starts with a recognized SSH key prefix
func ValidateSSHPublicKey(key string) error {
	// List of valid SSH public key prefixes
	validPrefixes := []string{
		"ssh-rsa ",
		"ssh-dss ",
		"ssh-ed25519 ",
		"ecdsa-sha2-nistp256 ",
		"ecdsa-sha2-nistp384 ",
		"ecdsa-sha2-nistp521 ",
	}

	// Trim whitespace
	key = strings.TrimSpace(key)

	// Check if the key starts with any of the valid prefixes
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(key, prefix) {
			return nil
		}
	}

	return errors.New("invalid SSH public key format: must begin with ssh-rsa, ssh-ed25519, ecdsa-sha2-*, or ssh-dss")
}

// Create handles the creation of a new SSH key
// @Summary Create a new SSH key
// @Description Create a new SSH key for the specified owner
// @Tags sshkey
// @Accept json
// @Produce json
// @Param request body RPCRequest true "SSH key creation request"
// @Success 200 {object} RPCResponse
// @Failure 400 {object} RPCResponse
// @Failure 500 {object} RPCResponse
// @Router / [post]
func (h *SSHKeyHandlers) Create(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[SSHKeyCreateParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, "Invalid parameters", err.Error(), req.ID)
	}

	// Validate required fields
	if params.Name == "" {
		return respondWithRPCError(c, fiber.StatusBadRequest, "Name is required", nil, req.ID)
	}
	if params.PublicKey == "" {
		return respondWithRPCError(c, fiber.StatusBadRequest, "Public key is required", nil, req.ID)
	}
	if params.OwnerID == 0 {
		return respondWithRPCError(c, fiber.StatusBadRequest, "Owner ID is required", nil, req.ID)
	}

	// Validate SSH public key format
	if err := ValidateSSHPublicKey(params.PublicKey); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	// Create the SSH key
	key := &models.SSHKey{
		Name:      params.Name,
		PublicKey: params.PublicKey,
		OwnerID:   params.OwnerID,
	}

	err = h.SSHKeyService.CreateSSHKey(c.Context(), key)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, "Failed to create SSH key", err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Data:    key,
		Success: true,
		ID:      req.ID,
	})
}

// List handles listing SSH keys for an owner
// @Summary List SSH keys
// @Description List all SSH keys for the specified owner
// @Tags sshkey
// @Accept json
// @Produce json
// @Param request body RPCRequest true "SSH key list request"
// @Success 200 {object} RPCResponse
// @Failure 400 {object} RPCResponse
// @Failure 500 {object} RPCResponse
// @Router / [post]
func (h *SSHKeyHandlers) List(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[SSHKeyListParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, "Invalid parameters", err.Error(), req.ID)
	}

	// List SSH keys
	keys, err := h.SSHKeyService.ListKeys(c.Context(), params.OwnerID)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, "Failed to list SSH keys", err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Data:    keys,
		Success: true,
		ID:      req.ID,
	})
}

// Delete handles deleting an SSH key
// @Summary Delete an SSH key
// @Description Delete an SSH key for the specified owner
// @Tags sshkey
// @Accept json
// @Produce json
// @Param request body RPCRequest true "SSH key delete request"
// @Success 200 {object} RPCResponse
// @Failure 400 {object} RPCResponse
// @Failure 500 {object} RPCResponse
// @Router / [post]
func (h *SSHKeyHandlers) Delete(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[SSHKeyDeleteParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, "Invalid parameters", err.Error(), req.ID)
	}

	// Delete SSH key
	err = h.SSHKeyService.DeleteKey(c.Context(), params.OwnerID, params.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return respondWithRPCError(c, fiber.StatusNotFound, "SSH key not found", err.Error(), req.ID)
		}
		return respondWithRPCError(c, fiber.StatusInternalServerError, "Failed to delete SSH key", err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Success: true,
		ID:      req.ID,
	})
}
