package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ipStore 存储每个 IP 的请求时间戳
var ipStore = sync.Map{}

// RateLimit IP 限速中间件（滑动窗口算法）
func RateLimit(maxPerMin int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		windowStart := now.Add(-time.Minute)

		// 获取该 IP 的请求时间戳列表
		raw, _ := ipStore.LoadOrStore(ip, &[]time.Time{})
		timestamps := raw.(*[]time.Time)

		// 过滤掉 60 秒之前的记录
		valid := make([]time.Time, 0, len(*timestamps))
		for _, ts := range *timestamps {
			if ts.After(windowStart) {
				valid = append(valid, ts)
			}
		}

		// 判断是否超过限制
		if len(valid) >= maxPerMin {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁"})
			c.Abort()
			return
		}

		// 添加当前请求时间戳
		valid = append(valid, now)
		ipStore.Store(ip, &valid)

		c.Next()
	}
}
