package infrastructure

import (
	"fmt"
	"strings"
)

// Validate validates the infrastructure request
func (r *JobRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

// Validate validates the infrastructure request
func (r *InstancesRequest) Validate() error {
	if r.JobName == "" {
		return fmt.Errorf("job_name is required")
	}
	if len(r.Instances) == 0 {
		return fmt.Errorf("at least one instance configuration is required")
	}

	for i, instance := range r.Instances {
		if err := instance.Validate(); err != nil {
			return fmt.Errorf("invalid instance configuration at index %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates the instance configuration
func (i *InstanceRequest) Validate() error {
	if i.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if i.NumberOfInstances < 1 {
		return fmt.Errorf("number_of_instances must be greater than 0")
	}
	if i.Region == "" {
		return fmt.Errorf("region is required")
	}
	if i.Size == "" {
		return fmt.Errorf("size is required")
	}
	if i.Image == "" {
		return fmt.Errorf("image is required")
	}
	if i.SSHKeyName == "" {
		return fmt.Errorf("ssh_key_name is required")
	}

	i.SSHKeyName = strings.ToLower(i.SSHKeyName)
	return nil
}

// Validate validates the delete request
func (r *DeleteRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}
	if len(r.Instances) == 0 {
		return fmt.Errorf("at least one instance configuration is required")
	}

	for i, instance := range r.Instances {
		if err := instance.Validate(); err != nil {
			return fmt.Errorf("invalid instance configuration at index %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates the delete instance configuration
func (i *DeleteInstance) Validate() error {
	if i.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if i.NumberOfInstances < 1 {
		return fmt.Errorf("number_of_instances must be greater than 0")
	}
	if i.Region == "" {
		return fmt.Errorf("region is required")
	}
	if i.Size == "" {
		return fmt.Errorf("size is required")
	}
	return nil
}
