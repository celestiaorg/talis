// Package handlers provides HTTP request handling
package handlers

// RPC method constants for standardized method naming
const (
	// Project methods
	ProjectCreate        = "project.create"
	ProjectGet           = "project.get"
	ProjectList          = "project.list"
	ProjectDelete        = "project.delete"
	ProjectListInstances = "project.listInstances"

	// Task methods
	TaskGet          = "task.get"
	TaskList         = "task.list"
	TaskTerminate    = "task.terminate"
	TaskUpdateStatus = "task.updateStatus"

	// User methods
	UserCreate  = "user.create"
	UserGet     = "user.get"
	UserGetByID = "user.get.id"
	UserDelete  = "user.delete"
)

// IsProjectMethod checks if the given method is a project operation
func IsProjectMethod(method string) bool {
	switch method {
	case ProjectCreate, ProjectGet, ProjectList, ProjectDelete, ProjectListInstances:
		return true
	default:
		return false
	}
}

// IsTaskMethod checks if the given method is a task operation
func IsTaskMethod(method string) bool {
	switch method {
	case TaskGet, TaskList, TaskTerminate, TaskUpdateStatus:
		return true
	default:
		return false
	}
}

// IsUserMethod checks if the given method is a user operation
func IsUserMethod(method string) bool {
	switch method {
	case UserCreate, UserGet, UserGetByID, UserDelete:
		return true
	default:
		return false
	}
}
