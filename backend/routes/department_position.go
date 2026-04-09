package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

// departmentPositionRoutes 执行相关业务逻辑
func departmentPositionRoutes(rg *gin.RouterGroup) {
	rg.GET("/department-positions", handlers.GetDepartmentPositions)
	rg.POST("/department-positions", handlers.CreateDepartmentPosition)
	rg.DELETE("/department-positions/:id", handlers.DeleteDepartmentPosition)
}
