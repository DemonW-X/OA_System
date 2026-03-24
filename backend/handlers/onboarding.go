package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OnboardingRequest struct {
	EmployeeName   string `json:"employee_name" binding:"required"`
	OnboardDate    string `json:"onboard_date" binding:"required"` // yyyy-MM-dd
	OnboardType    string `json:"onboard_type"`                    // new/rehire/transfer
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

func GetOnboardings(c *gin.Context) {
	var list []models.Onboarding
	query := database.DB.Model(&models.Onboarding{})

	if name := c.Query("employee_name"); name != "" {
		query = query.Where("employee_name LIKE ?", "%"+name+"%")
	}
	if status := c.Query("approve_status"); status != "" {
		query = query.Where("approve_status = ?", status)
	}
	if onboardType := c.Query("onboard_type"); onboardType != "" {
		query = query.Where("onboard_type = ?", onboardType)
	}

	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Order("onboard_date desc, id desc").Offset(offset).Limit(pageSize).Find(&list)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
	writeLog(c, "入职管理", "查询", "查询入职列表")
}

func GetOnboarding(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Onboarding
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "入职记录不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "入职管理", "查询", "查询入职详情")
}

func CreateOnboarding(c *gin.Context) {
	var req OnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	onboardDate, err := time.ParseInLocation(dateLayout, req.OnboardDate, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "入职日期格式错误，请使用 yyyy-MM-dd"})
		return
	}

	onboardType := req.OnboardType
	if onboardType == "" {
		onboardType = "new"
	}
	probationDays := req.ProbationDays
	if probationDays <= 0 {
		probationDays = 90
	}
	probationEnd := onboardDate.AddDate(0, 0, probationDays)

	// 检查是否存在相同身份证号的记录（无论任何状态）
	var existCount int64
	database.DB.Model(&models.Onboarding{}).Where("id_card = ?", req.IDCard).Count(&existCount)
	if existCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该身份证号已存在入职申请记录，请勿重复提交"})
		return
	}

	item := models.Onboarding{
		EmployeeName:   req.EmployeeName,
		OnboardDate:    onboardDate,
		OnboardType:    onboardType,
		ProbationDays:  probationDays,
		ProbationEnd:   &probationEnd,
		IDCard:         req.IDCard,
		Phone:          req.Phone,
		Email:          req.Email,
		NativePlace:    req.NativePlace,
		Address:        req.Address,
		EmergencyName:  req.EmergencyName,
		EmergencyPhone: req.EmergencyPhone,
		Education:      req.Education,
		School:         req.School,
		Major:          req.Major,
		WorkYears:      req.WorkYears,
		DepartmentID:   req.DepartmentID,
		PositionID:     req.PositionID,
		Remark:         req.Remark,
		ApproveStatus:  "draft",
	}
	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "入职管理", "新增", "新增入职记录")
}

func UpdateOnboarding(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Onboarding
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "入职记录不存在"})
		return
	}
	if item.ApproveStatus != "draft" && item.ApproveStatus != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅草稿/已拒绝状态可修改"})
		return
	}

	var req OnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	onboardDate, err := time.ParseInLocation(dateLayout, req.OnboardDate, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "入职日期格式错误，请使用 yyyy-MM-dd"})
		return
	}

	onboardType := req.OnboardType
	if onboardType == "" {
		onboardType = "new"
	}
	probationDays := req.ProbationDays
	if probationDays <= 0 {
		probationDays = 90
	}
	probationEnd := onboardDate.AddDate(0, 0, probationDays)

	item.EmployeeName = req.EmployeeName
	item.OnboardDate = onboardDate
	item.OnboardType = onboardType
	item.ProbationDays = probationDays
	item.ProbationEnd = &probationEnd
	item.IDCard = req.IDCard
	item.Phone = req.Phone
	item.Email = req.Email
	item.NativePlace = req.NativePlace
	item.Address = req.Address
	item.EmergencyName = req.EmergencyName
	item.EmergencyPhone = req.EmergencyPhone
	item.Education = req.Education
	item.School = req.School
	item.Major = req.Major
	item.WorkYears = req.WorkYears
	item.DepartmentID = req.DepartmentID
	item.PositionID = req.PositionID
	item.Remark = req.Remark

	if err := database.DB.Select(
		"employee_name", "onboard_date", "onboard_type", "probation_days", "probation_end",
		"id_card", "phone", "email", "native_place", "address",
		"emergency_name", "emergency_phone", "education", "school", "major",
		"work_years", "department_id", "position_id", "remark",
	).Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "入职管理", "修改", "修改入职记录")
}

func DeleteOnboarding(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Onboarding
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "入职记录不存在"})
		return
	}
	if item.ApproveStatus == "pending" || item.ApproveStatus == "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "待审批/已通过记录不可删除"})
		return
	}
	if err := database.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "入职管理", "删除", "删除入职记录")
}

func SubmitOnboarding(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Onboarding
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "入职记录不存在"})
		return
	}
	if item.ApproveStatus != "draft" && item.ApproveStatus != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅草稿/已拒绝状态可提交"})
		return
	}

	op := currentOperator(c)
	ret, err := submitApprovalFlow("onboarding", item.ID, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "流程实例启动失败: " + err.Error()})
		return
	}
	item.ApproveStatus = ret.Status
	item.ApprovedBy = ret.ApprovedBy
	item.ApprovedAt = ret.ApprovedAt
	item.ApproveRemark = ret.ApproveRemark

	tx := database.DB.Begin()
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "提交失败"})
		return
	}

	// 无流程定义自动通过时，复用员工创建逻辑
	if ret.Status == "approved" {
		if err := syncEmployeeFromOnboarding(tx, &item); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "入职管理", "提交", "提交入职审核")
}

func WithdrawOnboarding(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Onboarding
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "入职记录不存在"})
		return
	}
	if item.ApproveStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅待审批状态可撤回"})
		return
	}

	op := currentOperator(c)
	if err := withdrawApprovalFlow("onboarding", item.ID, op); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	item.ApproveStatus = "draft"
	item.ApprovedBy = ""
	item.ApprovedAt = nil
	item.ApproveRemark = ""
	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "撤回失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "入职管理", "撤回", "撤回入职审核")
}

func ApproveOnboarding(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Onboarding
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "入职记录不存在"})
		return
	}
	if item.ApproveStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅待审批状态可审批"})
		return
	}
	var req struct {
		Action string `json:"action" binding:"required"` // approved/rejected
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if req.Action != "approved" && req.Action != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "action 只能为 approved 或 rejected"})
		return
	}

	approver := c.GetString("realName")
	if approver == "" {
		approver = c.GetString("username")
	}
	ret, err := approveApprovalFlow("onboarding", item.ID, approver, req.Action, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "审批失败: " + err.Error()})
		return
	}
	item.ApproveStatus = ret.Status
	item.ApprovedBy = ret.ApprovedBy
	item.ApprovedAt = ret.ApprovedAt
	item.ApproveRemark = ret.ApproveRemark

	tx := database.DB.Begin()
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "保存失败"})
		return
	}

	// 审批通过时联动员工数据
	if ret.Status == "approved" {
		if err := syncEmployeeFromOnboarding(tx, &item); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
			return
		}
	}

	tx.Commit()
	action := "审批通过"
	if req.Action == "rejected" {
		action = "审批拒绝"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "入职管理", "审批", action+"入职申请")
}

// syncEmployeeFromOnboarding 审批通过后联动创建/恢复员工，供 SubmitOnboarding 和 ApproveOnboarding 共用
func syncEmployeeFromOnboarding(tx *gorm.DB, item *models.Onboarding) error {
	if item.OnboardType == "rehire" {
		var emp models.Employee
		err := tx.Where("name = ? AND phone = ? AND status = 0", item.EmployeeName, item.Phone).First(&emp).Error
		if err == nil {
			return tx.Model(&emp).Update("status", 1).Error
		}
	}
	newEmp := models.Employee{
		Name:         item.EmployeeName,
		Phone:        item.Phone,
		Email:        item.Email,
		DepartmentID: item.DepartmentID,
		PositionID:   item.PositionID,
		Status:       1,
	}
	// 只插入业务需要的字段，避免触发 employees 表中不存在的审批相关列
	return tx.Select("name", "phone", "email", "department_id", "position_id", "status").Create(&newEmp).Error
}

func CancelApproveOnboarding(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item models.Onboarding
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "入职记录不存在"})
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
	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "取消审核失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item, "msg": "已取消审核，状态恢复为草稿"})
	writeLog(c, "入职管理", "取消审核", "取消入职审核")
}
