package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	InstanceCreatedAtField = "created_at"
	InstanceDeletedField   = "deleted"
)

type InstanceStatus int

const (
	// we need unknown to be the first status to avoid conflicts with the default value
	// Also allow us to search for all instances no matter their status
	InstanceStatusUnknown InstanceStatus = iota
	InstanceStatusPending
	InstanceStatusProvisioning
	InstanceStatusReady
	InstanceStatusTerminated
)

type Instance struct {
	gorm.Model
	JobID      uint           `json:"job_id" gorm:"not null;index"` // ID from the jobs table
	ProviderID ProviderID     `json:"provider_id" gorm:"not null"`
	Name       string         `json:"name" gorm:"not null;index"`
	PublicIP   string         `json:"public_ip" gorm:"not null;varchar(100)"`
	Region     string         `json:"region" gorm:"varchar(255)"`
	Size       string         `json:"size" gorm:"varchar(255)"`
	Image      string         `json:"image" gorm:"varchar(255)"`
	Tags       []string       `json:"tags" gorm:"type:json"`
	Status     InstanceStatus `json:"status" gorm:"index"`
	CreatedAt  time.Time      `json:"created_at" gorm:"index"`
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

func (s InstanceStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

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
