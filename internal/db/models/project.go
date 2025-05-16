package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Project represents a collection of related tasks and instances
type Project struct {
	gorm.Model
	OwnerID     uint      `json:"-" gorm:"not null; index"`
	Name        string    `json:"name" gorm:"not null; index; unique"`
	Description string    `json:"description" gorm:"type:text"`
	Config      string    `json:"config" gorm:"type:text"`
	Tasks       []Task    `json:"tasks" gorm:"foreignKey:ProjectID"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}

// MarshalJSON implements the json.Marshaler interface for Project
func (p Project) MarshalJSON() ([]byte, error) {
	type Alias Project // Create an alias to avoid infinite recursion
	return json.Marshal(Alias(p))
}
