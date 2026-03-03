package models

import (
	"time"

	"gorm.io/gorm"
)

type Notice struct {
	ID            int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Title         string         `json:"title" gorm:"not null"`
	Content       string         `json:"content" gorm:"type:longtext"`
	Author        string         `json:"author"`
	Status        int            `json:"status" gorm:"default:1"`
	Attachments   string         `json:"attachments" gorm:"type:text"`
	DepartmentID  int            `json:"department_id" gorm:"default:0"` // 0 表示全部部门
	Department    *Department    `json:"department,omitempty" gorm:"foreignKey:DepartmentID;references:ID"`
	ApproveStatus string         `json:"approve_status" gorm:"default:'draft'"` // draft/pending/approved/rejected
	ApprovedBy    string         `json:"approved_by"`
	ApprovedAt    *time.Time     `json:"approved_at"`
	ApproveRemark string         `json:"approve_remark"`
	WorkflowLogs  string         `json:"workflow_logs" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}
