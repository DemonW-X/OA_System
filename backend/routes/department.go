package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func departmentRoutes(rg *gin.RouterGroup) {
	rg.GET("/departments", handlers.GetDepartments)
	rg.POST("/departments", handlers.CreateDepartment)
	rg.PUT("/departments/:id", handlers.UpdateDepartment)
	rg.DELETE("/departments/:id", handlers.DeleteDepartment)
}
