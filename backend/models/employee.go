package models

import (
	"time"

	"gorm.io/gorm"
)

type Employee struct {
	ID             int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name           string         `json:"name" gorm:"not null"`
	Phone          string         `json:"phone"`
	Email          string         `json:"email"`
	DepartmentID   int            `json:"department_id"`
	Department     Department     `json:"department" gorm:"foreignKey:DepartmentID;references:ID"`
	PositionID     int            `json:"position_id"`
	PositionInfo   Position       `json:"position_info" gorm:"foreignKey:PositionID;references:ID"`
	Status         int            `json:"status" gorm:"default:1"`
	UserID         int            `json:"user_id" gorm:"default:0"`
	ApproveStatus  string         `json:"approve_status" gorm:"default:'pending'"` // pending/approved/rejected
	ApprovedBy     string         `json:"approved_by"`
	ApprovedAt     *time.Time     `json:"approved_at"`
	ApproveRemark  string         `json:"approve_remark"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}
