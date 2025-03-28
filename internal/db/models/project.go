package models

import (
	"time"

	"gorm.io/gorm"
)

// Project represents a project in the system
type Project struct {
	gorm.Model
	OwnerID     uint      `json:"-" gorm:"uniqueIndex:idx_owner_project_name"`
	Name        string    `json:"name" gorm:"uniqueIndex:idx_owner_project_name"`
	Description string    `json:"description" gorm:"type:text"`
	Config      string    `json:"config" gorm:"type:json"`
	Tasks       []Task    `json:"tasks" gorm:"foreignKey:ProjectID"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}
