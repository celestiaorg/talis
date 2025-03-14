package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// CreateInfrastructure creates new infrastructure
func (c *APIClient) CreateInfrastructure(ctx context.Context, req interface{}) (interface{}, error) {
	// Validate request
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	endpoint := "/api/v1/instances"

	// Create agent for the request
	agent, err := c.createAgent(ctx, http.MethodPost, endpoint, req)
	if err != nil {
		return nil, err
	}

	var response infrastructure.Response
	if err := c.doRequest(agent, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// DeleteInfrastructure deletes infrastructure
func (c *APIClient) DeleteInfrastructure(ctx context.Context, req interface{}) (interface{}, error) {
	// Validate request
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	endpoint := "/api/v1/instances"

	// Create agent for the request
	agent, err := c.createAgent(ctx, http.MethodDelete, endpoint, req)
	if err != nil {
		return nil, err
	}

	var response infrastructure.Response
	if err := c.doRequest(agent, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetInfrastructure gets information about infrastructure
func (c *APIClient) GetInfrastructure(ctx context.Context, id string) (interface{}, error) {
	endpoint := fmt.Sprintf("/api/v1/instances/%s", id)

	// Create agent for the request
	agent, err := c.createAgent(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response infrastructure.Response
	if err := c.doRequest(agent, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ListInfrastructure lists all infrastructure
func (c *APIClient) ListInfrastructure(ctx context.Context) (interface{}, error) {
	endpoint := "/api/v1/instances"

	// Create agent for the request
	agent, err := c.createAgent(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response []infrastructure.Response
	if err := c.doRequest(agent, &response); err != nil {
		return nil, err
	}

	return response, nil
}
