package handlers

import (
	"errors"
	"oa-system/database"
	"oa-system/models"
	"time"
)

type ApprovalSubmitResult struct {
	Status        string
	SubmittedBy   string
	SubmittedAt   *time.Time
	ApprovedBy    string
	ApprovedAt    *time.Time
	ApproveRemark string
	WorkflowLogs  string
}

type ApprovalDecisionResult struct {
	Status        string
	ApprovedBy    string
	ApprovedAt    *time.Time
	ApproveRemark string
	WorkflowLogs  string
}

// submitApprovalFlow 提交业务处理
func submitApprovalFlow(bizType string, bizID int, operator string) (*ApprovalSubmitResult, error) {
	now := time.Now()
	res := &ApprovalSubmitResult{
		SubmittedBy: operator,
		SubmittedAt: &now,
	}

	if hasOrchidWorkflowForBiz(bizType) {
		ins, err := startOrchidInstance(bizType, bizID, operator)
		if err != nil {
			return nil, err
		}
		if ins != nil && ins.Status == "approved" {
			res.Status = "approved"
			res.ApprovedBy = operator
			res.ApprovedAt = &now
			res.ApproveRemark = "流程到达结束节点，自动通过"
		} else {
			res.Status = "pending"
		}
	} else {
		res.Status = "approved"
		res.ApprovedBy = operator
		res.ApprovedAt = &now
		res.ApproveRemark = "无流程定义，自动通过"
	}

	res.WorkflowLogs = buildBizWorkflowLogs(bizType, bizID)
	return res, nil
}

// approveApprovalFlow 处理审批业务
func approveApprovalFlow(bizType string, bizID int, operator, action, remark string) (*ApprovalDecisionResult, error) {
	if action != "approved" && action != "rejected" {
		return nil, errors.New("action 只能为 approved 或 rejected")
	}
	finalStatus, err := approveOrRejectInstance(bizType, bizID, operator, action, remark)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &ApprovalDecisionResult{
		Status:        finalStatus,
		ApprovedBy:    operator,
		ApprovedAt:    &now,
		ApproveRemark: remark,
		WorkflowLogs:  buildBizWorkflowLogs(bizType, bizID),
	}, nil
}

// withdrawApprovalFlow 执行撤回处理
func withdrawApprovalFlow(bizType string, bizID int, operator string) error {
	ins, err := getInstanceByBiz(bizType, bizID)
	if err != nil {
		return nil // 没有流程实例视为可撤回
	}

	var doneCount int64
	database.DB.Model(&models.OrchidWorkflowHistory{}).
		Where("instance_id = ? AND action IN ?", ins.ID, []string{"approved", "approved_partial", "rejected"}).
		Count(&doneCount)
	if doneCount > 0 {
		return errors.New("已有节点审批，无法撤回")
	}

	if err := database.DB.Model(&models.OrchidWorkflowTask{}).
		Where("instance_id = ? AND status = 'open'", ins.ID).
		Update("status", "withdrawn").Error; err != nil {
		return err
	}

	_ = database.DB.Create(&models.OrchidWorkflowHistory{
		InstanceID: ins.ID,
		NodeKey:    "withdraw",
		Action:     "withdraw",
		Operator:   operator,
		Remark:     "申请人撤回",
	}).Error

	now := time.Now()
	ins.Status = "withdrawn"
	ins.FinishedAt = &now
	return database.DB.Save(ins).Error
}
