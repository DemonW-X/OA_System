package models

import (
	"time"

	"gorm.io/gorm"
)

type Menu struct {
	ID       int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name     string         `json:"name" gorm:"not null"`
	Icon     string         `json:"icon" gorm:"default:''"`
	Path     string         `json:"path" gorm:"default:''"`
	SortCode int            `json:"sort_code" gorm:"default:0"`
	ParentID int            `json:"parent_id" gorm:"default:0"`
	Visible  bool           `json:"visible" gorm:"default:true"`
	Remark   string         `json:"remark"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
