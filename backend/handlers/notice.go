package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/models"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

var richTextPolicy = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("src", "alt", "width", "height").OnElements("img")
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("class", "style").Globally()
	return p
}()

type NoticeRequest = dto.NoticeRequestDTO
type NoticeSubmitReq = dto.NoticeSubmitRequestDTO
type NoticeApproveReq = dto.NoticeApproveRequestDTO

func noticeApproveStatusTagToInt(status string) int {
	switch status {
	case "draft":
		return 0
	case "pending":
		return 2
	case "approved":
		return 1
	case "rejected":
		return 3
	default:
		return 0
	}
}

func noticeStatusIntToApproveTag(status int) string {
	switch status {
	case 0:
		return "draft"
	case 1:
		return "approved"
	case 2:
		return "pending"
	case 3:
		return "rejected"
	default:
		return "draft"
	}
}

func GetNotices(c *gin.Context) {
	var list []models.Notice
	query := database.DB.Model(&models.Notice{}).Preload("Department")
	if title := c.Query("title"); title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if approveStatus := c.Query("approve_status"); approveStatus != "" {
		query = query.Where("approve_status = ?", approveStatus)
	}
	if deptID := c.Query("department_id"); deptID != "" {
		query = query.Where("department_id = ?", deptID)
	}
	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Offset(offset).Limit(pageSize).Find(&list)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
	writeLog(c, "公告管理", "查询", "查询公告列表")
}

func GetNotice(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var notice models.Notice
	if err := database.DB.Preload("Department").First(&notice, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "公告不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": notice})
	writeLog(c, "公告管理", "查询", "查询公告详情")
}

func CreateNotice(c *gin.Context) {
	var req NoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	author := c.GetString("realName")
	if author == "" {
		author = c.GetString("username")
	}
	safeContent := richTextPolicy.Sanitize(req.Content)
	notice := models.Notice{
		Title:         req.Title,
		Content:       safeContent,
		Author:        author,
		Status:        req.Status,
		Attachments:   req.Attachments,
		DepartmentID:  req.DepartmentID,
		ApproveStatus: "draft",
	}
	if err := database.DB.Create(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": notice})
	writeLog(c, "公告管理", "新增", "新增公告："+req.Title)
}

func UpdateNotice(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var notice models.Notice
	if err := database.DB.First(&notice, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "公告不存在"})
		return
	}
	if notice.ApproveStatus != "draft" && notice.ApproveStatus != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅草稿/已拒绝状态可修改"})
		return
	}
	var req NoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	safeContent := richTextPolicy.Sanitize(req.Content)
	notice.Title = req.Title
	notice.Content = safeContent
	notice.Status = req.Status
	notice.Attachments = req.Attachments
	notice.DepartmentID = req.DepartmentID
	if err := database.DB.Save(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": notice})
	writeLog(c, "公告管理", "修改", "修改公告："+req.Title)
}

func SubmitNotice(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var notice models.Notice
	if err := database.DB.First(&notice, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "公告不存在"})
		return
	}
	if notice.ApproveStatus != "draft" && notice.ApproveStatus != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅草稿/已拒绝状态可提交"})
		return
	}
	var req NoticeSubmitReq
	_ = c.ShouldBindJSON(&req)

	op := currentOperator(c)
	ret, err := submitApprovalFlow("notice", notice.ID, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "流程实例启动失败: " + err.Error()})
		return
	}
	notice.ApproveStatus = ret.Status
	notice.ApprovedBy = ret.ApprovedBy
	notice.ApprovedAt = ret.ApprovedAt
	notice.ApproveRemark = ret.ApproveRemark
	notice.WorkflowLogs = ret.WorkflowLogs
	if ret.Status == "approved" {
		notice.Status = 1
	} else if ret.Status == "rejected" {
		notice.Status = 0
	}

	if err := database.DB.Save(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": notice})
	writeLog(c, "公告管理", "提交", "提交公告审核："+notice.Title)
}

func ApproveNotice(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var notice models.Notice
	if err := database.DB.First(&notice, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "公告不存在"})
		return
	}
	if notice.ApproveStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅待审批状态可审批"})
		return
	}

	var req NoticeApproveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	op := currentOperator(c)
	ret, err := approveApprovalFlow("notice", notice.ID, op, req.Action, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "审批失败: " + err.Error()})
		return
	}
	notice.ApproveStatus = ret.Status
	notice.ApprovedBy = ret.ApprovedBy
	notice.ApprovedAt = ret.ApprovedAt
	notice.ApproveRemark = ret.ApproveRemark
	notice.WorkflowLogs = ret.WorkflowLogs
	notice.Status = noticeApproveStatusTagToInt(ret.Status)

	if err := database.DB.Save(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "审批保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": notice})
	writeLog(c, "公告管理", "审批", "审批公告："+notice.Title)
}

func WithdrawNotice(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var notice models.Notice
	if err := database.DB.First(&notice, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "公告不存在"})
		return
	}
	if notice.ApproveStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅待审批状态可撤回"})
		return
	}
	op := currentOperator(c)
	if err := withdrawApprovalFlow("notice", notice.ID, op); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	notice.ApproveStatus = "draft"
	notice.ApprovedBy = ""
	notice.ApprovedAt = nil
	notice.ApproveRemark = ""
	notice.Status = 0
	notice.WorkflowLogs = buildBizWorkflowLogs("notice", notice.ID)
	if err := database.DB.Save(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "撤回失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": notice})
	writeLog(c, "公告管理", "撤回", "撤回公告审核："+notice.Title)
}

func CancelApproveNotice(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var notice models.Notice
	if err := database.DB.First(&notice, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "公告不存在"})
		return
	}
	if notice.ApproveStatus != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅已通过状态可取消审核"})
		return
	}
	notice.ApproveStatus = "draft"
	notice.ApprovedBy = ""
	notice.ApprovedAt = nil
	notice.ApproveRemark = ""
	notice.Status = 0
	if err := database.DB.Save(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "取消审核失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": notice, "msg": "已取消审核"})
	writeLog(c, "公告管理", "取消审核", "取消公告审核："+notice.Title)
}

func DeleteNotice(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var notice models.Notice
	if err := database.DB.First(&notice, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "公告不存在"})
		return
	}
	if notice.ApproveStatus == "pending" || notice.ApproveStatus == "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "待审批/已通过公告不可删除"})
		return
	}
	if err := database.DB.Delete(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "公告管理", "删除", "删除公告："+notice.Title)
}
