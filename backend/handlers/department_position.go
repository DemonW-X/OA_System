package handlers

import (
	"errors"
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetDepartmentPositions 获取数据
func GetDepartmentPositions(c *gin.Context) {
	var list []models.DepartmentPosition
	query := database.DB.Model(&models.DepartmentPosition{}).
		Preload("Department").
		Preload("Position")

	if deptID := c.Query("department_id"); deptID != "" {
		query = query.Where("department_id = ?", deptID)
	}
	if posID := c.Query("position_id"); posID != "" {
		query = query.Where("position_id = ?", posID)
	}

	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Order("id asc").Offset(offset).Limit(pageSize).Find(&list)

	listResp := make([]dto.DepartmentPositionListItemDTO, 0, len(list))
	for _, item := range list {
		listResp = append(listResp, dto.DepartmentPositionListItemDTO{
			ID:           item.ID,
			DepartmentID: item.DepartmentID,
			Department:   item.Department,
			PositionID:   item.PositionID,
			Position: dto.PositionLiteDTO{
				ID:   item.Position.ID,
				Name: item.Position.Name,
			},
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, dto.DepartmentPositionListResponseDTO{
		Code: 0,
		Data: dto.DepartmentPositionListDataDTO{
			List:     listResp,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
	writeLog(c, "角色管理", "查询", "查询部门-职位关系列表")
}

// CreateDepartmentPosition 创建数据
func CreateDepartmentPosition(c *gin.Context) {
	var req dto.DepartmentPositionRequestDTO
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

	var rel models.DepartmentPosition
	err := database.DB.Unscoped().
		Where("department_id = ? AND position_id = ?", req.DepartmentID, req.PositionID).
		First(&rel).Error
	if err == nil {
		if rel.DeletedAt.Valid {
			if err := database.DB.Unscoped().
				Model(&models.DepartmentPosition{}).
				Where("id = ?", rel.ID).
				Update("deleted_at", nil).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "恢复关系失败: " + err.Error()})
				return
			}
			database.DB.Preload("Department").Preload("Position").First(&rel, rel.ID)
			c.JSON(http.StatusOK, gin.H{"code": 0, "data": rel, "msg": "关系已恢复"})
			writeLog(c, "角色管理", "新增", "恢复部门-职位关系")
			return
		}

		database.DB.Preload("Department").Preload("Position").First(&rel, rel.ID)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": rel, "msg": "该部门-职位关系已存在"})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "查询关系失败: " + err.Error()})
		return
	}

	rel = models.DepartmentPosition{DepartmentID: req.DepartmentID, PositionID: req.PositionID}
	if err := database.DB.Create(&rel).Error; err != nil {
		// 并发场景下可能出现重复插入，兜底返回已存在关系（幂等）
		var existing models.DepartmentPosition
		if findErr := database.DB.
			Where("department_id = ? AND position_id = ?", req.DepartmentID, req.PositionID).
			First(&existing).Error; findErr == nil {
			database.DB.Preload("Department").Preload("Position").First(&existing, existing.ID)
			c.JSON(http.StatusOK, gin.H{"code": 0, "data": existing, "msg": "该部门-职位关系已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	database.DB.Preload("Department").Preload("Position").First(&rel, rel.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rel})
	writeLog(c, "角色管理", "新增", "新增部门-职位关系")
}

// DeleteDepartmentPosition 删除数据
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
	if err := database.DB.Unscoped().Delete(&rel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "角色管理", "删除", "删除部门-职位关系")
}
