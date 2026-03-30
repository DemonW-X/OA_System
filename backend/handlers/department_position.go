package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/models"

	"github.com/gin-gonic/gin"
)

type DepartmentPositionRequest struct {
	DepartmentID int `json:"department_id" binding:"required"`
	PositionID   int `json:"position_id" binding:"required"`
}

func CreateDepartmentPosition(c *gin.Context) {
	var req DepartmentPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	var dept models.Department
	if err := database.DB.First(&dept, "id = ?", req.DepartmentID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "部门不存在"})
		return
	}
	var pos models.Position
	if err := database.DB.First(&pos, "id = ?", req.PositionID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "职位不存在"})
		return
	}

	var exists int64
	database.DB.Model(&models.DepartmentPosition{}).
		Where("department_id = ? AND position_id = ?", req.DepartmentID, req.PositionID).
		Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该部门-职位关系已存在"})
		return
	}

	rel := models.DepartmentPosition{DepartmentID: req.DepartmentID, PositionID: req.PositionID}
	if err := database.DB.Create(&rel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	database.DB.Preload("Department").Preload("Position").First(&rel, rel.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rel})
	writeLog(c, "角色管理", "新增", "新增部门-职位关系")
}

func DeleteDepartmentPosition(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var rel models.DepartmentPosition
	if err := database.DB.First(&rel, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "关系不存在"})
		return
	}
	if err := database.DB.Delete(&rel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "角色管理", "删除", "删除部门-职位关系")
}
