// Package models contains database models and related utility functions
package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Field names for instance model
const (
	InstanceCreatedAtField = "created_at"
	InstanceDeletedField   = "deleted"
	InstanceStatusField    = "status"
	InstancePublicIPField  = "public_ip"
	InstanceNameField      = "name"
)

// InstanceStatus represents the current state of an instance
type InstanceStatus int

// Instance status constants
const (
	// InstanceStatusUnknown represents an unknown or invalid instance status
	InstanceStatusUnknown InstanceStatus = iota
	// InstanceStatusPending indicates the instance is being created
	InstanceStatusPending
	// InstanceStatusCreated indicates the instance has been created but not provisioned
	InstanceStatusCreated
	// InstanceStatusProvisioning indicates the instance is being provisioned
	InstanceStatusProvisioning
	// InstanceStatusReady indicates the instance is operational
	InstanceStatusReady
	// InstanceStatusTerminated indicates the instance is terminated
	InstanceStatusTerminated
)

// PayloadStatus represents the state of a payload operation on an instance
type PayloadStatus int

// Payload status constants
const (
	// PayloadStatusNone indicates no payload operation has been initiated
	PayloadStatusNone PayloadStatus = iota
	// PayloadStatusPendingCopy indicates the payload is waiting to be copied
	PayloadStatusPendingCopy
	// PayloadStatusCopyFailed indicates the payload copy operation failed
	PayloadStatusCopyFailed
	// PayloadStatusCopied indicates the payload has been successfully copied
	PayloadStatusCopied
	// PayloadStatusPendingExecution indicates the payload is waiting to be executed
	PayloadStatusPendingExecution
	// PayloadStatusExecutionFailed indicates the payload execution failed
	PayloadStatusExecutionFailed
	// PayloadStatusExecuted indicates the payload was executed successfully
	PayloadStatusExecuted
)

// Instance represents a compute instance in the system
type Instance struct {
	gorm.Model
	OwnerID            uint           `json:"owner_id" gorm:"not null;index"`
	ProjectID          uint           `json:"project_id" gorm:"not null;index"`
	LastTaskID         uint           `json:"last_task_id" gorm:"not null;index"`
	ProviderID         ProviderID     `json:"provider_id" gorm:"not null"`
	ProviderInstanceID int            `json:"provider_instance_id" gorm:"not null"`
	Name               string         `json:"name" gorm:"not null;index"`
	PublicIP           string         `json:"public_ip" gorm:"varchar(100)"`
	Region             string         `json:"region" gorm:"varchar(255)"`
	Size               string         `json:"size" gorm:"varchar(255)"`
	Image              string         `json:"image" gorm:"varchar(255)"`
	Tags               pq.StringArray `json:"tags" gorm:"type:text[]"`
	Status             InstanceStatus `json:"status" gorm:"index"`
	VolumeIDs          pq.StringArray `json:"volume_ids" gorm:"type:text[]"`
	VolumeDetails      VolumeDetails  `json:"volume_details" gorm:"type:jsonb"`
	CreatedAt          time.Time      `json:"created_at" gorm:"index"`
	PayloadStatus      PayloadStatus  `json:"payload_status" gorm:"default:0;index"` // Default to PayloadStatusNone
}

func (s InstanceStatus) String() string {
	return []string{
		"unknown",
		"pending",
		"provisioning",
		"ready",
		"terminated",
	}[s]
}

// String returns the string representation of PayloadStatus
func (ps PayloadStatus) String() string {
	return []string{
		"none",
		"pending_copy",
		"copy_failed",
		"copied",
		"pending_execution",
		"execution_failed",
		"executed",
	}[ps]
}

// ParseInstanceStatus converts a string representation of an instance status to InstanceStatus type
func ParseInstanceStatus(str string) (InstanceStatus, error) {
	for i, status := range []string{
		"unknown",
		"pending",
		"provisioning",
		"ready",
		"terminated",
	} {
		if status == str {
			return InstanceStatus(i), nil
		}
	}

	return InstanceStatus(0), fmt.Errorf("invalid instance status: %s", str)
}

// ParsePayloadStatus converts a string representation to PayloadStatus type
func ParsePayloadStatus(str string) (PayloadStatus, error) {
	for i, status := range []string{
		"none",
		"pending_copy",
		"copy_failed",
		"copied",
		"pending_execution",
		"execution_failed",
		"executed",
	} {
		if status == str {
			return PayloadStatus(i), nil
		}
	}
	return PayloadStatus(0), fmt.Errorf("invalid payload status: %s", str)
}

// MarshalJSON implements the json.Marshaler interface for InstanceStatus
func (s InstanceStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// MarshalJSON implements the json.Marshaler interface for PayloadStatus
func (ps PayloadStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(ps.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for InstanceStatus
func (s *InstanceStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status, err := ParseInstanceStatus(str)
	if err != nil {
		return err
	}

	*s = status
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for PayloadStatus
func (ps *PayloadStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status, err := ParsePayloadStatus(str)
	if err != nil {
		return err
	}

	*ps = status
	return nil
}

// MarshalJSON implements the json.Marshaler interface for Instance
func (i Instance) MarshalJSON() ([]byte, error) {
	type Alias Instance // Create an alias to avoid infinite recursion
	return json.Marshal(Alias(i))
}
