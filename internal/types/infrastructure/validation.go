package infrastructure

import (
	"fmt"
	"regexp"
	"strings"
)

// Hostname validation constants
const (
	maxHostnameLength = 63
	hostnameRegex     = "^[a-z0-9][a-z0-9-]*[a-z0-9]$"
)

// validateHostname validates if a string is a valid hostname according to RFC standards
func validateHostname(hostname string) error {
	// Check length
	if len(hostname) > maxHostnameLength {
		return fmt.Errorf("hostname length must be less than or equal to %d characters", maxHostnameLength)
	}

	// Check if empty
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	// Convert to lowercase for validation
	hostname = strings.ToLower(hostname)

	// Check format using regex
	valid, err := regexp.MatchString(hostnameRegex, hostname)
	if err != nil {
		return fmt.Errorf("internal error validating hostname: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid hostname format: must contain only lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen")
	}

	return nil
}

// Validate validates the infrastructure request
func (r *JobRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("job_name is required")
	}
	return nil
}

// Validate validates the infrastructure request
func (r *InstancesRequest) Validate() error {
	if r.JobName == "" {
		return fmt.Errorf("job_name is required")
	}

	if r.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}
	if len(r.Instances) == 0 {
		return fmt.Errorf("at least one instance configuration is required")
	}

	for i, instance := range r.Instances {
		if instance.Name == "" && r.InstanceName == "" {
			return fmt.Errorf("instance_name or instance.name is required")
		}

		// Validate hostname format
		nameToValidate := instance.Name
		if nameToValidate == "" {
			nameToValidate = r.InstanceName
		}
		if err := validateHostname(nameToValidate); err != nil {
			return fmt.Errorf("invalid hostname at index %d: %w", i, err)
		}

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
	if r.InstanceName == "" {
		return fmt.Errorf("instance_name is required")
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

func (u CreateUserRequest) Validate() error {
	if u.Username == "" {
		return fmt.Errorf("provider is required")
	}
	return nil
}
