package model

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

// ErrorLog 错误日志
type ErrorLog struct {
	ID          uint   `gorm:"primaryKey"`
	Fingerprint string `gorm:"uniqueIndex"` // 去重查询
	Type        string `gorm:"index"`       // 按类型筛选
	Message     string
	Stack       string
	URL         string
	AppID       string `gorm:"index"` // 按应用筛选
	UserAgent   string
	Count       int `gorm:"default:1"`
	FirstSeen   time.Time
	LastSeen    time.Time `gorm:"index"` // 按最近出现排序
}

// GenFingerprint 生成错误指纹
// 规则：MD5(appId + type + message)
func GenFingerprint(appID, errType, message string) string {
	raw := appID + errType + message
	hash := md5.Sum([]byte(raw))
	return hex.EncodeToString(hash[:])
}
