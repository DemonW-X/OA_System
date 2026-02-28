package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/models"
	"strings"

	"github.com/gin-gonic/gin"
)

type PositionRequest struct {
	Name         string `json:"name" binding:"required"`
	DepartmentID int    `json:"department_id"`
	Remark       string `json:"remark"`
}

func GetPositions(c *gin.Context) {
	var list []models.Position
	query := database.DB.Model(&models.Position{}).Preload("Department")
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		keyword = strings.TrimSpace(c.Query("name"))
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR remark LIKE ?", like, like)
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
	writeLog(c, "职位管理", "查询", "查询职位列表")
}

func GetPosition(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var pos models.Position
	if err := database.DB.Preload("Department").First(&pos, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "职位不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": pos})
	writeLog(c, "职位管理", "查询", "查询职位详情")
}

func CreatePosition(c *gin.Context) {
	var req PositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	// 硬删除同名软删除记录，避免唯一索引冲突
	database.DB.Unscoped().Where("name = ? AND deleted_at IS NOT NULL", req.Name).Delete(&models.Position{})

	pos := models.Position{Name: req.Name, DepartmentID: req.DepartmentID, Remark: req.Remark}
	if err := database.DB.Create(&pos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": pos})
	writeLog(c, "职位管理", "新增", "新增职位："+req.Name)
}

func UpdatePosition(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var pos models.Position
	if err := database.DB.First(&pos, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "职位不存在"})
		return
	}
	var req PositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	pos.Name = req.Name
	pos.DepartmentID = req.DepartmentID
	pos.Remark = req.Remark
	if err := database.DB.Save(&pos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": pos})
	writeLog(c, "职位管理", "修改", "修改职位："+req.Name)
}

func DeletePosition(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var pos models.Position
	if err := database.DB.First(&pos, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "职位不存在"})
		return
	}
	if err := database.DB.Delete(&pos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "职位管理", "删除", "删除职位："+pos.Name)
}
