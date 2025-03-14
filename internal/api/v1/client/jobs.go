package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// GetJob gets a specific job by ID
func (c *APIClient) GetJob(ctx context.Context, id string) (interface{}, error) {
	endpoint := fmt.Sprintf("/api/v1/jobs/%s", id)

	httpReq, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	if err := c.doRequest(httpReq, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ListJobs lists all jobs with optional filtering
func (c *APIClient) ListJobs(ctx context.Context, limit int, status string) (interface{}, error) {
	endpoint := "/api/v1/jobs"

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

	httpReq, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response []map[string]interface{}
	if err := c.doRequest(httpReq, &response); err != nil {
		return nil, err
	}

	return response, nil
}
