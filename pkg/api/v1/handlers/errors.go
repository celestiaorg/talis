// Package handlers provides HTTP request handling
package handlers

// Common error messages
const (
	ErrMsgInvalidParams     = "Invalid parameters"
	ErrMsgInvalidReqFormat  = "Invalid request format"
	ErrMsgMethodRequired    = "Method is required"
	ErrMsgUnknownMethod     = "Unknown method"
	ErrMsgUnknownProjMethod = "Unknown project method"
	ErrMsgUnknownTaskMethod = "Unknown task method"
)

// Project error messages
const (
	ErrMsgProjNameRequired = "Project name is required"
	ErrMsgProjNotFound     = "Project not found"
	ErrMsgProjCreateFailed = "Failed to create project"
	ErrMsgProjListFailed   = "Failed to list projects"
	ErrMsgProjDeleteFailed = "Failed to delete project"
	ErrMsgProjGetFailed    = "Failed to get project"
)

// Task error messages
const (
	ErrMsgTaskNameRequired    = "Task name is required"
	ErrMsgTaskNotFound        = "Task not found"
	ErrMsgTaskListFailed      = "Failed to list tasks"
	ErrMsgTaskTerminateFailed = "Failed to terminate task"
	ErrMsgTaskStatusFailed    = "Failed to update task status"
	ErrMsgTaskStatusReqd      = "Status is required"
	ErrMsgInvalidReqBody      = "Invalid request body"
	ErrMsgTaskStatusInvalid   = "Invalid task status"
	ErrMsgTaskGetFailed       = "Failed to get task"
)
