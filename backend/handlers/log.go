package handlers

import (
	"oa-system/database"
	"oa-system/models"

	"github.com/gin-gonic/gin"
)

// writeLog 执行相关业务逻辑
func writeLog(c *gin.Context, module, action, remark string) {
	entry := models.OperationLog{
		UserID:     c.GetInt("userID"),
		Username:   c.GetString("username"),
		Module:     module,
		Action:     action,
		Method:     c.Request.Method,
		Path:       c.Request.URL.Path,
		StatusCode: c.Writer.Status(),
		IP:         c.ClientIP(),
		Remark:     remark,
	}

	if enqueueOperationLog(entry) {
		return
	}
	database.DB.Create(&entry)
}
