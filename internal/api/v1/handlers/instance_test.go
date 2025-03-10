package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockInstanceService is a mock implementation of InstanceServiceInterface
type MockInstanceService struct {
	mock.Mock
}

func (m *MockInstanceService) GetInstance(ctx context.Context, id uint) (*models.Instance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Instance), args.Error(1)
}

func (m *MockInstanceService) GetPublicIPs(ctx context.Context) ([]models.Instance, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Instance), args.Error(1)
}

func (m *MockInstanceService) GetInstancesByJobID(ctx context.Context, jobID uint) ([]models.Instance, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Instance), args.Error(1)
}

func (m *MockInstanceService) ListInstances(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Instance), args.Error(1)
}

func (m *MockInstanceService) CreateInstance(ctx context.Context, name, projectName, webhookURL string, instances []infrastructure.InstanceRequest) (*models.Job, error) {
	args := m.Called(ctx, name, projectName, webhookURL, instances)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockInstanceService) DeleteInstance(ctx context.Context, jobID uint, name, projectName string, instances []infrastructure.InstanceRequest) (*models.Job, error) {
	args := m.Called(ctx, jobID, name, projectName, instances)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

// MockJobService is a mock implementation of the JobServiceInterface
type MockJobService struct {
	mock.Mock
}

func (m *MockJobService) CreateJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	args := m.Called(ctx, job)
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobService) UpdateJobStatus(ctx context.Context, id uint, status models.JobStatus, result interface{}, errMsg string) error {
	args := m.Called(ctx, id, status, result, errMsg)
	return args.Error(0)
}

func (m *MockJobService) GetByProjectName(ctx context.Context, projectName string) (*models.Job, error) {
	args := m.Called(ctx, projectName)
	return args.Get(0).(*models.Job), args.Error(1)
}

type mockInstanceService struct {
	instances []models.Instance
}

func (m *mockInstanceService) GetInstance(ctx context.Context, id uint) (*models.Instance, error) {
	for _, instance := range m.instances {
		if instance.ID == id {
			return &instance, nil
		}
	}
	return nil, fmt.Errorf("instance not found")
}

func (m *mockInstanceService) GetPublicIPs(ctx context.Context) ([]models.Instance, error) {
	return m.instances, nil
}

func (m *mockInstanceService) GetInstancesByJobID(ctx context.Context, jobID uint) ([]models.Instance, error) {
	var jobInstances []models.Instance
	for _, instance := range m.instances {
		if instance.JobID == jobID {
			jobInstances = append(jobInstances, instance)
		}
	}
	return jobInstances, nil
}

func (m *mockInstanceService) ListInstances(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	return m.instances, nil
}

func (m *mockInstanceService) CreateInstance(ctx context.Context, name, projectName, webhookURL string, instances []infrastructure.InstanceRequest) (*models.Job, error) {
	return nil, nil
}

func (m *mockInstanceService) DeleteInstance(ctx context.Context, jobID uint, name, projectName string, instances []infrastructure.InstanceRequest) (*models.Job, error) {
	return nil, nil
}

func TestGetPublicIPs(t *testing.T) {
	// Create mock services
	mockInstanceService := new(MockInstanceService)
	mockJobService := new(MockJobService)

	// Create handler with mock services
	handler := NewInstanceHandler(mockInstanceService, mockJobService)

	// Create test app
	app := fiber.New()
	app.Get("/public-ips", handler.GetPublicIPs)

	// Create test instances
	testInstances := []models.Instance{
		{
			Model: gorm.Model{
				ID:        1,
				CreatedAt: time.Now(),
			},
			JobID:      1,
			ProviderID: "digitalocean",
			Name:       "test-instance-1",
			PublicIP:   "192.168.1.1",
			Region:     "nyc1",
			Size:       "s-1vcpu-1gb",
			Image:      "ubuntu-20-04-x64",
			Tags:       []string{"test"},
			Status:     models.InstanceStatusReady,
		},
		{
			Model: gorm.Model{
				ID:        2,
				CreatedAt: time.Now(),
			},
			JobID:      2,
			ProviderID: "digitalocean",
			Name:       "test-instance-2",
			PublicIP:   "192.168.1.2",
			Region:     "nyc1",
			Size:       "s-1vcpu-1gb",
			Image:      "ubuntu-20-04-x64",
			Tags:       []string{"test"},
			Status:     models.InstanceStatusReady,
		},
	}

	// Set up expectations
	mockInstanceService.On("GetPublicIPs", mock.Anything).Return(testInstances, nil)

	// Create test request
	req := httptest.NewRequest("GET", "/public-ips", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err)

	// Verify response
	assert.Equal(t, float64(2), result["total"])
	instances, ok := result["instances"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(instances))

	// Verify first instance
	instance1 := instances[0].(map[string]interface{})
	assert.Equal(t, "192.168.1.1", instance1["public_ip"])
	assert.Equal(t, float64(1), instance1["job_id"])

	// Verify second instance
	instance2 := instances[1].(map[string]interface{})
	assert.Equal(t, "192.168.1.2", instance2["public_ip"])
	assert.Equal(t, float64(2), instance2["job_id"])

	// Verify mock expectations
	mockInstanceService.AssertExpectations(t)
}

func TestGetPublicIPsError(t *testing.T) {
	// Create mock services
	mockInstanceService := new(MockInstanceService)
	mockJobService := new(MockJobService)

	// Create handler with mock services
	handler := NewInstanceHandler(mockInstanceService, mockJobService)

	// Create test app
	app := fiber.New()
	app.Get("/public-ips", handler.GetPublicIPs)

	// Set up expectations for error case
	mockInstanceService.On("GetPublicIPs", mock.Anything).Return([]models.Instance{}, errors.New("database error"))

	// Create test request
	req := httptest.NewRequest("GET", "/public-ips", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err)

	// Verify error message
	assert.Contains(t, result["error"], "failed to get public IPs")

	// Verify mock expectations
	mockInstanceService.AssertExpectations(t)
}

func TestGetInstancesByJobID(t *testing.T) {
	// Create a mock service
	mockService := new(MockInstanceService)
	mockJobService := new(MockJobService)

	// Create test instance
	testInstance := models.Instance{
		Model: gorm.Model{
			ID: 1,
		},
		JobID:      123,
		ProviderID: models.ProviderID("digitalocean"),
		Name:       "test-instance",
		PublicIP:   "1.2.3.4",
		Region:     "nyc1",
		Size:       "s-1vcpu-1gb",
		Image:      "ubuntu-20-04-x64",
		Tags:       []string{"test"},
		Status:     models.InstanceStatusReady,
		CreatedAt:  time.Now(),
	}

	// Set up expectations
	mockService.On("GetInstancesByJobID", mock.Anything, uint(123)).Return([]models.Instance{testInstance}, nil)

	// Create the handler with the mock service
	handler := NewInstanceHandler(mockService, mockJobService)

	// Create a test request
	app := fiber.New()
	app.Get("/job/:jobId", handler.GetInstancesByJobID)

	// Test successful case
	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/job/123", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		instances, ok := result["instances"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 1, len(instances))

		instance := instances[0].(map[string]interface{})
		assert.Equal(t, float64(1), instance["id"])
		assert.Equal(t, "1.2.3.4", instance["public_ip"])
		assert.Equal(t, float64(123), instance["job_id"])
		assert.Equal(t, "test-instance", instance["name"])
		assert.Equal(t, "nyc1", instance["region"])
		assert.Equal(t, "s-1vcpu-1gb", instance["size"])
		assert.Equal(t, "ubuntu-20-04-x64", instance["image"])
		assert.Equal(t, []interface{}{"test"}, instance["tags"])
		assert.Equal(t, "ready", instance["status"])
	})

	// Test invalid job ID
	t.Run("invalid job id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/job/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Contains(t, result["error"], "invalid job id")
	})

	// Test missing job ID
	t.Run("missing job id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/job", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})

	// Verify expectations
	mockService.AssertExpectations(t)
}

func TestGetAllMetadata(t *testing.T) {
	// Create mock services
	mockInstanceService := new(MockInstanceService)
	mockJobService := new(MockJobService)

	// Create handler with mock services
	handler := NewInstanceHandler(mockInstanceService, mockJobService)

	// Create test app
	app := fiber.New()
	app.Get("/all-metadata", handler.GetAllMetadata)

	// Create test instances
	testInstances := []models.Instance{
		{
			Model: gorm.Model{
				ID:        1,
				CreatedAt: time.Now(),
			},
			JobID:      1,
			ProviderID: "digitalocean",
			Name:       "test-instance-1",
			PublicIP:   "192.168.1.1",
			Region:     "nyc1",
			Size:       "s-1vcpu-1gb",
			Image:      "ubuntu-20-04-x64",
			Tags:       []string{"test"},
			Status:     models.InstanceStatusReady,
		},
		{
			Model: gorm.Model{
				ID:        2,
				CreatedAt: time.Now(),
			},
			JobID:      2,
			ProviderID: "digitalocean",
			Name:       "test-instance-2",
			PublicIP:   "192.168.1.2",
			Region:     "nyc1",
			Size:       "s-1vcpu-1gb",
			Image:      "ubuntu-20-04-x64",
			Tags:       []string{"test"},
			Status:     models.InstanceStatusReady,
		},
	}

	// Set up expectations
	mockInstanceService.On("GetPublicIPs", mock.Anything).Return(testInstances, nil)

	// Create test request
	req := httptest.NewRequest("GET", "/all-metadata", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err)

	// Verify response
	assert.Equal(t, float64(2), result["total"])
	instances, ok := result["instances"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(instances))

	// Verify first instance
	instance1 := instances[0].(map[string]interface{})
	assert.Equal(t, float64(1), instance1["id"])
	assert.Equal(t, float64(1), instance1["job_id"])
	assert.Equal(t, "test-instance-1", instance1["name"])
	assert.Equal(t, "192.168.1.1", instance1["public_ip"])
	assert.Equal(t, "nyc1", instance1["region"])
	assert.Equal(t, "s-1vcpu-1gb", instance1["size"])
	assert.Equal(t, "ubuntu-20-04-x64", instance1["image"])
	assert.Equal(t, []interface{}{"test"}, instance1["tags"])
	assert.Equal(t, "ready", instance1["status"])

	// Verify second instance
	instance2 := instances[1].(map[string]interface{})
	assert.Equal(t, float64(2), instance2["id"])
	assert.Equal(t, float64(2), instance2["job_id"])
	assert.Equal(t, "test-instance-2", instance2["name"])
	assert.Equal(t, "192.168.1.2", instance2["public_ip"])
	assert.Equal(t, "nyc1", instance2["region"])
	assert.Equal(t, "s-1vcpu-1gb", instance2["size"])
	assert.Equal(t, "ubuntu-20-04-x64", instance2["image"])
	assert.Equal(t, []interface{}{"test"}, instance2["tags"])
	assert.Equal(t, "ready", instance2["status"])

	// Verify mock expectations
	mockInstanceService.AssertExpectations(t)
}

func TestGetAllMetadataError(t *testing.T) {
	// Create mock services
	mockInstanceService := new(MockInstanceService)
	mockJobService := new(MockJobService)

	// Create handler with mock services
	handler := NewInstanceHandler(mockInstanceService, mockJobService)

	// Create test app
	app := fiber.New()
	app.Get("/all-metadata", handler.GetAllMetadata)

	// Set up expectations for error case
	mockInstanceService.On("GetPublicIPs", mock.Anything).Return([]models.Instance{}, errors.New("database error"))

	// Create test request
	req := httptest.NewRequest("GET", "/all-metadata", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err)

	// Verify error message
	assert.Contains(t, result["error"], "failed to get instance metadata")

	// Verify mock expectations
	mockInstanceService.AssertExpectations(t)
}
