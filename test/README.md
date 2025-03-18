# Talis Integration Test Package

This package provides infrastructure and utilities for integration testing in Talis. It enables testing of API contracts and component interactions while maintaining control over external dependencies. The package can be used both within Talis and by external packages that want to test their integration with Talis.

## Features

- **TestEnvironment**: Complete test setup including:
  - In-memory database
  - Real API server
  - Real API client
  - Mocked external providers

- **Mock Management**: 
  - Centralized mock implementations
  - Standardized configurations
  - Provider factory functions

- **Test Utilities**:
  - Common test scenarios
  - Assertion helpers
  - Request builders

## Quick Start

```go
import "github.com/celestiaorg/talis/test"

func TestExample(t *testing.T) {
    // Create a test environment
    env := test.NewTestEnvironment(t)
    defer env.Cleanup()

    // Configure mock behavior
    test.SetupMockDOForInstanceCreation(env.MockDOClient, test.DefaultInstanceCreationConfig())

    // Use the environment
    instance, err := env.APIClient.CreateInstance(env.Context(), instanceRequest)
    require.NoError(t, err)
    
    // Make assertions
    test.AssertInstanceEquals(t, expected, instance)
}
```

## Package Structure

```
test/
├── doc.go           # Package documentation
├── test.go          # Core types and interfaces
├── environment.go   # TestEnvironment implementation
├── mocks/           # Mock implementations
│   ├── config.go    # Mock configurations
│   └── providers/   # Provider-specific mocks
└── utils/           # Test utilities
```

## Best Practices

1. **Resource Cleanup**
   - Always defer `env.Cleanup()` after creating a test environment
   - Use `env.Context()` for proper timeout management

2. **Mock Configuration**
   - Use provided helper functions to configure mocks
   - Avoid direct manipulation of mock internals
   - Keep mock setup close to test cases

3. **Test Organization**
   - Group related test cases using subtests
   - Use descriptive test names
   - Document complex test scenarios

## Using in External Projects

To use this package in your project:

1. Add the dependency:
   ```bash
   go get github.com/celestiaorg/talis/test
   ```

2. Import the package:
   ```go
   import "github.com/celestiaorg/talis/test"
   ```

3. Use the test environment in your tests as shown in the Quick Start section.

## Contributing

When adding new features to the test package:

1. Update package documentation in `doc.go`
2. Add tests for new functionality
3. Update this README with examples if needed
4. Follow existing patterns for consistency

## Version History

- v0.1.0: Initial implementation
  - Basic TestEnvironment structure
  - Package documentation
  - Initial type definitions 