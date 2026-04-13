package services

import (
	"context"
	"encoding/json"
	"fmt"
	"oa-system/database"
	"oa-system/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	workflowCacheTTL        = 24 * time.Hour
	workflowWarmThrottleKey = "workflow:active:warm:throttle"
)

// workflowCacheKey 生成业务数据
func workflowCacheKey(bizType string) string {
	return fmt.Sprintf("workflow:active:def:%s", strings.ToLower(strings.TrimSpace(bizType)))
}

// queryLatestActiveDefinition 查询最新激活流程定义
func queryLatestActiveDefinition(bizType string) (*models.OrchidWorkflowDefinition, error) {
	var def models.OrchidWorkflowDefinition
	err := database.DB.
		Where("biz_type = ? AND is_active = ?", bizType, true).
		Order("id desc").
		First(&def).Error
	if err != nil {
		return nil, err
	}
	return &def, nil
}

// RefreshOrchidWorkflowCacheByBizType 刷新单个业务类型流程缓存
func RefreshOrchidWorkflowCacheByBizType(bizType string) {
	bizType = strings.ToLower(strings.TrimSpace(bizType))
	if bizType == "" {
		return
	}

	ctx := context.Background()
	def, err := queryLatestActiveDefinition(bizType)
	if err != nil {
		if database.RDB != nil {
			database.RDB.Del(ctx, workflowCacheKey(bizType))
		}
		return
	}
	if database.RDB == nil {
		return
	}
	b, err := json.Marshal(def)
	if err != nil {
		return
	}
	database.RDB.Set(ctx, workflowCacheKey(bizType), b, workflowCacheTTL)
}

// WarmActiveOrchidWorkflowCache 预热全部激活流程缓存
func WarmActiveOrchidWorkflowCache() {
	var bizTypes []string
	database.DB.Model(&models.OrchidWorkflowDefinition{}).
		Where("is_active = ?", true).
		Distinct("biz_type").
		Pluck("biz_type", &bizTypes)
	for _, bizType := range bizTypes {
		RefreshOrchidWorkflowCacheByBizType(bizType)
	}
}

// WarmActiveOrchidWorkflowCacheIfDue 按间隔节流预热流程缓存
func WarmActiveOrchidWorkflowCacheIfDue(interval time.Duration) {
	if database.RDB == nil {
		WarmActiveOrchidWorkflowCache()
		return
	}
	if interval <= 0 {
		WarmActiveOrchidWorkflowCache()
		return
	}

	ok, err := database.RDB.SetNX(context.Background(), workflowWarmThrottleKey, "1", interval).Result()
	if err != nil || !ok {
		return
	}
	WarmActiveOrchidWorkflowCache()
}

// LoadActiveOrchidWorkflowDefinitionByBiz 从缓存优先加载流程定义
func LoadActiveOrchidWorkflowDefinitionByBiz(bizType string) (*models.OrchidWorkflowDefinition, error) {
	bizType = strings.ToLower(strings.TrimSpace(bizType))
	if bizType == "" {
		return nil, gorm.ErrRecordNotFound
	}

	if database.RDB != nil {
		val, err := database.RDB.Get(context.Background(), workflowCacheKey(bizType)).Result()
		if err == nil && val != "" {
			var def models.OrchidWorkflowDefinition
			if err := json.Unmarshal([]byte(val), &def); err == nil && def.ID > 0 {
				return &def, nil
			}
		}
	}

	def, err := queryLatestActiveDefinition(bizType)
	if err != nil {
		return nil, err
	}
	if database.RDB != nil {
		if b, mErr := json.Marshal(def); mErr == nil {
			database.RDB.Set(context.Background(), workflowCacheKey(bizType), b, workflowCacheTTL)
		}
	}
	return def, nil
}
