package handlers

import (
	"encoding/json"
	"net/http"
	"oa-system/database"
	"oa-system/models"
	"time"

	"github.com/gin-gonic/gin"
)

type LeaveRequestCreate struct {
	EmployeeID int     `json:"employee_id" binding:"required"`
	Type       string  `json:"type" binding:"required"`
	StartDate  string  `json:"start_date" binding:"required"`
	EndDate    string  `json:"end_date" binding:"required"`
	Days       float64 `json:"days" binding:"required"`
	Reason     string  `json:"reason"`
}

type LeaveRequestApprove struct {
	Status       string `json:"status" binding:"required"` // approved/rejected
	RejectReason string `json:"reject_reason"`
}

type LeaveRequestSubmit struct {
	Remark string `json:"remark"`
}

const dateLayout = "2006-01-02"

type flowLogEntry struct {
	Time     string `json:"time"`
	Node     string `json:"node"`
	Action   string `json:"action"`
	Operator string `json:"operator"`
	Remark   string `json:"remark,omitempty"`
}

func appendFlowLog(raw string, entry flowLogEntry) string {
	logs := []flowLogEntry{}
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &logs)
	}
	logs = append(logs, entry)
	b, _ := json.Marshal(logs)
	return string(b)
}

func currentOperator(c *gin.Context) string {
	op := c.GetString("realName")
	if op == "" {
		op = c.GetString("username")
	}
	return op
}

func GetLeaveRequests(c *gin.Context) {
	var list []models.LeaveRequest
	query := database.DB.Model(&models.LeaveRequest{}).Preload("Employee")
	if empID := c.Query("employee_id"); empID != "" {
		query = query.Where("employee_id = ?", empID)
	}
	if empName := c.Query("employee_name"); empName != "" {
		query = query.Joins("JOIN employees ON employees.id = leave_requests.employee_id").Where("employees.name LIKE ?", "%"+empName+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if t := c.Query("type"); t != "" {
		query = query.Where("type = ?", t)
	}
	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
	writeLog(c, "请假管理", "查询", "查询请假列表")
}

func GetLeaveRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var leave models.LeaveRequest
	if err := database.DB.Preload("Employee").First(&leave, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "请假记录不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": leave})
}

func CreateLeaveRequest(c *gin.Context) {
	var req LeaveRequestCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	startDate, err := time.ParseInLocation(dateLayout, req.StartDate, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "开始日期格式错误，请使用 yyyy-MM-dd"})
		return
	}
	endDate, err := time.ParseInLocation(dateLayout, req.EndDate, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "结束日期格式错误，请使用 yyyy-MM-dd"})
		return
	}
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "结束日期不能早于开始日期"})
		return
	}
	leave := models.LeaveRequest{
		EmployeeID:   req.EmployeeID,
		Type:         req.Type,
		StartDate:    startDate,
		EndDate:      endDate,
		Days:         req.Days,
		Reason:       req.Reason,
		Status:       "draft",
		WorkflowLogs: "[]",
	}
	if err := database.DB.Create(&leave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": leave})
	writeLog(c, "请假管理", "新增", "新增请假申请")
}

func UpdateLeaveRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var leave models.LeaveRequest
	if err := database.DB.First(&leave, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "请假记录不存在"})
		return
	}
	if leave.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅草稿状态可修改"})
		return
	}
	var req LeaveRequestCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	startDate, err := time.ParseInLocation(dateLayout, req.StartDate, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "开始日期格式错误"})
		return
	}
	endDate, err := time.ParseInLocation(dateLayout, req.EndDate, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "结束日期格式错误"})
		return
	}
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "结束日期不能早于开始日期"})
		return
	}
	leave.EmployeeID = req.EmployeeID
	leave.Type = req.Type
	leave.StartDate = startDate
	leave.EndDate = endDate
	leave.Days = req.Days
	leave.Reason = req.Reason
	if err := database.DB.Save(&leave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": leave})
	writeLog(c, "请假管理", "修改", "修改请假申请")
}

func SubmitLeaveRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var leave models.LeaveRequest
	if err := database.DB.First(&leave, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "请假记录不存在"})
		return
	}
	if leave.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅草稿状态可提交"})
		return
	}
	var req LeaveRequestSubmit
	_ = c.ShouldBindJSON(&req)
	op := currentOperator(c)
	ret, err := submitApprovalFlow("leave_request", leave.ID, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "流程实例启动失败: " + err.Error()})
		return
	}
	leave.Status = ret.Status
	leave.SubmittedBy = ret.SubmittedBy
	leave.SubmittedAt = ret.SubmittedAt
	leave.ApprovedBy = ret.ApprovedBy
	leave.ApprovedAt = ret.ApprovedAt
	leave.WorkflowLogs = ret.WorkflowLogs
	if err := database.DB.Save(&leave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": leave})
	writeLog(c, "请假管理", "提交", "提交请假申请")
}

func ApproveLeaveRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var leave models.LeaveRequest
	if err := database.DB.First(&leave, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "请假记录不存在"})
		return
	}
	if leave.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅待审核状态可审批"})
		return
	}
	var req LeaveRequestApprove
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if req.Status != "approved" && req.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "状态只能为 approved 或 rejected"})
		return
	}
	op := currentOperator(c)
	ret, err := approveApprovalFlow("leave_request", leave.ID, op, req.Status, req.RejectReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "审批失败"})
		return
	}
	leave.Status = ret.Status
	leave.ApprovedBy = ret.ApprovedBy
	leave.ApprovedAt = ret.ApprovedAt
	leave.RejectReason = ret.ApproveRemark
	action := "审批通过"
	if req.Status == "rejected" {
		action = "审批拒绝"
	}
	leave.WorkflowLogs = ret.WorkflowLogs
	if err := database.DB.Save(&leave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "审批失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": leave})
	writeLog(c, "请假管理", "审批", action)
}

func DeleteLeaveRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var leave models.LeaveRequest
	if err := database.DB.First(&leave, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "请假记录不存在"})
		return
	}
	if leave.Status == "approved" || leave.Status == "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "待审核/已通过记录不能删除"})
		return
	}
	if err := database.DB.Delete(&leave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "请假管理", "删除", "删除请假申请")
}
