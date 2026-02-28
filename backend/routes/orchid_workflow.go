package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func orchidWorkflowRoutes(rg *gin.RouterGroup) {
	rg.GET("/orchid-workflows", handlers.GetOrchidWorkflowDefinitions)
	rg.GET("/orchid-workflows/:id", handlers.GetOrchidWorkflowDefinition)
	rg.POST("/orchid-workflows", handlers.CreateOrchidWorkflowDefinition)
	rg.PUT("/orchid-workflows/:id", handlers.UpdateOrchidWorkflowDefinition)
	rg.DELETE("/orchid-workflows/:id", handlers.DeleteOrchidWorkflowDefinition)
	rg.GET("/orchid-workflow-histories", handlers.GetOrchidWorkflowHistories)
	rg.POST("/orchid-workflow-transfer", handlers.TransferOrchidWorkflowTask)
	rg.POST("/orchid-workflow-skip", handlers.SkipOrchidWorkflowNode)
	rg.POST("/orchid-workflow-seed", handlers.SeedOrchidWorkflowTemplates)
}
