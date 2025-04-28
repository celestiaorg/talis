package compute

import (
	"context"
	"fmt"
	"net/http"

	"github.com/celestiaorg/talis/internal/types"
)

// ServerService handles server-related operations
type ServerService struct {
	client *Client
}

// newServerService creates a new server service
func newServerService(client *Client) *ServerService {
	return &ServerService{
		client: client,
	}
}

// Create creates a new server
func (s *ServerService) Create(ctx context.Context, req *types.ServerCreateRequest) (*types.Server, *types.APIResponse, error) {
	resp, err := s.client.doRequest(ctx, http.MethodPost, "/servers", req)
	if err != nil {
		return nil, nil, err
	}

	var server types.Server
	err = s.client.parseResponse(resp, &server)
	if err != nil {
		return nil, &types.APIResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
		}, err
	}

	return &server, &types.APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}, nil
}

// Get returns details of a specific server
func (s *ServerService) Get(ctx context.Context, id int) (*types.Server, *types.APIResponse, error) {
	path := fmt.Sprintf("/servers/%d", id)
	resp, err := s.client.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get server: %w", err)
	}

	var server types.Server
	if err := s.client.parseResponse(resp, &server); err != nil {
		return nil, nil, fmt.Errorf("failed to parse server response: %w", err)
	}

	apiResp := &types.APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}

	return &server, apiResp, nil
}

// Delete deletes a server
func (s *ServerService) Delete(ctx context.Context, id int) (*types.APIResponse, error) {
	path := fmt.Sprintf("/servers/%d", id)
	resp, err := s.client.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to delete server: %w", err)
	}

	apiResp := &types.APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}

	return apiResp, nil
}

// List returns a list of all servers
func (s *ServerService) List(ctx context.Context) ([]types.Server, *types.APIResponse, error) {
	resp, err := s.client.doRequest(ctx, http.MethodGet, "/servers", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list servers: %w", err)
	}

	var result struct {
		Servers []types.Server `json:"servers"`
	}
	if err := s.client.parseResponse(resp, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to parse servers response: %w", err)
	}

	apiResp := &types.APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}

	return result.Servers, apiResp, nil
}

// Build builds a server after creation
func (s *ServerService) Build(ctx context.Context, id int) (*types.APIResponse, error) {
	path := fmt.Sprintf("/servers/%d/build", id)
	resp, err := s.client.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build server: %w", err)
	}

	apiResp := &types.APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}

	return apiResp, nil
}

// Suspend suspends a server
func (s *ServerService) Suspend(ctx context.Context, id int) (*types.APIResponse, error) {
	path := fmt.Sprintf("/servers/%d/suspend", id)
	resp, err := s.client.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to suspend server: %w", err)
	}

	apiResp := &types.APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}

	return apiResp, nil
}

// Unsuspend unsuspends a server
func (s *ServerService) Unsuspend(ctx context.Context, id int) (*types.APIResponse, error) {
	path := fmt.Sprintf("/servers/%d/unsuspend", id)
	resp, err := s.client.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to unsuspend server: %w", err)
	}

	apiResp := &types.APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}

	return apiResp, nil
}

// GetStatus returns the current status of a server
func (s *ServerService) GetStatus(ctx context.Context, id int) (string, *types.APIResponse, error) {
	server, apiResp, err := s.Get(ctx, id)
	if err != nil {
		return "", apiResp, fmt.Errorf("failed to get server status: %w", err)
	}
	return server.Status, apiResp, nil
}
