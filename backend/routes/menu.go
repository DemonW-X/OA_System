package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func menuRoutes(rg *gin.RouterGroup) {
	rg.GET("/menus", handlers.GetMenus)
	rg.GET("/menus/:id", handlers.GetMenu)
	rg.POST("/menus", handlers.CreateMenu)
	rg.PUT("/menus/:id", handlers.UpdateMenu)
	rg.DELETE("/menus/:id", handlers.DeleteMenu)

	rg.GET("/employees/:id/menu-permissions", handlers.GetEmployeeMenuPermissions)
	rg.PUT("/employees/:id/menu-permissions", handlers.SetEmployeeMenuPermissions)


}
