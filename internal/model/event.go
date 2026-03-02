package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// 内置事件类型常量
const (
	EVENT_ACTIVE = "_active" // 用户活跃事件
)

// Event 统一事件模型
type Event struct {
	ID        uint            `gorm:"primaryKey"`
	EventName string          `gorm:"index:idx_event_name;index:idx_app_event_time"` // 事件名称
	Metadata  json.RawMessage `gorm:"type:text"`                                     // 元数据（JSON 格式）
	AppID     string          `gorm:"index:idx_app_event_time"`                      // 应用 ID
	UserID    string          `gorm:"index"`                                         // 用户 ID
	CreatedAt time.Time       `gorm:"index:idx_app_event_time"`                      // 创建时间
}

// CreateEvent 创建事件记录
func CreateEvent(db *gorm.DB, eventName string, metadata map[string]interface{}, appID, userID string) error {
	// 将 metadata 转换为 JSON
	var metadataJSON json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metadataJSON = data
	}

	event := Event{
		EventName: eventName,
		Metadata:  metadataJSON,
		AppID:     appID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	return db.Create(&event).Error
}

// EventStats 事件统计结果
type EventStats struct {
	EventName string `json:"eventName"`
	Count     int64  `json:"count"`
}

// GetEventStats 获取事件统计（按事件名称分组）
func GetEventStats(db *gorm.DB, appID string, eventName string, days int) ([]EventStats, error) {
	since := time.Now().AddDate(0, 0, -days)

	query := db.Model(&Event{}).
		Select("event_name, COUNT(*) as count").
		Where("created_at >= ?", since)

	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	if eventName != "" {
		query = query.Where("event_name = ?", eventName)
	}

	var stats []EventStats
	err := query.Group("event_name").Order("count DESC").Scan(&stats).Error
	return stats, err
}

// TopEvent Top 事件结果
type TopEvent struct {
	EventName string `json:"eventName"`
	Count     int64  `json:"count"`
}

// GetTopEvents 获取 Top 事件排行
func GetTopEvents(db *gorm.DB, appID string, days int, limit int) ([]TopEvent, error) {
	since := time.Now().AddDate(0, 0, -days)

	query := db.Model(&Event{}).
		Select("event_name, COUNT(*) as count").
		Where("created_at >= ?", since)

	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	var events []TopEvent
	err := query.Group("event_name").Order("count DESC").Limit(limit).Scan(&events).Error
	return events, err
}

// DailyEvent 每日事件统计
type DailyEvent struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// GetDailyEvents 获取每日事件统计
func GetDailyEvents(db *gorm.DB, appID string, eventName string, days int) ([]DailyEvent, error) {
	since := time.Now().AddDate(0, 0, -days)

	query := db.Model(&Event{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", since)

	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	if eventName != "" {
		query = query.Where("event_name = ?", eventName)
	}

	var daily []DailyEvent
	err := query.Group("DATE(created_at)").Order("date ASC").Scan(&daily).Error
	return daily, err
}

// GetEventCount 获取事件总数
func GetEventCount(db *gorm.DB, appID string, eventName string, since time.Time) (int64, error) {
	query := db.Model(&Event{}).Where("created_at >= ?", since)

	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	if eventName != "" {
		query = query.Where("event_name = ?", eventName)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// GetUniqueUserCount 获取唯一用户数（UV）
func GetUniqueUserCount(db *gorm.DB, appID string, eventName string, since time.Time) (int64, error) {
	query := db.Model(&Event{}).
		Select("COUNT(DISTINCT user_id)").
		Where("created_at >= ?", since)

	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	if eventName != "" {
		query = query.Where("event_name = ?", eventName)
	}

	var count int64
	err := query.Scan(&count).Error
	return count, err
}

// EventDetail 事件详情
type EventDetail struct {
	ID        uint            `json:"id"`
	EventName string          `json:"eventName"`
	Metadata  json.RawMessage `json:"metadata"`
	AppID     string          `json:"appId"`
	UserID    string          `json:"userId"`
	CreatedAt time.Time       `json:"createdAt"`
}

// GetEventList 获取事件列表（分页）
func GetEventList(db *gorm.DB, appID string, eventName string, page int, pageSize int) ([]EventDetail, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	query := db.Model(&Event{})

	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	if eventName != "" {
		query = query.Where("event_name = ?", eventName)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取列表
	var events []EventDetail
	err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&events).Error

	return events, total, err
}

// EventStatsSummary 事件统计摘要
type EventStatsSummary struct {
	TotalCount int64 `json:"totalCount"`
	TodayCount int64 `json:"todayCount"`
	UV         int64 `json:"uv"`
}

// GetEventStatsSummary 获取事件统计摘要
func GetEventStatsSummary(db *gorm.DB, appID string, eventName string, days int) (*EventStatsSummary, error) {
	since := time.Now().AddDate(0, 0, -days)
	today := time.Now().Truncate(24 * time.Hour)

	// 获取总次数
	totalCount, err := GetEventCount(db, appID, eventName, since)
	if err != nil {
		return nil, err
	}

	// 获取今日次数
	todayCount, err := GetEventCount(db, appID, eventName, today)
	if err != nil {
		return nil, err
	}

	// 获取 UV
	uv, err := GetUniqueUserCount(db, appID, eventName, since)
	if err != nil {
		return nil, err
	}

	return &EventStatsSummary{
		TotalCount: totalCount,
		TodayCount: todayCount,
		UV:         uv,
	}, nil
}
