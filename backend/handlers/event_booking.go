package handlers

import (
	"net/http"
	"oa-system/database"
	"oa-system/models"
	"time"

	"github.com/gin-gonic/gin"
)

type EventBookingRequest struct {
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description"`
	Type          string `json:"type" binding:"required"`
	StartTime     string `json:"start_time" binding:"required"`
	EndTime       string `json:"end_time" binding:"required"`
	MeetingRoomID int    `json:"meeting_room_id"`
	Participants  string `json:"participants"`
}

type EventBookingSubmit struct {
	Remark string `json:"remark"`
}

func GetEventBookings(c *gin.Context) {
	var list []models.EventBooking
	query := database.DB.Model(&models.EventBooking{}).Preload("MeetingRoom")
	if title := c.Query("title"); title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if t := c.Query("type"); t != "" {
		query = query.Where("type = ?", t)
	}
	if date := c.Query("date"); date != "" {
		query = query.Where("DATE(start_time) <= ? AND DATE(end_time) >= ?", date, date)
	}
	var total int64
	query.Count(&total)
	page, pageSize, offset := getPagination(c)
	query.Order("start_time ASC").Offset(offset).Limit(pageSize).Find(&list)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"list": list, "total": total, "page": page, "page_size": pageSize},
	})
	writeLog(c, "事件预定", "查询", "查询事件预定列表")
}

func GetEventBooking(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var booking models.EventBooking
	if err := database.DB.Preload("MeetingRoom").First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "预定不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": booking})
}

func CreateEventBooking(c *gin.Context) {
	var req EventBookingRequest
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
	// 会议室时间冲突检测
	if req.MeetingRoomID > 0 {
		var conflict int64
		database.DB.Model(&models.EventBooking{}).
			Where("meeting_room_id = ? AND start_time < ? AND end_time > ?",
				req.MeetingRoomID, endTime, startTime).
			Count(&conflict)
		if conflict > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该会议室在所选时间段已有预定"})
			return
		}
	}
	createdBy := c.GetString("realName")
	if createdBy == "" {
		createdBy = c.GetString("username")
	}
	booking := models.EventBooking{
		Title:         req.Title,
		Description:   req.Description,
		Type:          req.Type,
		StartTime:     startTime,
		EndTime:       endTime,
		MeetingRoomID: req.MeetingRoomID,
		Participants:  req.Participants,
		Status:        "draft",
		WorkflowLogs:  "[]",
		CreatedBy:     createdBy,
	}
	if err := database.DB.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": booking})
	writeLog(c, "事件预定", "新增", "新增预定："+req.Title)
}

func UpdateEventBooking(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var booking models.EventBooking
	if err := database.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "预定不存在"})
		return
	}
	if booking.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅草稿状态可修改"})
		return
	}
	var req EventBookingRequest
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
	if req.MeetingRoomID > 0 {
		var conflict int64
		database.DB.Model(&models.EventBooking{}).
			Where("meeting_room_id = ? AND id != ? AND start_time < ? AND end_time > ?",
				req.MeetingRoomID, id, endTime, startTime).
			Count(&conflict)
		if conflict > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "该会议室在所选时间段已有预定"})
			return
		}
	}
	booking.Title = req.Title
	booking.Description = req.Description
	booking.Type = req.Type
	booking.StartTime = startTime
	booking.EndTime = endTime
	booking.MeetingRoomID = req.MeetingRoomID
	booking.Participants = req.Participants
	if err := database.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": booking})
	writeLog(c, "事件预定", "修改", "修改预定："+req.Title)
}

func SubmitEventBooking(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var booking models.EventBooking
	if err := database.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "预定不存在"})
		return
	}
	if booking.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅草稿状态可提交"})
		return
	}
	var req EventBookingSubmit
	_ = c.ShouldBindJSON(&req)
	now := time.Now()
	op := currentOperator(c)
	booking.SubmittedBy = op
	booking.SubmittedAt = &now
	if hasOrchidWorkflowForBiz("event_booking") {
		if _, err := startOrchidInstance("event_booking", booking.ID, op); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "流程实例启动失败: " + err.Error()})
			return
		}
		booking.Status = "pending"
	} else {
		booking.Status = "approved"
		booking.ApprovedBy = op
		booking.ApprovedAt = &now
	}
	booking.WorkflowLogs = buildBizWorkflowLogs("event_booking", booking.ID)
	if err := database.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "提交失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": booking})
	writeLog(c, "事件预定", "提交", "提交预定："+booking.Title)
}

func ApproveEventBooking(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var booking models.EventBooking
	if err := database.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "预定不存在"})
		return
	}
	if booking.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "仅待审核状态可审批"})
		return
	}
	var req struct {
		Status       string `json:"status" binding:"required"` // approved/rejected
		RejectReason string `json:"reject_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if req.Status != "approved" && req.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "状态只能为 approved 或 rejected"})
		return
	}
	now := time.Now()
	op := currentOperator(c)
	finalStatus, err := approveOrRejectInstance("event_booking", booking.ID, op, req.Status, req.RejectReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "审批失败"})
		return
	}
	booking.Status = finalStatus
	booking.ApprovedBy = op
	booking.ApprovedAt = &now
	booking.RejectReason = req.RejectReason
	action := "审批通过"
	if req.Status == "rejected" {
		action = "审批拒绝"
	}
	booking.WorkflowLogs = buildBizWorkflowLogs("event_booking", booking.ID)
	if err := database.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "审批失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": booking})
	writeLog(c, "事件预定", "审批", action+"："+booking.Title)
}

func DeleteEventBooking(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var booking models.EventBooking
	if err := database.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "预定不存在"})
		return
	}
	if booking.Status == "approved" || booking.Status == "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "待审核/已通过记录不能删除"})
		return
	}
	if err := database.DB.Delete(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	writeLog(c, "事件预定", "删除", "删除预定："+booking.Title)
}
