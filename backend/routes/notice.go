package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func noticeRoutes(rg *gin.RouterGroup) {
	rg.GET("/notices", handlers.GetNotices)
	rg.GET("/notices/:id", handlers.GetNotice)
	rg.POST("/notices", handlers.CreateNotice)
	rg.PUT("/notices/:id", handlers.UpdateNotice)
	rg.DELETE("/notices/:id", handlers.DeleteNotice)
}
