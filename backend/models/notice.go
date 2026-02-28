package models

import (
	"time"

	"gorm.io/gorm"
)

type Notice struct {
	ID           int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Title        string         `json:"title" gorm:"not null"`
	Content      string         `json:"content" gorm:"type:longtext"`
	Author       string         `json:"author"`
	Status       int            `json:"status" gorm:"default:1"`
	Attachments  string         `json:"attachments" gorm:"type:text"`
	DepartmentID int            `json:"department_id" gorm:"default:0"` // 0 表示全部部门
	Department   *Department    `json:"department,omitempty" gorm:"foreignKey:DepartmentID;references:ID"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
