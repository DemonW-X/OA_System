package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func positionRoutes(rg *gin.RouterGroup) {
	rg.GET("/positions", handlers.GetPositions)
	rg.GET("/positions/:id", handlers.GetPosition)
	rg.POST("/positions", handlers.CreatePosition)
	rg.PUT("/positions/:id", handlers.UpdatePosition)
	rg.DELETE("/positions/:id", handlers.DeletePosition)
	rg.PUT("/positions/:id/employees", handlers.SetPositionEmployees)
	rg.GET("/positions/:id/menu-permissions", handlers.GetPositionMenuPermissions)
	rg.PUT("/positions/:id/menu-permissions", handlers.SetPositionMenuPermissions)
}
