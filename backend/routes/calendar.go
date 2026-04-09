package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

// calendarRoutes 执行相关业务逻辑
func calendarRoutes(rg *gin.RouterGroup) {
	rg.GET("/calendar", handlers.GetCalendarEvents)
}
