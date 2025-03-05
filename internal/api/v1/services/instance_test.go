package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockJobService is a mock implementation of JobServiceInterface
type MockJobService struct {
	mock.Mock
}

func (m *MockJobService) CreateJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	args := m.Called(ctx, job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobService) UpdateJobStatus(ctx context.Context, id uint, status models.JobStatus, result interface{}, errMsg string) error {
	args := m.Called(ctx, id, status, result, errMsg)
	return args.Error(0)
}

// MockInstanceRepository is a mock implementation of InstanceRepositoryInterface
type MockInstanceRepository struct {
	mock.Mock
}

func (m *MockInstanceRepository) List(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Instance), args.Error(1)
}

func (m *MockInstanceRepository) Get(ctx context.Context, id uint) (*models.Instance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Instance), args.Error(1)
}

func TestCreateInstance(t *testing.T) {
	tests := []struct {
		name       string
		jobService *MockJobService
		repo       *MockInstanceRepository
		request    struct {
			name        string
			projectName string
			webhookURL  string
			instances   []infrastructure.InstanceRequest
		}
		expectedJob   *models.Job
		expectedError error
	}{
		{
			name: "successful instance creation",
			jobService: func() *MockJobService {
				m := new(MockJobService)
				m.On("CreateJob", mock.Anything, mock.Anything).Return(&models.Job{
					Model:       gorm.Model{ID: 1, CreatedAt: time.Now()},
					Name:        "test-instance",
					ProjectName: "test-project",
					Status:      models.JobStatusPending,
				}, nil)
				// We don't need to expect UpdateJobStatus calls since they happen in a goroutine
				return m
			}(),
			repo: new(MockInstanceRepository),
			request: struct {
				name        string
				projectName string
				webhookURL  string
				instances   []infrastructure.InstanceRequest
			}{
				name:        "test-instance",
				projectName: "test-project",
				webhookURL:  "http://test.com",
				instances: []infrastructure.InstanceRequest{
					{
						Provider:          "digitalocean",
						NumberOfInstances: 1,
						Region:            "nyc3",
						Size:              "s-1vcpu-1gb",
						Image:             "ubuntu-20-04-x64",
						SSHKeyName:        "test-key",
					},
				},
			},
			expectedJob: &models.Job{
				Model:       gorm.Model{ID: 1, CreatedAt: time.Now()},
				Name:        "test-instance",
				ProjectName: "test-project",
				Status:      models.JobStatusPending,
			},
			expectedError: nil,
		},
		{
			name: "job creation fails",
			jobService: func() *MockJobService {
				m := new(MockJobService)
				m.On("CreateJob", mock.Anything, mock.Anything).Return(nil, errors.New("job creation failed"))
				return m
			}(),
			repo: new(MockInstanceRepository),
			request: struct {
				name        string
				projectName string
				webhookURL  string
				instances   []infrastructure.InstanceRequest
			}{
				name:        "test-instance",
				projectName: "test-project",
				webhookURL:  "http://test.com",
				instances: []infrastructure.InstanceRequest{
					{
						Provider:          "digitalocean",
						NumberOfInstances: 1,
						Region:            "nyc3",
						Size:              "s-1vcpu-1gb",
						Image:             "ubuntu-20-04-x64",
						SSHKeyName:        "test-key",
					},
				},
			},
			expectedJob:   nil,
			expectedError: errors.New("failed to create job: job creation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInstanceService(tt.repo, tt.jobService)

			job, err := service.CreateInstance(context.Background(), tt.request.name, tt.request.projectName, tt.request.webhookURL, tt.request.instances)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedJob.ID, job.ID)
			assert.Equal(t, tt.expectedJob.Name, job.Name)
			assert.Equal(t, tt.expectedJob.ProjectName, job.ProjectName)
			assert.Equal(t, tt.expectedJob.Status, job.Status)

			tt.jobService.AssertExpectations(t)
		})
	}
}

func TestDeleteInstance(t *testing.T) {
	tests := []struct {
		name          string
		jobService    *MockJobService
		repo          *MockInstanceRepository
		instanceID    string
		expectedJob   *models.Job
		expectedError error
	}{
		{
			name: "successful instance deletion",
			jobService: func() *MockJobService {
				m := new(MockJobService)
				m.On("CreateJob", mock.Anything, mock.Anything).Return(&models.Job{
					Model:       gorm.Model{ID: 1, CreatedAt: time.Now()},
					Name:        "delete-test-instance",
					ProjectName: "talis",
					Status:      models.JobStatusPending,
				}, nil)
				// We don't need to expect UpdateJobStatus calls since they happen in a goroutine
				return m
			}(),
			repo:       new(MockInstanceRepository),
			instanceID: "test-instance",
			expectedJob: &models.Job{
				Model:       gorm.Model{ID: 1, CreatedAt: time.Now()},
				Name:        "delete-test-instance",
				ProjectName: "talis",
				Status:      models.JobStatusPending,
			},
			expectedError: nil,
		},
		{
			name: "job creation fails",
			jobService: func() *MockJobService {
				m := new(MockJobService)
				m.On("CreateJob", mock.Anything, mock.Anything).Return(nil, errors.New("job creation failed"))
				return m
			}(),
			repo:          new(MockInstanceRepository),
			instanceID:    "test-instance",
			expectedJob:   nil,
			expectedError: errors.New("failed to create job: job creation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInstanceService(tt.repo, tt.jobService)

			job, err := service.DeleteInstance(context.Background(), tt.instanceID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedJob.ID, job.ID)
			assert.Equal(t, tt.expectedJob.Name, job.Name)
			assert.Equal(t, tt.expectedJob.ProjectName, job.ProjectName)
			assert.Equal(t, tt.expectedJob.Status, job.Status)

			tt.jobService.AssertExpectations(t)
		})
	}
}

func TestGetInstance(t *testing.T) {
	tests := []struct {
		name             string
		repo             *MockInstanceRepository
		instanceID       uint
		expectedInstance *models.Instance
		expectedError    error
	}{
		{
			name: "successful instance retrieval",
			repo: func() *MockInstanceRepository {
				m := new(MockInstanceRepository)
				m.On("Get", mock.Anything, uint(1)).Return(&models.Instance{
					Model: gorm.Model{ID: 1, CreatedAt: time.Now()},
					Name:  "test-instance",
				}, nil)
				return m
			}(),
			instanceID: 1,
			expectedInstance: &models.Instance{
				Model: gorm.Model{ID: 1, CreatedAt: time.Now()},
				Name:  "test-instance",
			},
			expectedError: nil,
		},
		{
			name: "instance not found",
			repo: func() *MockInstanceRepository {
				m := new(MockInstanceRepository)
				m.On("Get", mock.Anything, uint(1)).Return(nil, errors.New("instance not found"))
				return m
			}(),
			instanceID:       1,
			expectedInstance: nil,
			expectedError:    errors.New("instance not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInstanceService(tt.repo, nil)

			instance, err := service.GetInstance(context.Background(), tt.instanceID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedInstance.ID, instance.ID)
			assert.Equal(t, tt.expectedInstance.Name, instance.Name)

			tt.repo.AssertExpectations(t)
		})
	}
}

func TestListInstances(t *testing.T) {
	tests := []struct {
		name              string
		repo              *MockInstanceRepository
		opts              *models.ListOptions
		expectedInstances []models.Instance
		expectedError     error
	}{
		{
			name: "successful instance listing",
			repo: func() *MockInstanceRepository {
				m := new(MockInstanceRepository)
				m.On("List", mock.Anything, &models.ListOptions{Limit: 10, Offset: 0}).Return([]models.Instance{
					{
						Model: gorm.Model{ID: 1, CreatedAt: time.Now()},
						Name:  "test-instance-1",
					},
					{
						Model: gorm.Model{ID: 2, CreatedAt: time.Now()},
						Name:  "test-instance-2",
					},
				}, nil)
				return m
			}(),
			opts: &models.ListOptions{Limit: 10, Offset: 0},
			expectedInstances: []models.Instance{
				{
					Model: gorm.Model{ID: 1, CreatedAt: time.Now()},
					Name:  "test-instance-1",
				},
				{
					Model: gorm.Model{ID: 2, CreatedAt: time.Now()},
					Name:  "test-instance-2",
				},
			},
			expectedError: nil,
		},
		{
			name: "list fails",
			repo: func() *MockInstanceRepository {
				m := new(MockInstanceRepository)
				m.On("List", mock.Anything, &models.ListOptions{Limit: 10, Offset: 0}).Return(nil, errors.New("list failed"))
				return m
			}(),
			opts:              &models.ListOptions{Limit: 10, Offset: 0},
			expectedInstances: nil,
			expectedError:     errors.New("list failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInstanceService(tt.repo, nil)

			instances, err := service.ListInstances(context.Background(), tt.opts)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expectedInstances), len(instances))
			for i, instance := range instances {
				assert.Equal(t, tt.expectedInstances[i].ID, instance.ID)
				assert.Equal(t, tt.expectedInstances[i].Name, instance.Name)
			}

			tt.repo.AssertExpectations(t)
		})
	}
}
