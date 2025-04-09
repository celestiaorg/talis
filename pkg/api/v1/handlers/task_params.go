// Package handlers provides HTTP request handling
package handlers

import (
	"fmt"
	"strings"

	"github.com/celestiaorg/talis/internal/db/models"
)

// TaskGetParams defines the parameters for retrieving a task
type TaskGetParams struct {
	ProjectName string `json:"projectName"`
	TaskName    string `json:"taskName"`
}

// Validate validates the parameters for retrieving a task
func (p TaskGetParams) Validate() error {
	if p.ProjectName == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgProjNameRequired))
	}
	if p.TaskName == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgTaskNameRequired))
	}
	return nil
}

// TaskListParams defines the parameters for listing tasks
type TaskListParams struct {
	ProjectName string `json:"projectName"`
	Page        int    `json:"page,omitempty"`
}

// Validate validates the parameters for listing tasks
func (p TaskListParams) Validate() error {
	if p.ProjectName == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgProjNameRequired))
	}
	if p.Page < 0 {
		return fmt.Errorf("page must be a positive number")
	}
	return nil
}

// TaskTerminateParams defines the parameters for terminating a task
type TaskTerminateParams struct {
	ProjectName string `json:"projectName"`
	TaskName    string `json:"taskName"`
}

// Validate validates the parameters for terminating a task
func (p TaskTerminateParams) Validate() error {
	if p.ProjectName == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgProjNameRequired))
	}
	if p.TaskName == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgTaskNameRequired))
	}
	return nil
}

// TaskUpdateStatusParams defines the parameters for updating a task's status
type TaskUpdateStatusParams struct {
	ProjectName string            `json:"projectName"`
	TaskName    string            `json:"taskName"`
	Status      models.TaskStatus `json:"status"`
}

// Validate validates the parameters for updating a task's status
func (p TaskUpdateStatusParams) Validate() error {
	if p.ProjectName == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgProjNameRequired))
	}
	if p.TaskName == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgTaskNameRequired))
	}
	if p.Status == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgTaskStatusReqd))
	}
	return nil
}
