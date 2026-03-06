package models

import (
	"time"

	"gorm.io/gorm"
)

type Position struct {
	ID        int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name      string         `json:"name" gorm:"not null"`
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	Remark    string         `json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 兼容前端现有展示字段（非持久化）
	DepartmentID int         `json:"department_id" gorm:"-"`
	Department   *Department `json:"department,omitempty" gorm:"-"`
}
