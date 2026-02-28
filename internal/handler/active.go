package handler

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hanxi/tracely/internal/model"
	"gorm.io/gorm"
)

// ActiveRequest 活跃上报请求
type ActiveRequest struct {
	AppID    string `json:"appId" binding:"required"`
	UserID   string `json:"userId" binding:"required"`
	Page     string `json:"page" binding:"required"`
	Duration int    `json:"duration"`
}

// ReportActive 上报活跃接口
func ReportActive(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ActiveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}

		// 插入数据库
		activeLog := model.ActiveLog{
			AppID:     req.AppID,
			UserID:    req.UserID,
			Page:      req.Page,
			Duration:  req.Duration,
			UserAgent: c.GetHeader("User-Agent"),
			CreatedAt: time.Now(),
		}

		if err := db.Create(&activeLog).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库操作失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// Stats 获取统计接口
func Stats(db *gorm.DB) gin.HandlerFunc {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
		appID := c.Query("appID") // 支持按 appID 筛选

		if days < 1 || days > 30 {
			days = 7
		}

		since := time.Now().AddDate(0, 0, -days)

		// 添加调试日志
		logger.Info("[Stats] Request", "days", days, "appID", appID, "since", since.Format("2006-01-02 15:04:05"))

		// 构建基础查询
		query := db.Model(&model.ActiveLog{})
		if appID != "" {
			query = query.Where("app_id = ?", appID)
			logger.Info("[Stats] Filtering by appID", "appID", appID)
		}

		// 查询总记录数
		var totalCount int64
		query.Count(&totalCount)
		logger.Info("[Stats] Total records in time range", "count", totalCount)

		// 查询每日 PV/UV 数据
		type DailyStats struct {
			Date string `json:"date"`
			PV   int64  `json:"pv"`
			UV   int64  `json:"uv"`
		}
		var daily []DailyStats
		query.
			Select("DATE(created_at) as date, COUNT(*) as pv, COUNT(DISTINCT user_id) as uv").
			Where("created_at >= ?", since).
			Group("DATE(created_at)").
			Order("date ASC").
			Scan(&daily)
		logger.Info("[Stats] Daily stats", "days", len(daily))

		// 查询热门页面排行（按 PV 降序，取前 10）
		type PageStats struct {
			Page        string  `json:"page"`
			PV          int64   `json:"pv"`
			AvgDuration float64 `json:"avgDuration"`
		}
		var topPages []PageStats
		topPagesQuery := db.Table("active_logs").
			Select("page, COUNT(*) as pv, AVG(duration) as avgDuration").
			Where("created_at >= ?", since)
		if appID != "" {
			topPagesQuery = topPagesQuery.Where("app_id = ?", appID)
		}
		topPagesQuery.
			Group("page").
			Order("pv DESC").
			Limit(10).
			Scan(&topPages)
		logger.Info("[Stats] TopPages result", "rows", len(topPages))
		for i, p := range topPages {
			logger.Info("[Stats] TopPage", "index", i, "page", p.Page, "pv", p.PV, "avgDuration", p.AvgDuration)
		}

		// 确保返回空数组而不是 null
		if topPages == nil {
			topPages = []PageStats{}
		}

		c.JSON(http.StatusOK, gin.H{
			"daily":    daily,
			"topPages": topPages,
		})
	}
}
