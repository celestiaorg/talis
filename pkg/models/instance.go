// Package models contains database models and related utility functions.
// NOTE: This package currently uses type aliases to internal definitions
// as a temporary measure to support external tools. This should be revisited
// during a proper refactoring to define stable public types.
package models

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
)

// Public aliases for internal types and constants

// Field name constants (aliased)
const (
	InstanceCreatedAtField = internalmodels.InstanceCreatedAtField
	InstanceDeletedField   = internalmodels.InstanceDeletedField
	InstanceStatusField    = internalmodels.InstanceStatusField
	InstancePublicIPField  = internalmodels.InstancePublicIPField
	InstanceNameField      = internalmodels.InstanceNameField
)

// Type Aliases

// InstanceStatus represents the status of an instance.
type InstanceStatus = internalmodels.InstanceStatus

// PayloadStatus represents the status of payload execution on an instance.
type PayloadStatus = internalmodels.PayloadStatus

// Instance represents a single compute instance.
type Instance = internalmodels.Instance

// Constant Aliases
const (
	// InstanceStatus
	InstanceStatusUnknown      = internalmodels.InstanceStatusUnknown
	InstanceStatusPending      = internalmodels.InstanceStatusPending
	InstanceStatusCreated      = internalmodels.InstanceStatusCreated
	InstanceStatusProvisioning = internalmodels.InstanceStatusProvisioning
	InstanceStatusReady        = internalmodels.InstanceStatusReady
	InstanceStatusTerminated   = internalmodels.InstanceStatusTerminated

	// PayloadStatus
	PayloadStatusNone             = internalmodels.PayloadStatusNone
	PayloadStatusPendingCopy      = internalmodels.PayloadStatusPendingCopy
	PayloadStatusCopyFailed       = internalmodels.PayloadStatusCopyFailed
	PayloadStatusCopied           = internalmodels.PayloadStatusCopied
	PayloadStatusPendingExecution = internalmodels.PayloadStatusPendingExecution
	PayloadStatusExecutionFailed  = internalmodels.PayloadStatusExecutionFailed
	PayloadStatusExecuted         = internalmodels.PayloadStatusExecuted
)

// Function Aliases (assigned from internal package)
var (
	ParseInstanceStatus = internalmodels.ParseInstanceStatus
	ParsePayloadStatus  = internalmodels.ParsePayloadStatus
)

// NOTE: All methods (String, Parse*, MarshalJSON, UnmarshalJSON) are defined
// on the original internal types and are used via the aliases.
// DO NOT DEFINE METHODS ON ALIAS TYPES HERE.
