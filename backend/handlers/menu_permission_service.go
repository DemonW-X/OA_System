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

func menuCacheKey(positionID int) string {
	if positionID <= 0 {
		return "menu:tree:admin"
	}
	return fmt.Sprintf("menu:tree:position:%d", positionID)
}

func getMenuTreeCache(positionID int) []MenuTreeItem {
	if database.RDB == nil {
		return nil
	}
	val, err := database.RDB.Get(context.Background(), menuCacheKey(positionID)).Result()
	if err != nil {
		return nil
	}
	var items []MenuTreeItem
	if err := json.Unmarshal([]byte(val), &items); err != nil {
		return nil
	}
	return items
}

func setMenuTreeCache(positionID int, items []MenuTreeItem) {
	if database.RDB == nil {
		return
	}
	b, err := json.Marshal(items)
	if err != nil {
		return
	}
	database.RDB.Set(context.Background(), menuCacheKey(positionID), b, menuCacheTTL)
}

func InvalidateMenuCache(positionID int) {
	if database.RDB == nil {
		return
	}
	database.RDB.Del(context.Background(), menuCacheKey(positionID))
}

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

func resolveMenuFilterPositionID(c *gin.Context) (int, bool) {
	if positionIDStr := strings.TrimSpace(c.Query("position_id")); positionIDStr != "" {
		positionID, err := strconv.Atoi(positionIDStr)
		if err != nil || positionID <= 0 {
			return 0, true
		}
		return positionID, true
	}

	if employeeIDStr := strings.TrimSpace(c.Query("employee_id")); employeeIDStr != "" {
		employeeID, err := strconv.Atoi(employeeIDStr)
		if err != nil || employeeID <= 0 {
			return 0, true
		}
		var emp models.Employee
		err = database.DB.Select("id", "position_id").Where("id = ?", employeeID).First(&emp).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, true
			}
			return 0, false
		}
		if emp.PositionID <= 0 {
			return 0, true
		}
		return emp.PositionID, true
	}

	if c.GetString("role") == "admin" {
		return 0, false
	}

	userID := c.GetInt("userID")
	if userID <= 0 {
		return 0, true
	}

	var emp models.Employee
	err := database.DB.Select("id", "position_id").Where("user_id = ?", userID).First(&emp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, true
		}
		return 0, false
	}
	if emp.PositionID <= 0 {
		return 0, true
	}
	return emp.PositionID, true
}

// Compatibility wrapper: existing callers still use employee naming.
func getEmployeeAssignedMenuIDs(employeeID int) []int {
	return getPositionAssignedMenuIDs(employeeID)
}

// Compatibility wrapper: existing callers still use employee naming.
func resolveMenuFilterEmployeeID(c *gin.Context) (int, bool) {
	return resolveMenuFilterPositionID(c)
}

func applyMenuPermissionScope(c *gin.Context, query *gorm.DB) (*gorm.DB, bool) {
	positionID, needFilter := resolveMenuFilterPositionID(c)
	if !needFilter {
		return query, false
	}
	ids := getPositionAssignedMenuIDs(positionID)
	if len(ids) == 0 {
		return nil, true
	}
	return query.Where("id IN ?", ids), false
}
