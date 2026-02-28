package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func leaveRequestRoutes(rg *gin.RouterGroup) {
	rg.GET("/leave-requests", handlers.GetLeaveRequests)
	rg.POST("/leave-requests", handlers.CreateLeaveRequest)
	rg.PUT("/leave-requests/:id", handlers.UpdateLeaveRequest)
	rg.PUT("/leave-requests/:id/submit", handlers.SubmitLeaveRequest)
	rg.PUT("/leave-requests/:id/approve", handlers.ApproveLeaveRequest)
	rg.DELETE("/leave-requests/:id", handlers.DeleteLeaveRequest)
}
