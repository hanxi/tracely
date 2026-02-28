package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hanxi/tracely/internal/model"
	"gorm.io/gorm"
)

// ErrorTrend 错误趋势数据
type ErrorTrend struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// DashboardOverview 概览数据响应
type DashboardOverview struct {
	TodayPV     int64        `json:"todayPV"`     // 今日 PV
	TodayUV     int64        `json:"todayUV"`     // 今日 UV
	TotalErrors int64        `json:"totalErrors"` // 错误总数
	TodayErrors int64        `json:"todayErrors"` // 今日新增错误
	TopErrors   []TopError   `json:"topErrors"`   // Top 5 错误
	ErrorTrend  []ErrorTrend `json:"errorTrend"`  // 近 7 日错误趋势
}

// TopError 顶部错误
type TopError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// Overview 概览数据接口
func Overview(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		appID := c.Query("appID")

		// 构建基础查询
		activeQuery := db.Model(&model.ActiveLog{})
		errorQuery := db.Model(&model.ErrorLog{})

		if appID != "" {
			activeQuery = activeQuery.Where("app_id = ?", appID)
			errorQuery = errorQuery.Where("app_id = ?", appID)
		}

		// 计算今日 0 点时间
		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		// 今日 PV/UV
		var todayPV, todayUV int64
		activeQuery.Where("created_at >= ?", todayStart).Count(&todayPV)
		activeQuery.Where("created_at >= ?", todayStart).Distinct("user_id").Count(&todayUV)

		// 错误统计
		var totalErrors, todayErrors int64
		errorQuery.Count(&totalErrors)
		errorQuery.Where("first_seen >= ?", todayStart).Count(&todayErrors)

		// Top 5 错误 - 按指纹分组统计
		var topErrors []TopError
		errorQuery.Select("type, message, SUM(count) as count").
			Group("fingerprint, type, message").
			Order("count DESC").
			Limit(5).
			Find(&topErrors)

		// 近 7 日错误趋势
		var errorTrend []ErrorTrend
		for i := 6; i >= 0; i-- {
			dayStart := time.Date(now.Year(), now.Month(), now.Day()-i, 0, 0, 0, 0, now.Location())
			dayEnd := dayStart.Add(24 * time.Hour)
			var count int64
			errorQuery.Where("first_seen >= ? AND first_seen < ?", dayStart, dayEnd).Count(&count)
			errorTrend = append(errorTrend, ErrorTrend{
				Date:  dayStart.Format("01/02"),
				Count: int(count),
			})
		}

		c.JSON(http.StatusOK, DashboardOverview{
			TodayPV:     todayPV,
			TodayUV:     todayUV,
			TotalErrors: totalErrors,
			TodayErrors: todayErrors,
			TopErrors:   topErrors,
			ErrorTrend:  errorTrend,
		})
	}
}
