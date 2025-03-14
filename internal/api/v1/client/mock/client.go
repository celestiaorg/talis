package mock

import (
	"context"

	"github.com/celestiaorg/talis/internal/api/v1/client"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// MockClient implements the Client interface for testing
type MockClient struct {
	// Function fields that can be set to mock behavior
	GetJobFn              func(ctx context.Context, id string) (*infrastructure.JobStatus, error)
	CreateJobFn           func(ctx context.Context, req infrastructure.CreateRequest) (*infrastructure.Response, error)
	ListJobsFn            func(ctx context.Context, limit int, status string) ([]infrastructure.JobStatus, error)
	GetJobInstancesFn     func(ctx context.Context, jobID string) ([]infrastructure.InstanceInfo, error)
	GetJobPublicIPsFn     func(ctx context.Context, jobID string) ([]string, error)
	GetJobInstanceFn      func(ctx context.Context, jobID string, instanceID string) (*infrastructure.InstanceInfo, error)
	CreateJobInstanceFn   func(ctx context.Context, jobID string, req infrastructure.InstanceRequest) (*infrastructure.InstanceInfo, error)
	DeleteJobInstanceFn   func(ctx context.Context, jobID string, req infrastructure.DeleteInstanceRequest) (*infrastructure.Response, error)
	ListInstancesFn       func(ctx context.Context) ([]infrastructure.InstanceInfo, error)
	GetInstanceMetadataFn func(ctx context.Context) (map[string]interface{}, error)
	GetInstanceFn         func(ctx context.Context, id string) (*infrastructure.InstanceInfo, error)
	HealthCheckFn         func(ctx context.Context) (map[string]string, error)

	// Call tracking for verification
	GetJobCalls []struct {
		Ctx context.Context
		ID  string
	}
	CreateJobCalls []struct {
		Ctx context.Context
		Req infrastructure.CreateRequest
	}
	ListJobsCalls []struct {
		Ctx    context.Context
		Limit  int
		Status string
	}
	GetJobInstancesCalls []struct {
		Ctx   context.Context
		JobID string
	}
	GetJobPublicIPsCalls []struct {
		Ctx   context.Context
		JobID string
	}
	GetJobInstanceCalls []struct {
		Ctx        context.Context
		JobID      string
		InstanceID string
	}
	CreateJobInstanceCalls []struct {
		Ctx   context.Context
		JobID string
		Req   infrastructure.InstanceRequest
	}
	DeleteJobInstanceCalls []struct {
		Ctx   context.Context
		JobID string
		Req   infrastructure.DeleteInstanceRequest
	}
	ListInstancesCalls []struct {
		Ctx context.Context
	}
	GetInstanceMetadataCalls []struct {
		Ctx context.Context
	}
	GetInstanceCalls []struct {
		Ctx context.Context
		ID  string
	}
	HealthCheckCalls []struct {
		Ctx context.Context
	}
}

// Ensure MockClient implements Client interface
var _ client.Client = (*MockClient)(nil)

// GetJob mocks the GetJob method
func (m *MockClient) GetJob(ctx context.Context, id string) (*infrastructure.JobStatus, error) {
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
	return &infrastructure.JobStatus{
		JobID:     id,
		Status:    "completed",
		CreatedAt: "2023-01-01T00:00:00Z",
	}, nil
}

// CreateJob mocks the CreateJob method
func (m *MockClient) CreateJob(ctx context.Context, req infrastructure.CreateRequest) (*infrastructure.Response, error) {
	// Record this call
	m.CreateJobCalls = append(m.CreateJobCalls, struct {
		Ctx context.Context
		Req infrastructure.CreateRequest
	}{
		Ctx: ctx,
		Req: req,
	})

	// Return mock implementation if provided
	if m.CreateJobFn != nil {
		return m.CreateJobFn(ctx, req)
	}

	// Default mock implementation
	return &infrastructure.Response{
		ID:     1,
		Status: "created",
	}, nil
}

// ListJobs mocks the ListJobs method
func (m *MockClient) ListJobs(ctx context.Context, limit int, status string) ([]infrastructure.JobStatus, error) {
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
	return []infrastructure.JobStatus{
		{
			JobID:     "1",
			Status:    "completed",
			CreatedAt: "2023-01-01T00:00:00Z",
		},
		{
			JobID:     "2",
			Status:    "running",
			CreatedAt: "2023-01-02T00:00:00Z",
		},
	}, nil
}

// GetJobInstances mocks the GetJobInstances method
func (m *MockClient) GetJobInstances(ctx context.Context, jobID string) ([]infrastructure.InstanceInfo, error) {
	// Record this call
	m.GetJobInstancesCalls = append(m.GetJobInstancesCalls, struct {
		Ctx   context.Context
		JobID string
	}{
		Ctx:   ctx,
		JobID: jobID,
	})

	// Return mock implementation if provided
	if m.GetJobInstancesFn != nil {
		return m.GetJobInstancesFn(ctx, jobID)
	}

	// Default mock implementation
	return []infrastructure.InstanceInfo{
		{
			Name:     "instance-1",
			IP:       "192.168.1.1",
			Provider: "aws",
			Region:   "us-west-2",
			Size:     "t2.micro",
		},
	}, nil
}

// GetJobPublicIPs mocks the GetJobPublicIPs method
func (m *MockClient) GetJobPublicIPs(ctx context.Context, jobID string) ([]string, error) {
	// Record this call
	m.GetJobPublicIPsCalls = append(m.GetJobPublicIPsCalls, struct {
		Ctx   context.Context
		JobID string
	}{
		Ctx:   ctx,
		JobID: jobID,
	})

	// Return mock implementation if provided
	if m.GetJobPublicIPsFn != nil {
		return m.GetJobPublicIPsFn(ctx, jobID)
	}

	// Default mock implementation
	return []string{"192.168.1.1", "192.168.1.2"}, nil
}

// GetJobInstance mocks the GetJobInstance method
func (m *MockClient) GetJobInstance(ctx context.Context, jobID string, instanceID string) (*infrastructure.InstanceInfo, error) {
	// Record this call
	m.GetJobInstanceCalls = append(m.GetJobInstanceCalls, struct {
		Ctx        context.Context
		JobID      string
		InstanceID string
	}{
		Ctx:        ctx,
		JobID:      jobID,
		InstanceID: instanceID,
	})

	// Return mock implementation if provided
	if m.GetJobInstanceFn != nil {
		return m.GetJobInstanceFn(ctx, jobID, instanceID)
	}

	// Default mock implementation
	return &infrastructure.InstanceInfo{
		Name:     "instance-1",
		IP:       "192.168.1.1",
		Provider: "aws",
		Region:   "us-west-2",
		Size:     "t2.micro",
	}, nil
}

// CreateJobInstance mocks the CreateJobInstance method
func (m *MockClient) CreateJobInstance(ctx context.Context, jobID string, req infrastructure.InstanceRequest) (*infrastructure.InstanceInfo, error) {
	// Record this call
	m.CreateJobInstanceCalls = append(m.CreateJobInstanceCalls, struct {
		Ctx   context.Context
		JobID string
		Req   infrastructure.InstanceRequest
	}{
		Ctx:   ctx,
		JobID: jobID,
		Req:   req,
	})

	// Return mock implementation if provided
	if m.CreateJobInstanceFn != nil {
		return m.CreateJobInstanceFn(ctx, jobID, req)
	}

	// Default mock implementation
	return &infrastructure.InstanceInfo{
		Name:     "instance-1",
		IP:       "192.168.1.1",
		Provider: req.Provider,
		Region:   req.Region,
		Size:     req.Size,
	}, nil
}

// DeleteJobInstance mocks the DeleteJobInstance method
func (m *MockClient) DeleteJobInstance(ctx context.Context, jobID string, req infrastructure.DeleteInstanceRequest) (*infrastructure.Response, error) {
	// Record this call
	m.DeleteJobInstanceCalls = append(m.DeleteJobInstanceCalls, struct {
		Ctx   context.Context
		JobID string
		Req   infrastructure.DeleteInstanceRequest
	}{
		Ctx:   ctx,
		JobID: jobID,
		Req:   req,
	})

	// Return mock implementation if provided
	if m.DeleteJobInstanceFn != nil {
		return m.DeleteJobInstanceFn(ctx, jobID, req)
	}

	// Default mock implementation
	return &infrastructure.Response{
		ID:     1,
		Status: "deleted",
	}, nil
}

// ListInstances mocks the ListInstances method
func (m *MockClient) ListInstances(ctx context.Context) ([]infrastructure.InstanceInfo, error) {
	// Record this call
	m.ListInstancesCalls = append(m.ListInstancesCalls, struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	})

	// Return mock implementation if provided
	if m.ListInstancesFn != nil {
		return m.ListInstancesFn(ctx)
	}

	// Default mock implementation
	return []infrastructure.InstanceInfo{
		{
			Name:     "instance-1",
			IP:       "192.168.1.1",
			Provider: "aws",
			Region:   "us-west-2",
			Size:     "t2.micro",
		},
		{
			Name:     "instance-2",
			IP:       "192.168.1.2",
			Provider: "aws",
			Region:   "us-west-2",
			Size:     "t2.micro",
		},
	}, nil
}

// GetInstanceMetadata mocks the GetInstanceMetadata method
func (m *MockClient) GetInstanceMetadata(ctx context.Context) (map[string]interface{}, error) {
	// Record this call
	m.GetInstanceMetadataCalls = append(m.GetInstanceMetadataCalls, struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	})

	// Return mock implementation if provided
	if m.GetInstanceMetadataFn != nil {
		return m.GetInstanceMetadataFn(ctx)
	}

	// Default mock implementation
	return map[string]interface{}{
		"instance-1": map[string]interface{}{
			"ip":       "192.168.1.1",
			"provider": "aws",
			"region":   "us-west-2",
		},
		"instance-2": map[string]interface{}{
			"ip":       "192.168.1.2",
			"provider": "aws",
			"region":   "us-west-2",
		},
	}, nil
}

// GetInstance mocks the GetInstance method
func (m *MockClient) GetInstance(ctx context.Context, id string) (*infrastructure.InstanceInfo, error) {
	// Record this call
	m.GetInstanceCalls = append(m.GetInstanceCalls, struct {
		Ctx context.Context
		ID  string
	}{
		Ctx: ctx,
		ID:  id,
	})

	// Return mock implementation if provided
	if m.GetInstanceFn != nil {
		return m.GetInstanceFn(ctx, id)
	}

	// Default mock implementation
	return &infrastructure.InstanceInfo{
		Name:     id,
		IP:       "192.168.1.1",
		Provider: "aws",
		Region:   "us-west-2",
		Size:     "t2.micro",
	}, nil
}

// HealthCheck mocks the HealthCheck method
func (m *MockClient) HealthCheck(ctx context.Context) (map[string]string, error) {
	// Record this call
	m.HealthCheckCalls = append(m.HealthCheckCalls, struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	})

	// Return mock implementation if provided
	if m.HealthCheckFn != nil {
		return m.HealthCheckFn(ctx)
	}

	// Default mock implementation
	return map[string]string{
		"status": "healthy",
	}, nil
}
