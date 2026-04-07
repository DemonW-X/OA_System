package dto

type ResignationRequestDTO struct {
	EmployeeID int    `json:"employee_id" binding:"required"`
	ResignDate string `json:"resign_date" binding:"required"` // yyyy-MM-dd
	Reason     string `json:"reason"`
	Remark     string `json:"remark"`
}

type ResignationApproveRequestDTO struct {
	Action string `json:"action" binding:"required"` // approved / rejected
	Remark string `json:"remark"`
}
