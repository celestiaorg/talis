package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/celestiaorg/talis/internal/api/v1/routes"
)

// GetJob gets a specific job by ID
func (c *APIClient) GetJob(ctx context.Context, id string) (interface{}, error) {
	// Create agent for the request
	agent, err := c.createAgent(ctx, http.MethodGet, routes.GetJobURL(id), nil)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	if err := c.doRequest(agent, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ListJobs lists all jobs with optional filtering
func (c *APIClient) ListJobs(ctx context.Context, limit int, status string) (interface{}, error) {
	endpoint := routes.ListJobsURL()

	// Add query parameters if provided
	query := url.Values{}
	if limit > 0 {
		query.Add("limit", strconv.Itoa(limit))
	}
	if status != "" {
		query.Add("status", status)
	}

	if len(query) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, query.Encode())
	}

	// Create agent for the request
	agent, err := c.createAgent(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response []map[string]interface{}
	if err := c.doRequest(agent, &response); err != nil {
		return nil, err
	}

	return response, nil
}
