[![GitHub License](https://img.shields.io/github/license/hanxi/tracely)](https://github.com/hanxi/tracely)
[![Docker Image Version](https://img.shields.io/docker/v/hanxi/tracely?sort=semver&label=docker%20image)](https://hub.docker.com/r/hanxi/tracely)
[![Docker Pulls](https://img.shields.io/docker/pulls/hanxi/tracely)](https://hub.docker.com/r/hanxi/tracely)
[![GitHub Release](https://img.shields.io/github/v/release/hanxi/tracely)](https://github.com/hanxi/tracely/releases)
[![Visitors](https://api.visitorbadge.io/api/daily?path=hanxi%2Ftracely&label=daily%20visitor&countColor=%232ccce4&style=flat)](https://visitorbadge.io/status?path=hanxi%2Ftracely)
[![Visitors](https://api.visitorbadge.io/api/visitors?path=hanxi%2Ftracely&label=total%20visitor&countColor=%232ccce4&style=flat)](https://visitorbadge.io/status?path=hanxi%2Ftracely)

# Tracely

一个轻量级的前端监控平台，支持 **错误收集** 和 **用户活跃统计**，可自托管部署。

## 功能特性

- 🐛 **错误收集**：自动捕获 JS 运行时错误、Promise 异常、Vue 组件错误
- 📊 **自定义事件**：支持自定义事件上报和统计，灵活的元数据支持
- 📈 **数据概览**：实时展示今日 PV/UV、错误总数、Top 事件排行
- 🔐 **安全认证**：AppID + HMAC 签名验证，时间戳防过期，Nonce 防重放；Dashboard 支持 JWT 登录
- 🚦 **限速保护**：IP 维度限速，防止恶意刷数据
- 🗂️ **错误去重**：相同错误合并记录，统计出现次数
- 🎯 **事件白名单**：配置文件控制允许上报的事件类型，防止随意上报
- 🏗️ **多平台构建**：支持 Linux 多架构（amd64, arm64）
- 🎨 **内嵌 Dashboard**：前端资源打包到后端，单个二进制文件即可运行
- 🌙 **现代化 UI**：基于 Nuxt UI，支持明暗色模式、响应式布局
- 🔄 **多应用支持**：支持多应用配置，可在 Dashboard 中切换查看
- 🧹 **数据清理**：自动定期清理历史事件数据，错误数据永久保留


**在线体验：**
- 体验地址：https://tracely.hanxi.cc/
- 用户名：`admin`
- 密码：`admin123`

---

## 快速开始

### 1. Docker Compose 部署

**一键部署：**

```bash
# 1. 下载 Docker Compose 配置
mkdir tracely && cd tracely
curl -o docker-compose.yaml https://raw.githubusercontent.com/hanxi/tracely/main/docker-compose.yaml
# 2. 启动服务
docker compose up -d
# 3. 访问 Dashboard
# http://localhost:3001
# 用户名：admin
# 密码：你在脚本运行时设置的密码（默认：admin123）
```

**配置说明：**
- `gen-config.sh` 脚本会自动生成 JWT Secret、App Secret 和密码哈希
- 配置文件保存在 `./config/config.yaml`
- 数据持久化到 `./data` 目录
- **无需本地 Go 环境**：所有操作都在 Docker 容器中执行

### 2. 构建

#### 方式一：本地构建

```bash
# 一键构建全部
make build

# 或分步构建
make build-frontend  # 构建 Dashboard
make build-backend   # 编译后端
```

#### 方式二：Docker 构建

```bash
make docker
```

### 3. 运行

```bash
# 本地运行
./tracely

# Docker 运行
docker run -d -p 3001:3001 -v $(pwd)/data:/app/data hanxi/tracely:latest
```

访问 Dashboard：http://localhost:3001

**登录信息：**
- 用户名：`admin`（或你在配置中设置的用户名）
- 密码：你在配置中设置的密码

---

## 项目结构

```
tracely/
├── cmd/
│   ├── server/          # 后端入口
│   └── hashpwd/         # 密码哈希生成工具
├── internal/
│   ├── config/          # 配置加载
│   ├── middleware/      # 中间件（认证、限速、JWT）
│   ├── handler/         # 业务接口（错误、活跃、概览、认证）
│   └── model/           # 数据模型 + 定时清理任务
├── sdk/
│   └── go/              # Go SDK
├── dashboard/           # Vue 3 + Nuxt UI Dashboard
│   ├── src/
│   │   ├── pages/       # 页面（登录、概览、错误、统计）
│   │   ├── components/  # 组件（应用切换、用户菜单）
│   │   ├── stores/      # Pinia 状态管理
│   │   ├── api/         # API 请求封装
│   │   └── layouts/     # 布局
│   └── dist/            # 构建产物（嵌入后端）
├── config.example.yaml  # 配置模板
├── Makefile             # 构建脚本
├── Dockerfile           # Docker 镜像
└── README.md
```

---

## 技术栈

| 模块 | 技术 |
|------|------|
| 后端 | Go + Gin + GORM（支持 Linux）|
| 数据库 | SQLite |
| 后端 SDK | Go |
| 可视化面板 | Vue 3 + Nuxt UI + Vite |

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

### 事件表 `events`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键 |
| event_name | TEXT | 事件名称（如 `_active`、`click_button` 等） |
| metadata | TEXT | 元数据（JSON 格式，可包含 page、duration 等字段） |
| app_id | TEXT | 应用 ID |
| user_id | TEXT | 用户唯一标识 |
| created_at | DATETIME | 创建时间 |

**内置事件**：
- `_active`：用户活跃事件，用于统计 PV/UV

**Metadata 建议格式**：
```json
{
  "page": "/home",
  "duration": 120,
  "custom_field": "value"
}
```

---

## API 设计

### 上报接口（SDK 调用）

所有上报接口需要在请求头中携带以下认证信息：

| 请求头 | 说明 |
|--------|------|
| X-App-Id | 应用 ID |
| X-Timestamp | 当前 Unix 时间戳（秒） |
| X-Nonce | 随机字符串（UUID 去掉横线） |
| X-Signature | HMAC-SHA256 签名 |

**签名算法：** `HMAC-SHA256(appId + timestamp + nonce, appSecret)`

**安全规则：**
- 时间戳与服务器时间差超过 300 秒则拒绝
- 同一 Nonce 只能使用一次（服务端内存存储，5 分钟后清理）
- 同一 IP 每分钟最多请求 60 次

#### POST `/report/error` 上报错误

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

#### POST `/report/event` 上报事件

**请求体：**
```json
{
  "eventName": "_active",
  "metadata": {
    "page": "/home",
    "duration": 30,
    "custom_field": "value"
  },
  "appId": "my-app-id",
  "userId": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应：**
```json
{ "message": "上报成功" }
```

**说明**：
- `eventName` 必须在 `config.yaml` 的事件白名单中
- `metadata` 为可选字段，支持任意 JSON 对象
- `_active` 是内置的活跃事件类型

---

### Dashboard 接口（JWT 认证）

所有接口需要在请求头中携带：`Authorization: Bearer <JWT_TOKEN>`

#### GET `/api/apps` 获取应用列表

获取配置中的应用列表（用于 Dashboard 切换应用）。

**响应：**
```json
{
  "apps": [
    {
      "appId": "my-app-id",
      "appName": "我的应用"
    }
  ]
}
```

#### GET `/api/overview` 获取概览数据

Dashboard 首页数据，展示实时统计信息。

**Query 参数：**

| 参数 | 说明 | 默认值 |
|------|------|--------|
| appID | 应用 ID 筛选 | 全部 |

**响应：**
```json
{
  "todayPV": 1500,
  "todayUV": 420,
  "totalErrors": 85,
  "todayErrors": 12,
  "topErrors": [
    {
      "type": "jsError",
      "message": "Cannot read properties of undefined",
      "count": 25
    }
  ],
  "errorTrend": [
    { "date": "01/01", "count": 5 },
    { "date": "01/02", "count": 8 }
  ]
}
```

#### GET `/api/errors` 获取错误列表

**Query 参数：**

| 参数 | 说明 | 默认值 |
|------|------|--------|
| page | 页码 | 1 |
| pageSize | 每页条数 | 20 |
| type | 错误类型筛选 | 全部 |
| appID | 应用 ID 筛选 | 全部 |

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

#### GET `/api/stats` 获取活跃统计

**Query 参数：**

| 参数 | 说明 | 默认值 |
|------|------|--------|
| days | 统计最近几天 | 7 |
| appID | 应用 ID 筛选 | 全部 |

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

#### POST `/auth/login` 登录

**请求体：**
```json
{
  "username": "admin",
  "password": "yourpassword"
}
```

**响应：**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "admin"
}
```

---

## Go SDK 使用

### 安装

```bash
go get github.com/hanxi/tracely/sdk/go
```

### 快速使用

```go
import "github.com/hanxi/tracely/sdk/go"

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

    // 上报活跃事件（内置事件类型 _active）
    client.ReportEvent("_active", nil, "user-123", "/api/user/login", 0)
}
```

### 手动上报事件

```go
// 上报自定义事件
client.ReportEvent("click_button", map[string]interface{}{
    "button_id": "submit",
    "page": "/checkout",
}, "user-123")

// 上报活跃事件（内置事件类型 _active）
client.ReportEvent("_active", map[string]interface{}{
    "page": "/home",
    "duration": 30,
}, "user-123")
```

### Gin 框架集成示例

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/hanxi/tracely/sdk/go"
    "time"
)

func main() {
    client := tracely.New(tracely.Config{
        AppID:     "my-app-id",
        AppSecret: "my-app-secret",
        Host:      "http://你的服务器:3001",
    })

    r := gin.New()

    // 自定义中间件：捕获 panic
    r.Use(func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                client.ReportError(tracely.ErrorPayload{
                    Type:    "panicError",
                    Message: fmt.Sprintf("%v", err),
                    Stack:   string(debug.Stack()),
                    URL:     c.FullPath(),
                })
                c.AbortWithStatus(500)
            }
        }()
        c.Next()
    })

    // 自定义中间件：统计接口访问
    r.Use(func(c *gin.Context) {
        start := time.Now()
        c.Next()
        duration := int(time.Since(start).Seconds())
        
        // 上报活跃事件
        client.ReportEvent("_active", map[string]interface{}{
            "page": c.FullPath(),
            "duration": duration,
        }, "user-id")
    })

    r.Run(":8080")
}
```

### SDK 特性

- **异步上报**：内置缓冲队列，上报失败不影响主业务
- **自动重试**：上报失败自动重试，最多重试 3 次
- **无框架依赖**：纯函数接口，可集成到任意 Go 框架（Gin、Echo、Fiber 等）
- **灵活的事件系统**：支持自定义事件名称和元数据

---

## Dashboard 面板页面

### 📊 概览页 `/`
- **数据卡片**：今日事件总数、今日活跃 PV、今日活跃 UV、错误总数
- **Top 5 事件**：展示出现次数最多的事件列表（事件名称、次数）
- 快速跳转到错误列表页和事件统计页

### 🐛 错误列表页 `/errors`
- 表格展示所有错误，字段：错误类型、错误信息、出现次数、最近出现
- 支持按错误类型筛选（全部 / jsError / promiseError / vueError）
- 支持分页（每页 20 条）
- 点击"详情"按钮查看完整错误信息（类型、消息、堆栈、URL、首次/最近出现时间）
- 支持多应用切换查看

### 📈 事件统计页 `/events`
- **事件类型分布**：展示所有事件类型及其数量
- **每日事件趋势**：表格展示每日事件统计数据
- **Top 10 事件排行**：展示最热门的事件（支持筛选事件类型）
- 支持切换统计天数（7 天 / 14 天 / 30 天）
- 支持按事件类型筛选
- 支持多应用切换查看

### 🔐 登录页 `/login`
- 用户名 + 密码登录
- JWT Token 认证
- 登录状态持久化（localStorage）
- 路由守卫保护

### 🎨 通用功能
- **明暗色模式**：基于 Nuxt UI 自动适配
- **响应式布局**：基于 Tailwind CSS
- **Hash 路由**：使用 `createWebHashHistory`
- **用户菜单**：显示当前用户，支持退出登录
- **应用切换**：多应用配置时显示切换下拉框（从 `/api/apps` 接口加载）

---

## 数据清理策略

- **事件数据**：根据 `config.yaml` 中每个事件的 `retentionDays` 配置自动清理（0 表示永久保留）
- **错误日志**：永久保留（不清理），方便历史问题排查和趋势分析

**注意**：活跃事件（`_active`）是一种特殊的自定义事件，默认保留 90 天。

---

## 部署

### Docker 部署

```bash
docker run -d -p 3001:3001 -v $(pwd)/data:/app/data -v $(pwd)/config.yaml:/app/config.yaml hanxi/tracely:latest
```

### 注意事项

- AppSecret 在前端是可见的，建议对打包产物进行代码混淆
- SQLite 适合中小流量，日上报量建议不超过 10 万条
- 生产环境建议在前面挂 Nginx 做反向代理并配置 HTTPS
- 定期备份 `data/tracely.db` 数据库文件
- Dashboard 构建产物已嵌入后端二进制文件

---

## 相关文档

- [架构文档](./ARCHITECTURE.md) - AI 友好的架构说明文档
- [配置模板](./config.example.yaml) - 完整的配置示例

---

## 🙏 致谢

感谢使用 Tracely！如有问题或建议，欢迎提交 Issue 或 PR。
