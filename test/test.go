// Package test provides integration testing infrastructure for Talis
package test

import (
	"testing"
)

// Version represents the current version of the test package.
// This will be updated as new features are added.
const Version = "0.1.0"

// Option represents a configuration option for the test environment.
type Option func(*TestEnvironment)

// TestEnvironment represents a complete test environment for integration testing.
// This is a placeholder that will be expanded in Task 1.2.
type TestEnvironment struct {
	t *testing.T //nolint:unused
}

// WithOption applies the given option to the test environment.
// This is a placeholder that will be expanded as we add more options.
func WithOption(opt Option) Option {
	return opt
}
