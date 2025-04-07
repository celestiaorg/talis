package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// VolumeDetail represents the details of a volume attached to an instance
type VolumeDetail struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Region     string `json:"region"`
	SizeGB     int    `json:"size_gb"`
	MountPoint string `json:"mount_point"`
}

// VolumeDetails is a slice of VolumeDetail
type VolumeDetails []VolumeDetail

// Value implements the driver.Valuer interface
func (vd VolumeDetails) Value() (driver.Value, error) {
	if vd == nil {
		return nil, nil
	}
	return json.Marshal(vd)
}

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
