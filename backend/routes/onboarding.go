package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func onboardingRoutes(rg *gin.RouterGroup) {
	rg.GET("/onboardings", handlers.GetOnboardings)
	rg.GET("/onboardings/:id", handlers.GetOnboarding)
	rg.POST("/onboardings", handlers.CreateOnboarding)
	rg.PUT("/onboardings/:id", handlers.UpdateOnboarding)
	rg.PUT("/onboardings/:id/submit", handlers.SubmitOnboarding)
	rg.PUT("/onboardings/:id/withdraw", handlers.WithdrawOnboarding)
	rg.PUT("/onboardings/:id/approve", handlers.ApproveOnboarding)
	rg.PUT("/onboardings/:id/cancel-approve", handlers.CancelApproveOnboarding)
	rg.DELETE("/onboardings/:id", handlers.DeleteOnboarding)
}
