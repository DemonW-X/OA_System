package dto

type EmployeeRequestDTO struct {
	Name         string `json:"name" binding:"required"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	DepartmentID int    `json:"department_id"`
	PositionID   int    `json:"position_id"`
	Status       int    `json:"status"`
}

type EmployeeApproveRequestDTO struct {
	Action string `json:"action" binding:"required"` // approved / rejected
	Remark string `json:"remark"`
}

type EmployeeNameItemDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
