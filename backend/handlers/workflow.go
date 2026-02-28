package handlers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"oa-system/database"
	"oa-system/models"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

var routeRegisterRegex = regexp.MustCompile(`rg\.(GET|POST|PUT|DELETE)\("([^"]+)",\s*handlers\.([A-Za-z0-9_]+)\)`)

// InitBizTypes 根据现有 routes + handlers 动态初始化业务类型（已存在则更新）
func InitBizTypes() {
	list := discoverBizTypesFromRoutes()
	for _, b := range list {
		var existed models.BizType
		err := database.DB.Where("code = ?", b.Code).First(&existed).Error
		if err != nil {
			database.DB.Create(&b)
			continue
		}
		existed.Name = b.Name
		existed.Sort = b.Sort
		database.DB.Save(&existed)
	}
}

func discoverBizTypesFromRoutes() []models.BizType {
	handlerModuleMap := buildHandlerModuleMap()

	type routeBiz struct {
		pathBase string
		hasPost  bool
		handlers []string
	}
	bizMap := map[string]*routeBiz{}

	routeFiles, err := os.ReadDir("./routes")
	if err != nil {
		return []models.BizType{}
	}
	for _, f := range routeFiles {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".go") {
			continue
		}
		content, err := os.ReadFile(filepath.Join("./routes", f.Name()))
		if err != nil {
			continue
		}
		matches := routeRegisterRegex.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			method, path, handlerFn := m[1], m[2], m[3]
			base := routeBase(path)
			if base == "" || shouldIgnoreBizBase(base) {
				continue
			}
			item, ok := bizMap[base]
			if !ok {
				item = &routeBiz{pathBase: base}
				bizMap[base] = item
			}
			if method == "POST" {
				item.hasPost = true
			}
			if !contains(item.handlers, handlerFn) {
				item.handlers = append(item.handlers, handlerFn)
			}
		}
	}

	bases := make([]string, 0, len(bizMap))
	for base, item := range bizMap {
		if item.hasPost {
			bases = append(bases, base)
		}
	}
	sort.Strings(bases)

	result := make([]models.BizType, 0, len(bases))
	for i, base := range bases {
		item := bizMap[base]
		name := displayNameFromHandlers(item.handlers, handlerModuleMap)
		if name == "" {
			name = strings.ReplaceAll(base, "-", " ")
		}
		result = append(result, models.BizType{
			Code: normalizeBizCode(base),
			Name: name,
			Sort: i + 1,
		})
	}
	return result
}

func buildHandlerModuleMap() map[string]string {
	result := map[string]string{}
	files, err := os.ReadDir("./handlers")
	if err != nil {
		return result
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".go") {
			continue
		}
		path := filepath.Join("./handlers", f.Name())
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		for _, decl := range node.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			module := findWriteLogModule(fn)
			if module != "" {
				result[fn.Name.Name] = module
			}
		}
	}
	return result
}

func findWriteLogModule(fn *ast.FuncDecl) string {
	var module string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != "writeLog" {
			return true
		}
		lit, ok := call.Args[1].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		module = strings.Trim(lit.Value, "\"")
		return false
	})
	return module
}

func routeBase(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func normalizeBizCode(base string) string {
	parts := strings.Split(base, "-")
	for i, p := range parts {
		parts[i] = singularize(p)
	}
	return strings.Join(parts, "_")
}

func singularize(s string) string {
	if strings.HasSuffix(s, "ies") && len(s) > 3 {
		return s[:len(s)-3] + "y"
	}
	if strings.HasSuffix(s, "ses") && len(s) > 3 {
		return s[:len(s)-2]
	}
	if strings.HasSuffix(s, "s") && len(s) > 1 {
		return s[:len(s)-1]
	}
	return s
}

func shouldIgnoreBizBase(base string) bool {
	ignore := map[string]bool{
		"login":     true,
		"workflows": true,
		"biz-types": true,
		"logs":      true,
		"profile":   true,
	}
	return ignore[base]
}

func contains(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func displayNameFromHandlers(handlerFns []string, moduleMap map[string]string) string {
	for _, fn := range handlerFns {
		if name := strings.TrimSpace(moduleMap[fn]); name != "" {
			return name
		}
	}
	return ""
}

// GetBizTypes 获取所有业务类型（供适用业务下拉使用）
func GetBizTypes(c *gin.Context) {
	var list []models.BizType
	database.DB.Order("sort asc, id asc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

// CreateBizType 新增业务类型
func CreateBizType(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
		Name string `json:"name" binding:"required"`
		Sort int    `json:"sort"`
	}
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

type WorkflowNodeInput struct {
	Sort          int    `json:"sort"`
	Name          string `json:"name" binding:"required"`
	ApproveType   string `json:"approve_type"`
	Approvers     string `json:"approvers"`
	Conditions    string `json:"conditions"`
	AllowSkip     bool   `json:"allow_skip"`
	AllowTransfer bool   `json:"allow_transfer"`
	ParentIDs     string `json:"parent_ids"` // JSON数组，存父节点sort值
}

type WorkflowTemplateRequest struct {
	Name        string              `json:"name" binding:"required"`
	Description string              `json:"description"`
	BizType     string              `json:"biz_type"`
	Nodes       []WorkflowNodeInput `json:"nodes"`
}

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
