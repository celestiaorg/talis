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
	TaskAbort        = "task.abort"
	TaskUpdateStatus = "task.updateStatus"
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
	case TaskGet, TaskList, TaskAbort, TaskUpdateStatus:
		return true
	default:
		return false
	}
}
