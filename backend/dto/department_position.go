package dto

import (
	"oa-system/models"
	"time"
)

type DepartmentPositionRequestDTO struct {
	DepartmentID int `json:"department_id" binding:"required"`
	PositionID   int `json:"position_id" binding:"required"`
}

// PositionLiteDTO is a trimmed position shape for list display.
type PositionLiteDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DepartmentPositionListItemDTO controls response fields for list rows.
type DepartmentPositionListItemDTO struct {
	ID           int               `json:"id"`
	DepartmentID int               `json:"department_id"`
	Department   models.Department `json:"department"`
	PositionID   int               `json:"position_id"`
	Position     PositionLiteDTO   `json:"position"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type DepartmentPositionListDataDTO struct {
	List     []DepartmentPositionListItemDTO `json:"list"`
	Total    int64                           `json:"total"`
	Page     int                             `json:"page"`
	PageSize int                             `json:"page_size"`
}

type DepartmentPositionListResponseDTO struct {
	Code int                           `json:"code"`
	Data DepartmentPositionListDataDTO `json:"data"`
}
