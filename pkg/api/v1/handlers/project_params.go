// Package handlers provides HTTP request handling
package handlers

import (
	"fmt"
	"strings"
)

// ProjectConfig defines the configuration options for a project
type ProjectConfig struct {
	// Add any project-specific configuration fields here
	Resources map[string]interface{} `json:"resources,omitempty"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
}

// ProjectCreateParams defines the parameters for creating a project
type ProjectCreateParams struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Config      string `json:"config,omitempty"`
}

// Validate validates the parameters for creating a project
func (p ProjectCreateParams) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgProjNameRequired))
	}
	return nil
}

// ProjectGetParams defines the parameters for retrieving a project
type ProjectGetParams struct {
	Name string `json:"name"`
}

// Validate validates the parameters for retrieving a project
func (p ProjectGetParams) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgProjNameRequired))
	}
	return nil
}

// ProjectListParams defines the parameters for listing projects
type ProjectListParams struct {
	Page int `json:"page,omitempty"`
}

// Validate validates the parameters for listing projects
func (p ProjectListParams) Validate() error {
	if p.Page < 0 {
		return fmt.Errorf("page must be a positive number")
	}
	return nil
}

// ProjectDeleteParams defines the parameters for deleting a project
type ProjectDeleteParams struct {
	Name string `json:"name"`
}

// Validate validates the parameters for deleting a project
func (p ProjectDeleteParams) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgProjNameRequired))
	}
	return nil
}

// ProjectListInstancesParams defines the parameters for listing project instances
type ProjectListInstancesParams struct {
	Name string `json:"name"`
	Page int    `json:"page,omitempty"`
}

// Validate validates the parameters for listing project instances
func (p ProjectListInstancesParams) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgProjNameRequired))
	}
	if p.Page < 0 {
		return fmt.Errorf("page must be a positive number")
	}
	return nil
}
