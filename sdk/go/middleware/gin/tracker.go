package gin

import (
	"time"

	"github.com/gin-gonic/gin"
	tracely "github.com/hanxi/tracely/sdk/go"
)

// Tracker 自动统计接口访问的中间件
func Tracker(client *tracely.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		c.Next()

		// 计算请求耗时（毫秒）
		duration := int(time.Since(start).Milliseconds())

		// 上报活跃数据
		client.ReportActive(tracely.ActivePayload{
			UserID:   c.GetHeader("X-User-Id"), // 从请求头读取，不存在则为空
			Page:     c.FullPath(),
			Duration: duration,
		})
	}
}
