package handlers

import (
	"fmt"
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/models"
	"oa-system/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kyodo-tech/orchid"
)

type OrchidWorkflowDefinitionReq = dto.OrchidWorkflowDefinitionRequestDTO
type TransferOrSkipReq = dto.OrchidTransferOrSkipRequestDTO

// validateOrchidDag 校验输入或状态
func validateOrchidDag(name, dagJSON string) error {
	wf := orchid.NewWorkflow(name)
	return wf.Import([]byte(dagJSON))
}

// GetOrchidWorkflowDefinition 获取数据
func GetOrchidWorkflowDefinition(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var def models.OrchidWorkflowDefinition
	if err := database.DB.First(&def, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "流程定义不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": def})
}

// GetOrchidWorkflowDefinitions 获取数据
func GetOrchidWorkflowDefinitions(c *gin.Context) {
	var list []models.OrchidWorkflowDefinition
	query := database.DB.Model(&models.OrchidWorkflowDefinition{})
	if bizType := c.Query("biz_type"); bizType != "" {
		query = query.Where("biz_type = ?", bizType)
	}
	query.Order("id desc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

// CreateOrchidWorkflowDefinition 创建数据
func CreateOrchidWorkflowDefinition(c *gin.Context) {
	var req OrchidWorkflowDefinitionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if err := validateOrchidDag(req.Name, req.DagJSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "DAG JSON不合法: " + err.Error()})
		return
	}
	def := models.OrchidWorkflowDefinition{
		Name:        req.Name,
		BizType:     req.BizType,
		Description: req.Description,
		DagJSON:     req.DagJSON,
		IsActive:    req.IsActive,
	}
	if !req.IsActive {
		def.IsActive = false
	}
	if err := database.DB.Create(&def).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	services.RefreshOrchidWorkflowCacheByBizType(def.BizType)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": def})
	writeLog(c, "流程引擎", "新增", "新增Orchid流程定义："+req.Name)
}

// UpdateOrchidWorkflowDefinition 更新数据
func UpdateOrchidWorkflowDefinition(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var def models.OrchidWorkflowDefinition
	if err := database.DB.First(&def, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "流程定义不存在"})
		return
	}
	oldBizType := def.BizType
	var req OrchidWorkflowDefinitionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if err := validateOrchidDag(req.Name, req.DagJSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "DAG JSON不合法: " + err.Error()})
		return
	}
	def.Name = req.Name
	def.BizType = req.BizType
	def.Description = req.Description
	def.DagJSON = req.DagJSON
	def.IsActive = req.IsActive
	if err := database.DB.Save(&def).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败"})
		return
	}
	services.RefreshOrchidWorkflowCacheByBizType(oldBizType)
	services.RefreshOrchidWorkflowCacheByBizType(def.BizType)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": def})
	writeLog(c, "流程引擎", "修改", "修改Orchid流程定义："+def.Name)
}

// DeleteOrchidWorkflowDefinition 删除数据
func DeleteOrchidWorkflowDefinition(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var def models.OrchidWorkflowDefinition
	if err := database.DB.First(&def, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "流程定义不存在"})
		return
	}
	if err := database.DB.Delete(&def).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	services.RefreshOrchidWorkflowCacheByBizType(def.BizType)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "流程引擎", "删除", "删除Orchid流程定义："+def.Name)
}

// GetOrchidWorkflowHistories 获取数据
func GetOrchidWorkflowHistories(c *gin.Context) {
	bizType := c.Query("biz_type")
	bizID := c.Query("biz_id")
	if bizType == "" || bizID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "biz_type 和 biz_id 必填"})
		return
	}
	bizIDInt := atoiDefault(bizID, 0)

	// 查该业务对象的所有流程实例（按时间正序）
	var instances []models.OrchidWorkflowInstance
	database.DB.Where("biz_type = ? AND biz_id = ?", bizType, bizIDInt).Order("id asc").Find(&instances)

	if len(instances) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
			"instance":  nil,
			"instances": []interface{}{},
			"histories": []models.OrchidWorkflowHistory{},
			"tasks":     []models.OrchidWorkflowTask{},
		}})
		return
	}

	// 最新实例作为当前实例
	latest := instances[len(instances)-1]

	// 收集所有实例ID
	insIDs := make([]int, len(instances))
	for i, ins := range instances {
		insIDs[i] = ins.ID
	}

	// 查所有历史记录
	var allHistories []models.OrchidWorkflowHistory
	database.DB.Where("instance_id IN ?", insIDs).Order("id asc").Find(&allHistories)

	// 查当前实例的待办任务
	var tasks []models.OrchidWorkflowTask
	database.DB.Where("instance_id = ? AND status = 'open'", latest.ID).Find(&tasks)

	// 按实例分组历史，方便前端分轮次展示
	type instanceWithHistory struct {
		Instance  models.OrchidWorkflowInstance  `json:"instance"`
		Histories []models.OrchidWorkflowHistory `json:"histories"`
	}
	rounds := make([]instanceWithHistory, 0, len(instances))
	histMap := map[int][]models.OrchidWorkflowHistory{}
	for _, h := range allHistories {
		histMap[h.InstanceID] = append(histMap[h.InstanceID], h)
	}
	for _, ins := range instances {
		rounds = append(rounds, instanceWithHistory{
			Instance:  ins,
			Histories: histMap[ins.ID],
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"instance":  latest,
		"instances": rounds,
		"histories": allHistories,
		"tasks":     tasks,
	}})
}

type PendingApprovalItem = dto.OrchidPendingApprovalItemDTO

// buildApprovalTitle 构建业务数据
func buildApprovalTitle(ins models.OrchidWorkflowInstance) string {
	switch ins.BizType {
	case "employee":
		var emp models.Employee
		if err := database.DB.First(&emp, ins.BizID).Error; err == nil {
			return fmt.Sprintf("员工入职：%s", emp.Name)
		}
	case "leave_request":
		var leave models.LeaveRequest
		if err := database.DB.Preload("Employee").First(&leave, ins.BizID).Error; err == nil {
			if leave.Employee != nil && leave.Employee.Name != "" {
				return fmt.Sprintf("请假申请：%s（%s）", leave.Employee.Name, leave.Type)
			}
			return fmt.Sprintf("请假申请：#%d", leave.ID)
		}
	case "event_booking":
		var booking models.EventBooking
		if err := database.DB.First(&booking, ins.BizID).Error; err == nil {
			return fmt.Sprintf("事件预定：%s", booking.Title)
		}
	case "notice":
		var notice models.Notice
		if err := database.DB.First(&notice, ins.BizID).Error; err == nil {
			return fmt.Sprintf("公告审批：%s", notice.Title)
		}
	case "resignation":
		var rz models.Resignation
		if err := database.DB.Preload("Employee").First(&rz, ins.BizID).Error; err == nil {
			if rz.Employee != nil {
				return fmt.Sprintf("离职审批：%s", rz.Employee.Name)
			}
			return fmt.Sprintf("离职审批：#%d", rz.ID)
		}
	}
	return fmt.Sprintf("%s #%d", ins.BizType, ins.BizID)
}

// buildApprovalDetailPath 构建业务数据
func buildApprovalDetailPath(bizType string, bizID int) string {
	switch bizType {
	case "employee":
		return fmt.Sprintf("/employee?id=%d", bizID)
	case "leave_request":
		return fmt.Sprintf("/leave-request?id=%d", bizID)
	case "event_booking":
		return fmt.Sprintf("/event-booking?id=%d", bizID)
	case "notice":
		return fmt.Sprintf("/notice?id=%d", bizID)
	case "resignation":
		return fmt.Sprintf("/resignation?id=%d", bizID)
	default:
		return "/dashboard"
	}
}

// GetMyPendingApprovals 获取数据
func GetMyPendingApprovals(c *gin.Context) {
	userID := c.GetInt("userID")
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录"})
		return
	}

	page := atoiDefault(c.DefaultQuery("page", "1"), 1)
	pageSize := atoiDefault(c.DefaultQuery("page_size", "10"), 10)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var total int64
	database.DB.Model(&models.OrchidWorkflowTask{}).
		Where("assignee_id = ? AND status = 'open' AND (task_type = 'approve' OR task_type = '')", userID).
		Count(&total)

	var tasks []models.OrchidWorkflowTask
	database.DB.Where("assignee_id = ? AND status = 'open' AND (task_type = 'approve' OR task_type = '')", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&tasks)

	if len(tasks) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": []PendingApprovalItem{}, "total": total}})
		return
	}

	instanceIDs := make([]int, 0, len(tasks))
	for _, t := range tasks {
		instanceIDs = append(instanceIDs, t.InstanceID)
	}

	var instances []models.OrchidWorkflowInstance
	database.DB.Where("id IN ?", instanceIDs).Find(&instances)
	insMap := map[int]models.OrchidWorkflowInstance{}
	for _, ins := range instances {
		insMap[ins.ID] = ins
	}

	items := make([]PendingApprovalItem, 0, len(tasks))
	for _, t := range tasks {
		ins, ok := insMap[t.InstanceID]
		if !ok {
			continue
		}
		items = append(items, PendingApprovalItem{
			TaskID:     t.ID,
			BizType:    ins.BizType,
			BizID:      ins.BizID,
			NodeKey:    t.NodeKey,
			Title:      buildApprovalTitle(ins),
			Status:     ins.Status,
			CreatedAt:  t.CreatedAt.Format(timeLayout),
			DetailPath: buildApprovalDetailPath(ins.BizType, ins.BizID),
			InstanceID: ins.ID,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": items, "total": total}})
}

// 我的已审
func GetMyApprovedApprovals(c *gin.Context) {
	userID := c.GetInt("userID")
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录"})
		return
	}

	page := atoiDefault(c.DefaultQuery("page", "1"), 1)
	pageSize := atoiDefault(c.DefaultQuery("page_size", "10"), 10)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var total int64
	database.DB.Model(&models.OrchidWorkflowTask{}).
		Where("assignee_id = ? AND status IN ('done','transferred','skipped') AND (task_type = 'approve' OR task_type = '')", userID).
		Count(&total)

	var tasks []models.OrchidWorkflowTask
	database.DB.Where("assignee_id = ? AND status IN ('done','transferred','skipped') AND (task_type = 'approve' OR task_type = '')", userID).
		Order("updated_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&tasks)

	if len(tasks) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": []PendingApprovalItem{}, "total": total}})
		return
	}

	instanceIDs := make([]int, 0, len(tasks))
	for _, t := range tasks {
		instanceIDs = append(instanceIDs, t.InstanceID)
	}
	var instances []models.OrchidWorkflowInstance
	database.DB.Where("id IN ?", instanceIDs).Find(&instances)
	insMap := map[int]models.OrchidWorkflowInstance{}
	for _, ins := range instances {
		insMap[ins.ID] = ins
	}

	items := make([]PendingApprovalItem, 0, len(tasks))
	for _, t := range tasks {
		ins, ok := insMap[t.InstanceID]
		if !ok {
			continue
		}
		items = append(items, PendingApprovalItem{
			TaskID:     t.ID,
			BizType:    ins.BizType,
			BizID:      ins.BizID,
			NodeKey:    t.NodeKey,
			Title:      buildApprovalTitle(ins),
			Status:     ins.Status,
			CreatedAt:  t.CreatedAt.Format(timeLayout),
			DetailPath: buildApprovalDetailPath(ins.BizType, ins.BizID),
			InstanceID: ins.ID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": items, "total": total}})
}

// 我的待阅
func GetMyPendingReads(c *gin.Context) {
	userID := c.GetInt("userID")
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录"})
		return
	}

	page := atoiDefault(c.DefaultQuery("page", "1"), 1)
	pageSize := atoiDefault(c.DefaultQuery("page_size", "10"), 10)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var total int64
	database.DB.Model(&models.OrchidWorkflowTask{}).
		Where("assignee_id = ? AND task_type = 'read' AND read_at IS NULL", userID).
		Count(&total)

	var tasks []models.OrchidWorkflowTask
	database.DB.Where("assignee_id = ? AND task_type = 'read' AND read_at IS NULL", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&tasks)

	if len(tasks) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": []PendingApprovalItem{}, "total": total}})
		return
	}

	instanceIDs := make([]int, 0, len(tasks))
	for _, t := range tasks {
		instanceIDs = append(instanceIDs, t.InstanceID)
	}
	var instances []models.OrchidWorkflowInstance
	database.DB.Where("id IN ?", instanceIDs).Find(&instances)
	insMap := map[int]models.OrchidWorkflowInstance{}
	for _, ins := range instances {
		insMap[ins.ID] = ins
	}

	items := make([]PendingApprovalItem, 0, len(tasks))
	for _, t := range tasks {
		ins, ok := insMap[t.InstanceID]
		if !ok {
			continue
		}
		items = append(items, PendingApprovalItem{
			TaskID:     t.ID,
			BizType:    ins.BizType,
			BizID:      ins.BizID,
			NodeKey:    t.NodeKey,
			Title:      buildApprovalTitle(ins),
			Status:     ins.Status,
			CreatedAt:  t.CreatedAt.Format(timeLayout),
			DetailPath: buildApprovalDetailPath(ins.BizType, ins.BizID),
			InstanceID: ins.ID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": items, "total": total})
}

// 我的已阅
func GetMyReadItems(c *gin.Context) {
	userID := c.GetInt("userID")
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录"})
		return
	}

	page := atoiDefault(c.DefaultQuery("page", "1"), 1)
	pageSize := atoiDefault(c.DefaultQuery("page_size", "10"), 10)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var total int64
	database.DB.Model(&models.OrchidWorkflowTask{}).
		Where("assignee_id = ? AND task_type = 'read' AND read_at IS NOT NULL", userID).
		Count(&total)

	var tasks []models.OrchidWorkflowTask
	database.DB.Where("assignee_id = ? AND task_type = 'read' AND read_at IS NOT NULL", userID).
		Order("read_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&tasks)

	if len(tasks) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": []PendingApprovalItem{}, "total": total}})
		return
	}

	instanceIDs := make([]int, 0, len(tasks))
	for _, t := range tasks {
		instanceIDs = append(instanceIDs, t.InstanceID)
	}
	var instances []models.OrchidWorkflowInstance
	database.DB.Where("id IN ?", instanceIDs).Find(&instances)
	insMap := map[int]models.OrchidWorkflowInstance{}
	for _, ins := range instances {
		insMap[ins.ID] = ins
	}

	items := make([]PendingApprovalItem, 0, len(tasks))
	for _, t := range tasks {
		ins, ok := insMap[t.InstanceID]
		if !ok {
			continue
		}
		items = append(items, PendingApprovalItem{
			TaskID:     t.ID,
			BizType:    ins.BizType,
			BizID:      ins.BizID,
			NodeKey:    t.NodeKey,
			Title:      buildApprovalTitle(ins),
			Status:     ins.Status,
			CreatedAt:  t.CreatedAt.Format(timeLayout),
			DetailPath: buildApprovalDetailPath(ins.BizType, ins.BizID),
			InstanceID: ins.ID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": items, "total": total}})
}

// TransferOrchidWorkflowTask 执行相关业务逻辑
func TransferOrchidWorkflowTask(c *gin.Context) {
	bizType := c.Query("biz_type")
	bizID := atoiDefault(c.Query("biz_id"), 0)
	if bizType == "" || bizID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "biz_type/biz_id 必填"})
		return
	}
	var req TransferOrSkipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	op := c.GetString("realName")
	if op == "" {
		op = c.GetString("username")
	}
	if err := transferTaskForInstance(bizType, bizID, req.FromUserID, req.ToUserID, op, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "转交失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "转交成功"})
}

// SkipOrchidWorkflowNode 执行相关业务逻辑
func SkipOrchidWorkflowNode(c *gin.Context) {
	bizType := c.Query("biz_type")
	bizID := atoiDefault(c.Query("biz_id"), 0)
	if bizType == "" || bizID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "biz_type/biz_id 必填"})
		return
	}
	var req TransferOrSkipReq
	_ = c.ShouldBindJSON(&req)
	op := c.GetString("realName")
	if op == "" {
		op = c.GetString("username")
	}
	if err := skipCurrentNodeForInstance(bizType, bizID, op, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "跳过失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已跳过"})
}

// SeedOrchidWorkflowTemplates 执行相关业务逻辑
func SeedOrchidWorkflowTemplates(c *gin.Context) {
	templates := []OrchidWorkflowDefinitionReq{
		{
			Name:        "请假审批模板-标准",
			BizType:     "leave_request",
			Description: "请假：直属经理审批后结束",
			IsActive:    true,
			DagJSON: `{
  "name": "请假审批模板-标准",
  "nodes": {
    "manager": {"id": 1, "activity": "approve", "config": {"approver_position_ids": [10]} }
  },
  "edges": []
}`,
		},
		{
			Name:        "事件预定模板-标准",
			BizType:     "event_booking",
			Description: "事件：行政审批后结束",
			IsActive:    true,
			DagJSON: `{
  "name": "事件预定模板-标准",
  "nodes": {
    "admin": {"id": 1, "activity": "approve", "config": {"approver_position_ids": [14]} }
  },
  "edges": []
}`,
		},
	}

	created := 0
	for _, t := range templates {
		if err := validateOrchidDag(t.Name, t.DagJSON); err != nil {
			continue
		}
		var count int64
		database.DB.Model(&models.OrchidWorkflowDefinition{}).Where("name = ?", t.Name).Count(&count)
		if count > 0 {
			continue
		}
		def := models.OrchidWorkflowDefinition{Name: t.Name, BizType: t.BizType, Description: t.Description, DagJSON: t.DagJSON, IsActive: t.IsActive}
		if err := database.DB.Create(&def).Error; err == nil {
			created++
			services.RefreshOrchidWorkflowCacheByBizType(def.BizType)
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "导入完成", "data": gin.H{"created": created}})
}

// atoiDefault 执行相关业务逻辑
func atoiDefault(s string, d int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return d
	}
	return n
}
