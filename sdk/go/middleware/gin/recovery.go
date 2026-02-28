package gin

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	tracely "github.com/hanxi/tracely/sdk/go"
)

// Recovery 捕获 panic 并自动上报的中间件
func Recovery(client *tracely.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 上报 panic 错误
				client.ReportError(tracely.ErrorPayload{
					Type:    "panicError",
					Message: fmt.Sprintf("%v", err),
					Stack:   string(debug.Stack()),
					URL:     c.FullPath(),
				})

				// 返回 500 响应
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()

		c.Next()
	}
}
