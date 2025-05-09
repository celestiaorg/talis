package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
	// Assuming repos path for mock interface if it were real, but we'll define locally
)

// MockTaskRepository is a mock type for the TaskRepository type
type MockTaskRepository struct {
	mock.Mock
}

// ListByInstanceID mocks the ListByInstanceID method
func (m *MockTaskRepository) ListByInstanceID(ctx context.Context, ownerID uint, instanceID uint, actionFilter models.TaskAction, opts *models.ListOptions) ([]models.Task, error) {
	args := m.Called(ctx, ownerID, instanceID, actionFilter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

// --- Add other TaskRepository methods that might be called by the service under test ---
// For ListTasksByInstanceID, only ListByInstanceID is directly called from the repo.
// For other service methods, you'd add mocks for Create, GetByID, UpdateStatus etc.
func (m *MockTaskRepository) GetByID(ctx context.Context, ownerID uint, id uint) (*models.Task, error) {
	args := m.Called(ctx, ownerID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) Create(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) CreateBatch(ctx context.Context, tasks []*models.Task) error {
	args := m.Called(ctx, tasks)
	return args.Error(0)
}

func (m *MockTaskRepository) ListByProject(ctx context.Context, ownerID uint, projectID uint, opts *models.ListOptions) ([]models.Task, error) {
	args := m.Called(ctx, ownerID, projectID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskRepository) UpdateStatus(ctx context.Context, ownerID uint, id uint, status models.TaskStatus) error {
	args := m.Called(ctx, ownerID, id, status)
	return args.Error(0)
}

func (m *MockTaskRepository) Update(ctx context.Context, ownerID uint, task *models.Task) error {
	args := m.Called(ctx, ownerID, task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetSchedulableTasks(ctx context.Context, limit int) ([]models.Task, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

// MockProjectService is a mock type for the ProjectService type
// TaskService depends on ProjectService for ListByProject, but not for ListTasksByInstanceID.
// So, for this specific test, it might not be strictly needed unless other task service methods are tested.
// However, NewTaskService requires it.

type MockProjectService struct {
	mock.Mock
}

func (m *MockProjectService) GetByName(ctx context.Context, ownerID uint, name string) (*models.Project, error) {
	args := m.Called(ctx, ownerID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

// --- Add other ProjectService methods if needed for other tests ---
func (m *MockProjectService) Create(ctx context.Context, project *models.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectService) List(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Project, error) {
	args := m.Called(ctx, ownerID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectService) Delete(ctx context.Context, ownerID uint, name string) error {
	args := m.Called(ctx, ownerID, name)
	return args.Error(0)
}

func (m *MockProjectService) ListInstances(ctx context.Context, ownerID uint, projectName string, opts *models.ListOptions) ([]models.Instance, error) {
	args := m.Called(ctx, ownerID, projectName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Instance), args.Error(1)
}

func TestTaskService_ListTasksByInstanceID(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockProjectService := new(MockProjectService) // Needed for NewTaskService
	taskService := NewTaskService(mockRepo, mockProjectService)

	ctx := context.Background()
	ownerID := uint(1)
	instanceID := uint(101)
	actionFilter := models.TaskActionCreateInstances
	opts := &models.ListOptions{Limit: 10, Offset: 0}

	expectedTasks := []models.Task{
		{Model: gorm.Model{ID: 1}, InstanceID: instanceID, Action: actionFilter, OwnerID: ownerID, ProjectID: 1},
		{Model: gorm.Model{ID: 2}, InstanceID: instanceID, Action: actionFilter, OwnerID: ownerID, ProjectID: 1},
	}

	// Setup expectation
	mockRepo.On("ListByInstanceID", ctx, ownerID, instanceID, actionFilter, opts).Return(expectedTasks, nil)

	// Call the service method
	tasks, err := taskService.ListTasksByInstanceID(ctx, ownerID, instanceID, actionFilter, opts)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedTasks, tasks)
	mockRepo.AssertExpectations(t) // Verify that the expected mock calls were made
}

func TestTaskService_ListTasksByInstanceID_RepoError(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockProjectService := new(MockProjectService)
	taskService := NewTaskService(mockRepo, mockProjectService)

	ctx := context.Background()
	ownerID := uint(1)
	instanceID := uint(101)
	expectedError := errors.New("repository error")

	// Setup expectation for error
	mockRepo.On("ListByInstanceID", ctx, ownerID, instanceID, models.TaskAction(""), (*models.ListOptions)(nil)).Return(nil, expectedError)

	// Call the service method
	_, err := taskService.ListTasksByInstanceID(ctx, ownerID, instanceID, "", nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}
