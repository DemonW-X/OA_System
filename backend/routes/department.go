package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

// departmentRoutes 执行相关业务逻辑
func departmentRoutes(rg *gin.RouterGroup) {
	rg.GET("/departments", handlers.GetDepartments)
	rg.POST("/departments", handlers.CreateDepartment)
	rg.PUT("/departments/:id", handlers.UpdateDepartment)
	rg.DELETE("/departments/:id", handlers.DeleteDepartment)
	rg.GET("/departments/:id/menu-permissions", handlers.GetDepartmentMenuPermissions)
	rg.PUT("/departments/:id/menu-permissions", handlers.SetDepartmentMenuPermissions)
}
