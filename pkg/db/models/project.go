// Package models contains PUBLIC aliases for database models and related types.
//
// NOTE: This package uses type aliases to internal definitions
// as a temporary measure. This should be revisited
// during a proper refactoring to define stable public types.
package models

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
)

// Project represents a project in the system (public alias).
type Project = internalmodels.Project

// NOTE: Methods are defined on the original internal types.
