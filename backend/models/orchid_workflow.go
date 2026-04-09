package models

import "time"

type OrchidWorkflowDefinition struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"not null"`
	BizType     string    `json:"biz_type" gorm:"not null;index"`
	Description string    `json:"description"`
	DagJSON     string    `json:"dag_json" gorm:"type:longtext"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OrchidWorkflowInstance struct {
	ID           int        `json:"id" gorm:"primaryKey;autoIncrement"`
	DefinitionID int        `json:"definition_id" gorm:"index"`
	BizType      string     `json:"biz_type" gorm:"index"`
	BizID        int        `json:"biz_id" gorm:"index"`
	Status       string     `json:"status" gorm:"default:'pending'"` // pending/approved/rejected
	CurrentNode  string     `json:"current_node"`
	StartedBy    string     `json:"started_by"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type OrchidWorkflowHistory struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	InstanceID int       `json:"instance_id" gorm:"index"`
	NodeKey    string    `json:"node_key"`
	Action     string    `json:"action"` // submit/approved/rejected/transfer/skip
	Operator   string    `json:"operator"`
	Remark     string    `json:"remark"`
	CreatedAt  time.Time `json:"created_at"`
}

type OrchidWorkflowTask struct {
	ID         int        `json:"id" gorm:"primaryKey;autoIncrement"`
	InstanceID int        `json:"instance_id" gorm:"index"`
	NodeKey    string     `json:"node_key" gorm:"index"`
	AssigneeID int        `json:"assignee_id" gorm:"index"`
	Status     string     `json:"status" gorm:"default:'open'"`       // open/done/transferred/skipped
	TaskType   string     `json:"task_type" gorm:"default:'approve'"` // approve/read
	ReadAt     *time.Time `json:"read_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
