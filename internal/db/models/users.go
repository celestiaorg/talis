package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `json:"username" gorm:"not null;unique"`
	PublicSshKey string `json:"public_ssh_key" gorm:""`
}
