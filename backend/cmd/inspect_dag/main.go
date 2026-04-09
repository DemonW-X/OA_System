package main

import (
	"fmt"
	"oa-system/config"
	"oa-system/database"
	"oa-system/models"
)

// main 程序入口
func main() {
	config.Init()
	database.InitDB()

	var defs []models.OrchidWorkflowDefinition
	database.DB.Where("is_active = ?", true).Order("id desc").Limit(10).Find(&defs)
	for _, d := range defs {
		fmt.Printf("=== ID:%d  BizType:%s  Name:%s ===\n%s\n\n", d.ID, d.BizType, d.Name, d.DagJSON)
	}
}
