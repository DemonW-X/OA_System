package main

import (
	"log"
	"oa-system/config"
	"oa-system/database"
	"oa-system/handlers"
	"oa-system/models"
	"oa-system/routes"
)

func main() {
	config.Init()
	database.InitDB()
	database.InitRedis()

	database.DB.AutoMigrate(
		&models.User{},
		&models.Department{},
		&models.Position{},
		&models.Employee{},
		&models.Notice{},
		&models.OperationLog{},
		&models.MeetingRoom{},
		&models.EventBooking{},
		&models.LeaveRequest{},
		&models.Resignation{},
		&models.WorkflowTemplate{},
		&models.WorkflowNode{},
		&models.BizType{},
		&models.OrchidWorkflowDefinition{},
		&models.OrchidWorkflowInstance{},
		&models.OrchidWorkflowHistory{},
		&models.OrchidWorkflowTask{},
		&models.Menu{},
		&models.EmployeeMenuPermission{},
	)

	handlers.InitAdmin()
	handlers.InitBizTypes()

	r := routes.SetupRouter()
	log.Printf("服务启动在 :%s", config.AppConfig.ServerPort)
	r.Run(":" + config.AppConfig.ServerPort)
}
