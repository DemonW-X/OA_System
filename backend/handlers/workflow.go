package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/models"

	"github.com/gin-gonic/gin"
)

// InitBizTypes 启动时校验 menu_workflow_configs 关联的 biz_types 记录是否完整
// 关联表是主数据源，biz_types 由关联表的增删来维护，这里只做孤立记录清理
func InitBizTypes() {
	// 查询所有关联的 biz_type_id
	var configs []models.MenuWorkflowConfig
	database.DB.Find(&configs)
	linkedIDs := map[int]bool{}
	for _, cfg := range configs {
		linkedIDs[cfg.BizTypeID] = true
	}

	// 删除 biz_types 中没有被任何菜单关联的孤立记录
	var allBiz []models.BizType
	database.DB.Find(&allBiz)
	for _, b := range allBiz {
		if !linkedIDs[b.ID] {
			database.DB.Delete(&models.BizType{}, b.ID)
		}
	}
}

// contains 执行相关业务逻辑
func contains(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

// GetBizTypes 获取所有业务类型（供适用业务下拉使用）
func GetBizTypes(c *gin.Context) {
	var list []models.BizType
	database.DB.Order("sort asc, id asc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

// CreateBizType 新增业务类型
func CreateBizType(c *gin.Context) {
	var req dto.WorkflowBizTypeCreateRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	biz := models.BizType{Code: req.Code, Name: req.Name, Sort: req.Sort}
	if err := database.DB.Create(&biz).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "业务类型编码已存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": biz})
	writeLog(c, "业务类型", "新增", "新增业务类型："+req.Name)
}

// DeleteBizType 删除业务类型
func DeleteBizType(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var biz models.BizType
	if err := database.DB.First(&biz, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "业务类型不存在"})
		return
	}
	database.DB.Delete(&biz)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "业务类型", "删除", "删除业务类型："+biz.Name)
}

type WorkflowNodeInput = dto.WorkflowNodeInputDTO
type WorkflowTemplateRequest = dto.WorkflowTemplateRequestDTO

// GetWorkflowTemplates 获取数据
func GetWorkflowTemplates(c *gin.Context) {
	var list []models.WorkflowTemplate
	query := database.DB.Model(&models.WorkflowTemplate{}).Preload("Nodes")
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Offset(offset).Limit(pageSize).Find(&list)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
	writeLog(c, "流程管理", "查询", "查询流程模板列表")
}

// GetWorkflowTemplate 获取数据
func GetWorkflowTemplate(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var tpl models.WorkflowTemplate
	if err := database.DB.Preload("Nodes").First(&tpl, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "流程模板不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": tpl})
}

// CreateWorkflowTemplate 创建数据
func CreateWorkflowTemplate(c *gin.Context) {
	var req WorkflowTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	tpl := models.WorkflowTemplate{
		Name:        req.Name,
		Description: req.Description,
		BizType:     req.BizType,
	}
	tx := database.DB.Begin()
	if err := tx.Create(&tpl).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	for i, n := range req.Nodes {
		approveType := n.ApproveType
		if approveType == "" {
			approveType = "or"
		}
		node := models.WorkflowNode{
			TemplateID:    tpl.ID,
			Sort:          i,
			Name:          n.Name,
			ApproveType:   approveType,
			Approvers:     n.Approvers,
			Conditions:    n.Conditions,
			AllowSkip:     n.AllowSkip,
			AllowTransfer: n.AllowTransfer,
			ParentIDs:     n.ParentIDs,
		}
		if err := tx.Create(&node).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建节点失败"})
			return
		}
	}
	tx.Commit()
	database.DB.Preload("Nodes").First(&tpl, tpl.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": tpl})
	writeLog(c, "流程管理", "新增", "新增流程模板："+req.Name)
}

// UpdateWorkflowTemplate 更新数据
func UpdateWorkflowTemplate(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var tpl models.WorkflowTemplate
	if err := database.DB.First(&tpl, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "流程模板不存在"})
		return
	}
	var req WorkflowTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	tx := database.DB.Begin()
	tpl.Name = req.Name
	tpl.Description = req.Description
	tpl.BizType = req.BizType
	if err := tx.Save(&tpl).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败"})
		return
	}
	// 删除旧节点，重新插入
	if err := tx.Where("template_id = ?", id).Delete(&models.WorkflowNode{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新节点失败"})
		return
	}
	for i, n := range req.Nodes {
		approveType := n.ApproveType
		if approveType == "" {
			approveType = "or"
		}
		node := models.WorkflowNode{
			TemplateID:    id,
			Sort:          i,
			Name:          n.Name,
			ApproveType:   approveType,
			Approvers:     n.Approvers,
			Conditions:    n.Conditions,
			AllowSkip:     n.AllowSkip,
			AllowTransfer: n.AllowTransfer,
			ParentIDs:     n.ParentIDs,
		}
		if err := tx.Create(&node).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新节点失败"})
			return
		}
	}
	tx.Commit()
	database.DB.Preload("Nodes").First(&tpl, id)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": tpl})
	writeLog(c, "流程管理", "修改", "修改流程模板："+req.Name)
}

// DeleteWorkflowTemplate 删除数据
func DeleteWorkflowTemplate(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var tpl models.WorkflowTemplate
	if err := database.DB.First(&tpl, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "流程模板不存在"})
		return
	}
	tx := database.DB.Begin()
	tx.Where("template_id = ?", id).Delete(&models.WorkflowNode{})
	if err := tx.Delete(&tpl).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "流程管理", "删除", "删除流程模板："+tpl.Name)
}
