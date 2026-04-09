package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/models"

	"github.com/gin-gonic/gin"
)

// GetLogs 获取数据
func GetLogs(c *gin.Context) {
	var list []models.OperationLog
	query := database.DB.Model(&models.OperationLog{})

	if username := c.Query("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if module := c.Query("module"); module != "" {
		query = query.Where("module = ?", module)
	}
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}
	if startTime := c.Query("start_time"); startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime := c.Query("end_time"); endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	var total int64
	query.Count(&total)

	page, pageSize, offset := getPagination(c)
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
}

// GetLogModules 获取数据
func GetLogModules(c *gin.Context) {
	var modules []string
	database.DB.Model(&models.OperationLog{}).
		Distinct("module").
		Where("module IS NOT NULL AND module <> ''").
		Order("module ASC").
		Pluck("module", &modules)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": modules,
	})
}
