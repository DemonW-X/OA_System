package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/models"

	"github.com/gin-gonic/gin"
)

type MeetingRoomRequest struct {
	Name     string `json:"name" binding:"required"`
	Location string `json:"location"`
	Capacity int    `json:"capacity"`
	Status   int    `json:"status"`
	Remark   string `json:"remark"`
}

func GetMeetingRooms(c *gin.Context) {
	var list []models.MeetingRoom
	query := database.DB.Model(&models.MeetingRoom{})
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Offset(offset).Limit(pageSize).Find(&list)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
	writeLog(c, "会议室管理", "查询", "查询会议室列表")
}

func CreateMeetingRoom(c *gin.Context) {
	var req MeetingRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	// 检查未删除的同名记录
	var count int64
	database.DB.Model(&models.MeetingRoom{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "会议室名称已存在"})
		return
	}
	room := models.MeetingRoom{
		Name: req.Name, Location: req.Location,
		Capacity: req.Capacity, Status: req.Status, Remark: req.Remark,
	}
	if room.Status == 0 {
		room.Status = 1
	}
	if err := database.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": room})
	writeLog(c, "会议室管理", "新增", "新增会议室："+req.Name)
}

func UpdateMeetingRoom(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var room models.MeetingRoom
	if err := database.DB.First(&room, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "会议室不存在"})
		return
	}
	var req MeetingRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var count int64
	database.DB.Model(&models.MeetingRoom{}).Where("name = ? AND id != ?", req.Name, id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "会议室名称已存在"})
		return
	}
	room.Name = req.Name
	room.Location = req.Location
	room.Capacity = req.Capacity
	room.Status = req.Status
	room.Remark = req.Remark
	if err := database.DB.Save(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": room})
	writeLog(c, "会议室管理", "修改", "修改会议室："+req.Name)
}

func DeleteMeetingRoom(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var room models.MeetingRoom
	if err := database.DB.First(&room, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "会议室不存在"})
		return
	}
	// 检查是否有未来的预定
	var bookingCount int64
	database.DB.Model(&models.EventBooking{}).
		Where("meeting_room_id = ? AND end_time > NOW()", id).Count(&bookingCount)
	if bookingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该会议室存在未来预定，无法删除"})
		return
	}
	if err := database.DB.Delete(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "会议室管理", "删除", "删除会议室："+room.Name)
}
