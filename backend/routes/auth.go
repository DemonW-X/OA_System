package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

// authRoutes 执行相关业务逻辑
func authRoutes(rg *gin.RouterGroup) {
	rg.GET("/profile", handlers.GetProfile)
	rg.PUT("/profile", handlers.UpdateProfile)
	rg.PUT("/profile/password", handlers.ChangePassword)
	rg.POST("/logout", handlers.Logout)
}
