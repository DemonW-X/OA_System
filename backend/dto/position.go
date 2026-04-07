package dto

type PositionRequestDTO struct {
	Name         string `json:"name" binding:"required"`
	DepartmentID int    `json:"department_id"`
	Remark       string `json:"remark"`
}

type PositionMenuPermissionRequestDTO struct {
	MenuIDs []int `json:"menu_ids"`
}

type PositionEmployeeRelationRequestDTO struct {
	DepartmentID int   `json:"department_id"`
	EmployeeIDs  []int `json:"employee_ids"`
}
