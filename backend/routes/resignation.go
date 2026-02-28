package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func resignationRoutes(rg *gin.RouterGroup) {
	rg.GET("/resignations", handlers.GetResignations)
	rg.GET("/resignations/:id", handlers.GetResignation)
	rg.POST("/resignations", handlers.CreateResignation)
	rg.PUT("/resignations/:id", handlers.UpdateResignation)
	rg.DELETE("/resignations/:id", handlers.DeleteResignation)
}
