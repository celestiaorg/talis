// Package models contains PUBLIC aliases for database models and related types.
//
// NOTE: This package uses type aliases to internal definitions
// as a temporary measure. This should be revisited
// during a proper refactoring to define stable public types.
package models

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
)

// TaskStatus represents the status of a task.
type TaskStatus = internalmodels.TaskStatus

// Task status constants.
const (
	TaskStatusUnknown    TaskStatus = internalmodels.TaskStatusUnknown
	TaskStatusPending    TaskStatus = internalmodels.TaskStatusPending
	TaskStatusRunning    TaskStatus = internalmodels.TaskStatusRunning
	TaskStatusCompleted  TaskStatus = internalmodels.TaskStatusCompleted
	TaskStatusFailed     TaskStatus = internalmodels.TaskStatusFailed
	TaskStatusTerminated TaskStatus = internalmodels.TaskStatusTerminated
)

// TaskAction defines the type of action a task performs.
type TaskAction = internalmodels.TaskAction

// Task action constants.
const (
	// TaskActionUnknown              TaskAction = internalmodels.TaskActionUnknown // REMOVED - Not defined internally
	TaskActionCreateInstances    TaskAction = internalmodels.TaskActionCreateInstances
	TaskActionTerminateInstances TaskAction = internalmodels.TaskActionTerminateInstances
	TaskActionDeleteUpload       TaskAction = internalmodels.TaskActionDeleteUpload
)

// Task represents a background task in the system (public alias).
type Task = internalmodels.Task

// TaskResult represents the result of a completed task (public alias).
// type TaskResult = internalmodels.TaskResult // REMOVED - Result is json.RawMessage in internal Task struct

// NOTE: Methods like String(), ParseTaskStatus(), MarshalJSON(), UnmarshalJSON()
// are defined on the original internal types and are used via the aliases.
// DO NOT REDEFINE METHODS ON ALIAS TYPES HERE.
