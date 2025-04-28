package compute

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/celestiaorg/talis/internal/config"
	"github.com/celestiaorg/talis/internal/logger"
)

// APIError represents a VirtFusion API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

// Error implements the error interface for APIError
func (e *APIError) Error() string {
	return fmt.Sprintf("VirtFusion API error: %s - %s (status: %d)", e.Code, e.Message, e.Status)
}

// IsNotFound returns true if the error is a not found error
func (e *APIError) IsNotFound() bool {
	return e.Status == http.StatusNotFound
}

// IsRateLimited returns true if the error is a rate limit error
func (e *APIError) IsRateLimited() bool {
	return e.Status == http.StatusTooManyRequests
}

// IsServerError returns true if the error is a server error
func (e *APIError) IsServerError() bool {
	return e.Status >= http.StatusInternalServerError
}

const (
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
	retryDelay     = 2 * time.Second
)

// Client represents a VirtFusion API client
type Client struct {
	httpClient *http.Client
	config     *config.VirtFusionConfig
	baseURL    string
}

// NewClient creates a new VirtFusion API client
func NewClient(cfg *config.VirtFusionConfig) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Ensure baseURL doesn't end with a slash
	baseURL := cfg.Host
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		config:  cfg,
		baseURL: baseURL,
	}, nil
}

// doRequest performs an HTTP request with retries
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyData []byte
	var err error

	if body != nil {
		bodyData, err = json.Marshal(body)
		if err != nil {
			logger.Errorf("Failed to marshal request body: %v", err)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	// Ensure path starts with a slash
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}

	logger.Debugf("Preparing API request: method=%s, path=%s, body=%s", method, path, string(bodyData))

	url := fmt.Sprintf("%s%s", c.baseURL, path)
	var resp *http.Response

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Debugf("Making API request: attempt=%d, url=%s", attempt, url)

		var req *http.Request
		if body != nil {
			req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyData))
		} else {
			req, err = http.NewRequestWithContext(ctx, method, url, nil)
		}
		if err != nil {
			logger.Errorf("Failed to create request: %v", err)
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIToken))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			logger.Errorf("Request failed: attempt=%d, error=%v", attempt, err)
			if attempt == maxRetries {
				return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
			}
			time.Sleep(retryDelay)
			continue
		}

		logger.Debugf("Received API response: status_code=%d", resp.StatusCode)

		if !shouldRetry(resp.StatusCode) {
			break
		}

		logger.Warnf("Retrying request due to status code: attempt=%d, status_code=%d", attempt, resp.StatusCode)

		resp.Body.Close()
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return resp, nil
}

// parseResponse parses the response body and handles errors
func (c *Client) parseResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Failed to read response body: %v", err)
		return fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Debugf("Parsing response body: status_code=%d, body=%s", resp.StatusCode, string(body))

	// Check for error response
	if resp.StatusCode >= 400 {
		var apiError APIError
		if err := json.Unmarshal(body, &apiError); err != nil {
			logger.Errorf("Failed to parse error response: status_code=%d, body=%s, error=%v", resp.StatusCode, string(body), err)
			return fmt.Errorf("failed to parse error response (status: %d): %w", resp.StatusCode, err)
		}
		apiError.Status = resp.StatusCode
		return &apiError
	}

	// Parse success response
	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			logger.Errorf("Failed to parse response body: body=%s, error=%v", string(body), err)
			return fmt.Errorf("failed to parse response body: %w", err)
		}
		logger.Debug("Successfully parsed response body")
	}

	return nil
}

// shouldRetry determines if a request should be retried based on the status code
func shouldRetry(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests: // Rate limit
		return true
	case http.StatusInternalServerError: // Server error
		return true
	case http.StatusBadGateway: // Bad gateway
		return true
	case http.StatusServiceUnavailable: // Service unavailable
		return true
	case http.StatusGatewayTimeout: // Gateway timeout
		return true
	default:
		return false
	}
}

// TestConnection tests the connection to the VirtFusion API
func (c *Client) TestConnection(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodGet, "/connect", nil)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	return c.parseResponse(resp, nil)
}

// Hypervisors returns the hypervisor service
func (c *Client) Hypervisors() *HypervisorService {
	return newHypervisorService(c)
}

// Servers returns the server service
func (c *Client) Servers() *ServerService {
	return newServerService(c)
}
