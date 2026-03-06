package models

import (
	"time"

	"gorm.io/gorm"
)

// DepartmentPosition 部门-职位关系（解耦职位与部门）
type DepartmentPosition struct {
	ID           int            `json:"id" gorm:"primaryKey;autoIncrement"`
	DepartmentID int            `json:"department_id" gorm:"index;uniqueIndex:uk_dept_position"`
	Department   Department     `json:"department" gorm:"foreignKey:DepartmentID;references:ID"`
	PositionID   int            `json:"position_id" gorm:"index;uniqueIndex:uk_dept_position"`
	Position     Position       `json:"position" gorm:"foreignKey:PositionID;references:ID"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
