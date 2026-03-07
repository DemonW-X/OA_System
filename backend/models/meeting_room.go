package models

import "time"

type MeetingRoom struct {
	ID       int    `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name     string `json:"name" gorm:"not null"`
	Location string `json:"location"`
	Capacity int    `json:"capacity" gorm:"default:0"`
	Status   int    `json:"status" gorm:"default:1"` // 1:可用 0:停用
	Remark   string `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
