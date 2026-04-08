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
	OnboardDate    *time.Time     `json:"onboard_date"`
	OnboardType    string         `json:"onboard_type" gorm:"default:'new'"`
	ProbationDays  int            `json:"probation_days" gorm:"default:90"`
	ProbationEnd   *time.Time     `json:"probation_end"`
	IDCard         string         `json:"id_card"`
	NativePlace    string         `json:"native_place"`
	Address        string         `json:"address"`
	EmergencyName  string         `json:"emergency_name"`
	EmergencyPhone string         `json:"emergency_phone"`
	Education      string         `json:"education"`
	School         string         `json:"school"`
	Major          string         `json:"major"`
	WorkYears      int            `json:"work_years" gorm:"default:0"`
	Remark         string         `json:"remark" gorm:"type:text"`
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
