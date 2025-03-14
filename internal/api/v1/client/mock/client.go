package mock

import (
	"context"

	"github.com/celestiaorg/talis/internal/api/v1/client"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// MockClient implements the Client interface for testing
type MockClient struct {
	// Function fields that can be set to mock behavior
	CreateInfrastructureFn func(ctx context.Context, req interface{}) (interface{}, error)
	DeleteInfrastructureFn func(ctx context.Context, req interface{}) (interface{}, error)
	GetInfrastructureFn    func(ctx context.Context, id string) (interface{}, error)
	ListInfrastructureFn   func(ctx context.Context) (interface{}, error)
	GetJobFn               func(ctx context.Context, id string) (interface{}, error)
	ListJobsFn             func(ctx context.Context, limit int, status string) (interface{}, error)

	// Call tracking for verification
	CreateInfrastructureCalls []struct {
		Ctx context.Context
		Req interface{}
	}
	DeleteInfrastructureCalls []struct {
		Ctx context.Context
		Req interface{}
	}
	GetInfrastructureCalls []struct {
		Ctx context.Context
		ID  string
	}
	ListInfrastructureCalls []struct {
		Ctx context.Context
	}
	GetJobCalls []struct {
		Ctx context.Context
		ID  string
	}
	ListJobsCalls []struct {
		Ctx    context.Context
		Limit  int
		Status string
	}
}

// Ensure MockClient implements Client interface
var _ client.Client = (*MockClient)(nil)

// CreateInfrastructure mocks the CreateInfrastructure method
func (m *MockClient) CreateInfrastructure(ctx context.Context, req interface{}) (interface{}, error) {
	// Record this call
	m.CreateInfrastructureCalls = append(m.CreateInfrastructureCalls, struct {
		Ctx context.Context
		Req interface{}
	}{
		Ctx: ctx,
		Req: req,
	})

	// Return mock implementation if provided
	if m.CreateInfrastructureFn != nil {
		return m.CreateInfrastructureFn(ctx, req)
	}

	// Default mock implementation
	return &infrastructure.Response{
		ID:     1,
		Status: "created",
	}, nil
}

// DeleteInfrastructure mocks the DeleteInfrastructure method
func (m *MockClient) DeleteInfrastructure(ctx context.Context, req interface{}) (interface{}, error) {
	// Record this call
	m.DeleteInfrastructureCalls = append(m.DeleteInfrastructureCalls, struct {
		Ctx context.Context
		Req interface{}
	}{
		Ctx: ctx,
		Req: req,
	})

	// Return mock implementation if provided
	if m.DeleteInfrastructureFn != nil {
		return m.DeleteInfrastructureFn(ctx, req)
	}

	// Default mock implementation
	return &infrastructure.Response{
		ID:     1,
		Status: "deleted",
	}, nil
}

// GetInfrastructure mocks the GetInfrastructure method
func (m *MockClient) GetInfrastructure(ctx context.Context, id string) (interface{}, error) {
	// Record this call
	m.GetInfrastructureCalls = append(m.GetInfrastructureCalls, struct {
		Ctx context.Context
		ID  string
	}{
		Ctx: ctx,
		ID:  id,
	})

	// Return mock implementation if provided
	if m.GetInfrastructureFn != nil {
		return m.GetInfrastructureFn(ctx, id)
	}

	// Default mock implementation
	return &infrastructure.Response{
		ID:     1,
		Status: "active",
	}, nil
}

// ListInfrastructure mocks the ListInfrastructure method
func (m *MockClient) ListInfrastructure(ctx context.Context) (interface{}, error) {
	// Record this call
	m.ListInfrastructureCalls = append(m.ListInfrastructureCalls, struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	})

	// Return mock implementation if provided
	if m.ListInfrastructureFn != nil {
		return m.ListInfrastructureFn(ctx)
	}

	// Default mock implementation
	return []infrastructure.Response{
		{
			ID:     1,
			Status: "active",
		},
		{
			ID:     2,
			Status: "active",
		},
	}, nil
}

// GetJob mocks the GetJob method
func (m *MockClient) GetJob(ctx context.Context, id string) (interface{}, error) {
	// Record this call
	m.GetJobCalls = append(m.GetJobCalls, struct {
		Ctx context.Context
		ID  string
	}{
		Ctx: ctx,
		ID:  id,
	})

	// Return mock implementation if provided
	if m.GetJobFn != nil {
		return m.GetJobFn(ctx, id)
	}

	// Default mock implementation
	return map[string]interface{}{
		"job_id":     id,
		"status":     "completed",
		"created_at": "2023-01-01T00:00:00Z",
	}, nil
}

// ListJobs mocks the ListJobs method
func (m *MockClient) ListJobs(ctx context.Context, limit int, status string) (interface{}, error) {
	// Record this call
	m.ListJobsCalls = append(m.ListJobsCalls, struct {
		Ctx    context.Context
		Limit  int
		Status string
	}{
		Ctx:    ctx,
		Limit:  limit,
		Status: status,
	})

	// Return mock implementation if provided
	if m.ListJobsFn != nil {
		return m.ListJobsFn(ctx, limit, status)
	}

	// Default mock implementation
	return []map[string]interface{}{
		{
			"job_id":     "1",
			"status":     "completed",
			"created_at": "2023-01-01T00:00:00Z",
		},
		{
			"job_id":     "2",
			"status":     "running",
			"created_at": "2023-01-02T00:00:00Z",
		},
	}, nil
}
