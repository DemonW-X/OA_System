package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/dto"
	"oa-system/models"
	"time"

	"github.com/gin-gonic/gin"
)

const timeLayout = "2006-01-02 15:04:05"

func GetCalendarEvents(c *gin.Context) {
	var list []models.CalendarEvent
	query := database.DB.Model(&models.CalendarEvent{})
	if month := c.Query("month"); month != "" {
		// month 格式：2026-02，查询该月所有事件
		query = query.Where("DATE_FORMAT(start_time, '%Y-%m') = ?", month)
	}
	query.Order("start_time ASC").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
	writeLog(c, "行事历", "查询", "查询行事历事件")
}

func CreateCalendarEvent(c *gin.Context) {
	var req dto.CalendarEventRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	startTime, err := time.ParseInLocation(timeLayout, req.StartTime, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "开始时间格式错误，请使用 yyyy-MM-dd HH:mm:ss"})
		return
	}
	endTime, err := time.ParseInLocation(timeLayout, req.EndTime, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "结束时间格式错误，请使用 yyyy-MM-dd HH:mm:ss"})
		return
	}
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "结束时间不能早于开始时间"})
		return
	}
	createdBy := c.GetString("realName")
	if createdBy == "" {
		createdBy = c.GetString("username")
	}
	eventType := req.Type
	if eventType == "" {
		eventType = "other"
	}
	event := models.CalendarEvent{
		Title:       req.Title,
		Description: req.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Type:        eventType,
		CreatedBy:   createdBy,
	}
	if err := database.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": event})
	writeLog(c, "行事历", "新增", "新增事件："+req.Title)
}

func UpdateCalendarEvent(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var event models.CalendarEvent
	if err := database.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "事件不存在"})
		return
	}
	var req dto.CalendarEventRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	startTime, err := time.ParseInLocation(timeLayout, req.StartTime, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "开始时间格式错误"})
		return
	}
	endTime, err := time.ParseInLocation(timeLayout, req.EndTime, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "结束时间格式错误"})
		return
	}
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "结束时间不能早于开始时间"})
		return
	}
	event.Title = req.Title
	event.Description = req.Description
	event.StartTime = startTime
	event.EndTime = endTime
	event.Type = req.Type
	if err := database.DB.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": event})
	writeLog(c, "行事历", "修改", "修改事件："+req.Title)
}

func DeleteCalendarEvent(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var event models.CalendarEvent
	if err := database.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "事件不存在"})
		return
	}
	if err := database.DB.Delete(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "行事历", "删除", "删除事件："+event.Title)
}
