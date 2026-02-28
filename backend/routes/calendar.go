package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func calendarRoutes(rg *gin.RouterGroup) {
	rg.GET("/calendar", handlers.GetCalendarEvents)
}
