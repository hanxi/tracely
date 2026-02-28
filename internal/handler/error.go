package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hanxi/tracely/internal/model"
	"gorm.io/gorm"
)

// ErrorRequest 错误上报请求
type ErrorRequest struct {
	Type    string `json:"type" binding:"required"`
	Message string `json:"message" binding:"required"`
	Stack   string `json:"stack"`
	URL     string `json:"url"`
	AppID   string `json:"appId" binding:"required"`
}

// ReportError 上报错误接口
func ReportError(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ErrorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}

		// 生成错误指纹
		fingerprint := model.GenFingerprint(req.AppID, req.Type, req.Message)

		// 查询是否存在相同指纹
		var existing model.ErrorLog
		if err := db.Where("fingerprint = ?", fingerprint).First(&existing).Error; err == nil {
			// 存在则更新
			existing.Count++
			existing.LastSeen = model.GetDB().NowFunc()
			existing.Stack = req.Stack
			existing.URL = req.URL
			if err := db.Save(&existing).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库操作失败"})
				return
			}
		} else {
			// 不存在则插入
			now := model.GetDB().NowFunc()
			newLog := model.ErrorLog{
				Fingerprint: fingerprint,
				Type:        req.Type,
				Message:     req.Message,
				Stack:       req.Stack,
				URL:         req.URL,
				AppID:       req.AppID,
				UserAgent:   c.GetHeader("User-Agent"),
				Count:       1,
				FirstSeen:   now,
				LastSeen:    now,
			}
			if err := db.Create(&newLog).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库操作失败"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "上报成功"})
	}
}

// ErrorList 获取错误列表接口
func ErrorList(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
		errType := c.Query("type")
		appID := c.Query("appID") // 支持按 appID 筛选

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		offset := (page - 1) * pageSize

		// 构建查询
		query := db.Model(&model.ErrorLog{})
		if errType != "" {
			query = query.Where("type = ?", errType)
		}
		if appID != "" {
			query = query.Where("app_id = ?", appID)
		}

		// 查询总数
		var total int64
		query.Count(&total)

		// 查询列表
		var list []model.ErrorLog
		if err := query.Order("count DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total": total,
			"list":  list,
		})
	}
}
