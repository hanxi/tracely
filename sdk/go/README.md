# Tracely Go SDK

Tracely Go SDK 是一个轻量级的后端监控客户端，支持错误上报和事件追踪。

## 功能特性

- 🚀 **异步上报**：基于 Channel 的异步队列，不阻塞业务逻辑
- 🔄 **自动重试**：失败请求自动重试 3 次，提高上报成功率
- 🔒 **签名认证**：HMAC-SHA256 签名确保请求安全
- 📊 **灵活上报**：支持错误和自定义事件上报
- 💓 **心跳上报**：定时自动上报服务活跃状态，支持实例标识和自定义标签
- ⚡ **高性能**：缓冲队列容量 100，队列满时自动丢弃，不阻塞
- 🛡️ **静默失败**：上报失败不影响业务逻辑

## 安装

```bash
go get github.com/hanxi/tracely/sdk/go
```

在你的 `go.mod` 文件中添加依赖：

```go
require github.com/hanxi/tracely/sdk/go v0.1.0
```

## 配置

在使用 SDK 前，需要先在 Tracely 服务器配置中获取以下信息：

1. **AppID**：应用唯一标识（从 `config.yaml` 的 `apps[].appId` 获取）
2. **AppSecret**：应用密钥（从 `config.yaml` 的 `apps[].appSecret` 获取）
3. **Host**：Tracely 服务器地址（如：`https://tracely.example.com`）

## 快速开始

### 基础用法

```go
package main

import (
    "github.com/hanxi/tracely/sdk/go"
)

func main() {
    // 创建 SDK 客户端
    client := tracely.New(tracely.Config{
        AppID:     "my-app-id",
        AppSecret: "my-app-secret",
        Host:      "https://tracely.example.com",
        Timeout:   5 * time.Second, // 可选，默认 5s
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
}
```

### Web 框架集成

#### Gin

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/hanxi/tracely/sdk/go"
)

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

    r.Use(func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 捕获 panic 并上报
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

    r.GET("/api/user", func(c *gin.Context) {
        // 业务逻辑
        if err != nil {
            // 上报错误
            tracelyClient.ReportError(tracely.ErrorPayload{
                Type:    "apiError",
                Message: err.Error(),
                Stack:   "handler.go:25",
                URL:     c.Request.URL.String(),
            })
            c.JSON(500, gin.H{"error": "internal error"})
            return
        }

        // 上报成功事件
        tracelyClient.ReportEvent("api_call", map[string]interface{}{
            "endpoint": "/api/user",
            "method":   "GET",
        }, "user-123")

        c.JSON(200, gin.H{"data": "..."})
    })

    r.Run(":8080")
}
```

#### Echo

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/hanxi/tracely/sdk/go"
)

var tracelyClient *tracely.Client

func init() {
    tracelyClient = tracely.New(tracely.Config{
        AppID:     "my-app-id",
        AppSecret: "my-app-secret",
        Host:      "https://tracely.example.com",
    })
}

func main() {
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

    e.GET("/api/data", func(c echo.Context) error {
        // 业务逻辑
        return c.JSON(200, map[string]interface{}{"data": "..."})
    })

    e.Start(":8080")
}
```

## API 参考

### Config 结构

```go
type Config struct {
    AppID     string        // 应用 ID（必填）
    AppSecret string        // 应用密钥（必填）
    Host      string        // Tracely 服务器地址（必填）
    Timeout   time.Duration // 请求超时时间，默认 5s

    // 心跳上报配置
    EnableHeartbeat   bool              // 是否启用自动心跳上报，默认 false
    HeartbeatInterval time.Duration     // 心跳上报间隔，默认 60s
    InstanceID        string            // 实例标识，为空时自动生成（主机名+PID+时间戳）
    Tags              map[string]string // 自定义标签（如 env、version 等）
}
```

### Client 类型

#### New() 函数

```go
func New(config Config) *Client
```

创建新的 Tracely 客户端实例。

**示例：**
```go
client := tracely.New(tracely.Config{
    AppID:     "my-app-id",
    AppSecret: "my-app-secret",
    Host:      "https://tracely.example.com",
})
```

#### ReportError() 方法

```go
func (c *Client) ReportError(payload ErrorPayload)
```

上报错误信息。

**参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| Type | string | ✅ | 错误类型（如：`runtimeError`, `apiError`, `panic`） |
| Message | string | ✅ | 错误消息 |
| Stack | string | ❌ | 错误堆栈 |
| URL | string | ❌ | 错误发生的 URL |
| AppID | string | ❌ | 自动填充，无需设置 |

**示例：**
```go
client.ReportError(tracely.ErrorPayload{
    Type:    "databaseError",
    Message: "连接数据库失败",
    Stack:   "db.go:100",
    URL:     "https://example.com/api/users",
})
```

#### ReportEvent() 方法

```go
func (c *Client) ReportEvent(eventName string, metadata map[string]interface{}, userID string)
```

上报自定义事件。

**参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| eventName | string | ✅ | 事件名称（如：`user_action`, `page_view`） |
| metadata | map[string]interface{} | ❌ | 事件元数据 |
| userID | string | ✅ | 用户唯一标识 |

**示例：**
```go
client.ReportEvent("purchase", map[string]interface{}{
    "product_id": "12345",
    "amount":     99.99,
    "currency":   "CNY",
}, "user-123")
```

### 数据结构

#### ErrorPayload

```go
type ErrorPayload struct {
    Type    string `json:"type"`
    Message string `json:"message"`
    Stack   string `json:"stack"`
    URL     string `json:"url"`
    AppID   string `json:"appId"`
}
```

#### EventPayload

```go
type EventPayload struct {
    EventName string                 `json:"eventName"`
    Metadata  map[string]interface{} `json:"metadata"`
    AppID     string                 `json:"appId"`
    UserID    string                 `json:"userId"`
}
```

## 安全说明

### 签名机制

SDK 使用 HMAC-SHA256 签名确保请求安全：

```go
signature = HMAC-SHA256(appId + timestamp + nonce, appSecret)
```

**请求头：**
- `X-App-Id`：应用 ID
- `X-Timestamp`：当前时间戳（秒）
- `X-Nonce`：16 字节随机数（十六进制）
- `X-Signature`：签名值

### AppSecret 安全性

**注意**：AppSecret 在后端代码中也应妥善保管，建议：

1. 使用环境变量存储 AppSecret
2. 不要将配置文件提交到版本控制
3. 定期更换 AppSecret
4. 配合服务端的限速策略使用

**示例：**
```go
client := tracely.New(tracely.Config{
    AppID:     os.Getenv("TRACELY_APP_ID"),
    AppSecret: os.Getenv("TRACELY_APP_SECRET"),
    Host:      os.Getenv("TRACELY_HOST"),
})
```

## 异步队列机制

### 工作原理

SDK 使用 Go Channel 实现异步上报：

1. **缓冲队列**：容量 100 的缓冲 Channel
2. **非阻塞**：队列满时直接丢弃，不阻塞业务
3. **后台消费**：独立的 Goroutine 消费队列
4. **自动重试**：失败请求自动重试 3 次（每次间隔 1 秒）

### 性能特点

- ✅ 不阻塞主线程
- ✅ 高并发友好
- ✅ 失败不影响业务
- ⚠️ 极端情况下可能丢失数据（队列满时）

## 错误处理

SDK 采用**静默失败**策略：

- 上报失败不会影响业务逻辑
- 自动重试 3 次后放弃
- 不抛出异常，不阻塞流程
- 如需调试，可在服务端查看日志

## 最佳实践

### 1. 全局单例

推荐在应用启动时创建全局客户端：

```go
var tracelyClient *tracely.Client

func init() {
    tracelyClient = tracely.New(tracely.Config{
        AppID:     os.Getenv("TRACELY_APP_ID"),
        AppSecret: os.Getenv("TRACELY_APP_SECRET"),
        Host:      os.Getenv("TRACELY_HOST"),
    })
}
```

### 2. 中间件集成

在 Web 框架中使用中间件统一捕获错误：

```go
func TracelyMiddleware(client *tracely.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                client.ReportError(tracely.ErrorPayload{
                    Type:    "panic",
                    Message: fmt.Sprintf("%v", err),
                    Stack:   string(debug.Stack()),
                    URL:     c.Request.URL.String(),
                })
            }
        }()
        c.Next()
    }
}
```

### 3. 结构化日志

结合日志库使用：

```go
func (s *Service) DoSomething() error {
    err := s.doWork()
    if err != nil {
        log.Errorf("do work failed: %v", err)
        
        tracelyClient.ReportError(tracely.ErrorPayload{
            Type:    "serviceError",
            Message: err.Error(),
            Stack:   fmt.Sprintf("%+v", errors.Stack(err)),
            URL:     "service.go:50",
        })
        
        return err
    }
    return nil
}
```

### 4. 环境隔离

根据环境决定是否启用：

```go
if os.Getenv("ENV") == "production" {
    tracelyClient = tracely.New(config)
} else {
    // 开发环境使用空实现或本地日志
    tracelyClient = &NoopClient{}
}
```

## 心跳机制

### 工作原理

SDK 使用 `_active` 内置事件实现服务心跳上报，与 TS SDK 保持一致：

1. **启动即上报**：客户端创建后立即上报一次活跃状态
2. **定时上报**：使用 `time.Ticker` 按配置间隔定时上报（默认 60 秒）
3. **复用事件接口**：底层调用 `ReportEvent("_active", metadata, instanceID)`
4. **自动实例 ID**：未配置 `InstanceID` 时，自动基于主机名 + PID + 时间戳生成唯一标识

### 上报数据格式

心跳事件的 metadata 包含以下字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `instanceId` | string | 实例唯一标识 |
| `duration` | int64 | 服务运行时长（秒，从客户端创建开始计算） |
| `tags` | map | 自定义标签（仅在配置了 Tags 时存在） |

### 使用示例

```go
client := tracely.New(tracely.Config{
    AppID:             os.Getenv("TRACELY_APP_ID"),
    AppSecret:         os.Getenv("TRACELY_APP_SECRET"),
    Host:              os.Getenv("TRACELY_HOST"),
    EnableHeartbeat:   true,
    HeartbeatInterval: 60 * time.Second, // 每 60 秒上报一次
    InstanceID:        "server-01",       // 可选，为空时自动生成
    Tags: map[string]string{
        "env":     "production",
        "version": "1.0.0",
    },
})
```

## 与 TypeScript SDK 的对比

| 特性 | Go SDK | TypeScript SDK |
|------|--------|----------------|
| 上报方式 | 异步队列 | 异步队列 |
| 错误捕获 | 手动上报 | 自动 + 手动 |
| 活跃统计 | ✅（心跳上报） | ✅（页面停留时长） |
| 重试机制 | 3 次 | ❌ |
| 节流控制 | ❌ | ✅（1 分钟） |
| 适用场景 | 后端服务 | 前端应用 |

## 常见问题

### Q1: 为什么上报失败没有报错？

SDK 采用静默失败策略，上报错误不会影响业务逻辑。如需调试，可在 Tracely 服务端查看日志。

### Q2: 如何确保数据不丢失？

SDK 设计为**尽力而为**的上报策略：
- 队列满时会丢弃数据
- 重试 3 次后放弃
- 适用于监控场景，不适用于关键业务数据

如需可靠传输，建议使用消息队列或其他可靠机制。

### Q3: 如何调试 SDK？

1. 检查网络连接
2. 验证 AppID 和 AppSecret 是否正确
3. 查看 Tracely 服务端日志
4. 使用 `net/http` 的 `Transport` 自定义日志

### Q4: 支持哪些 Go 版本？

支持 Go 1.26+ 版本。

### Q5: 如何自定义 HTTP 客户端？

当前版本不支持自定义 HTTP 客户端，如需自定义可 Fork 项目修改。

## 构建和测试

```bash
# 进入 sdk/go 目录
cd sdk/go

# 运行测试
go test -v ./...

# 格式化代码
go fmt ./...

# 检查依赖
go mod tidy
```

## 许可证

MIT License
