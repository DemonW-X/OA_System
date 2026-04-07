package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/models"

	"github.com/gin-gonic/gin"
)

func calcDepartmentLevel(parentID *int) (int, error) {
	if parentID == nil {
		return 1, nil
	}
	var parent models.Department
	if err := database.DB.First(&parent, "id = ?", *parentID).Error; err != nil {
		return 0, err
	}
	return parent.Level + 1, nil
}

func hasDepartmentCycle(currentID int, parentID *int) bool {
	if parentID == nil {
		return false
	}
	if *parentID == currentID {
		return true
	}
	seen := map[int]bool{}
	pid := parentID
	for pid != nil {
		if seen[*pid] {
			return true
		}
		seen[*pid] = true
		if *pid == currentID {
			return true
		}
		var p models.Department
		if err := database.DB.Select("id", "parent_id").First(&p, "id = ?", *pid).Error; err != nil {
			return false
		}
		pid = p.ParentID
	}
	return false
}

func updateChildrenDepartmentLevel(parentID int, parentLevel int) {
	var children []models.Department
	database.DB.Where("parent_id = ?", parentID).Find(&children)
	for _, ch := range children {
		newLevel := parentLevel + 1
		if ch.Level != newLevel {
			database.DB.Model(&models.Department{}).Where("id = ?", ch.ID).Update("level", newLevel)
		}
		updateChildrenDepartmentLevel(ch.ID, newLevel)
	}
}

func GetDepartments(c *gin.Context) {
	var list []models.Department
	query := database.DB.Model(&models.Department{}).Preload("Parent")
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if parentID := c.Query("parent_id"); parentID != "" {
		if parentID == "0" {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", parentID)
		}
	}
	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Order("id asc").Offset(offset).Limit(pageSize).Find(&list)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
	writeLog(c, "部门管理", "查询", "查询部门列表")
}

func CreateDepartment(c *gin.Context) {
	var req dto.DepartmentRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if req.ParentID != nil && *req.ParentID <= 0 {
		req.ParentID = nil
	}
	level, err := calcDepartmentLevel(req.ParentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "上级部门不存在"})
		return
	}
	dept := models.Department{Name: req.Name, ParentID: req.ParentID, Level: level, Remark: req.Remark}
	if err := database.DB.Create(&dept).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": dept})
	writeLog(c, "部门管理", "新增", "新增部门："+req.Name)
}

func UpdateDepartment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var dept models.Department
	if err := database.DB.First(&dept, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "部门不存在"})
		return
	}
	var req dto.DepartmentRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if req.ParentID != nil && *req.ParentID <= 0 {
		req.ParentID = nil
	}
	if hasDepartmentCycle(id, req.ParentID) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "上级部门设置非法：不能形成循环层级"})
		return
	}
	level, err := calcDepartmentLevel(req.ParentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "上级部门不存在"})
		return
	}
	dept.Name = req.Name
	dept.ParentID = req.ParentID
	dept.Level = level
	dept.Remark = req.Remark
	if err := database.DB.Save(&dept).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败: " + err.Error()})
		return
	}
	updateChildrenDepartmentLevel(dept.ID, dept.Level)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": dept})
	writeLog(c, "部门管理", "修改", "修改部门："+req.Name)
}

func DeleteDepartment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var dept models.Department
	if err := database.DB.First(&dept, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "部门不存在"})
		return
	}

	var childCount int64
	database.DB.Model(&models.Department{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "请先删除或迁移下级部门后再删除"})
		return
	}

	tx := database.DB.Begin()

	// 1) 找出该部门下关联的职位
	var positionIDs []int
	if err := tx.Model(&models.DepartmentPosition{}).
		Where("department_id = ?", id).
		Pluck("position_id", &positionIDs).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "查询部门职位失败"})
		return
	}

	// 2) 找出该部门下需要删除的员工（部门下员工 + 这些职位下员工）
	empQuery := tx.Model(&models.Employee{}).Where("department_id = ?", id)
	if len(positionIDs) > 0 {
		empQuery = empQuery.Or("position_id IN ?", positionIDs)
	}

	var employees []models.Employee
	if err := empQuery.Find(&employees).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "查询部门员工失败"})
		return
	}

	// 3) 删除员工对应的登录用户
	var userIDs []int
	for _, e := range employees {
		if e.UserID > 0 {
			userIDs = append(userIDs, e.UserID)
		}
	}
	if len(userIDs) > 0 {
		if err := tx.Where("id IN ?", userIDs).Delete(&models.User{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除员工账号失败"})
			return
		}
	}

	// 4) 删除员工
	if err := empQuery.Delete(&models.Employee{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除部门员工失败"})
		return
	}

	// 5) 删除部门-职位关系（不删除职位主数据）
	if err := tx.Where("department_id = ?", id).Delete(&models.DepartmentPosition{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除部门职位关系失败"})
		return
	}

	// 6) 删除部门
	if err := tx.Delete(&dept).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除部门失败"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "部门管理", "删除", "删除部门："+dept.Name)
}
