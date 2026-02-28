package handlers

import (
	"net/http"
	"oa-system/database"
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

type NoticeRequest struct {
	Title        string `json:"title" binding:"required"`
	Content      string `json:"content"`
	Status       int    `json:"status"`
	Attachments  string `json:"attachments"`
	DepartmentID int    `json:"department_id"`
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
		Title:        req.Title,
		Content:      safeContent,
		Author:       author,
		Status:       req.Status,
		Attachments:  req.Attachments,
		DepartmentID: req.DepartmentID,
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
	if err := database.DB.Delete(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "公告管理", "删除", "删除公告："+notice.Title)
}
