package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/models"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type MenuRequest = dto.MenuRequestDTO
type MenuTreeItem = dto.MenuTreeItemDTO

// buildMenuTree 构建业务数据
func buildMenuTree(list []models.Menu) []MenuTreeItem {
	// 预加载所有审批流配置，避免 N+1
	var configs []models.MenuWorkflowConfig
	database.DB.Find(&configs)
	cfgByMenu := map[int]models.MenuWorkflowConfig{}
	for _, cfg := range configs {
		cfgByMenu[cfg.MenuID] = cfg
	}
	var bizTypes []models.BizType
	database.DB.Find(&bizTypes)
	bizByID := map[int]models.BizType{}
	for _, b := range bizTypes {
		bizByID[b.ID] = b
	}

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
			item := MenuTreeItem{
				ID:       m.ID,
				Name:     m.Name,
				Icon:     m.Icon,
				Path:     m.Path,
				SortCode: m.SortCode,
				ParentID: m.ParentID,
				Visible:  m.Visible,
				Remark:   m.Remark,
				Children: build(m.ID),
			}
			if cfg, ok := cfgByMenu[m.ID]; ok {
				if biz, ok2 := bizByID[cfg.BizTypeID]; ok2 {
					item.BizCode = biz.Code
					item.BizName = biz.Name
					item.BizSort = biz.Sort
				}
			}
			res = append(res, item)
		}
		return res
	}

	return build(0)
}

// GetMenus 获取数据
func GetMenus(c *gin.Context) {
	kw := strings.TrimSpace(c.Query("keyword"))
	useCache := kw == "" && c.DefaultQuery("tree", "1") == "1"

	if useCache {
		permCtx := resolveMenuPermissionContext(c)
		cacheToken := permCtx.cacheToken()
		if cached := getMenuTreeCache(cacheToken); cached != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "data": cached})
			return
		}
		var list []models.Menu
		query := database.DB.Model(&models.Menu{})
		if permCtx.needFilter() {
			ids := getMenuIDsByContext(permCtx)
			if len(ids) == 0 {
				c.JSON(http.StatusOK, gin.H{"code": 0, "data": []MenuTreeItem{}})
				return
			}
			query = query.Where("id IN ?", ids)
		}
		query.Order("sort_code asc, id asc").Find(&list)
		tree := buildMenuTree(list)
		setMenuTreeCache(cacheToken, tree)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": tree})
		return
	}

	// 有关键词或非树形模式，不走缓存
	var list []models.Menu
	query := database.DB.Model(&models.Menu{})
	if kw != "" {
		like := "%" + kw + "%"
		query = query.Where("name LIKE ? OR path LIKE ?", like, like)
	}
	if q, empty := applyMenuPermissionScope(c, query); empty {
		if c.DefaultQuery("tree", "1") == "1" {
			c.JSON(http.StatusOK, gin.H{"code": 0, "data": []MenuTreeItem{}})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 0, "data": []models.Menu{}})
		}
		return
	} else {
		query = q
	}
	query.Order("sort_code asc, id asc").Find(&list)
	if c.DefaultQuery("tree", "1") == "1" {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": buildMenuTree(list)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

// GetMenu 获取数据
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

// CreateMenu 创建数据
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
	if (req.BizCode == "" && req.BizName != "") || (req.BizCode != "" && req.BizName == "") {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "业务编码和业务名称需同时填写或同时留空"})
		return
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
	// 同步审批流配置
	syncMenuWorkflowConfig(m.ID, req.BizCode, req.BizName, req.BizSort)
	InvalidateAllMenuCache()
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": m})
	writeLog(c, "菜单管理", "新增", "新增菜单："+req.Name)
}

// UpdateMenu 更新数据
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
	if (req.BizCode == "" && req.BizName != "") || (req.BizCode != "" && req.BizName == "") {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "业务编码和业务名称需同时填写或同时留空"})
		return
	}
	if err := database.DB.Save(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败: " + err.Error()})
		return
	}
	// 同步审批流配置
	syncMenuWorkflowConfig(m.ID, req.BizCode, req.BizName, req.BizSort)
	InvalidateAllMenuCache()
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": m})
	writeLog(c, "菜单管理", "修改", "修改菜单："+req.Name)
}

// DeleteMenu 删除数据
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
	DeleteMenuWorkflowConfigByMenuID(id)
	if err := database.DB.Delete(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	InvalidateAllMenuCache()
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "菜单管理", "删除", "删除菜单："+m.Name)
}
