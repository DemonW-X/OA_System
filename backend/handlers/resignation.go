package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/models"
	"time"

	"github.com/gin-gonic/gin"
)

type ResignationRequest = dto.ResignationRequestDTO
type ResignationApproveReq = dto.ResignationApproveRequestDTO

func GetResignations(c *gin.Context) {
	var list []models.Resignation
	query := database.DB.Model(&models.Resignation{}).Preload("Employee").Preload("Employee.Department").Preload("Employee.PositionInfo")

	if empID := c.Query("employee_id"); empID != "" {
		query = query.Where("employee_id = ?", empID)
	}
	if empName := c.Query("employee_name"); empName != "" {
		query = query.Joins("JOIN employees ON employees.id = resignations.employee_id").Where("employees.name LIKE ?", "%"+empName+"%")
	}
	if approveStatus := c.Query("approve_status"); approveStatus != "" {
		query = query.Where("approve_status = ?", approveStatus)
	}

	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Order("resign_date desc, id desc").Offset(offset).Limit(pageSize).Find(&list)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
	writeLog(c, "离职管理", "查询", "查询离职列表")
}

func GetResignation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Resignation
	if err := database.DB.Preload("Employee").Preload("Employee.Department").Preload("Employee.PositionInfo").First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "离职记录不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
}

func CreateResignation(c *gin.Context) {
	var req ResignationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	resignDate, err := time.ParseInLocation(dateLayout, req.ResignDate, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "离职日期格式错误，请使用 yyyy-MM-dd"})
		return
	}

	var emp models.Employee
	if err := database.DB.First(&emp, "id = ?", req.EmployeeID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "员工不存在"})
		return
	}

	item := models.Resignation{
		EmployeeID:    req.EmployeeID,
		ResignDate:    resignDate,
		Reason:        req.Reason,
		Remark:        req.Remark,
		ApproveStatus: "draft",
	}
	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "离职管理", "新增", "新增离职记录")
}

func UpdateResignation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var item models.Resignation
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "离职记录不存在"})
		return
	}
	if item.ApproveStatus != "draft" && item.ApproveStatus != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅草稿/已拒绝状态可修改"})
		return
	}

	var req ResignationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	resignDate, err := time.ParseInLocation(dateLayout, req.ResignDate, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "离职日期格式错误，请使用 yyyy-MM-dd"})
		return
	}

	var emp models.Employee
	if err := database.DB.First(&emp, "id = ?", req.EmployeeID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "员工不存在"})
		return
	}

	item.EmployeeID = req.EmployeeID
	item.ResignDate = resignDate
	item.Reason = req.Reason
	item.Remark = req.Remark
	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "离职管理", "修改", "修改离职记录")
}

func SubmitResignation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Resignation
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "离职记录不存在"})
		return
	}
	if item.ApproveStatus != "draft" && item.ApproveStatus != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅草稿/已拒绝状态可提交"})
		return
	}

	op := currentOperator(c)
	ret, err := submitApprovalFlow("resignation", item.ID, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "流程实例启动失败: " + err.Error()})
		return
	}
	item.ApproveStatus = ret.Status
	item.ApprovedBy = ret.ApprovedBy
	item.ApprovedAt = ret.ApprovedAt
	item.ApproveRemark = ret.ApproveRemark
	item.WorkflowLogs = ret.WorkflowLogs

	tx := database.DB.Begin()
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "提交失败"})
		return
	}
	if ret.Status == "approved" {
		if err := tx.Model(&models.Employee{}).Where("id = ?", item.EmployeeID).Update("status", 0).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新员工状态失败"})
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "离职管理", "提交", "提交离职审核")
}

func ApproveResignation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Resignation
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "离职记录不存在"})
		return
	}
	if item.ApproveStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅待审批状态可审批"})
		return
	}
	var req ResignationApproveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	op := currentOperator(c)
	ret, err := approveApprovalFlow("resignation", item.ID, op, req.Action, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "审批失败: " + err.Error()})
		return
	}
	item.ApproveStatus = ret.Status
	item.ApprovedBy = ret.ApprovedBy
	item.ApprovedAt = ret.ApprovedAt
	item.ApproveRemark = ret.ApproveRemark
	item.WorkflowLogs = ret.WorkflowLogs

	tx := database.DB.Begin()
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "审批保存失败"})
		return
	}
	if ret.Status == "approved" {
		if err := tx.Model(&models.Employee{}).Where("id = ?", item.EmployeeID).Update("status", 0).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新员工状态失败"})
			return
		}
	} else if ret.Status == "rejected" {
		if err := tx.Model(&models.Employee{}).Where("id = ?", item.EmployeeID).Update("status", 1).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新员工状态失败"})
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "离职管理", "审批", "审批离职记录")
}

func WithdrawResignation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Resignation
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "离职记录不存在"})
		return
	}
	if item.ApproveStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅待审批状态可撤回"})
		return
	}
	op := currentOperator(c)
	if err := withdrawApprovalFlow("resignation", item.ID, op); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	item.ApproveStatus = "draft"
	item.ApprovedBy = ""
	item.ApprovedAt = nil
	item.ApproveRemark = ""
	item.WorkflowLogs = buildBizWorkflowLogs("resignation", item.ID)

	tx := database.DB.Begin()
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "撤回失败"})
		return
	}
	if err := tx.Model(&models.Employee{}).Where("id = ?", item.EmployeeID).Update("status", 1).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "恢复员工状态失败"})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "离职管理", "撤回", "撤回离职审核")
}

func CancelApproveResignation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Resignation
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "离职记录不存在"})
		return
	}
	if item.ApproveStatus != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅已通过状态可取消审核"})
		return
	}
	item.ApproveStatus = "draft"
	item.ApprovedBy = ""
	item.ApprovedAt = nil
	item.ApproveRemark = ""

	tx := database.DB.Begin()
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "取消审核失败"})
		return
	}
	if err := tx.Model(&models.Employee{}).Where("id = ?", item.EmployeeID).Update("status", 1).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "恢复员工状态失败"})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item, "msg": "已取消审核"})
	writeLog(c, "离职管理", "取消审核", "取消离职审核")
}

func DeleteResignation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var item models.Resignation
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "离职记录不存在"})
		return
	}
	if item.ApproveStatus == "pending" || item.ApproveStatus == "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "待审批/已通过记录不可删除"})
		return
	}

	tx := database.DB.Begin()
	empID := item.EmployeeID
	if err := tx.Delete(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}

	var count int64
	if err := tx.Model(&models.Resignation{}).Where("employee_id = ?", empID).Count(&count).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "校验失败"})
		return
	}
	if count == 0 {
		if err := tx.Model(&models.Employee{}).Where("id = ?", empID).Update("status", 1).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "恢复员工状态失败"})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "离职管理", "删除", "删除离职记录")
}
