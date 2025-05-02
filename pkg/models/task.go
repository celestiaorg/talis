package models

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
)

// TaskStatus represents the current state of a task
type TaskStatus = internalmodels.TaskStatus

// Task status constants
const (
	// TaskStatusUnknown represents an unknown or invalid task status
	TaskStatusUnknown TaskStatus = internalmodels.TaskStatusUnknown
	// TaskStatusPending indicates the task is waiting to be processed
	TaskStatusPending TaskStatus = internalmodels.TaskStatusPending
	// TaskStatusRunning indicates the task is currently being processed
	TaskStatusRunning TaskStatus = internalmodels.TaskStatusRunning
	// TaskStatusCompleted indicates the task has been successfully completed
	TaskStatusCompleted TaskStatus = internalmodels.TaskStatusCompleted
	// TaskStatusFailed indicates the task has failed
	TaskStatusFailed TaskStatus = internalmodels.TaskStatusFailed
	// TaskStatusTerminated indicates the task was manually aborted
	TaskStatusTerminated TaskStatus = internalmodels.TaskStatusTerminated
)

// TaskAction represents the possible actions a task can perform.
type TaskAction = internalmodels.TaskAction

const (
	// TaskActionCreateInstances represents the action to create instances.
	TaskActionCreateInstances TaskAction = internalmodels.TaskActionCreateInstances
	// TaskActionTerminateInstances represents the action to terminate instances.
	TaskActionTerminateInstances TaskAction = internalmodels.TaskActionTerminateInstances
)

// Task represents an asynchronous operation that can be tracked
type Task = internalmodels.Task

// ParseTaskStatus converts a string to a TaskStatus.
var ParseTaskStatus = internalmodels.ParseTaskStatus
