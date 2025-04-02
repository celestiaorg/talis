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
	// InstanceCreatedAtField is the field name for instance creation timestamp
	InstanceCreatedAtField = "created_at"
	InstanceDeletedField   = "deleted"
	InstanceStatusField    = "status"
	InstancePublicIPField  = "public_ip"
)

// InstanceStatus represents the current state of an instance
type InstanceStatus int

// Instance status constants
const (
	// InstanceStatusUnknown represents an unknown or invalid instance status
	InstanceStatusUnknown InstanceStatus = iota
	// InstanceStatusPending indicates the instance is being created
	InstanceStatusPending
	// InstanceStatusProvisioning indicates the instance is being provisioned
	InstanceStatusProvisioning
	// InstanceStatusReady indicates the instance is operational
	InstanceStatusReady
	// InstanceStatusTerminated indicates the instance is terminated
	InstanceStatusTerminated
)

// Instance represents a compute instance in the system
type Instance struct {
	gorm.Model
	JobID         uint           `json:"job_id" gorm:"not null;index"`
	ProviderID    ProviderID     `json:"provider_id" gorm:"not null"`
	Name          string         `json:"name" gorm:"not null;index"`
	PublicIP      string         `json:"public_ip" gorm:"varchar(100)"`
	Region        string         `json:"region" gorm:"varchar(255)"`
	Size          string         `json:"size" gorm:"varchar(255)"`
	Image         string         `json:"image" gorm:"varchar(255)"`
	Tags          pq.StringArray `json:"tags" gorm:"type:text[]"`
	Status        InstanceStatus `json:"status" gorm:"index"`
	IsProvisioned bool           `json:"is_provisioned" gorm:"default:false"`
	Volumes       pq.StringArray `json:"volumes" gorm:"type:text[]"`
	CreatedAt     time.Time      `json:"created_at" gorm:"index"`
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

// MarshalJSON implements the json.Marshaler interface for InstanceStatus
func (s InstanceStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
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

// MarshalJSON implements the json.Marshaler interface for Instance
func (i Instance) MarshalJSON() ([]byte, error) {
	type Alias Instance
	return json.Marshal(struct {
		ID uint `json:"id"`
		Alias
	}{
		ID:    i.Model.ID,
		Alias: Alias(i),
	})
}
