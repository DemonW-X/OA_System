package routes

import (
	"oa-system/handlers"

	"github.com/gin-gonic/gin"
)

// dictionaryRoutes 配置数据字典路由
func dictionaryRoutes(rg *gin.RouterGroup) {
	rg.GET("/data-dictionaries", handlers.GetDataDictionaries)
	rg.GET("/data-dictionaries/:id", handlers.GetDataDictionary)
	rg.POST("/data-dictionaries", handlers.CreateDataDictionary)
	rg.PUT("/data-dictionaries/:id", handlers.UpdateDataDictionary)
	rg.DELETE("/data-dictionaries/:id", handlers.DeleteDataDictionary)

	rg.GET("/data-dictionaries/:id/items", handlers.GetDataDictionaryItems)
	rg.POST("/data-dictionaries/:id/items", handlers.CreateDataDictionaryItem)
	rg.PUT("/data-dictionary-items/:id", handlers.UpdateDataDictionaryItem)
	rg.DELETE("/data-dictionary-items/:id", handlers.DeleteDataDictionaryItem)
}
