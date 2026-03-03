package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/models"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MenuRequest struct {
	Name     string `json:"name" binding:"required"`
	Icon     string `json:"icon"`
	Path     string `json:"path"`
	SortCode int    `json:"sort_code"`
	ParentID int    `json:"parent_id"`
	Visible  *bool  `json:"visible"`
	Remark   string `json:"remark"`
}

type MenuTreeItem struct {
	ID       int            `json:"id"`
	Name     string         `json:"name"`
	Icon     string         `json:"icon"`
	Path     string         `json:"path"`
	SortCode int            `json:"sort_code"`
	ParentID int            `json:"parent_id"`
	Visible  bool           `json:"visible"`
	Remark   string         `json:"remark"`
	Children []MenuTreeItem `json:"children"`
}

type EmployeeMenuPermissionRequest struct {
	MenuIDs []int `json:"menu_ids"`
}

func buildMenuTree(list []models.Menu) []MenuTreeItem {
	byParent := map[int][]models.Menu{}
	for _, m := range list {
		byParent[m.ParentID] = append(byParent[m.ParentID], m)
	}

	var build func(parentID int) []MenuTreeItem
	build = func(parentID int) []MenuTreeItem {
		src := byParent[parentID]
		sort.Slice(src, func(i, j int) bool {
			if src[i].SortCode == src[j].SortCode {
				return src[i].ID < src[j].ID
			}
			return src[i].SortCode < src[j].SortCode
		})

		res := make([]MenuTreeItem, 0, len(src))
		for _, m := range src {
			res = append(res, MenuTreeItem{
				ID:       m.ID,
				Name:     m.Name,
				Icon:     m.Icon,
				Path:     m.Path,
				SortCode: m.SortCode,
				ParentID: m.ParentID,
				Visible:  m.Visible,
				Remark:   m.Remark,
				Children: build(m.ID),
			})
		}
		return res
	}

	return build(0)
}

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

func GetMenus(c *gin.Context) {
	var list []models.Menu
	query := database.DB.Model(&models.Menu{})
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("name LIKE ? OR path LIKE ?", like, like)
	}

	if employeeID, needFilter := resolveMenuFilterEmployeeID(c); needFilter {
		ids := getEmployeeAssignedMenuIDs(employeeID)
		if len(ids) == 0 {
			if c.DefaultQuery("tree", "1") == "1" {
				c.JSON(http.StatusOK, gin.H{"code": 0, "data": []MenuTreeItem{}})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": 0, "data": []models.Menu{}})
			}
			return
		}
		query = query.Where("id IN ?", ids)
	}

	query.Order("sort_code asc, id asc").Find(&list)

	if c.DefaultQuery("tree", "1") == "1" {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": buildMenuTree(list)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func GetMenu(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var m models.Menu
	if err := database.DB.First(&m, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "菜单不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": m})
}

func CreateMenu(c *gin.Context) {
	var req MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if req.ParentID > 0 {
		var p models.Menu
		if err := database.DB.First(&p, req.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "父级菜单不存在"})
			return
		}
	}
	visible := true
	if req.Visible != nil {
		visible = *req.Visible
	}
	m := models.Menu{
		Name:     req.Name,
		Icon:     req.Icon,
		Path:     req.Path,
		SortCode: req.SortCode,
		ParentID: req.ParentID,
		Visible:  visible,
		Remark:   req.Remark,
	}
	if err := database.DB.Create(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": m})
	writeLog(c, "菜单管理", "新增", "新增菜单："+req.Name)
}

func UpdateMenu(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var m models.Menu
	if err := database.DB.First(&m, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "菜单不存在"})
		return
	}
	var req MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if req.ParentID == id {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "父级菜单不能是自己"})
		return
	}
	if req.ParentID > 0 {
		var p models.Menu
		if err := database.DB.First(&p, req.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "父级菜单不存在"})
			return
		}
	}

	m.Name = req.Name
	m.Icon = req.Icon
	m.Path = req.Path
	m.SortCode = req.SortCode
	m.ParentID = req.ParentID
	if req.Visible != nil {
		m.Visible = *req.Visible
	}
	m.Remark = req.Remark
	if err := database.DB.Save(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": m})
	writeLog(c, "菜单管理", "修改", "修改菜单："+req.Name)
}

func DeleteMenu(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var m models.Menu
	if err := database.DB.First(&m, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "菜单不存在"})
		return
	}
	var childCount int64
	database.DB.Model(&models.Menu{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "请先删除子菜单"})
		return
	}
	if err := database.DB.Delete(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "菜单管理", "删除", "删除菜单："+m.Name)
}

func GetEmployeeMenuPermissions(c *gin.Context) {
	if c.GetString("role") != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "无权限"})
		return
	}
	employeeID, ok := parseID(c)
	if !ok {
		return
	}
	var emp models.Employee
	if err := database.DB.Select("id", "name").First(&emp, employeeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "员工不存在"})
		return
	}

	var menus []models.Menu
	database.DB.Order("sort_code asc, id asc").Find(&menus)
	ids := getEmployeeAssignedMenuIDs(employeeID)

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"employee_id":      employeeID,
		"employee_name":    emp.Name,
		"menu_tree":        buildMenuTree(menus),
		"checked_menu_ids": ids,
	}})
}

func SetEmployeeMenuPermissions(c *gin.Context) {
	if c.GetString("role") != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "无权限"})
		return
	}
	employeeID, ok := parseID(c)
	if !ok {
		return
	}
	var emp models.Employee
	if err := database.DB.Select("id", "name").First(&emp, employeeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "员工不存在"})
		return
	}

	var req EmployeeMenuPermissionRequest
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
	if err := tx.Where("employee_id = ?", employeeID).Delete(&models.EmployeeMenuPermission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "保存失败"})
		return
	}
	if len(ids) > 0 {
		links := make([]models.EmployeeMenuPermission, 0, len(ids))
		for _, id := range ids {
			links = append(links, models.EmployeeMenuPermission{EmployeeID: employeeID, MenuID: id})
		}
		if err := tx.Create(&links).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "保存失败: " + err.Error()})
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"employee_id": employeeID, "checked_menu_ids": ids}})
	writeLog(c, "员工权限", "分配", "设置员工菜单权限："+emp.Name)
}
