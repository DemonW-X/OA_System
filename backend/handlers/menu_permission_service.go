package handlers

import (
	"oa-system/database"
	"oa-system/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getEmployeeAssignedMenuIDs 读取员工已分配菜单ID（公共方法）
func getEmployeeAssignedMenuIDs(employeeID int) []int {
	if employeeID <= 0 {
		return []int{}
	}
	var links []models.EmployeeMenuPermission
	database.DB.Where("employee_id = ?", employeeID).Find(&links)
	if len(links) == 0 {
		return []int{}
	}
	ids := make([]int, 0, len(links))
	for _, l := range links {
		ids = append(ids, l.MenuID)
	}
	return ids
}

// resolveMenuFilterEmployeeID 解析当前请求应按哪个员工ID过滤菜单
// 返回: employeeID, needFilter
func resolveMenuFilterEmployeeID(c *gin.Context) (int, bool) {
	if employeeIDStr := strings.TrimSpace(c.Query("employee_id")); employeeIDStr != "" {
		employeeID, err := strconv.Atoi(employeeIDStr)
		if err != nil || employeeID <= 0 {
			return 0, false
		}
		return employeeID, true
	}

	role := c.GetString("role")
	if role == "admin" {
		return 0, false
	}
	userID := c.GetInt("userID")
	if userID <= 0 {
		return 0, true
	}
	var emp models.Employee
	err := database.DB.Select("id").Where("user_id = ?", userID).First(&emp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, true
		}
		return 0, false
	}
	return emp.ID, true
}

// applyMenuPermissionScope 将员工菜单权限范围应用到查询（公共方法）
func applyMenuPermissionScope(c *gin.Context, query *gorm.DB) (*gorm.DB, bool) {
	employeeID, needFilter := resolveMenuFilterEmployeeID(c)
	if !needFilter {
		return query, false
	}
	ids := getEmployeeAssignedMenuIDs(employeeID)
	if len(ids) == 0 {
		return nil, true
	}
	return query.Where("id IN ?", ids), false
}
