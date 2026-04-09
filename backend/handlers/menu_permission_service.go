package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"oa-system/database"
	"oa-system/models"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const menuCacheTTL = 30 * time.Minute

const (
	menuScopeAdmin    = "admin"
	menuScopeEmployee = "employee"
	menuScopePosition = "position"
	menuScopeEmpty    = "empty"
)

type menuPermissionContext struct {
	scope        string
	employeeID   int
	departmentID int
	positionID   int
}

// needFilter 执行相关业务逻辑
func (ctx menuPermissionContext) needFilter() bool {
	return ctx.scope != menuScopeAdmin
}

// cacheToken 执行相关业务逻辑
func (ctx menuPermissionContext) cacheToken() string {
	switch ctx.scope {
	case menuScopeAdmin:
		return "admin"
	case menuScopePosition:
		return fmt.Sprintf("position:%d", ctx.positionID)
	case menuScopeEmployee:
		return fmt.Sprintf("employee:%d", ctx.employeeID)
	default:
		return "empty"
	}
}

// menuCacheKey 执行相关业务逻辑
func menuCacheKey(scopeToken string) string {
	return fmt.Sprintf("menu:tree:scope:%s", scopeToken)
}

// getMenuTreeCache 获取数据
func getMenuTreeCache(scopeToken string) []MenuTreeItem {
	if database.RDB == nil {
		return nil
	}
	val, err := database.RDB.Get(context.Background(), menuCacheKey(scopeToken)).Result()
	if err != nil {
		return nil
	}
	var items []MenuTreeItem
	if err := json.Unmarshal([]byte(val), &items); err != nil {
		return nil
	}
	return items
}

// setMenuTreeCache 更新数据
func setMenuTreeCache(scopeToken string, items []MenuTreeItem) {
	if database.RDB == nil {
		return
	}
	b, err := json.Marshal(items)
	if err != nil {
		return
	}
	database.RDB.Set(context.Background(), menuCacheKey(scopeToken), b, menuCacheTTL)
}

// InvalidateMenuCache 执行相关业务逻辑
func InvalidateMenuCache(scopeToken string) {
	if database.RDB == nil {
		return
	}
	database.RDB.Del(context.Background(), menuCacheKey(scopeToken))
}

// InvalidateAllMenuCache 执行相关业务逻辑
func InvalidateAllMenuCache() {
	if database.RDB == nil {
		return
	}
	ctx := context.Background()
	keys, err := database.RDB.Keys(ctx, "menu:tree:*").Result()
	if err != nil || len(keys) == 0 {
		return
	}
	database.RDB.Del(ctx, keys...)
}

// getPositionAssignedMenuIDs 获取数据
func getPositionAssignedMenuIDs(positionID int) []int {
	if positionID <= 0 {
		return []int{}
	}
	var links []models.PositionMenuPermission
	database.DB.Where("position_id = ?", positionID).Find(&links)
	if len(links) == 0 {
		return []int{}
	}
	ids := make([]int, 0, len(links))
	for _, l := range links {
		ids = append(ids, l.MenuID)
	}
	return ids
}

// getDepartmentAssignedMenuIDs 获取数据
func getDepartmentAssignedMenuIDs(departmentID int) []int {
	if departmentID <= 0 {
		return []int{}
	}
	var links []models.DepartmentMenuPermission
	database.DB.Where("department_id = ?", departmentID).Find(&links)
	if len(links) == 0 {
		return []int{}
	}
	ids := make([]int, 0, len(links))
	for _, l := range links {
		ids = append(ids, l.MenuID)
	}
	return ids
}

// getEmployeeAssignedMenuIDs 获取数据
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

// mergeUniqueMenuIDs 执行相关业务逻辑
func mergeUniqueMenuIDs(groups ...[]int) []int {
	if len(groups) == 0 {
		return []int{}
	}
	set := map[int]struct{}{}
	for _, group := range groups {
		for _, id := range group {
			if id > 0 {
				set[id] = struct{}{}
			}
		}
	}
	if len(set) == 0 {
		return []int{}
	}
	ids := make([]int, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

// includeParentMenuIDs 执行相关业务逻辑
func includeParentMenuIDs(ids []int) []int {
	if len(ids) == 0 {
		return []int{}
	}
	var menus []models.Menu
	database.DB.Select("id", "parent_id").Find(&menus)
	if len(menus) == 0 {
		return ids
	}

	parentMap := make(map[int]int, len(menus))
	valid := make(map[int]struct{}, len(menus))
	for _, m := range menus {
		parentMap[m.ID] = m.ParentID
		valid[m.ID] = struct{}{}
	}

	set := map[int]struct{}{}
	for _, id := range ids {
		if _, ok := valid[id]; !ok {
			continue
		}
		set[id] = struct{}{}
		pid := parentMap[id]
		for pid > 0 {
			if _, ok := valid[pid]; !ok {
				break
			}
			if _, exists := set[pid]; exists {
				break
			}
			set[pid] = struct{}{}
			pid = parentMap[pid]
		}
	}

	if len(set) == 0 {
		return []int{}
	}
	result := make([]int, 0, len(set))
	for id := range set {
		result = append(result, id)
	}
	sort.Ints(result)
	return result
}

// resolveMenuPermissionContext 执行相关业务逻辑
func resolveMenuPermissionContext(c *gin.Context) menuPermissionContext {
	if positionIDStr := strings.TrimSpace(c.Query("position_id")); positionIDStr != "" {
		positionID, err := strconv.Atoi(positionIDStr)
		if err != nil || positionID <= 0 {
			return menuPermissionContext{scope: menuScopeEmpty}
		}
		return menuPermissionContext{
			scope:      menuScopePosition,
			positionID: positionID,
		}
	}

	if employeeIDStr := strings.TrimSpace(c.Query("employee_id")); employeeIDStr != "" {
		employeeID, err := strconv.Atoi(employeeIDStr)
		if err != nil || employeeID <= 0 {
			return menuPermissionContext{scope: menuScopeEmpty}
		}
		var emp models.Employee
		if err := database.DB.Select("id", "department_id", "position_id").Where("id = ?", employeeID).First(&emp).Error; err != nil {
			return menuPermissionContext{scope: menuScopeEmpty}
		}
		return menuPermissionContext{
			scope:        menuScopeEmployee,
			employeeID:   emp.ID,
			departmentID: emp.DepartmentID,
			positionID:   emp.PositionID,
		}
	}

	if c.GetString("role") == "admin" {
		return menuPermissionContext{scope: menuScopeAdmin}
	}

	userID := c.GetInt("userID")
	if userID <= 0 {
		return menuPermissionContext{scope: menuScopeEmpty}
	}

	var emp models.Employee
	if err := database.DB.Select("id", "department_id", "position_id").Where("user_id = ?", userID).First(&emp).Error; err != nil {
		return menuPermissionContext{scope: menuScopeEmpty}
	}
	return menuPermissionContext{
		scope:        menuScopeEmployee,
		employeeID:   emp.ID,
		departmentID: emp.DepartmentID,
		positionID:   emp.PositionID,
	}
}

// getMenuIDsByContext 获取数据
func getMenuIDsByContext(ctx menuPermissionContext) []int {
	switch ctx.scope {
	case menuScopeAdmin:
		return []int{}
	case menuScopePosition:
		return includeParentMenuIDs(getPositionAssignedMenuIDs(ctx.positionID))
	case menuScopeEmployee:
		ids := mergeUniqueMenuIDs(
			getDepartmentAssignedMenuIDs(ctx.departmentID),
			getPositionAssignedMenuIDs(ctx.positionID),
			getEmployeeAssignedMenuIDs(ctx.employeeID),
		)
		return includeParentMenuIDs(ids)
	default:
		return []int{}
	}
}

// Compatibility wrapper: existing callers still use employee naming.
func resolveMenuFilterEmployeeID(c *gin.Context) (int, bool) {
	ctx := resolveMenuPermissionContext(c)
	if !ctx.needFilter() {
		return 0, false
	}
	if ctx.scope == menuScopePosition {
		return ctx.positionID, true
	}
	return ctx.employeeID, true
}

// applyMenuPermissionScope 执行相关业务逻辑
func applyMenuPermissionScope(c *gin.Context, query *gorm.DB) (*gorm.DB, bool) {
	ctx := resolveMenuPermissionContext(c)
	if !ctx.needFilter() {
		return query, false
	}
	ids := getMenuIDsByContext(ctx)
	if len(ids) == 0 {
		return nil, true
	}
	return query.Where("id IN ?", ids), false
}
