package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/celestiaorg/talis/internal/api/v1/routes"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// CreateInfrastructure creates new infrastructure
func (c *APIClient) CreateInfrastructure(ctx context.Context, req interface{}) (interface{}, error) {
	// Validate request
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Create agent for the request
	agent, err := c.createAgent(ctx, http.MethodPost, routes.CreateInstanceURL(), req)
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

	// Create agent for the request
	agent, err := c.createAgent(ctx, http.MethodDelete, routes.DeleteInstanceURL(), req)
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
	// Create agent for the request
	agent, err := c.createAgent(ctx, http.MethodGet, routes.GetInstanceURL(id), nil)
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
	// Create agent for the request
	agent, err := c.createAgent(ctx, http.MethodGet, routes.ListInstancesURL(), nil)
	if err != nil {
		return nil, err
	}

	var response []infrastructure.Response
	if err := c.doRequest(agent, &response); err != nil {
		return nil, err
	}

	return response, nil
}
