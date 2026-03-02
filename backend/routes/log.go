package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func logRoutes(rg *gin.RouterGroup) {
	rg.GET("/logs", handlers.GetLogs)
	rg.GET("/logs/modules", handlers.GetLogModules)
}
