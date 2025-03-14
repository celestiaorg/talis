package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client defines the interface for interacting with the Talis API
type Client interface {
	// Infrastructure methods
	CreateInfrastructure(ctx context.Context, req interface{}) (interface{}, error)
	DeleteInfrastructure(ctx context.Context, req interface{}) (interface{}, error)
	GetInfrastructure(ctx context.Context, id string) (interface{}, error)
	ListInfrastructure(ctx context.Context) (interface{}, error)

	// Job methods
	GetJob(ctx context.Context, id string) (interface{}, error)
	ListJobs(ctx context.Context, limit int, status string) (interface{}, error)
}

// ClientOptions contains configuration options for the API client
type ClientOptions struct {
	// BaseURL is the base URL of the API
	BaseURL string

	// HTTPClient is the HTTP client to use for requests
	HTTPClient *http.Client

	// Timeout is the request timeout
	Timeout time.Duration
}

// DefaultOptions returns the default client options
func DefaultOptions() *ClientOptions {
	return &ClientOptions{
		BaseURL: "http://localhost:8080",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Timeout: 30 * time.Second,
	}
}

// APIClient implements the Client interface
type APIClient struct {
	baseURL    *url.URL
	httpClient *http.Client
	timeout    time.Duration
}

// NewClient creates a new API client with the given options
func NewClient(opts *ClientOptions) (Client, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	baseURL, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: opts.Timeout,
		}
	}

	return &APIClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		timeout:    opts.Timeout,
	}, nil
}

// newRequest creates a new HTTP request with the given method, endpoint, and body
func (c *APIClient) newRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	// Resolve the endpoint against the base URL
	resolvedURL := c.baseURL.ResolveReference(u)

	req, err := http.NewRequestWithContext(ctx, method, resolvedURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set common headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// doRequest sends the HTTP request and decodes the response
func (c *APIClient) doRequest(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Check for non-success status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Try to decode the error response
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			return NewAPIError(resp.StatusCode, errResp.Message, string(body))
		}

		// If we can't decode the error response, return a generic error
		return NewAPIError(resp.StatusCode, "unknown error", string(body))
	}

	// Decode the response body if a target is provided
	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}
	}

	return nil
}

// marshalRequest marshals the request body to JSON
func marshalRequest(req interface{}) (io.Reader, error) {
	if req == nil {
		return nil, nil
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	return bytes.NewBuffer(jsonData), nil
}
