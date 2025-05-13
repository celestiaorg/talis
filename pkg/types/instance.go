// Package types contains PUBLIC aliases for internal request/response structs.
//
// NOTE: This package uses type aliases to internal definitions
// as a temporary measure. This should be revisited
// during a proper refactoring to define stable public types.
package types

import (
	internaltypes "github.com/celestiaorg/talis/internal/types"
)

// InstanceRequest defines the structure for requesting instance creation (public alias).
type InstanceRequest = internaltypes.InstanceRequest

// DeleteInstancesRequest defines the structure for requesting instance deletion (public alias).
type DeleteInstancesRequest = internaltypes.DeleteInstancesRequest

// PublicIPsResponse defines the structure for the response containing public IPs (public alias).
type PublicIPsResponse = internaltypes.PublicIPsResponse
