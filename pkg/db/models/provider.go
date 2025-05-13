// Package models contains PUBLIC aliases for database models and related types.
//
// NOTE: This package uses type aliases to internal definitions
// as a temporary measure. This should be revisited
// during a proper refactoring to define stable public types.
package models

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
)

// ProviderID represents a compute provider
type ProviderID = internalmodels.ProviderID

// Provider constants
const (
	ProviderAWS      ProviderID = internalmodels.ProviderAWS
	ProviderGCP      ProviderID = internalmodels.ProviderGCP
	ProviderAzure    ProviderID = internalmodels.ProviderAzure
	ProviderDO       ProviderID = internalmodels.ProviderDO
	ProviderScaleway ProviderID = internalmodels.ProviderScaleway
	ProviderVultr    ProviderID = internalmodels.ProviderVultr
	ProviderLinode   ProviderID = internalmodels.ProviderLinode
	ProviderHetzner  ProviderID = internalmodels.ProviderHetzner
	ProviderOVH      ProviderID = internalmodels.ProviderOVH

	// Mock Providers
	ProviderDOMock1 ProviderID = internalmodels.ProviderDOMock1
	ProviderDOMock2 ProviderID = internalmodels.ProviderDOMock2
	ProviderMock3   ProviderID = internalmodels.ProviderMock3
)

// NOTE: Methods like IsValid, MarshalJSON, UnmarshalJSON are defined on the
// original internal types and are used via the aliases.
// DO NOT REDEFINE METHODS ON ALIAS TYPES HERE.
