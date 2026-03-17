package models

import (
	"time"

	"gorm.io/gorm"
)

// MenuWorkflowConfig menus 与 biz_types 的关联表
// 有记录即代表该菜单关联了对应的业务类型（启用审批流）
type MenuWorkflowConfig struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	MenuID    int       `json:"menu_id" gorm:"not null;uniqueIndex"` // 关联菜单ID
	BizTypeID int       `json:"biz_type_id" gorm:"not null;index"`  // 关联 biz_types.id
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Menu struct {
	ID       int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name     string         `json:"name" gorm:"not null"`
	Icon     string         `json:"icon" gorm:"default:''"`
	Path     string         `json:"path" gorm:"default:''"`
	SortCode int            `json:"sort_code" gorm:"default:0"`
	ParentID int            `json:"parent_id" gorm:"default:0"`
	Visible  bool           `json:"visible" gorm:"default:true"`
	Remark   string         `json:"remark"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
