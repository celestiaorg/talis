// Package mocks provides mock implementations of various interfaces used in Talis.
//
// The mocks are organized by component type:
//   - compute/: Mocks for compute providers (e.g., DigitalOcean)
//   - providers/: Mocks for other infrastructure providers
//
// Each mock implementation follows these principles:
//  1. Implements the same interface as the real component
//  2. Provides configurable behavior through function fields
//  3. Includes helper functions for common testing scenarios
//  4. Uses consistent naming: Mock{Interface} for the mock type
//
// Example usage:
//
//	mockClient := compute.NewMockDOClient()
//	mockClient.MockDropletService.CreateFunc = func(ctx context.Context, req *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
//		return &godo.Droplet{ID: 12345, Name: req.Name}, nil, nil
//	}
package mocks