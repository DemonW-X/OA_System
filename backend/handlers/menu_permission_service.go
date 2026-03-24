package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"oa-system/database"
	"oa-system/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const menuCacheTTL = 30 * time.Minute

func menuCacheKey(employeeID int) string {
	if employeeID <= 0 {
		return "menu:tree:admin"
	}
	return fmt.Sprintf("menu:tree:employee:%d", employeeID)
}

// getMenuTreeCache 从 Redis 读菜单树缓存，未命中返回 nil
func getMenuTreeCache(employeeID int) []MenuTreeItem {
	if database.RDB == nil {
		return nil
	}
	val, err := database.RDB.Get(context.Background(), menuCacheKey(employeeID)).Result()
	if err != nil {
		return nil
	}
	var items []MenuTreeItem
	if err := json.Unmarshal([]byte(val), &items); err != nil {
		return nil
	}
	return items
}

// setMenuTreeCache 将菜单树写入 Redis 缓存
func setMenuTreeCache(employeeID int, items []MenuTreeItem) {
	if database.RDB == nil {
		return
	}
	b, err := json.Marshal(items)
	if err != nil {
		return
	}
	database.RDB.Set(context.Background(), menuCacheKey(employeeID), b, menuCacheTTL)
}

// InvalidateMenuCache 删除指定员工的菜单缓存；employeeID<=0 时删除 admin 缓存
func InvalidateMenuCache(employeeID int) {
	if database.RDB == nil {
		return
	}
	database.RDB.Del(context.Background(), menuCacheKey(employeeID))
}

// InvalidateAllMenuCache 删除所有员工菜单缓存（菜单结构变更时使用）
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
