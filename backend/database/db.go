package database

import (
	"fmt"
	"log"

	"oa-system/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Query 返回一个默认过滤软删除记录的查询对象（deleted_at IS NULL）
// 所有业务查询应优先使用此函数，避免查出已删除数据
func Query() *gorm.DB {
	return DB.Where("deleted_at IS NULL")
}

func InitDB() {
	cfg := config.AppConfig
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	DB = db
	log.Println("数据库连接成功")
}
