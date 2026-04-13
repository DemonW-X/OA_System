package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/models"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetDataDictionaries 获取数据字典主表列表
func GetDataDictionaries(c *gin.Context) {
	var list []models.DataDictionary
	query := database.DB.Model(&models.DataDictionary{})

	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("code LIKE ? OR name LIKE ? OR remark LIKE ?", like, like, like)
	}

	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Order("id asc").Offset(offset).Limit(pageSize).Find(&list)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      list,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
	writeLog(c, "数据字典", "查询", "查询数据字典主表列表")
}

// GetDataDictionary 获取单个数据字典（含子表）
func GetDataDictionary(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var dict models.DataDictionary
	if err := database.DB.
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("id asc")
		}).
		First(&dict, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "数据字典不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": dict})
	writeLog(c, "数据字典", "查询", "查询数据字典详情")
}

// CreateDataDictionary 创建数据字典主表
func CreateDataDictionary(c *gin.Context) {
	var req dto.DataDictionaryRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "编码和名称不能为空"})
		return
	}

	var cnt int64
	database.DB.Model(&models.DataDictionary{}).Where("code = ?", req.Code).Count(&cnt)
	if cnt > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "编码已存在"})
		return
	}

	dict := models.DataDictionary{
		Code:   req.Code,
		Name:   req.Name,
		Remark: req.Remark,
	}
	if err := database.DB.Create(&dict).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": dict})
	writeLog(c, "数据字典", "新增", "新增数据字典："+dict.Name)
}

// UpdateDataDictionary 更新数据字典主表
func UpdateDataDictionary(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var dict models.DataDictionary
	if err := database.DB.First(&dict, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "数据字典不存在"})
		return
	}

	var req dto.DataDictionaryRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "编码和名称不能为空"})
		return
	}

	var cnt int64
	database.DB.Model(&models.DataDictionary{}).
		Where("id <> ? AND code = ?", id, req.Code).
		Count(&cnt)
	if cnt > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "编码已存在"})
		return
	}

	dict.Code = req.Code
	dict.Name = req.Name
	dict.Remark = req.Remark
	if err := database.DB.Save(&dict).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": dict})
	writeLog(c, "数据字典", "修改", "修改数据字典："+dict.Name)
}

// DeleteDataDictionary 删除数据字典主表及子表
func DeleteDataDictionary(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var dict models.DataDictionary
	if err := database.DB.First(&dict, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "数据字典不存在"})
		return
	}

	tx := database.DB.Begin()
	if err := tx.Where("dictionary_id = ?", id).Delete(&models.DataDictionaryItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除子表失败: " + err.Error()})
		return
	}
	if err := tx.Delete(&dict).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除主表失败: " + err.Error()})
		return
	}
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "提交事务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "数据字典", "删除", "删除数据字典："+dict.Name)
}

// GetDataDictionaryItems 获取数据字典子表列表
func GetDataDictionaryItems(c *gin.Context) {
	dictID, ok := parseID(c)
	if !ok {
		return
	}

	var dict models.DataDictionary
	if err := database.DB.Select("id").First(&dict, "id = ?", dictID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "数据字典不存在"})
		return
	}

	var items []models.DataDictionaryItem
	query := database.DB.Model(&models.DataDictionaryItem{}).Where("dictionary_id = ?", dictID)

	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Order("id asc").Offset(offset).Limit(pageSize).Find(&items)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
	writeLog(c, "数据字典", "查询", "查询数据字典子表列表")
}

// CreateDataDictionaryItem 创建数据字典子表
func CreateDataDictionaryItem(c *gin.Context) {
	dictID, ok := parseID(c)
	if !ok {
		return
	}

	var dict models.DataDictionary
	if err := database.DB.Select("id", "name").First(&dict, "id = ?", dictID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "数据字典不存在"})
		return
	}

	var req dto.DataDictionaryItemRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	item := models.DataDictionaryItem{
		DictionaryID: dictID,
		ExtField:     req.ExtField,
		Remark:       req.Remark,
	}
	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建明细失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "数据字典", "新增", "新增字典明细："+dict.Name)
}

// UpdateDataDictionaryItem 更新数据字典子表
func UpdateDataDictionaryItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var item models.DataDictionaryItem
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "明细不存在"})
		// return
	}

	var req dto.DataDictionaryItemRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	item.ExtField = req.ExtField
	item.Remark = req.Remark
	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新明细失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	writeLog(c, "数据字典", "修改", "修改字典明细")
}

// DeleteDataDictionaryItem 删除数据字典子表
func DeleteDataDictionaryItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var item models.DataDictionaryItem
	if err := database.DB.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "明细不存在"})
		return
	}

	if err := database.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除明细失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "数据字典", "删除", "删除字典明细")
}
