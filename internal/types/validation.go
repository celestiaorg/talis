package types

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
