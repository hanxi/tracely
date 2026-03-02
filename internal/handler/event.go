package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hanxi/tracely/internal/config"
	"github.com/hanxi/tracely/internal/model"
	"gorm.io/gorm"
)

// EventRequest 事件上报请求
type EventRequest struct {
	EventName string                 `json:"eventName" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
	AppID     string                 `json:"appId" binding:"required"`
	UserID    string                 `json:"userId" binding:"required"`
}

// ReportEvent 上报事件接口
func ReportEvent(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req EventRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}

		// 验证事件是否在白名单中
		if !cfg.IsEventAllowed(req.EventName) {
			c.JSON(http.StatusForbidden, gin.H{"error": "事件未在白名单中"})
			return
		}

		// 创建事件记录
		if err := model.CreateEvent(db, req.EventName, req.Metadata, req.AppID, req.UserID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库操作失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "上报成功"})
	}
}

// GetEventStats 获取事件统计接口
func GetEventStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
		appID := c.Query("appID")
		eventName := c.Query("eventName")

		if days < 1 || days > 365 {
			days = 7
		}

		stats, err := model.GetEventStats(db, appID, eventName, days)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		// 确保返回空数组而不是 null
		if stats == nil {
			stats = []model.EventStats{}
		}

		c.JSON(http.StatusOK, gin.H{"stats": stats})
	}
}

// GetTopEvents 获取 Top 事件排行接口
func GetTopEvents(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
		appID := c.Query("appID")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		if days < 1 || days > 365 {
			days = 7
		}
		if limit < 1 || limit > 100 {
			limit = 10
		}

		events, err := model.GetTopEvents(db, appID, days, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		// 确保返回空数组而不是 null
		if events == nil {
			events = []model.TopEvent{}
		}

		c.JSON(http.StatusOK, gin.H{"events": events})
	}
}

// GetDailyEvents 获取每日事件统计接口
func GetDailyEvents(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
		appID := c.Query("appID")
		eventName := c.Query("eventName")

		if days < 1 || days > 365 {
			days = 7
		}

		daily, err := model.GetDailyEvents(db, appID, eventName, days)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		// 确保返回空数组而不是 null
		if daily == nil {
			daily = []model.DailyEvent{}
		}

		c.JSON(http.StatusOK, gin.H{"daily": daily})
	}
}

// GetEventOverview 获取事件概览数据（用于 Dashboard 首页）
func GetEventOverview(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		appID := c.Query("appID")

		// 今日开始时间
		today := time.Now().Truncate(24 * time.Hour)

		// 今日事件总数
		todayCount, err := model.GetEventCount(db, appID, "", today)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		// 今日活跃事件数（PV）
		todayActivePV, err := model.GetEventCount(db, appID, model.EVENT_ACTIVE, today)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		// 今日活跃用户数（UV）
		todayActiveUV, err := model.GetUniqueUserCount(db, appID, model.EVENT_ACTIVE, today)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		// Top 5 事件（最近 7 天）
		topEvents, err := model.GetTopEvents(db, appID, 7, 5)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
		if topEvents == nil {
			topEvents = []model.TopEvent{}
		}

		c.JSON(http.StatusOK, gin.H{
			"todayEventCount": todayCount,
			"todayActivePV":   todayActivePV,
			"todayActiveUV":   todayActiveUV,
			"topEvents":       topEvents,
		})
	}
}

// GetEventList 获取事件列表接口
func GetEventList(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		appID := c.Query("appID")
		eventName := c.Query("eventName")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		events, total, err := model.GetEventList(db, appID, eventName, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		// 确保返回空数组而不是 null
		if events == nil {
			events = []model.EventDetail{}
		}

		c.JSON(http.StatusOK, gin.H{
			"list":  events,
			"total": total,
		})
	}
}

// GetEventStatsSummary 获取事件统计摘要接口
func GetEventStatsSummary(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		appID := c.Query("appID")
		eventName := c.Query("eventName")
		days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

		if days < 1 || days > 365 {
			days = 7
		}

		summary, err := model.GetEventStatsSummary(db, appID, eventName, days)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		c.JSON(http.StatusOK, summary)
	}
}
