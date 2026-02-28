package models

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	ID        int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name      string         `json:"name" gorm:"not null;unique"`
	Remark    string         `json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
