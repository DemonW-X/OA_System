package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func workflowRoutes(rg *gin.RouterGroup) {
	rg.GET("/workflows", handlers.GetWorkflowTemplates)
	rg.GET("/workflows/:id", handlers.GetWorkflowTemplate)
	rg.POST("/workflows", handlers.CreateWorkflowTemplate)
	rg.PUT("/workflows/:id", handlers.UpdateWorkflowTemplate)
	rg.DELETE("/workflows/:id", handlers.DeleteWorkflowTemplate)

	// 业务类型
	rg.GET("/biz-types", handlers.GetBizTypes)
	rg.POST("/biz-types", handlers.CreateBizType)
	rg.DELETE("/biz-types/:id", handlers.DeleteBizType)

}
