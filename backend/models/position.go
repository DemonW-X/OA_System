package models

import (
	"time"

	"gorm.io/gorm"
)

type Position struct {
	ID           int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name         string         `json:"name" gorm:"not null;unique"`
	DepartmentID int            `json:"department_id"`
	Department   Department     `json:"department" gorm:"foreignKey:DepartmentID;references:ID"`
	Remark       string         `json:"remark"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
