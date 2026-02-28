package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func meetingRoomRoutes(rg *gin.RouterGroup) {
	rg.GET("/meeting-rooms", handlers.GetMeetingRooms)
	rg.POST("/meeting-rooms", handlers.CreateMeetingRoom)
	rg.PUT("/meeting-rooms/:id", handlers.UpdateMeetingRoom)
	rg.DELETE("/meeting-rooms/:id", handlers.DeleteMeetingRoom)
}
