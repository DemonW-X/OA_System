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
	SortOrder    int    `json:"sort_order"`
	DepartmentID int    `json:"department_id"`
	Remark       string `json:"remark"`
}

func GetPositions(c *gin.Context) {
	var list []models.Position
	query := database.DB.Model(&models.Position{})
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		keyword = strings.TrimSpace(c.Query("name"))
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR remark LIKE ?", like, like)
	}

	// 按部门过滤：通过关系表筛选职位
	if deptID := c.Query("department_id"); deptID != "" {
		query = query.Where("id IN (?)",
			database.DB.Model(&models.DepartmentPosition{}).
				Select("position_id").
				Where("department_id = ?", deptID),
		)
	}

	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Order("sort_order asc, id asc").Offset(offset).Limit(pageSize).Find(&list)

	// 回填前端兼容字段：department_id / department
	for i := range list {
		var rel models.DepartmentPosition
		err := database.DB.Preload("Department").
			Where("position_id = ?", list[i].ID).
			Order("id asc").
			First(&rel).Error
		if err == nil {
			list[i].DepartmentID = rel.DepartmentID
			list[i].Department = &rel.Department
		}
	}

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
	if err := database.DB.First(&pos, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "职位不存在"})
		return
	}

	var rel models.DepartmentPosition
	err := database.DB.Preload("Department").
		Where("position_id = ?", pos.ID).
		Order("id asc").
		First(&rel).Error
	if err == nil {
		pos.DepartmentID = rel.DepartmentID
		pos.Department = &rel.Department
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

	// 若传了部门，则做“部门内同名唯一”校验
	if req.DepartmentID > 0 {
		var cnt int64
		database.DB.Model(&models.Position{}).
			Joins("JOIN department_positions dp ON dp.position_id = positions.id AND dp.deleted_at IS NULL").
			Where("positions.name = ? AND dp.department_id = ?", req.Name, req.DepartmentID).
			Count(&cnt)
		if cnt > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该部门下职位名称已存在"})
			return
		}
	}

	pos := models.Position{Name: req.Name, SortOrder: req.SortOrder, Remark: req.Remark}
	tx := database.DB.Begin()
	if err := tx.Create(&pos).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}

	if req.DepartmentID > 0 {
		rel := models.DepartmentPosition{DepartmentID: req.DepartmentID, PositionID: pos.ID}
		if err := tx.Create(&rel).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建部门-职位关系失败: " + err.Error()})
			return
		}
		var dept models.Department
		if err := tx.First(&dept, "id = ?", req.DepartmentID).Error; err == nil {
			pos.DepartmentID = req.DepartmentID
			pos.Department = &dept
		}
	}

	tx.Commit()
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

	if req.DepartmentID > 0 {
		var cnt int64
		database.DB.Model(&models.Position{}).
			Joins("JOIN department_positions dp ON dp.position_id = positions.id AND dp.deleted_at IS NULL").
			Where("positions.id <> ? AND positions.name = ? AND dp.department_id = ?", id, req.Name, req.DepartmentID).
			Count(&cnt)
		if cnt > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该部门下职位名称已存在"})
			return
		}
	}

	tx := database.DB.Begin()
	pos.Name = req.Name
	pos.SortOrder = req.SortOrder
	pos.Remark = req.Remark
	if err := tx.Save(&pos).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败: " + err.Error()})
		return
	}

	// 维护部门-职位关系（这里按“一个职位当前只挂一个部门”处理；未来可扩展多部门）
	if req.DepartmentID > 0 {
		if err := tx.Where("position_id = ?", pos.ID).Delete(&models.DepartmentPosition{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新部门-职位关系失败"})
			return
		}
		rel := models.DepartmentPosition{DepartmentID: req.DepartmentID, PositionID: pos.ID}
		if err := tx.Create(&rel).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新部门-职位关系失败: " + err.Error()})
			return
		}
	}

	tx.Commit()
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

	tx := database.DB.Begin()
	if err := tx.Where("position_id = ?", id).Delete(&models.DepartmentPosition{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除职位关系失败"})
		return
	}
	if err := tx.Delete(&pos).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "职位管理", "删除", "删除职位："+pos.Name)
}
