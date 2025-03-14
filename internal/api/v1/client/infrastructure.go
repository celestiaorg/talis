package client

import (
	"context"
	"fmt"
	"net/http"
)

// CreateInfrastructure creates new infrastructure
func (c *APIClient) CreateInfrastructure(ctx context.Context, req interface{}) (interface{}, error) {
	// Validate request
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	endpoint := "/api/v1/instances"

	body, err := marshalRequest(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := c.newRequest(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}

	var response CreateResponse
	if err := c.doRequest(httpReq, &response); err != nil {
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

	body, err := marshalRequest(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := c.newRequest(ctx, http.MethodDelete, endpoint, body)
	if err != nil {
		return nil, err
	}

	var response DeleteResponse
	if err := c.doRequest(httpReq, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetInfrastructure gets information about infrastructure
func (c *APIClient) GetInfrastructure(ctx context.Context, id string) (interface{}, error) {
	endpoint := fmt.Sprintf("/api/v1/instances/%s", id)

	httpReq, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response CreateResponse
	if err := c.doRequest(httpReq, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ListInfrastructure lists all infrastructure
func (c *APIClient) ListInfrastructure(ctx context.Context) (interface{}, error) {
	endpoint := "/api/v1/instances"

	httpReq, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response []CreateResponse
	if err := c.doRequest(httpReq, &response); err != nil {
		return nil, err
	}

	return response, nil
}
