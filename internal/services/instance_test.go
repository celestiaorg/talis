package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
)

var globalInstanceTestMockIDCounterForInstanceTests uint

// --- Mock InstanceRepositoryInterface specifically for instance_test.go ---
type InstanceTestMockInstanceRepository struct {
	mock.Mock
}

func (m *InstanceTestMockInstanceRepository) getID() uint {
	globalInstanceTestMockIDCounterForInstanceTests++
	return globalInstanceTestMockIDCounterForInstanceTests
}

func (m *InstanceTestMockInstanceRepository) Create(ctx context.Context, instance *models.Instance) (*models.Instance, error) {
	args := m.Called(ctx, instance)
	retInstance := args.Get(0).(*models.Instance)
	err := args.Error(1)
	if err == nil && retInstance != nil {
		if retInstance.ID == 0 {
			retInstance.ID = m.getID()
		}
		*instance = *retInstance
	}
	return retInstance, err
}

func (m *InstanceTestMockInstanceRepository) CreateBatch(ctx context.Context, instances []*models.Instance) ([]*models.Instance, error) {
	args := m.Called(ctx, instances)
	createdInstances := args.Get(0).([]*models.Instance)
	err := args.Error(1)
	if err == nil && createdInstances != nil {
		for _, inst := range createdInstances {
			if inst.ID == 0 {
				inst.ID = m.getID()
			}
		}
	}
	return createdInstances, err
}

func (m *InstanceTestMockInstanceRepository) Get(ctx context.Context, ownerID uint, id uint) (*models.Instance, error) {
	args := m.Called(ctx, ownerID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Instance), args.Error(1)
}

func (m *InstanceTestMockInstanceRepository) List(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Instance, error) {
	args := m.Called(ctx, ownerID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Instance), args.Error(1)
}

func (m *InstanceTestMockInstanceRepository) GetByProjectIDAndInstanceIDs(ctx context.Context, ownerID uint, projectID uint, instanceIDs []uint) ([]models.Instance, error) {
	args := m.Called(ctx, ownerID, projectID, instanceIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Instance), args.Error(1)
}

func (m *InstanceTestMockInstanceRepository) Update(ctx context.Context, ownerID uint, instanceID uint, instance *models.Instance) error {
	return m.Called(ctx, ownerID, instanceID, instance).Error(0)
}

func (m *InstanceTestMockInstanceRepository) Terminate(ctx context.Context, ownerID, instanceID uint) error {
	return m.Called(ctx, ownerID, instanceID).Error(0)
}

// --- Mock TaskServiceInterface specifically for instance_test.go ---
type InstanceTestMockTaskService struct {
	mock.Mock
}

func (m *InstanceTestMockTaskService) getID() uint {
	globalInstanceTestMockIDCounterForInstanceTests++
	return globalInstanceTestMockIDCounterForInstanceTests
}

func (m *InstanceTestMockTaskService) Create(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	if task != nil && task.ID == 0 {
		task.ID = m.getID()
	}
	return args.Error(0)
}

func (m *InstanceTestMockTaskService) CreateBatch(ctx context.Context, tasks []*models.Task) error {
	args := m.Called(ctx, tasks)
	for _, task := range tasks {
		if task.ID == 0 {
			task.ID = m.getID()
		}
	}
	return args.Error(0)
}
func (m *InstanceTestMockTaskService) GetByID(ctx context.Context, ownerID uint, taskID uint) (*models.Task, error) {
	args := m.Called(ctx, ownerID, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}
func (m *InstanceTestMockTaskService) ListByProject(ctx context.Context, ownerID uint, projectName string, opts *models.ListOptions) ([]models.Task, error) {
	args := m.Called(ctx, ownerID, projectName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}
func (m *InstanceTestMockTaskService) UpdateStatus(ctx context.Context, ownerID uint, taskID uint, status models.TaskStatus) error {
	return m.Called(ctx, ownerID, taskID, status).Error(0)
}
func (m *InstanceTestMockTaskService) Update(ctx context.Context, ownerID uint, task *models.Task) error {
	return m.Called(ctx, ownerID, task).Error(0)
}
func (m *InstanceTestMockTaskService) UpdateFailed(ctx context.Context, task *models.Task, errMsg, logMsg string) error {
	return m.Called(ctx, task, errMsg, logMsg).Error(0)
}
func (m *InstanceTestMockTaskService) AddLogs(ctx context.Context, ownerID uint, taskID uint, logs string) error {
	return m.Called(ctx, ownerID, taskID, logs).Error(0)
}
func (m *InstanceTestMockTaskService) SetResult(ctx context.Context, ownerID uint, taskID uint, result json.RawMessage) error {
	return m.Called(ctx, ownerID, taskID, result).Error(0)
}
func (m *InstanceTestMockTaskService) SetError(ctx context.Context, ownerID uint, taskID uint, errMsg string) error {
	return m.Called(ctx, ownerID, taskID, errMsg).Error(0)
}
func (m *InstanceTestMockTaskService) CompleteTask(ctx context.Context, ownerID uint, taskID uint, result json.RawMessage) error {
	return m.Called(ctx, ownerID, taskID, result).Error(0)
}
func (m *InstanceTestMockTaskService) GetSchedulableTasks(ctx context.Context, limit int) ([]models.Task, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}
func (m *InstanceTestMockTaskService) ListTasksByInstanceID(ctx context.Context, ownerID uint, instanceID uint, actionFilter models.TaskAction, opts *models.ListOptions) ([]models.Task, error) {
	args := m.Called(ctx, ownerID, instanceID, actionFilter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

// --- Mock ProjectServiceInterface specifically for instance_test.go ---
type InstanceTestMockProjectService struct {
	mock.Mock
}

func (m *InstanceTestMockProjectService) getID() uint {
	globalInstanceTestMockIDCounterForInstanceTests++
	return globalInstanceTestMockIDCounterForInstanceTests
}

func (m *InstanceTestMockProjectService) Create(ctx context.Context, project *models.Project) error {
	args := m.Called(ctx, project)
	if project != nil && project.ID == 0 {
		project.ID = m.getID()
	}
	return args.Error(0)
}

func (m *InstanceTestMockProjectService) GetByName(ctx context.Context, ownerID uint, name string) (*models.Project, error) {
	args := m.Called(ctx, ownerID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	proj := args.Get(0).(*models.Project)
	if proj.ID == 0 {
		proj.ID = m.getID()
	}
	return proj, args.Error(1)
}

func (m *InstanceTestMockProjectService) List(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Project, error) {
	args := m.Called(ctx, ownerID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *InstanceTestMockProjectService) Delete(ctx context.Context, ownerID uint, name string) error {
	return m.Called(ctx, ownerID, name).Error(0)
}

func (m *InstanceTestMockProjectService) ListInstances(ctx context.Context, ownerID uint, projectName string, opts *models.ListOptions) ([]models.Instance, error) {
	args := m.Called(ctx, ownerID, projectName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Instance), args.Error(1)
}

func TestInstanceService_CreateInstance_SetsTaskInstanceID(t *testing.T) {
	mockInstanceRepo := new(InstanceTestMockInstanceRepository) // Use renamed mock
	mockTaskService := new(InstanceTestMockTaskService)         // Use renamed mock
	mockProjectService := new(InstanceTestMockProjectService)   // Use renamed mock
	globalInstanceTestMockIDCounterForInstanceTests = 0

	instanceService := NewInstanceService(mockInstanceRepo, mockTaskService, mockProjectService)
	ctx := context.Background()

	ownerID := uint(1)
	projectName := "test-project"
	projectID := uint(10)

	instanceReqs := []types.InstanceRequest{
		{
			OwnerID: ownerID, ProjectName: projectName, Provider: models.ProviderDO,
			Region: "nyc1", Size: "s-1vcpu-1gb", Image: "ubuntu-20-04-x64",
			SSHKeyName: "test-key", NumberOfInstances: 1, Action: "create",
			Volumes: []types.VolumeConfig{{Name: "vol1", SizeGB: 10, MountPoint: "/mnt/vol1"}},
		},
	}

	mockedProject := &models.Project{OwnerID: ownerID, Name: projectName}
	mockedProject.ID = projectID
	mockProjectService.On("GetByName", ctx, ownerID, projectName).Return(mockedProject, nil)

	expectedInstanceID := uint(101)
	mockedCreatedInstancesResult := []*models.Instance{
		{Model: gorm.Model{ID: expectedInstanceID}, OwnerID: ownerID, ProjectID: projectID},
	}
	mockInstanceRepo.On("CreateBatch", ctx, mock.AnythingOfType("[]*models.Instance")).Return(mockedCreatedInstancesResult, nil).Run(func(args mock.Arguments) {
		inputInstances := args.Get(1).([]*models.Instance)
		for i, inst := range inputInstances {
			if inst.ID == 0 && i < len(mockedCreatedInstancesResult) {
				inst.ID = mockedCreatedInstancesResult[i].ID
			}
		}
	})

	var capturedTasks []*models.Task
	mockTaskService.On("CreateBatch", ctx, mock.AnythingOfType("[]*models.Task")).Return(nil).Run(func(args mock.Arguments) {
		capturedTasks = args.Get(1).([]*models.Task)
		for _, task := range capturedTasks {
			if task.ID == 0 {
				task.ID = mockTaskService.getID()
			}
		}
	})

	actualCreatedInstances, err := instanceService.CreateInstance(ctx, instanceReqs)

	assert.NoError(t, err)
	assert.NotNil(t, actualCreatedInstances)
	assert.Len(t, actualCreatedInstances, 1)
	assert.Equal(t, expectedInstanceID, actualCreatedInstances[0].ID)

	mockProjectService.AssertExpectations(t)
	mockInstanceRepo.AssertExpectations(t)
	mockTaskService.AssertExpectations(t)

	assert.Len(t, capturedTasks, 1)
	assert.Equal(t, expectedInstanceID, capturedTasks[0].InstanceID, "Task.InstanceID should match the created instance ID")

	var taskPayload types.InstanceRequest
	err = json.Unmarshal(capturedTasks[0].Payload, &taskPayload)
	assert.NoError(t, err)
	assert.Equal(t, expectedInstanceID, taskPayload.InstanceID, "Task payload InstanceID should be updated")
}

func TestInstanceService_Terminate_SetsTaskInstanceID(t *testing.T) {
	mockInstanceRepo := new(InstanceTestMockInstanceRepository) // Use renamed mock
	mockTaskService := new(InstanceTestMockTaskService)         // Use renamed mock
	mockProjectService := new(InstanceTestMockProjectService)   // Use renamed mock
	globalInstanceTestMockIDCounterForInstanceTests = 0

	instanceService := NewInstanceService(mockInstanceRepo, mockTaskService, mockProjectService)
	ctx := context.Background()

	ownerID := uint(1)
	projectName := "test-project-term"
	projectID := uint(11)
	instanceToTerminateID1 := mockInstanceRepo.getID()
	instanceToTerminateID2 := mockInstanceRepo.getID()

	mockedProject := &models.Project{Model: gorm.Model{ID: projectID}, OwnerID: ownerID, Name: projectName}
	mockProjectService.On("GetByName", ctx, ownerID, projectName).Return(mockedProject, nil)

	mockedInstancesToTerminate := []models.Instance{
		{Model: gorm.Model{ID: instanceToTerminateID1}, OwnerID: ownerID, ProjectID: projectID},
		{Model: gorm.Model{ID: instanceToTerminateID2}, OwnerID: ownerID, ProjectID: projectID},
	}
	mockInstanceRepo.On("GetByProjectIDAndInstanceIDs", ctx, ownerID, projectID, []uint{instanceToTerminateID1, instanceToTerminateID2}).Return(mockedInstancesToTerminate, nil)

	var capturedTasks []*models.Task
	mockTaskService.On("Create", ctx, mock.AnythingOfType("*models.Task")).Return(nil).Run(func(args mock.Arguments) {
		task := args.Get(1).(*models.Task)
		if task != nil && task.ID == 0 {
			task.ID = mockTaskService.getID()
		}
		if task != nil {
			capturedTasks = append(capturedTasks, task)
		}
	})

	err := instanceService.Terminate(ctx, ownerID, projectName, []uint{instanceToTerminateID1, instanceToTerminateID2})

	assert.NoError(t, err)
	mockProjectService.AssertExpectations(t)
	mockInstanceRepo.AssertExpectations(t)
	mockTaskService.AssertNumberOfCalls(t, "Create", 2)

	assert.Len(t, capturedTasks, 2)
	foundID1 := false
	foundID2 := false
	for _, task := range capturedTasks {
		if task.InstanceID == instanceToTerminateID1 {
			foundID1 = true
		}
		if task.InstanceID == instanceToTerminateID2 {
			foundID2 = true
		}
		assert.Equal(t, models.TaskActionTerminateInstances, task.Action)
		assert.Equal(t, ownerID, task.OwnerID)
		assert.Equal(t, projectID, task.ProjectID)

		var taskPayload types.DeleteInstanceRequest
		err = json.Unmarshal(task.Payload, &taskPayload)
		assert.NoError(t, err)
		assert.Equal(t, task.InstanceID, taskPayload.InstanceID)
	}
	assert.True(t, foundID1, "Task for instance ID %d not found", instanceToTerminateID1)
	assert.True(t, foundID2, "Task for instance ID %d not found", instanceToTerminateID2)
}
