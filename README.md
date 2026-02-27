# Tracely

一个轻量级的前端监控平台，支持 **错误收集** 和 **用户活跃统计**，可自托管部署。

## 功能特性

- 🐛 **错误收集**：自动捕获 JS 运行时错误、Promise 异常、Vue 组件错误
- 📊 **活跃统计**：统计 PV/UV、页面停留时长、热门页面排行
- 🔐 **安全认证**：AppID + HMAC 签名验证，时间戳防过期，Nonce 防重放
- 🚦 **限速保护**：IP 维度限速，防止恶意刷数据
- 🗂️ **错误去重**：相同错误合并记录，统计出现次数
- 🪶 **轻量部署**：Go + SQLite，单二进制文件，无外部依赖

---

## 项目结构

```
tracely/
├── server/                      # Go 后端服务
│   ├── main.go
│   ├── config/
│   │   └── config.go            # AppID/Secret、环境变量配置
│   ├── middleware/
│   │   ├── auth.go              # 签名验证中间件
│   │   └── ratelimit.go         # IP 限速中间件
│   ├── handler/
│   │   ├── error.go             # 报错收集接口
│   │   └── active.go            # 活跃统计接口
│   ├── model/
│   │   ├── error_log.go         # 报错数据模型
│   │   └── active_log.go        # 活跃数据模型
│   └── Dockerfile
│
├── sdk/
│   ├── ts/                      # 前端 SDK（TypeScript）
│   │   ├── src/
│   │   │   ├── index.ts         # 入口，统一导出
│   │   │   ├── error.ts         # 报错采集模块
│   │   │   ├── tracker.ts       # 活跃统计模块
│   │   │   └── request.ts       # 签名请求封装
│   │   ├── package.json
│   │   └── tsconfig.json
│   │
│   └── go/                      # Go SDK
│       ├── client.go            # 核心客户端，ReportError / ReportActive
│       ├── sign.go              # HMAC 签名生成
│       ├── payload.go           # ErrorPayload / ActivePayload 结构体定义
│       ├── queue.go             # 异步上报队列，失败重试
│       ├── middleware/
│       │   └── gin/
│       │       ├── recovery.go  # Gin panic 捕获中间件
│       │       └── tracker.go   # Gin 请求追踪中间件
│       ├── go.mod
│       └── README.md            # Go SDK 单独说明文档
│
├── dashboard/                   # 可视化面板（Vue3）
│   ├── src/
│   │   ├── views/
│   │   │   ├── ErrorList.vue    # 报错列表页
│   │   │   └── ActiveStats.vue  # 活跃统计页
│   │   └── main.ts
│   └── package.json
│
├── docker-compose.yml
└── README.md
```

---

## 技术栈

| 模块 | 技术 |
|------|------|
| 后端 | Go + Gin + GORM |
| 数据库 | SQLite |
| 前端 SDK | TypeScript |
| 可视化面板 | Vue3 + TypeScript |

---

## 数据库设计

### 错误表 `error_logs`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键 |
| fingerprint | TEXT | 错误唯一指纹（唯一索引），用于去重 |
| type | TEXT | 错误类型：jsError / promiseError / vueError |
| message | TEXT | 错误信息 |
| stack | TEXT | 错误堆栈 |
| url | TEXT | 发生错误的页面地址 |
| app_id | TEXT | 应用 ID |
| user_agent | TEXT | 浏览器 UA |
| count | INTEGER | 出现次数，默认 1 |
| first_seen | DATETIME | 首次出现时间 |
| last_seen | DATETIME | 最近出现时间 |

**指纹生成规则：** `MD5(appId + type + message)`

### 活跃表 `active_logs`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键 |
| app_id | TEXT | 应用 ID |
| user_id | TEXT | 用户唯一标识（前端 localStorage 生成的 UUID） |
| page | TEXT | 页面路径 |
| duration | INTEGER | 停留时长（秒） |
| user_agent | TEXT | 浏览器 UA |
| created_at | DATETIME | 上报时间 |

---

## API 设计

所有接口需要在请求头中携带以下认证信息：

| 请求头 | 说明 |
|--------|------|
| X-App-Id | 应用 ID |
| X-Timestamp | 当前 Unix 时间戳（秒） |
| X-Nonce | 随机字符串（UUID去掉横线） |
| X-Signature | HMAC-SHA256 签名 |

**签名算法：** `HMAC-SHA256(appId + timestamp + nonce, appSecret)`

**安全规则：**
- 时间戳与服务器时间差超过 300 秒则拒绝
- 同一 Nonce 只能使用一次（服务端内存存储，5分钟后清理）
- 同一 IP 每分钟最多请求 60 次

### POST `/api/error` 上报错误

**请求体：**
```json
{
  "type": "jsError",
  "message": "Cannot read properties of undefined",
  "stack": "TypeError: Cannot read...\n    at xxx.js:10:5",
  "url": "https://example.com/home",
  "appId": "my-app-id"
}
```

**响应：**
```json
{ "message": "上报成功" }
```

**逻辑：**
1. 根据 `appId + type + message` 生成 MD5 指纹
2. 查询数据库是否存在相同指纹
3. 存在则更新 `count + 1`、`last_seen`、`stack`、`url`
4. 不存在则新增记录

---

### POST `/api/active` 上报活跃

**请求体：**
```json
{
  "appId": "my-app-id",
  "userId": "550e8400-e29b-41d4-a716-446655440000",
  "page": "/home",
  "duration": 30
}
```

**响应：**
```json
{ "message": "ok" }
```

---

### GET `/api/errors` 获取错误列表

**Query 参数：**

| 参数 | 说明 | 默认值 |
|------|------|--------|
| page | 页码 | 1 |
| pageSize | 每页条数 | 20 |
| type | 错误类型筛选 | 全部 |

**响应：**
```json
{
  "total": 100,
  "list": [
    {
      "id": 1,
      "type": "jsError",
      "message": "Cannot read properties of undefined",
      "stack": "TypeError...",
      "url": "https://example.com/home",
      "count": 42,
      "firstSeen": "2024-01-01T00:00:00Z",
      "lastSeen": "2024-01-02T00:00:00Z"
    }
  ]
}
```

---

### GET `/api/stats` 获取活跃统计

**Query 参数：**

| 参数 | 说明 | 默认值 |
|------|------|--------|
| days | 统计最近几天 | 7 |

**响应：**
```json
{
  "daily": [
    { "date": "2024-01-01", "pv": 1000, "uv": 300 },
    { "date": "2024-01-02", "pv": 1200, "uv": 350 }
  ],
  "topPages": [
    { "page": "/home", "pv": 500, "avgDuration": 45 },
    { "page": "/about", "pv": 300, "avgDuration": 30 }
  ]
}
```

---

## 前端 SDK 设计

### 安装

```bash
npm install tracely-sdk
```

### 使用

```ts
import { Tracely } from "tracely-sdk";

const tracely = new Tracely({
  appId: "my-app-id",
  appSecret: "my-app-secret",
  host: "http://你的服务器:3001",
});

// Vue 中一键初始化（传入 app 实例和 router）
tracely.init(app, router);
```

### SDK 内部模块说明

**`request.ts`**
- 封装 `signedFetch` 方法
- 自动生成 `timestamp`、`nonce`、`signature` 并注入请求头
- 签名算法：`HMAC-SHA256(appId + timestamp + nonce, appSecret)`
- 使用 `crypto-js` 库实现 HMAC 签名

**`error.ts`**
- 监听 `window.error` 事件，捕获 JS 运行时错误
- 监听 `window.unhandledrejection` 事件，捕获 Promise 异常
- 注册 `app.config.errorHandler`，捕获 Vue 组件内部错误
- 调用 `signedFetch` 上报到 `/api/error`

**`tracker.ts`**
- 从 `localStorage` 读取或生成用户唯一 `userId`（UUID格式）
- 记录页面进入时间，路由切换或页面关闭时计算停留时长
- 配合 Vue Router 的 `afterEach` 钩子上报页面切换
- 监听 `beforeunload` 事件上报最后一个页面的停留时长
- 调用 `signedFetch` 上报到 `/api/active`

---

## Go SDK 设计

方便其他 Go 应用接入 Tracely，上报服务端错误和自定义事件。

### 安装

```bash
go get github.com/yourname/tracely-go
```

### 快速使用

```go
import "github.com/yourname/tracely-go"

func main() {
    client := tracely.New(tracely.Config{
        AppID:     "my-app-id",
        AppSecret: "my-app-secret",
        Host:      "http://你的服务器:3001",
    })

    // 手动上报错误
    client.ReportError(tracely.ErrorPayload{
        Type:    "serverError",
        Message: err.Error(),
        Stack:   string(debug.Stack()),
        URL:     "/api/user/login",
    })

    // 上报活跃/事件
    client.ReportActive(tracely.ActivePayload{
        UserID: "user-123",
        Page:   "/api/user/login",
    })
}
```

### Gin 中间件一键接入

```go
import (
    "github.com/yourname/tracely-go"
    tracely_gin "github.com/yourname/tracely-go/middleware/gin"
)

func main() {
    client := tracely.New(tracely.Config{
        AppID:     "my-app-id",
        AppSecret: "my-app-secret",
        Host:      "http://你的服务器:3001",
    })

    r := gin.New()

    // 自动捕获 panic 和请求信息
    r.Use(tracely_gin.Recovery(client))
    // 自动统计接口访问（上报到活跃统计）
    r.Use(tracely_gin.Tracker(client))

    r.Run(":8080")
}
```

### SDK 内部模块说明

**`client.go`**
- 核心客户端，持有 `Config`
- 提供 `ReportError(ErrorPayload)` 方法
- 提供 `ReportActive(ActivePayload)` 方法
- 内置异步上报队列，上报失败不影响主业务
- 上报失败自动重试，最多重试 3 次

**`sign.go`**
- 生成 `timestamp`、`nonce`、`signature`
- 签名算法：`HMAC-SHA256(appId + timestamp + nonce, appSecret)`

**`middleware/gin/recovery.go`**
- 捕获 `panic`，自动上报到 `/api/error`
- 上报字段：错误信息、堆栈、请求路径、请求方法、客户端 IP
- 上报完成后正常返回 500 响应，不影响服务

**`middleware/gin/tracker.go`**
- 每次请求结束后上报到 `/api/active`
- 上报字段：请求路径、耗时（作为 duration）、客户端 IP

### ErrorPayload 结构

| 字段 | 类型 | 说明 |
|------|------|------|
| Type | string | 错误类型，建议：serverError / panicError / dbError 等 |
| Message | string | 错误信息 |
| Stack | string | 堆栈信息，可用 `runtime/debug.Stack()` 获取 |
| URL | string | 发生错误的接口路径 |

### ActivePayload 结构

| 字段 | 类型 | 说明 |
|------|------|------|
| UserID | string | 用户唯一标识，没有可传空字符串 |
| Page | string | 接口路径或事件名称 |
| Duration | int | 耗时（毫秒） |

---

## 后端配置

通过环境变量或配置文件进行配置：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| PORT | 服务端口 | 3001 |
| DB_PATH | SQLite 文件路径 | ./tracely.db |
| RATE_LIMIT | 每IP每分钟最大请求数 | 60 |
| NONCE_TTL | Nonce 有效期（秒） | 300 |
| TIMESTAMP_TTL | 时间戳有效期（秒） | 300 |

**AppID 和 AppSecret 配置（支持多应用）：**

```yaml
# config.yaml
apps:
  - appId: "my-app-id"
    appSecret: "my-app-secret"
  - appId: "another-app-id"
    appSecret: "another-app-secret"
```

---

## 部署

### Docker Compose 一键部署

```bash
git clone https://github.com/yourname/tracely.git
cd tracely
docker-compose up -d
```

**`docker-compose.yml`：**

```yaml
services:
  server:
    build: ./server
    ports:
      - "3001:3001"
    volumes:
      - ./data:/app/data      # SQLite 数据持久化
      - ./config.yaml:/app/config.yaml
    restart: unless-stopped
    environment:
      - PORT=3001
      - DB_PATH=/app/data/tracely.db

  dashboard:
    build: ./dashboard
    ports:
      - "8080:80"
    restart: unless-stopped
```

---

## Dashboard 面板页面

### 错误列表页 `ErrorList.vue`
- 表格展示所有错误，字段：错误类型、错误信息、出现次数、首次出现、最近出现
- 支持按错误类型筛选（jsError / promiseError / vueError）
- 支持按出现次数排序
- 点击某条错误可展开查看完整 Stack Trace

### 活跃统计页 `ActiveStats.vue`
- 折线图展示最近 7 天 PV/UV 趋势
- 表格展示热门页面排行（页面路径、PV、平均停留时长）
- 支持切换统计天数（7天 / 14天 / 30天）

---

## 注意事项

- AppSecret 在前端是可见的，建议对打包产物进行代码混淆
- SQLite 适合中小流量，日上报量建议不超过 10 万条
- 生产环境建议在前面挂 Nginx 做反向代理并配置 HTTPS

