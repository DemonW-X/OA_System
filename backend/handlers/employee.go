package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/models"
	"regexp"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

type EmployeeRequest struct {
	Name         string `json:"name" binding:"required"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	DepartmentID int    `json:"department_id"`
	PositionID   int    `json:"position_id"`
	Status       int    `json:"status"`
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
	var req EmployeeRequest
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
	// 检查手机号是否已被注册为用户
	var existCount int64
	database.DB.Model(&models.User{}).Where("username = ?", req.Phone).Count(&existCount)
	if existCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该手机号已被注册"})
		return
	}

	// 开启事务，保证员工和用户同时创建成功
	tx := database.DB.Begin()

	// 创建登录用户，默认密码 123456
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

	// 创建员工，关联 UserID
	emp := models.Employee{
		Name:         req.Name,
		Phone:        req.Phone,
		Email:        req.Email,
		DepartmentID: req.DepartmentID,
		PositionID:   req.PositionID,
		Status:       req.Status,
		UserID:       user.ID,
	}
	if emp.Status == 0 {
		emp.Status = 1
	}
	if err := tx.Create(&emp).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建员工失败: " + err.Error()})
		return
	}

	tx.Commit()
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
	var req EmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if msg, ok := validateEmployee(req.Phone, req.Email); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": msg})
		return
	}

	tx := database.DB.Begin()

	// 同步更新关联用户的真实姓名
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
	emp.DepartmentID = req.DepartmentID
	emp.PositionID = req.PositionID
	emp.Status = req.Status
	if err := tx.Save(&emp).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败"})
		return
	}

	tx.Commit()
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

	// 同步删除关联的登录用户
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

	tx.Commit()
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
	var req struct {
		Action string `json:"action" binding:"required"` // approved / rejected
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
