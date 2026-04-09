package models

import "time"

type OperationLog struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	UserID     int       `json:"user_id"`
	Username   string    `json:"username"`
	Module     string    `json:"module"` // 模块：部门、职位、员工、公告
	Action     string    `json:"action"` // 操作：新增、修改、删除、查询
	Method     string    `json:"method"` // HTTP方法
	Path       string    `json:"path"`   // 请求路径
	StatusCode int       `json:"status_code"`
	IP         string    `json:"ip"`
	Remark     string    `json:"remark"` // 操作描述
	CreatedAt  time.Time `json:"created_at"`
}
