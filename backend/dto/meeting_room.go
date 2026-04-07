package dto

type MeetingRoomRequestDTO struct {
	Name     string `json:"name" binding:"required"`
	Location string `json:"location"`
	Capacity int    `json:"capacity"`
	Status   int    `json:"status"`
	Remark   string `json:"remark"`
}
