package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hanxi/tracely/internal/config"
)

// AppInfo 应用信息（返回给前端，不包含敏感信息）
type AppInfo struct {
	AppID   string `json:"appId"`
	AppName string `json:"appName"`
}

// GetApps 获取应用列表
func GetApps(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 构建应用列表（不包含 appSecret）
		appList := make([]AppInfo, 0, len(cfg.Apps))
		for _, app := range cfg.Apps {
			appList = append(appList, AppInfo{
				AppID:   app.AppID,
				AppName: app.AppName,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"apps": appList,
		})
	}
}
