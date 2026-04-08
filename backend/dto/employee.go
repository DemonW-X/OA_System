package dto

type EmployeeRequestDTO struct {
	Name           string `json:"name" binding:"required"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	OnboardDate    string `json:"onboard_date"`
	OnboardType    string `json:"onboard_type"`
	ProbationDays  int    `json:"probation_days"`
	ProbationEnd   string `json:"probation_end"`
	IDCard         string `json:"id_card"`
	NativePlace    string `json:"native_place"`
	Address        string `json:"address"`
	EmergencyName  string `json:"emergency_name"`
	EmergencyPhone string `json:"emergency_phone"`
	Education      string `json:"education"`
	School         string `json:"school"`
	Major          string `json:"major"`
	WorkYears      int    `json:"work_years"`
	Remark         string `json:"remark"`
	DepartmentID   int    `json:"department_id"`
	PositionID     int    `json:"position_id"`
	Status         int    `json:"status"`
}

type EmployeeApproveRequestDTO struct {
	Action string `json:"action" binding:"required"` // approved / rejected
	Remark string `json:"remark"`
}

type EmployeeNameItemDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
