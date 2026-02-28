package model

import (
	"log/slog"
	"os"
	"time"

	"gorm.io/gorm"
)

// StartActiveLogCleanup 启动活跃日志定时清理任务
// 每天凌晨 3 点清理 N 天前的活跃日志
func StartActiveLogCleanup(db *gorm.DB, retentionDays int) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if retentionDays <= 0 {
		logger.Info("[Tracely] Active log cleanup disabled", "reason", "retentionDays <= 0")
		return
	}

	go func() {
		for {
			now := time.Now()
			// 计算距离下一个凌晨 3 点的时间
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, now.Location())
			sleepDuration := time.Until(next)
			logger.Info("[Tracely] Next active log cleanup scheduled", "in", sleepDuration.String())
			time.Sleep(sleepDuration)

			// 执行清理
			cutoff := time.Now().AddDate(0, 0, -retentionDays)
			result := db.Where("created_at < ?", cutoff).Delete(&ActiveLog{})
			logger.Info("[Tracely] Cleaned up active logs", "count", result.RowsAffected, "older_than_days", retentionDays)
		}
	}()
}
