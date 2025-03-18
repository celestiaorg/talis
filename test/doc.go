// Package test provides infrastructure and utilities for integration testing in Talis.
//
// The test package implements a complete test environment that allows testing
// the interaction between different components while maintaining control over
// external dependencies. It can be used both within Talis and by external
// packages that want to test their integration with Talis.
//
// The package provides:
//
//   - TestEnvironment: A struct that manages a complete test setup including
//     an in-memory database, real API server, and mocked external providers
//
//   - Mock Management: Centralized mock implementations and configurations
//     for external dependencies like cloud providers
//
//   - Test Utilities: Helper functions for common testing scenarios and
//     assertions
//
// Example Usage:
//
//	func TestExample(t *testing.T) {
//	    env := test.NewTestEnvironment(t)
//	    defer env.Cleanup()
//
//	    // Use env.APIClient to make requests
//	    // Use env.MockDOClient to configure provider behavior
//	}
//
// The test package is designed to:
//  1. Enable testing of API contracts between client and server
//  2. Provide consistent mocking of external dependencies
//  3. Reduce test setup boilerplate
//  4. Make tests more maintainable and reliable
//  5. Support external package integration testing
package test
