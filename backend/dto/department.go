package dto

type DepartmentRequestDTO struct {
	Name     string `json:"name" binding:"required"`
	ParentID *int   `json:"parent_id"`
	Remark   string `json:"remark"`
}
