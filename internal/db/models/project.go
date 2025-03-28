package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	ID          uint      `json:"-" gorm:"primaryKey"`
	OwnerID     uint      `json:"-" gorm:"not null; index"`
	Name        string    `json:"name" gorm:"not null; index"`
	Description string    `json:"description" gorm:"type:text"`
	Config      string    `json:"config" gorm:"type:text"`
	Tasks       []Task    `json:"tasks" gorm:"foreignKey:ProjectID"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}
