package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/models"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
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

func GetMenus(c *gin.Context) {
	var list []models.Menu
	query := database.DB.Model(&models.Menu{})
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("name LIKE ? OR path LIKE ?", like, like)
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
