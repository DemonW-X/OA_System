package models

import "time"

type EmployeeMenuPermission struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	EmployeeID int       `json:"employee_id" gorm:"not null;uniqueIndex:uk_employee_menu,priority:1"`
	MenuID     int       `json:"menu_id" gorm:"not null;uniqueIndex:uk_employee_menu,priority:2"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
