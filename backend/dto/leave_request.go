package dto

type LeaveRequestCreateDTO struct {
	EmployeeID int     `json:"employee_id" binding:"required"`
	Type       string  `json:"type" binding:"required"`
	StartDate  string  `json:"start_date" binding:"required"`
	EndDate    string  `json:"end_date" binding:"required"`
	Days       float64 `json:"days" binding:"required"`
	Reason     string  `json:"reason"`
}

type LeaveRequestApproveDTO struct {
	Status       string `json:"status" binding:"required"` // approved / rejected
	RejectReason string `json:"reject_reason"`
}

type LeaveRequestSubmitDTO struct {
	Remark string `json:"remark"`
}

type LeaveFlowLogEntryDTO struct {
	Time     string `json:"time"`
	Node     string `json:"node"`
	Action   string `json:"action"`
	Operator string `json:"operator"`
	Remark   string `json:"remark,omitempty"`
}
