package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func eventBookingRoutes(rg *gin.RouterGroup) {
	rg.GET("/event-bookings", handlers.GetEventBookings)
	rg.GET("/event-bookings/:id", handlers.GetEventBooking)
	rg.POST("/event-bookings", handlers.CreateEventBooking)
	rg.PUT("/event-bookings/:id", handlers.UpdateEventBooking)
	rg.PUT("/event-bookings/:id/submit", handlers.SubmitEventBooking)
	rg.PUT("/event-bookings/:id/approve", handlers.ApproveEventBooking)
	rg.DELETE("/event-bookings/:id", handlers.DeleteEventBooking)
}
