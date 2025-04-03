package models

import (
	"encoding/json"
	"fmt"
	"time"

	"database/sql/driver"

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
	ProviderID    string         `json:"provider_id" gorm:"not null"`
	Name          string         `json:"name" gorm:"not null;index"`
	PublicIP      string         `json:"public_ip" gorm:"varchar(100)"`
	Region        string         `json:"region" gorm:"varchar(255)"`
	Size          string         `json:"size" gorm:"varchar(255)"`
	Image         string         `json:"image" gorm:"varchar(255)"`
	Tags          pq.StringArray `json:"tags" gorm:"type:text[]"`
	Status        InstanceStatus `json:"status" gorm:"index"`
	Volumes       pq.StringArray `json:"volumes" gorm:"type:text[]"`
	VolumeDetails VolumeDetails  `json:"volume_details" gorm:"type:jsonb"`
	CreatedAt     time.Time      `json:"created_at" gorm:"index"`
}

// VolumeDetail represents the details of a volume attached to an instance
type VolumeDetail struct {
	ID         string `json:"id"`          // Volume ID
	Name       string `json:"name"`        // Volume name
	SizeGB     int    `json:"size_gb"`     // Size in gigabytes
	Region     string `json:"region"`      // Region where the volume is created
	MountPoint string `json:"mount_point"` // Where the volume is mounted
}

// VolumeDetails type for handling JSONB in database
type VolumeDetails []VolumeDetail

// Scan implements the sql.Scanner interface
func (vd *VolumeDetails) Scan(value interface{}) error {
	if value == nil {
		*vd = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	// Try first as array
	var temp []VolumeDetail
	err := json.Unmarshal(bytes, &temp)
	if err == nil {
		*vd = temp
		return nil
	}

	// If array fails, try as single object
	var singleVolume VolumeDetail
	err = json.Unmarshal(bytes, &singleVolume)
	if err != nil {
		return fmt.Errorf("failed to unmarshal as array or object: %w", err)
	}

	*vd = []VolumeDetail{singleVolume}
	return nil
}

// Value implements the driver.Valuer interface
func (vd VolumeDetails) Value() (driver.Value, error) {
	if vd == nil {
		return nil, nil
	}

	return json.Marshal(vd)
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
