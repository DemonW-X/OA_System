package main

import (
	"fmt"
	"oa-system/config"
	"oa-system/database"
	"oa-system/models"
)

func main() {
	config.Init()
	database.InitDB()

	var emps []models.Employee
	database.DB.Preload("Department").Preload("PositionInfo").Limit(5).Find(&emps)
	for _, e := range emps {
		fmt.Printf("ID:%d Name:%s DeptID:%d DeptName:%s PosID:%d PosName:%s Status:%d\n",
			e.ID, e.Name, e.DepartmentID,
			e.Department.Name,
			e.PositionID,
			e.PositionInfo.Name,
			e.Status,
		)
	}
}
