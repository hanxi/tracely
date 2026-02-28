package model

import "time"

// ActiveLog 活跃日志
type ActiveLog struct {
	ID        uint   `gorm:"primaryKey"`
	AppID     string `gorm:"index"`
	UserID    string `gorm:"index:idx_user_page"` // UV 统计去重
	Page      string `gorm:"index:idx_user_page"` // 热门页面统计
	Duration  int
	UserAgent string
	CreatedAt time.Time `gorm:"index"` // 按日期分组
}
