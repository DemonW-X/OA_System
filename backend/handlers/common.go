package handlers

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	phoneRe  = regexp.MustCompile(`^1[3-9]\d{9}$`)
	emailRe  = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	idCardRe = regexp.MustCompile(`^\d{17}[\dXx]$`)
)

// ValidateFormat 公共格式校验接口
// POST /api/validate/format
// Body: { "type": "phone"|"email"|"id_card", "value": "..." }
// Response: { "code": 0, "valid": true } 或 { "code": 1, "valid": false, "msg": "..." }
func ValidateFormat(c *gin.Context) {
	var req struct {
		Type  string `json:"type"  binding:"required"`
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "valid": false, "msg": "参数错误: " + err.Error()})
		return
	}

	var ok bool
	var msg string
	switch req.Type {
	case "phone":
		ok = phoneRe.MatchString(req.Value)
		if !ok {
			msg = "手机号格式不正确，请输入11位有效手机号"
		}
	case "email":
		ok = emailRe.MatchString(req.Value)
		if !ok {
			msg = "邮箱格式不正确"
		}
	case "id_card":
		ok = idCardRe.MatchString(req.Value)
		if !ok {
			msg = "身份证号格式不正确"
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "valid": false, "msg": "不支持的校验类型: " + req.Type})
		return
	}

	if ok {
		c.JSON(http.StatusOK, gin.H{"code": 0, "valid": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 1, "valid": false, "msg": msg})
	}
}

// parseID 解析输入数据
func parseID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效的ID"})
		return 0, false
	}
	return id, true
}

type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// getPagination 获取数据
func getPagination(c *gin.Context) (page, pageSize, offset int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset = (page - 1) * pageSize
	return
}
