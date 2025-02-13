package job

import (
	"context"
	"database/sql"
)

// Repository defines operations for persisting jobs
type Repository interface {
	Create(ctx context.Context, job *Job) error
	GetByID(ctx context.Context, id string) (*Job, error)
	List(ctx context.Context, opts *ListOptions) ([]*Job, error)
	UpdateStatus(ctx context.Context, id string, status string, result interface{}, errMsg string) error
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}
