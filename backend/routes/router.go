package routes

import (
	"oa-system/handlers"
	"oa-system/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// 静态文件服务（上传的图片和附件）
	r.Static("/uploads", "./uploads")

	// 公开接口
	r.POST("/api/login", handlers.Login)

	// 鉴权接口
	api := r.Group("/api", middleware.JWTAuth())
	authRoutes(api)
	departmentRoutes(api)
	positionRoutes(api)
	employeeRoutes(api)
	noticeRoutes(api)
	logRoutes(api)
	meetingRoomRoutes(api)
	eventBookingRoutes(api)
	leaveRequestRoutes(api)
	resignationRoutes(api)
	workflowRoutes(api)
	menuRoutes(api)
	orchidWorkflowRoutes(api)

	// 文件上传接口
	api.POST("/upload/image", handlers.UploadImage)
	api.POST("/upload/attachment", handlers.UploadAttachment)

	return r
}
