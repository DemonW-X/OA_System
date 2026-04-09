package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

// noticeRoutes 执行相关业务逻辑
func noticeRoutes(rg *gin.RouterGroup) {
	rg.GET("/notices", handlers.GetNotices)
	rg.GET("/notices/:id", handlers.GetNotice)
	rg.POST("/notices", handlers.CreateNotice)
	rg.PUT("/notices/:id", handlers.UpdateNotice)
	rg.PUT("/notices/:id/submit", handlers.SubmitNotice)
	rg.PUT("/notices/:id/withdraw", handlers.WithdrawNotice)
	rg.PUT("/notices/:id/approve", handlers.ApproveNotice)
	rg.PUT("/notices/:id/cancel-approve", handlers.CancelApproveNotice)
	rg.DELETE("/notices/:id", handlers.DeleteNotice)
}
