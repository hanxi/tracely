package tracely

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// generateNonce 生成随机 Nonce（16 字节随机数转十六进制）
func generateNonce() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 降级方案：使用时间戳 + 随机数
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// generateSignature 生成 HMAC-SHA256 签名
// 算法：HMAC-SHA256(appId + timestamp + nonce, appSecret)
func generateSignature(appID, appSecret, timestamp, nonce string) string {
	raw := appID + timestamp + nonce
	h := hmac.New(sha256.New, []byte(appSecret))
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))
}

// buildHeaders 生成认证请求头
func buildHeaders(appID, appSecret string) map[string]string {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := generateNonce()
	signature := generateSignature(appID, appSecret, timestamp, nonce)

	return map[string]string{
		"X-App-Id":    appID,
		"X-Timestamp": timestamp,
		"X-Nonce":     nonce,
		"X-Signature": signature,
	}
}
