// Package models contains PUBLIC aliases for database models and related types.
//
// NOTE: This package uses type aliases to internal definitions
// as a temporary measure. This should be revisited
// during a proper refactoring to define stable public types.
package models

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
)

// VolumeDetails represents the details of a volume (public alias)
type VolumeDetails = internalmodels.VolumeDetails

// VolumeDetail represents a block storage volume (public alias)
type VolumeDetail = internalmodels.VolumeDetail

// NOTE: Methods are defined on the original internal types.
