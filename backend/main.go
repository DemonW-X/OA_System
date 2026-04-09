package main

import (
	"log"
	"oa-system/config"
	"oa-system/database"
	"oa-system/handlers"
	"oa-system/routes"
)

// main 程序入口
func main() {
	config.Init()
	database.InitDB()
	database.InitRedis()

	handlers.InitAdmin()
	handlers.InitBizTypes()
	handlers.InitOperationLogAsync()

	r := routes.SetupRouter()
	log.Printf("服务启动在 :%s", config.AppConfig.ServerPort)
	r.Run(":" + config.AppConfig.ServerPort)
}
