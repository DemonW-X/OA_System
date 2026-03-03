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
	rg.PUT("/resignations/:id/submit", handlers.SubmitResignation)
	rg.PUT("/resignations/:id/withdraw", handlers.WithdrawResignation)
	rg.PUT("/resignations/:id/approve", handlers.ApproveResignation)
	rg.PUT("/resignations/:id/cancel-approve", handlers.CancelApproveResignation)
	rg.DELETE("/resignations/:id", handlers.DeleteResignation)
}
