package dto

type NoticeRequestDTO struct {
	Title        string `json:"title" binding:"required"`
	Content      string `json:"content"`
	Status       int    `json:"status"`
	Attachments  string `json:"attachments"`
	DepartmentID int    `json:"department_id"`
}

type NoticeSubmitRequestDTO struct {
	Remark string `json:"remark"`
}

type NoticeApproveRequestDTO struct {
	Action string `json:"action" binding:"required"` // approved / rejected
	Remark string `json:"remark"`
}
