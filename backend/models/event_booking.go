package models

import "time"

type EventBooking struct {
	ID            int          `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Title         string       `json:"title" gorm:"not null"`
	Description   string       `json:"description"`
	Type          string       `json:"type" gorm:"default:'other'"`
	StartTime     time.Time    `json:"start_time"`
	EndTime       time.Time    `json:"end_time"`
	MeetingRoomID int          `json:"meeting_room_id" gorm:"default:0"`
	MeetingRoom   *MeetingRoom `json:"meeting_room,omitempty" gorm:"foreignKey:MeetingRoomID;references:ID"`
	Participants  string       `json:"participants" gorm:"type:text"` // JSON数组，存员工ID列表
	Status        string       `json:"status" gorm:"default:'draft'"` // draft/pending/approved/rejected
	SubmittedBy   string       `json:"submitted_by"`
	SubmittedAt   *time.Time   `json:"submitted_at"`
	ApprovedBy    string       `json:"approved_by"`
	ApprovedAt    *time.Time   `json:"approved_at"`
	RejectReason  string       `json:"reject_reason"`
	WorkflowLogs  string       `json:"workflow_logs" gorm:"type:text"` // 流程记录JSON
	CreatedBy     string       `json:"created_by"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}
