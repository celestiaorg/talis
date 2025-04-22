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

// User error messages
const (
	ErrMsgInvalidUserID          = "Invalid user id"
	ErrMsgUserIDRequired         = "User id is required"
	ErrMsgUsernameRequired       = "Username is required"
	ErrMsgInvalidUsername        = "Invalid username"
	ErrMsgInvalidUserEmail       = "Invalid user email format"
	ErrMsgUserNotFoundByID       = "User not found with provided id"
	ErrMsgUserNotFoundByUsername = "User not found with provided username"
	ErrMsgGetUsersFailed         = "Failed to get users"
	ErrMsgGetUserFailed          = "Failed to get user"
	ErrMsgCreateUserFailed       = "Failed to create user"
	ErrMsgDeleteUserFailed       = "Failed to delete user"
	ErrMsgNegativeUserID         = "User ID must be positive"
	ErrMsgNilUserObject          = "User object is nil"
)

// Pagination error messages
const (
	ErrMsgNegativePagination = "Page must be a positive number from 1"
)
