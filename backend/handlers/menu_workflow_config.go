package handlers

import (
	"oa-system/database"
	"oa-system/models"
)

// syncMenuWorkflowConfig 在新增/编辑菜单时同步 menu_workflow_configs 和 biz_types
// enableWorkflow=true  → 确保关联存在，biz_types 有对应记录
// enableWorkflow=false → 删除关联及对应 biz_types 记录
func syncMenuWorkflowConfig(menuID int, enableWorkflow bool, bizCode, bizName string, bizSort int) {
	if !enableWorkflow {
		DeleteMenuWorkflowConfigByMenuID(menuID)
		return
	}
	if bizCode == "" || bizName == "" {
		return
	}

	var cfg models.MenuWorkflowConfig
	exists := database.DB.Where("menu_id = ?", menuID).First(&cfg).Error == nil

	if exists {
		// 已有关联，更新 biz_types
		var biz models.BizType
		if database.DB.First(&biz, cfg.BizTypeID).Error == nil {
			if biz.Code != bizCode || biz.Name != bizName || biz.Sort != bizSort {
				biz.Code = bizCode
				biz.Name = bizName
				biz.Sort = bizSort
				database.DB.Save(&biz)
			}
		}
	} else {
		// 无关联，先确保 biz_types 有记录
		var biz models.BizType
		if database.DB.Where("code = ?", bizCode).First(&biz).Error != nil {
			biz = models.BizType{Code: bizCode, Name: bizName, Sort: bizSort}
			database.DB.Create(&biz)
		}
		// 创建关联
		database.DB.Create(&models.MenuWorkflowConfig{
			MenuID:    menuID,
			BizTypeID: biz.ID,
		})
	}
}

// DeleteMenuWorkflowConfigByMenuID 删除菜单时联动删除关联及 biz_type
func DeleteMenuWorkflowConfigByMenuID(menuID int) {
	var cfg models.MenuWorkflowConfig
	if database.DB.Where("menu_id = ?", menuID).First(&cfg).Error != nil {
		return
	}
	bizTypeID := cfg.BizTypeID
	database.DB.Delete(&cfg)
	database.DB.Delete(&models.BizType{}, bizTypeID)
}
