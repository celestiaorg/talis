package models

import (
	"time"
)

const (
	InstanceCreatedAtField = "created_at"
	InstanceDeletedField   = "deleted"
)

type Instance struct {
	ID        string    `json:"id" gorm:"primaryKey;varchar(50)"`
	JobID     string    `json:"job_id" gorm:"not null;varchar(50);index"`
	PublicIP  string    `json:"public_ip" gorm:"not null;varchar(100)"`
	Deleted   bool      `json:"-" gorm:"index"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at"`
}
