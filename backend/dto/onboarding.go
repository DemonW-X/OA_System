package dto

type OnboardingRequestDTO struct {
	EmployeeName   string `json:"employee_name" binding:"required"`
	OnboardDate    string `json:"onboard_date" binding:"required"` // yyyy-MM-dd
	OnboardType    string `json:"onboard_type"`                    // new / rehire / transfer
	ProbationDays  int    `json:"probation_days"`
	IDCard         string `json:"id_card" binding:"required"`
	Phone          string `json:"phone" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	NativePlace    string `json:"native_place"`
	Address        string `json:"address"`
	EmergencyName  string `json:"emergency_name"`
	EmergencyPhone string `json:"emergency_phone"`
	Education      string `json:"education"`
	School         string `json:"school"`
	Major          string `json:"major"`
	WorkYears      int    `json:"work_years"`
	DepartmentID   int    `json:"department_id"`
	PositionID     int    `json:"position_id"`
	Remark         string `json:"remark"`
}

type OnboardingApproveRequestDTO struct {
	Action string `json:"action" binding:"required"` // approved / rejected
	Remark string `json:"remark"`
}
