package models

import (
	"time"

	"gorm.io/gorm"
)

// Resignation 离职管理
// 创建/编辑离职记录时，需联动将员工状态改为离职（status=0）
// 删除最后一条离职记录时，可回滚为在职（status=1）
type Resignation struct {
	ID            int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	EmployeeID    int            `json:"employee_id"`
	Employee      *Employee      `json:"employee,omitempty" gorm:"foreignKey:EmployeeID;references:ID"`
	ResignDate    time.Time      `json:"resign_date"`
	Reason        string         `json:"reason"`
	Remark        string         `json:"remark"`
	ApproveStatus string         `json:"approve_status" gorm:"default:'draft'"` // draft/pending/approved/rejected
	ApprovedBy    string         `json:"approved_by"`
	ApprovedAt    *time.Time     `json:"approved_at"`
	ApproveRemark string         `json:"approve_remark"`
	WorkflowLogs  string         `json:"workflow_logs" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}
