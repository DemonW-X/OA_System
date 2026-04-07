package dto

type EventBookingRequestDTO struct {
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description"`
	Type          string `json:"type" binding:"required"`
	StartTime     string `json:"start_time" binding:"required"`
	EndTime       string `json:"end_time" binding:"required"`
	MeetingRoomID int    `json:"meeting_room_id"`
	Participants  string `json:"participants"`
}

type EventBookingSubmitRequestDTO struct {
	Remark string `json:"remark"`
}

type EventBookingApproveRequestDTO struct {
	Status       string `json:"status" binding:"required"` // approved / rejected
	RejectReason string `json:"reject_reason"`
}
