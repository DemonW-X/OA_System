package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

// logRoutes 执行相关业务逻辑
func logRoutes(rg *gin.RouterGroup) {
	rg.GET("/logs", handlers.GetLogs)
	rg.GET("/logs/modules", handlers.GetLogModules)
}
