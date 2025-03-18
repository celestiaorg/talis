// Package test provides integration testing infrastructure for Talis
package test

import (
	"context"
	"time"
)

// Version represents the current version of the test package.
// This will be updated as new features are added.
const Version = "0.1.0"

// DefaultTestTimeout is the default timeout for test environments.
const DefaultTestTimeout = 30 * time.Second

// Option represents a configuration option for the test environment.
type Option func(*TestEnvironment)

// WithTimeout returns an option that sets the test environment timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(env *TestEnvironment) {
		if env.cancelFunc != nil {
			env.cancelFunc()
		}
		env.ctx, env.cancelFunc = context.WithTimeout(context.Background(), timeout)
	}
}

// WithCleanupFunc returns an option that adds a cleanup function to be
// called when the environment is cleaned up.
func WithCleanupFunc(cleanup func()) Option {
	return func(env *TestEnvironment) {
		oldCleanup := env.cleanup
		env.cleanup = func() {
			if cleanup != nil {
				cleanup()
			}
			if oldCleanup != nil {
				oldCleanup()
			}
		}
	}
}

// WithOption applies the given option to the test environment.
// This is a placeholder that will be expanded as we add more options.
func WithOption(opt Option) Option {
	return opt
}
