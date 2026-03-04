package models

import "time"

type CalendarEvent struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Type        string    `json:"type" gorm:"default:'other'"` // meeting/activity/other
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
