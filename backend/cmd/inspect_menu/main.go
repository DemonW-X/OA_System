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

	// 删除重复的组织管理（保留 ID 最小的）
	result := database.DB.Unscoped().Where("id IN ?", []int{14, 15}).Delete(&models.Menu{})
	fmt.Printf("清理重复菜单：影响 %d 行\n", result.RowsAffected)

	// 验证结果
	var menus []models.Menu
	database.DB.Unscoped().Where("name = ?", "组织管理").Find(&menus)
	for _, m := range menus {
		fmt.Printf("ID:%-3d Name:%s Deleted:%v\n", m.ID, m.Name, m.DeletedAt.Valid)
	}
}
