package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

var AppConfig Config

// Init 初始化相关资源
func Init() {
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用环境变量")
	}

	AppConfig = Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3307"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "123456"),
		DBName:     getEnv("DB_NAME", "oa_system"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

// getEnv 获取数据
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
