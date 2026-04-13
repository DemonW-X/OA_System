package models

import (
	"time"

	"gorm.io/gorm"
)

// DataDictionary 数据字典主表
type DataDictionary struct {
	ID        int                  `json:"id" gorm:"primaryKey;autoIncrement"`
	Code      string               `json:"code" gorm:"size:64;not null;index"`
	Name      string               `json:"name" gorm:"size:100;not null"`
	Remark    string               `json:"remark"`
	Items     []DataDictionaryItem `json:"items,omitempty" gorm:"foreignKey:DictionaryID;references:ID"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
	DeletedAt gorm.DeletedAt       `json:"-" gorm:"index"`
}

// DataDictionaryItem 数据字典子表
type DataDictionaryItem struct {
	ID           int            `json:"id" gorm:"primaryKey;autoIncrement"`
	DictionaryID int            `json:"dictionary_id" gorm:"index;not null"`
	ExtField     string         `json:"extfield" gorm:"column:extfield"`
	Remark       string         `json:"remark"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
