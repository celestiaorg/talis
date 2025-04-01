package models

import (
	"time"

	"gorm.io/gorm"
)

// Project represents a collection of related tasks and instances
type Project struct {
	gorm.Model
	OwnerID     uint      `json:"-" gorm:"not null; index"`
	Name        string    `json:"name" gorm:"not null; index"`
	Description string    `json:"description" gorm:"type:text"`
	Config      string    `json:"config" gorm:"type:text"`
	Tasks       []Task    `json:"tasks" gorm:"foreignKey:ProjectID"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}
