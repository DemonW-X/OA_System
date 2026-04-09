package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/models"
	"sort"

	"github.com/gin-gonic/gin"
)

type DepartmentMenuPermissionRequest = dto.PositionMenuPermissionRequestDTO

// GetDepartmentMenuPermissions 获取数据
func GetDepartmentMenuPermissions(c *gin.Context) {
	if c.GetString("role") != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "No permission"})
		return
	}

	departmentID, ok := parseID(c)
	if !ok {
		return
	}

	var dept models.Department
	if err := database.DB.Select("id", "name").First(&dept, departmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "Department not found"})
		return
	}

	var menus []models.Menu
	database.DB.Order("id asc").Find(&menus)

	var links []models.DepartmentMenuPermission
	database.DB.Where("department_id = ?", departmentID).Find(&links)
	ids := make([]int, 0, len(links))
	for _, l := range links {
		ids = append(ids, l.MenuID)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"department_id":    departmentID,
		"department_name":  dept.Name,
		"menu_tree":        buildMenuTree(menus),
		"checked_menu_ids": ids,
	}})
}

// SetDepartmentMenuPermissions 更新数据
func SetDepartmentMenuPermissions(c *gin.Context) {
	if c.GetString("role") != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "No permission"})
		return
	}

	departmentID, ok := parseID(c)
	if !ok {
		return
	}

	var dept models.Department
	if err := database.DB.Select("id", "name").First(&dept, departmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "Department not found"})
		return
	}

	var req DepartmentMenuPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	var menus []models.Menu
	database.DB.Select("id", "parent_id").Find(&menus)
	parentMap := map[int]int{}
	valid := map[int]bool{}
	for _, m := range menus {
		parentMap[m.ID] = m.ParentID
		valid[m.ID] = true
	}

	selected := map[int]bool{}
	for _, id := range req.MenuIDs {
		if id <= 0 || !valid[id] {
			continue
		}
		selected[id] = true
		pid := parentMap[id]
		for pid > 0 {
			if !valid[pid] {
				break
			}
			if selected[pid] {
				break
			}
			selected[pid] = true
			pid = parentMap[pid]
		}
	}

	ids := make([]int, 0, len(selected))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	tx := database.DB.Begin()
	if err := tx.Where("department_id = ?", departmentID).Delete(&models.DepartmentMenuPermission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "Save failed: " + err.Error()})
		return
	}
	if len(ids) > 0 {
		links := make([]models.DepartmentMenuPermission, 0, len(ids))
		for _, id := range ids {
			links = append(links, models.DepartmentMenuPermission{
				DepartmentID: departmentID,
				MenuID:       id,
			})
		}
		if err := tx.Create(&links).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "Save failed: " + err.Error()})
			return
		}
	}
	tx.Commit()
	InvalidateAllMenuCache()

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"department_id":    departmentID,
		"checked_menu_ids": ids,
	}})
	writeLog(c, "Department Permission", "Assign", "Set department menu permissions: "+dept.Name)
}
