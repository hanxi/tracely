# Tracely Go SDK

轻量级后端监控客户端，支持错误上报和事件追踪。

完整文档请参阅 [docs/sdk-go.md](../../docs/sdk-go.md)。

## 安装

```bash
go get github.com/hanxi/tracely/sdk/go/tracely
```

## 快速开始

```go
client := tracely.New(tracely.Config{
    AppID:     "my-app-id",
    AppSecret: "my-app-secret",
    Host:      "https://tracely.example.com",
})

client.ReportError(tracely.ErrorPayload{
    Type:    "runtimeError",
    Message: "something went wrong",
})

client.ReportEvent("user_action", map[string]interface{}{
    "action": "click",
}, "user-123")
```
