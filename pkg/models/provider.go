// Package models contains database models and related utility functions.
// NOTE: This package currently uses type aliases to internal definitions
// as a temporary measure to support external tools. This should be revisited
// during a proper refactoring to define stable public types.
package models

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
	// Remove unused json import "encoding/json"
)

// ProviderID represents a unique identifier for a cloud provider.
// This aliases the internal definition.
type ProviderID = internalmodels.ProviderID

// Publicly exposed provider constants (aliased from internal)
const (
	ProviderAWS      = internalmodels.ProviderAWS
	ProviderGCP      = internalmodels.ProviderGCP
	ProviderAzure    = internalmodels.ProviderAzure
	ProviderDO       = internalmodels.ProviderDO
	ProviderScaleway = internalmodels.ProviderScaleway
	ProviderVultr    = internalmodels.ProviderVultr
	ProviderLinode   = internalmodels.ProviderLinode
	ProviderHetzner  = internalmodels.ProviderHetzner
	ProviderOVH      = internalmodels.ProviderOVH

	// Mock Providers (aliased from internal)
	ProviderDOMock1 = internalmodels.ProviderDOMock1
	ProviderDOMock2 = internalmodels.ProviderDOMock2
	ProviderMock3   = internalmodels.ProviderMock3
)

// NOTE: Methods like IsValid, MarshalJSON, UnmarshalJSON are defined on the
// original internal type (internal/db/models.ProviderID) and are used via the alias.
// DO NOT DEFINE METHODS ON ALIAS TYPES HERE.

// The IsValid, MarshalJSON, and UnmarshalJSON functions previously defined here
// are removed as they should be called on the internal type.
