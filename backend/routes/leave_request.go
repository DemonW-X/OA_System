package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

// leaveRequestRoutes 执行相关业务逻辑
func leaveRequestRoutes(rg *gin.RouterGroup) {
	rg.GET("/leave-requests", handlers.GetLeaveRequests)
	rg.GET("/leave-requests/:id", handlers.GetLeaveRequest)
	rg.POST("/leave-requests", handlers.CreateLeaveRequest)
	rg.PUT("/leave-requests/:id", handlers.UpdateLeaveRequest)
	rg.PUT("/leave-requests/:id/submit", handlers.SubmitLeaveRequest)
	rg.PUT("/leave-requests/:id/approve", handlers.ApproveLeaveRequest)
	rg.DELETE("/leave-requests/:id", handlers.DeleteLeaveRequest)
}
