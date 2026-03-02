package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func authRoutes(rg *gin.RouterGroup) {
	rg.GET("/profile", handlers.GetProfile)
	rg.PUT("/profile", handlers.UpdateProfile)
	rg.PUT("/profile/password", handlers.ChangePassword)
	rg.POST("/logout", handlers.Logout)
}
