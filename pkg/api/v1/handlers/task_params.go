// Package handlers provides HTTP request handling
package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/celestiaorg/talis/internal/db/models"
)

// TaskGetParams defines the parameters for retrieving a task
type TaskGetParams struct {
	TaskID  uint `json:"task_id"`
	OwnerID uint `json:"owner_id"`
}

// Validate validates the parameters for retrieving a task
func (p TaskGetParams) Validate() error {
	if p.TaskID == 0 {
		return fmt.Errorf("task_id is required and must be a positive number")
	}
	if p.OwnerID == 0 {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgTaskOwnerIDRequired))
	}
	return nil
}

// TaskListParams defines the parameters for listing tasks
type TaskListParams struct {
	ProjectName string `json:"projectName"`
	OwnerID     uint   `json:"owner_id"`
	Page        int    `json:"page,omitempty"`
}

// Validate validates the parameters for listing tasks
func (p TaskListParams) Validate() error {
	if p.ProjectName == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgProjNameRequired))
	}
	if p.OwnerID == 0 {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgTaskOwnerIDRequired))
	}
	if p.Page < 0 {
		return fmt.Errorf("page must be a positive number")
	}
	return nil
}

// TaskTerminateParams defines the parameters for terminating a task
type TaskTerminateParams struct {
	TaskID  uint `json:"task_id"`
	OwnerID uint `json:"owner_id"`
}

// Validate validates the parameters for terminating a task
func (p TaskTerminateParams) Validate() error {
	if p.TaskID == 0 {
		return fmt.Errorf("task_id is required and must be a positive number")
	}
	if p.OwnerID == 0 {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgTaskOwnerIDRequired))
	}
	return nil
}

// TaskUpdateStatusParams defines the parameters for updating a task's status
type TaskUpdateStatusParams struct {
	TaskID  uint              `json:"task_id"`
	Status  models.TaskStatus `json:"status"`
	OwnerID uint              `json:"owner_id"`
}

// Validate validates the parameters for updating a task's status
func (p TaskUpdateStatusParams) Validate() error {
	if p.TaskID == 0 {
		return fmt.Errorf("task_id is required and must be a positive number")
	}
	if p.OwnerID == 0 {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgTaskOwnerIDRequired))
	}
	if p.Status == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgTaskStatusReqd))
	}
	return nil
}

// TaskListByInstanceParams defines the parameters for listing tasks by instance ID.
type TaskListByInstanceParams struct {
	InstanceID uint   `json:"instance_id"`
	OwnerID    uint   `json:"owner_id"` // From context/auth, to be populated by handler/middleware
	Action     string `json:"action"`   // Optional query parameter for filtering by models.TaskAction
	Limit      int    `json:"limit"`    // Optional query parameter for pagination
	Offset     int    `json:"offset"`   // Optional query parameter for pagination
}

// Validate validates the parameters for listing tasks by instance ID.
func (p TaskListByInstanceParams) Validate() error {
	if p.InstanceID == 0 {
		return fmt.Errorf("instance_id is required and must be a positive number")
	}
	if p.OwnerID == 0 {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgTaskOwnerIDRequired))
	}
	if p.Limit < 0 {
		return fmt.Errorf("limit must be a non-negative number")
	}
	if p.Offset < 0 {
		return fmt.Errorf("offset must be a non-negative number")
	}

	// Validate that Limit is set when Offset is used
	if p.Offset > 0 && p.Limit == 0 {
		return errors.New("limit must be set when offset is used")
	}

	// Validate Action field if provided
	if p.Action != "" {
		switch models.TaskAction(p.Action) {
		case models.TaskActionCreateInstances, models.TaskActionTerminateInstances:
			// Valid actions
		default:
			return fmt.Errorf("invalid task action: %s", p.Action)
		}
	}

	return nil
}
