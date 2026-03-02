package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

func employeeRoutes(rg *gin.RouterGroup) {
	rg.GET("/employees", handlers.GetEmployees)
	rg.GET("/employees/:id", handlers.GetEmployee)
	rg.POST("/employees", handlers.CreateEmployee)
	rg.PUT("/employees/:id", handlers.UpdateEmployee)
	rg.DELETE("/employees/:id", handlers.DeleteEmployee)
	rg.PUT("/employees/:id/submit", handlers.SubmitEmployee)
	rg.PUT("/employees/:id/withdraw", handlers.WithdrawEmployee)
	rg.PUT("/employees/:id/approve", handlers.ApproveEmployee)
	rg.PUT("/employees/:id/cancel-approve", handlers.CancelEmployeeApproval)
}
