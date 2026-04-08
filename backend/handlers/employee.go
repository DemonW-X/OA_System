package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/models"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func parseDateOnly(value string) (*time.Time, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02", text, time.Local)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func validateEmployee(phone, email string) (string, bool) {
	if phone != "" && !phoneRegex.MatchString(phone) {
		return "手机号格式不正确", false
	}
	if email != "" && !emailRegex.MatchString(email) {
		return "邮箱格式不正确", false
	}
	return "", true
}

func GetEmployees(c *gin.Context) {
	nameOnly := c.Query("name_only") == "1"
	if nameOnly {
		var list []dto.EmployeeNameItemDTO
		query := database.DB.Model(&models.Employee{})
		keyword := c.Query("keyword")
		if keyword == "" {
			keyword = c.Query("name")
		}
		if keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where("name LIKE ?", like)
		}
		if deptID := c.Query("department_id"); deptID != "" {
			query = query.Where("department_id = ?", deptID)
		}
		if posID := c.Query("position_id"); posID != "" {
			query = query.Where("position_id = ?", posID)
		}
		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}
		var total int64
		query.Count(&total)
		page, pageSize, offset := getPagination(c)
		query.Select("id", "name").Order("id asc").Offset(offset).Limit(pageSize).Scan(&list)
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
		})
		writeLog(c, "员工管理", "查询", "查询员工列表")
		return
	}

	var list []models.Employee
	query := database.DB.Model(&models.Employee{}).Preload("Department").Preload("PositionInfo")
	keyword := c.Query("keyword")
	if keyword == "" {
		keyword = c.Query("name")
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR phone LIKE ? OR email LIKE ?", like, like, like)
	}
	if deptID := c.Query("department_id"); deptID != "" {
		query = query.Where("department_id = ?", deptID)
	}
	if posID := c.Query("position_id"); posID != "" {
		query = query.Where("position_id = ?", posID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Order("id asc").Offset(offset).Limit(pageSize).Find(&list)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
	writeLog(c, "员工管理", "查询", "查询员工列表")
}

func GetEmployee(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var emp models.Employee
	if err := database.DB.Preload("Department").Preload("PositionInfo").First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "员工不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": emp})
	writeLog(c, "员工管理", "查询", "查询员工详情")
}

func CreateEmployee(c *gin.Context) {
	var req dto.EmployeeRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if msg, ok := validateEmployee(req.Phone, req.Email); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": msg})
		return
	}
	if req.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "手机号不能为空，将作为登录账号"})
		return
	}

	var existCount int64
	database.DB.Model(&models.User{}).Where("username = ?", req.Phone).Count(&existCount)
	if existCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该手机号已被注册"})
		return
	}

	onboardDate, err := parseDateOnly(req.OnboardDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "onboard_date 格式应为 YYYY-MM-DD"})
		return
	}
	probationEnd, err := parseDateOnly(req.ProbationEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "probation_end 格式应为 YYYY-MM-DD"})
		return
	}

	tx := database.DB.Begin()

	hashed, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "密码加密失败"})
		return
	}
	user := models.User{
		Username: req.Phone,
		Password: string(hashed),
		RealName: req.Name,
		Role:     "employee",
	}
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建用户失败: " + err.Error()})
		return
	}

	emp := models.Employee{
		Name:           req.Name,
		Phone:          req.Phone,
		Email:          req.Email,
		OnboardDate:    onboardDate,
		OnboardType:    req.OnboardType,
		ProbationDays:  req.ProbationDays,
		ProbationEnd:   probationEnd,
		IDCard:         req.IDCard,
		NativePlace:    req.NativePlace,
		Address:        req.Address,
		EmergencyName:  req.EmergencyName,
		EmergencyPhone: req.EmergencyPhone,
		Education:      req.Education,
		School:         req.School,
		Major:          req.Major,
		WorkYears:      req.WorkYears,
		Remark:         req.Remark,
		DepartmentID:   req.DepartmentID,
		PositionID:     req.PositionID,
		Status:         req.Status,
		UserID:         user.ID,
	}
	if emp.OnboardType == "" {
		emp.OnboardType = "new"
	}
	if emp.ProbationDays <= 0 {
		emp.ProbationDays = 90
	}
	if emp.Status == 0 {
		emp.Status = 1
	}
	if err := tx.Create(&emp).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建员工失败: " + err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "提交事务失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": emp})
	writeLog(c, "员工管理", "新增", "新增员工："+req.Name)
}

func UpdateEmployee(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var emp models.Employee
	if err := database.DB.First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "员工不存在"})
		return
	}
	var req dto.EmployeeRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if msg, ok := validateEmployee(req.Phone, req.Email); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": msg})
		return
	}

	onboardDate, err := parseDateOnly(req.OnboardDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "onboard_date 格式应为 YYYY-MM-DD"})
		return
	}
	probationEnd, err := parseDateOnly(req.ProbationEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "probation_end 格式应为 YYYY-MM-DD"})
		return
	}

	tx := database.DB.Begin()

	if emp.UserID > 0 {
		if err := tx.Model(&models.User{}).Where("id = ?", emp.UserID).
			Update("real_name", req.Name).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新用户信息失败"})
			return
		}
	}

	emp.Name = req.Name
	emp.Phone = req.Phone
	emp.Email = req.Email
	emp.OnboardDate = onboardDate
	emp.OnboardType = req.OnboardType
	emp.ProbationDays = req.ProbationDays
	emp.ProbationEnd = probationEnd
	emp.IDCard = req.IDCard
	emp.NativePlace = req.NativePlace
	emp.Address = req.Address
	emp.EmergencyName = req.EmergencyName
	emp.EmergencyPhone = req.EmergencyPhone
	emp.Education = req.Education
	emp.School = req.School
	emp.Major = req.Major
	emp.WorkYears = req.WorkYears
	emp.Remark = req.Remark
	emp.DepartmentID = req.DepartmentID
	emp.PositionID = req.PositionID
	emp.Status = req.Status
	if emp.OnboardType == "" {
		emp.OnboardType = "new"
	}
	if emp.ProbationDays <= 0 {
		emp.ProbationDays = 90
	}
	if err := tx.Save(&emp).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "提交事务失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": emp})
	writeLog(c, "员工管理", "修改", "修改员工："+req.Name)
}

func DeleteEmployee(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var emp models.Employee
	if err := database.DB.First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "员工不存在"})
		return
	}

	tx := database.DB.Begin()

	if emp.UserID > 0 {
		if err := tx.Delete(&models.User{}, emp.UserID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除用户失败"})
			return
		}
	}

	if err := tx.Delete(&emp).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "提交事务失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "员工管理", "删除", "删除员工："+emp.Name)
}

// SubmitEmployee 提交审核
func SubmitEmployee(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var emp models.Employee
	if err := database.DB.First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "员工不存在"})
		return
	}
	if emp.ApproveStatus != "draft" && emp.ApproveStatus != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅草稿或已拒绝状态可提交审核"})
		return
	}

	op := currentOperator(c)
	ret, err := submitApprovalFlow("employee", emp.ID, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "流程实例启动失败: " + err.Error()})
		return
	}
	emp.ApproveStatus = ret.Status
	emp.ApprovedBy = ret.ApprovedBy
	emp.ApprovedAt = ret.ApprovedAt
	emp.ApproveRemark = ret.ApproveRemark

	if err := database.DB.Save(&emp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "提交审核失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": emp})
	writeLog(c, "员工管理", "提交", "提交员工审核："+emp.Name)
}

// WithdrawEmployee 撤回审核（仅第一个节点未审批时可撤回）
func WithdrawEmployee(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var emp models.Employee
	if err := database.DB.First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "员工不存在"})
		return
	}
	if emp.ApproveStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅待审批状态可撤回"})
		return
	}

	op := currentOperator(c)
	if err := withdrawApprovalFlow("employee", emp.ID, op); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	emp.ApproveStatus = "draft"
	emp.ApprovedBy = ""
	emp.ApprovedAt = nil
	emp.ApproveRemark = ""
	if err := database.DB.Save(&emp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "撤回失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": emp})
	writeLog(c, "员工管理", "撤回", "撤回员工审核："+emp.Name)
}

// ApproveEmployee 审批员工
func ApproveEmployee(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var emp models.Employee
	if err := database.DB.First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "员工不存在"})
		return
	}
	var req dto.EmployeeApproveRequestDTO
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
	ret, err := approveApprovalFlow("employee", emp.ID, approver, req.Action, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "审批失败: " + err.Error()})
		return
	}
	emp.ApproveStatus = ret.Status
	emp.ApprovedBy = ret.ApprovedBy
	emp.ApprovedAt = ret.ApprovedAt
	emp.ApproveRemark = ret.ApproveRemark
	if err := database.DB.Save(&emp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "审批失败"})
		return
	}
	action := "审批通过"
	if req.Action == "rejected" {
		action = "审批拒绝"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": emp})
	writeLog(c, "员工管理", "审批", action+"："+emp.Name)
}

// CancelEmployeeApproval 取消已通过审核（仅重置单据状态，不处理流程记录）
func CancelEmployeeApproval(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var emp models.Employee
	if err := database.DB.First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "员工不存在"})
		return
	}
	if emp.ApproveStatus != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅已通过状态可取消审核"})
		return
	}

	emp.ApproveStatus = "draft"
	emp.ApprovedBy = ""
	emp.ApprovedAt = nil
	emp.ApproveRemark = ""
	if err := database.DB.Save(&emp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "取消审核失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": emp, "msg": "已取消审核，状态恢复为草稿"})
	writeLog(c, "员工管理", "取消审核", "取消审核（不变更流程记录）："+emp.Name)
}
