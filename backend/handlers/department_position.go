package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DepartmentPositionRequest struct {
	DepartmentID int `json:"department_id" binding:"required"`
	PositionID   int `json:"position_id" binding:"required"`
}

func GetDepartmentPositions(c *gin.Context) {
	var list []models.DepartmentPosition
	query := database.DB.Model(&models.DepartmentPosition{}).
		Preload("Department").
		Preload("Position")

	if deptID := c.Query("department_id"); deptID != "" {
		query = query.Where("department_id = ?", deptID)
	}
	if posID := c.Query("position_id"); posID != "" {
		query = query.Where("position_id = ?", posID)
	}

	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Order("id desc").Offset(offset).Limit(pageSize).Find(&list)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
	writeLog(c, "角色管理", "查询", "查询部门-职位关系列表")
}

func CreateDepartmentPosition(c *gin.Context) {
	var req DepartmentPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	var dept models.Department
	if err := database.DB.First(&dept, "id = ?", req.DepartmentID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "部门不存在"})
		return
	}
	var pos models.Position
	if err := database.DB.First(&pos, "id = ?", req.PositionID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "职位不存在"})
		return
	}

	var exists int64
	database.DB.Model(&models.DepartmentPosition{}).
		Where("department_id = ? AND position_id = ?", req.DepartmentID, req.PositionID).
		Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该部门-职位关系已存在"})
		return
	}

	rel := models.DepartmentPosition{DepartmentID: req.DepartmentID, PositionID: req.PositionID}
	if err := database.DB.Create(&rel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	database.DB.Preload("Department").Preload("Position").First(&rel, rel.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rel})
	writeLog(c, "角色管理", "新增", "新增部门-职位关系")
}

func DeleteDepartmentPosition(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var rel models.DepartmentPosition
	if err := database.DB.First(&rel, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "关系不存在"})
		return
	}
	if err := database.DB.Delete(&rel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "角色管理", "删除", "删除部门-职位关系")
}

func GetDepartmentPositionTree(c *gin.Context) {
	var depts []models.Department
	var relations []models.DepartmentPosition
	var positions []models.Position

	database.DB.Order("level asc, id asc").Find(&depts)
	database.DB.Preload("Position").Order("id asc").Find(&relations)
	database.DB.Order("sort_order asc, id asc").Find(&positions)

	type treeNode struct {
		ID         string      `json:"id"`
		Type       string      `json:"type"`
		Name       string      `json:"name"`
		RefID      int         `json:"ref_id"`
		RelationID int         `json:"relation_id,omitempty"`
		Children   []*treeNode `json:"children,omitempty"`
	}

	deptNodeMap := map[int]*treeNode{}
	for _, d := range depts {
		deptNodeMap[d.ID] = &treeNode{
			ID:       "dept-" + strconv.Itoa(d.ID),
			Type:     "department",
			Name:     d.Name,
			RefID:    d.ID,
			Children: []*treeNode{},
		}
	}

	roots := []*treeNode{}
	for _, d := range depts {
		node := deptNodeMap[d.ID]
		if d.ParentID != nil {
			if p, ok := deptNodeMap[*d.ParentID]; ok {
				p.Children = append(p.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	usedPos := map[int]bool{}
	for _, r := range relations {
		if dn, ok := deptNodeMap[r.DepartmentID]; ok {
			dn.Children = append(dn.Children, &treeNode{
				ID:         "rel-" + strconv.Itoa(r.ID),
				Type:       "position",
				Name:       r.Position.Name,
				RefID:      r.PositionID,
				RelationID: r.ID,
			})
			usedPos[r.PositionID] = true
		}
	}

	if len(positions) > 0 {
		unbound := &treeNode{ID: "unbound", Type: "group", Name: "未关联职位", RefID: 0, Children: []*treeNode{}}
		for _, p := range positions {
			if !usedPos[p.ID] {
				unbound.Children = append(unbound.Children, &treeNode{
					ID:    "pos-" + strconv.Itoa(p.ID),
					Type:  "position-unbound",
					Name:  p.Name,
					RefID: p.ID,
				})
			}
		}
		if len(unbound.Children) > 0 {
			roots = append(roots, unbound)
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": roots})
	writeLog(c, "角色管理", "查询", "查询部门-职位关系树")
}
