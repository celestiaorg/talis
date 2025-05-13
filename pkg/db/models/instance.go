// Package models contains PUBLIC aliases for database models and related types.
//
// NOTE: This package uses type aliases to internal definitions
// as a temporary measure. This should be revisited
// during a proper refactoring to define stable public types.
package models

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
)

// InstanceStatus represents the current state of an instance
type InstanceStatus = internalmodels.InstanceStatus

// Instance status constants
const (
	InstanceStatusUnknown      InstanceStatus = internalmodels.InstanceStatusUnknown
	InstanceStatusPending      InstanceStatus = internalmodels.InstanceStatusPending
	InstanceStatusCreated      InstanceStatus = internalmodels.InstanceStatusCreated
	InstanceStatusProvisioning InstanceStatus = internalmodels.InstanceStatusProvisioning
	InstanceStatusReady        InstanceStatus = internalmodels.InstanceStatusReady
	InstanceStatusTerminated   InstanceStatus = internalmodels.InstanceStatusTerminated
)

// PayloadStatus represents the state of a payload operation on an instance
type PayloadStatus = internalmodels.PayloadStatus

// Payload status constants
const (
	PayloadStatusNone             PayloadStatus = internalmodels.PayloadStatusNone
	PayloadStatusPendingCopy      PayloadStatus = internalmodels.PayloadStatusPendingCopy
	PayloadStatusCopyFailed       PayloadStatus = internalmodels.PayloadStatusCopyFailed
	PayloadStatusCopied           PayloadStatus = internalmodels.PayloadStatusCopied
	PayloadStatusPendingExecution PayloadStatus = internalmodels.PayloadStatusPendingExecution
	PayloadStatusExecutionFailed  PayloadStatus = internalmodels.PayloadStatusExecutionFailed
	PayloadStatusExecuted         PayloadStatus = internalmodels.PayloadStatusExecuted
)

// Instance represents a compute instance in the system (public alias)
type Instance = internalmodels.Instance

// NOTE: Methods like String(), ParseInstanceStatus(), MarshalJSON(), UnmarshalJSON()
// are defined on the original internal types and are used via the aliases.
// DO NOT REDEFINE METHODS ON ALIAS TYPES HERE.
