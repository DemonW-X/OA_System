package dto

type WorkflowBizTypeCreateRequestDTO struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
	Sort int    `json:"sort"`
}

type WorkflowNodeInputDTO struct {
	Sort          int    `json:"sort"`
	Name          string `json:"name" binding:"required"`
	ApproveType   string `json:"approve_type"`
	Approvers     string `json:"approvers"`
	Conditions    string `json:"conditions"`
	AllowSkip     bool   `json:"allow_skip"`
	AllowTransfer bool   `json:"allow_transfer"`
	ParentIDs     string `json:"parent_ids"` // JSON array of parent node sorts
}

type WorkflowTemplateRequestDTO struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	BizType     string                 `json:"biz_type"`
	Nodes       []WorkflowNodeInputDTO `json:"nodes"`
}
