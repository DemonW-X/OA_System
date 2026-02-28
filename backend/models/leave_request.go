package models

import (
	"time"

	"gorm.io/gorm"
)

type LeaveRequest struct {
	ID           int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	EmployeeID   int            `json:"employee_id"`
	Employee     *Employee      `json:"employee,omitempty" gorm:"foreignKey:EmployeeID;references:ID"`
	Type         string         `json:"type" gorm:"default:'annual'"` // annual/sick/personal/other
	StartDate    time.Time      `json:"start_date"`
	EndDate      time.Time      `json:"end_date"`
	Days         float64        `json:"days"`
	Reason       string         `json:"reason"`
	Status       string         `json:"status" gorm:"default:'draft'"` // draft/pending/approved/rejected
	SubmittedBy  string         `json:"submitted_by"`
	SubmittedAt  *time.Time     `json:"submitted_at"`
	ApprovedBy   string         `json:"approved_by"`
	ApprovedAt   *time.Time     `json:"approved_at"`
	RejectReason string         `json:"reject_reason"`
	WorkflowLogs string         `json:"workflow_logs" gorm:"type:text"` // 流程记录JSON
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
