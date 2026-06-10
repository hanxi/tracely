# Tracely Go SDK

轻量级后端监控客户端，支持错误上报和事件追踪。

## 功能特性

- 异步上报：基于 Channel 的异步队列，不阻塞业务逻辑
- 自动重试：失败请求自动重试 3 次
- 签名认证：HMAC-SHA256 签名确保请求安全
- 灵活上报：支持错误、自定义事件、安装/升级上报
- 心跳上报：定时自动上报服务活跃状态
- 静默失败：上报失败不影响业务逻辑

## 安装

```bash
go get github.com/hanxi/tracely/sdk/go/tracely
```

## 快速开始

```go
package main

import (
    "github.com/hanxi/tracely/sdk/go/tracely"
)

func main() {
    client := tracely.New(tracely.Config{
        AppID:     "my-app-id",
        AppSecret: "my-app-secret",
        Host:      "https://tracely.example.com",
    })

    // 上报错误
    client.ReportError(tracely.ErrorPayload{
        Type:    "runtimeError",
        Message: "无法解析用户配置",
        Stack:   "main.go:42",
        URL:     "https://example.com/user/settings",
    })

    // 上报事件
    client.ReportEvent("user_action", map[string]interface{}{
        "action": "click_button",
        "page":   "/dashboard",
    }, "user-123")

    // 上报安装
    client.ReportInstall("1.0.0", "android", "user-123")

    // 上报升级
    client.ReportUpgrade("1.0.0", "2.0.0", "android", "user-123")
}
```

## Web 框架集成

### Gin

```go
var tracelyClient *tracely.Client

func init() {
    tracelyClient = tracely.New(tracely.Config{
        AppID:     "my-app-id",
        AppSecret: "my-app-secret",
        Host:      "https://tracely.example.com",
    })
}

func main() {
    r := gin.Default()

    // Panic 捕获中间件
    r.Use(func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                tracelyClient.ReportError(tracely.ErrorPayload{
                    Type:    "panic",
                    Message: fmt.Sprintf("%v", err),
                    Stack:   string(debug.Stack()),
                    URL:     c.Request.URL.String(),
                })
            }
        }()
        c.Next()
    })

    r.Run(":8080")
}
```

### Chi

```go
r := chi.NewRouter()
r.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                tracelyClient.ReportError(tracely.ErrorPayload{
                    Type:    "panic",
                    Message: fmt.Sprintf("%v", err),
                    Stack:   string(debug.Stack()),
                    URL:     req.URL.String(),
                })
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next.ServeHTTP(w, req)
    })
})
```

### Echo

```go
e := echo.New()
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        defer func() {
            if err := recover(); err != nil {
                tracelyClient.ReportError(tracely.ErrorPayload{
                    Type:    "panic",
                    Message: fmt.Sprintf("%v", err),
                    Stack:   string(debug.Stack()),
                    URL:     c.Request().URL.String(),
                })
            }
        }()
        return next(c)
    }
})
```

## API 参考

### Config 结构

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| AppID | string | 是 | - | 应用 ID |
| AppSecret | string | 是 | - | 应用密钥 |
| Host | string | 是 | - | 服务器地址 |
| Timeout | time.Duration | 否 | 5s | 请求超时 |
| EnableHeartbeat | bool | 否 | false | 启用心跳上报 |
| HeartbeatInterval | time.Duration | 否 | 60s | 心跳间隔 |
| InstanceID | string | 否 | 自动生成 | 实例标识 |
| Tags | map[string]string | 否 | nil | 自定义标签 |

### Client 方法

#### ReportError(payload ErrorPayload)

上报错误信息。`AppID` 字段自动填充。

#### ReportEvent(eventName string, metadata map[string]interface{}, userID string)

上报自定义事件。`eventName` 须在服务端白名单中。

#### ReportInstall(version, platform, userID string)

上报应用安装事件。metadata: `{"version": "...", "platform": "..."}`

#### ReportUpgrade(fromVersion, toVersion, platform, userID string)

上报应用升级事件。metadata: `{"version": "...", "from_version": "...", "to_version": "...", "platform": "..."}`

### 数据结构

```go
type ErrorPayload struct {
    Type    string `json:"type"`
    Message string `json:"message"`
    Stack   string `json:"stack"`
    URL     string `json:"url"`
    AppID   string `json:"appId"` // 自动填充
}

type EventPayload struct {
    EventName string                 `json:"eventName"`
    Metadata  map[string]interface{} `json:"metadata"`
    AppID     string                 `json:"appId"`
    UserID    string                 `json:"userId"`
}
```

## 心跳机制

启用心跳后，SDK 使用 `_active` 事件定时上报服务活跃状态：

```go
client := tracely.New(tracely.Config{
    AppID:             os.Getenv("TRACELY_APP_ID"),
    AppSecret:         os.Getenv("TRACELY_APP_SECRET"),
    Host:              os.Getenv("TRACELY_HOST"),
    EnableHeartbeat:   true,
    HeartbeatInterval: 60 * time.Second,
    InstanceID:        "server-01",
    Tags: map[string]string{
        "env":     "production",
        "version": "1.0.0",
    },
})
```

心跳 metadata：`instanceId`（实例标识）、`duration`（运行时长秒数）、`tags`（自定义标签）。

## 异步队列

- 缓冲 Channel 容量 100，队列满时丢弃（不阻塞）
- 后台 Goroutine 消费，失败自动重试 3 次（间隔 1 秒）
- 适用于监控场景，不适用于关键业务数据

## 最佳实践

1. **全局单例**：应用启动时创建一个客户端实例
2. **环境变量**：通过 `os.Getenv` 管理 AppID/Secret
3. **中间件集成**：在 Web 框架中统一捕获 panic
4. **环境隔离**：生产环境启用，开发环境禁用或使用空实现

## 与 TypeScript SDK 对比

| 特性 | Go SDK | TypeScript SDK |
|------|--------|----------------|
| 上报方式 | 异步队列 | 异步请求 |
| 错误捕获 | 手动上报 | 自动 + 手动 |
| 活跃统计 | 心跳上报 | 页面停留时长 |
| 重试机制 | 3 次 | 无 |
| 节流控制 | 无 | 1 分钟 |
| 适用场景 | 后端服务 | 前端应用 |

## 许可证

MIT License
