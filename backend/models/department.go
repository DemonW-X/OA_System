package models

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	ID        int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name      string         `json:"name" gorm:"not null;unique"`
	ParentID  *int           `json:"parent_id" gorm:"index"`
	Parent    *Department    `json:"parent,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Level     int            `json:"level" gorm:"default:1"`
	Remark    string         `json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
