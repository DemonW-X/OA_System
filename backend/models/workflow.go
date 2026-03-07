package models

import "time"

type BizType struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Code      string    `json:"code" gorm:"not null"`  // 业务编码，如 leave
	Name      string    `json:"name" gorm:"not null"`          // 业务名称，如 请假审批
	Sort      int       `json:"sort" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WorkflowTemplate struct {
	ID          int             `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string          `json:"name" gorm:"not null"`
	Description string          `json:"description"`
	BizType     string          `json:"biz_type"` // leave/event/other
	Nodes       []WorkflowNode  `json:"nodes" gorm:"foreignKey:TemplateID;references:ID"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type WorkflowNode struct {
	ID            int       `json:"id" gorm:"primaryKey;autoIncrement"`
	TemplateID    int       `json:"template_id"`
	Sort          int       `json:"sort" gorm:"default:0"`
	Name          string    `json:"name" gorm:"not null"`
	ApproveType   string    `json:"approve_type" gorm:"default:'or'"`
	Approvers     string    `json:"approvers" gorm:"type:text"`
	Conditions    string    `json:"conditions" gorm:"type:text"`
	AllowSkip     bool      `json:"allow_skip" gorm:"default:false"`
	AllowTransfer bool      `json:"allow_transfer" gorm:"default:false"`
	ParentIDs     string    `json:"parent_ids" gorm:"type:text;default:'[]'"` // JSON数组，存父节点sort值
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
