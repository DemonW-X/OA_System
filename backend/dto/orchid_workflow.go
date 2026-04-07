package dto

type OrchidWorkflowDefinitionRequestDTO struct {
	Name        string `json:"name" binding:"required"`
	BizType     string `json:"biz_type" binding:"required"`
	Description string `json:"description"`
	DagJSON     string `json:"dag_json" binding:"required"`
	IsActive    bool   `json:"is_active"`
}

type OrchidTransferOrSkipRequestDTO struct {
	FromUserID int    `json:"from_user_id"`
	ToUserID   int    `json:"to_user_id"`
	Remark     string `json:"remark"`
}

type OrchidPendingApprovalItemDTO struct {
	TaskID     int    `json:"task_id"`
	BizType    string `json:"biz_type"`
	BizID      int    `json:"biz_id"`
	NodeKey    string `json:"node_key"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	DetailPath string `json:"detail_path"`
	InstanceID int    `json:"instance_id"`
}
