package services

import (
	"context"
	"encoding/json"

	"github.com/celestiaorg/talis/internal/db/models"
	// "github.com/celestiaorg/talis/internal/db/repos" // No longer directly needed for type casting here
)

// TaskRepositoryInterface defines the interface for task repository operations.
// It mirrors the methods of repos.TaskRepository.
type TaskRepositoryInterface interface {
	Create(ctx context.Context, task *models.Task) error
	CreateBatch(ctx context.Context, tasks []*models.Task) error
	GetByID(ctx context.Context, ownerID uint, id uint) (*models.Task, error)
	ListByProject(ctx context.Context, ownerID uint, projectID uint, opts *models.ListOptions) ([]models.Task, error)
	UpdateStatus(ctx context.Context, ownerID uint, id uint, status models.TaskStatus) error
	Update(ctx context.Context, ownerID uint, task *models.Task) error
	GetSchedulableTasks(ctx context.Context, limit int) ([]models.Task, error)
	ListByInstanceID(ctx context.Context, ownerID uint, instanceID uint, actionFilter models.TaskAction, opts *models.ListOptions) ([]models.Task, error)
}

// ProjectRepositoryInterface defines the interface for project repository operations.
// It mirrors the methods of repos.ProjectRepository.
type ProjectRepositoryInterface interface {
	Create(ctx context.Context, project *models.Project) error
	CreateBatch(ctx context.Context, projects []*models.Project) error
	Get(ctx context.Context, projectID uint) (*models.Project, error)
	GetByName(ctx context.Context, ownerID uint, name string) (*models.Project, error)
	List(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Project, error)
	Delete(ctx context.Context, ownerID uint, name string) error
	ListInstances(ctx context.Context, projectID uint, opts *models.ListOptions) ([]models.Instance, error)
	// Add any other methods from repos.ProjectRepository
}

// ProjectServiceInterface defines the interface for project service operations.
// It mirrors the methods of the Project service.
type ProjectServiceInterface interface {
	Create(ctx context.Context, project *models.Project) error
	GetByName(ctx context.Context, ownerID uint, name string) (*models.Project, error)
	List(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Project, error)
	Delete(ctx context.Context, ownerID uint, name string) error
	ListInstances(ctx context.Context, ownerID uint, projectName string, opts *models.ListOptions) ([]models.Instance, error)
}

// InstanceRepositoryInterface defines the interface for instance repository operations.
// It mirrors methods of repos.InstanceRepository used by services.
type InstanceRepositoryInterface interface {
	List(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Instance, error)
	Create(ctx context.Context, instance *models.Instance) (*models.Instance, error)
	CreateBatch(ctx context.Context, instances []*models.Instance) ([]*models.Instance, error)
	Get(ctx context.Context, ownerID uint, id uint) (*models.Instance, error)
	GetByProjectIDAndInstanceIDs(ctx context.Context, ownerID uint, projectID uint, instanceIDs []uint) ([]models.Instance, error)
	Update(ctx context.Context, ownerID uint, instanceID uint, instance *models.Instance) error
	Terminate(ctx context.Context, ownerID, instanceID uint) error
}

// TaskServiceInterface defines the interface for task service operations.
// It mirrors methods of the Task service.
type TaskServiceInterface interface {
	Create(ctx context.Context, task *models.Task) error
	CreateBatch(ctx context.Context, tasks []*models.Task) error
	GetByID(ctx context.Context, ownerID uint, taskID uint) (*models.Task, error)
	ListByProject(ctx context.Context, ownerID uint, projectName string, opts *models.ListOptions) ([]models.Task, error)
	UpdateStatus(ctx context.Context, ownerID uint, taskID uint, status models.TaskStatus) error
	Update(ctx context.Context, ownerID uint, task *models.Task) error
	UpdateFailed(ctx context.Context, task *models.Task, errMsg, logMsg string) error
	AddLogs(ctx context.Context, ownerID uint, taskID uint, logs string) error
	SetResult(ctx context.Context, ownerID uint, taskID uint, result json.RawMessage) error
	SetError(ctx context.Context, ownerID uint, taskID uint, errMsg string) error
	CompleteTask(ctx context.Context, ownerID uint, taskID uint, result json.RawMessage) error
	GetSchedulableTasks(ctx context.Context, limit int) ([]models.Task, error)
	ListTasksByInstanceID(ctx context.Context, ownerID uint, instanceID uint, actionFilter models.TaskAction, opts *models.ListOptions) ([]models.Task, error)
}
