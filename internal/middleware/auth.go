package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hanxi/tracely/internal/config"
)

// nonceStore 存储已使用的 Nonce
var nonceStore = sync.Map{}

// SignAuth HMAC 签名验证中间件
func SignAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 检查请求头是否存在
		appID := c.GetHeader("X-App-Id")
		timestamp := c.GetHeader("X-Timestamp")
		nonce := c.GetHeader("X-Nonce")
		signature := c.GetHeader("X-Signature")

		if appID == "" || timestamp == "" || nonce == "" || signature == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证请求头"})
			c.Abort()
			return
		}

		// 2. 根据 AppID 查找 Secret
		secret, ok := cfg.GetSecret(appID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "非法 AppID"})
			c.Abort()
			return
		}

		// 3. 验证时间戳
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的时间戳"})
			c.Abort()
			return
		}

		now := time.Now().Unix()
		if abs(now-ts) > int64(cfg.TimestampTTL) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请求已过期"})
			c.Abort()
			return
		}

		// 4. 验证 Nonce 是否已使用（防重放）
		if _, exists := nonceStore.Load(nonce); exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "重放攻击"})
			c.Abort()
			return
		}
		nonceStore.Store(nonce, time.Now())

		// 5. 计算签名并比对
		raw := appID + timestamp + nonce
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(raw))
		expectedSig := hex.EncodeToString(h.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "签名错误"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StartNonceCleaner 启动定时清理过期 Nonce
func StartNonceCleaner(ttl int) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			nonceStore.Range(func(key, value interface{}) bool {
				if t, ok := value.(time.Time); ok {
					if now.Sub(t) > time.Duration(ttl)*time.Second {
						nonceStore.Delete(key)
					}
				}
				return true
			})
		}
	}()
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
