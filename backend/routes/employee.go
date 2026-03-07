package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func employeeRoutes(rg *gin.RouterGroup) {
	rg.GET("/employees", handlers.GetEmployees)
	rg.GET("/employees/:id", handlers.GetEmployee)
}
