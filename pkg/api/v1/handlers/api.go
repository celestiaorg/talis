package handlers

import "github.com/celestiaorg/talis/internal/services"

// APIHandler is a handler for the API
type APIHandler struct {
	instance *services.Instance
	project  *services.Project
	task     *services.Task
	user     *services.User
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(instance *services.Instance, project *services.Project, task *services.Task, user *services.User) *APIHandler {
	return &APIHandler{
		instance: instance,
		project:  project,
		task:     task,
		user:     user,
	}
}
