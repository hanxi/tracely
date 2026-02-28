# Tracely

一个轻量级的前端监控平台，支持 **错误收集** 和 **用户活跃统计**，可自托管部署。

## 功能特性

- 🐛 **错误收集**：自动捕获 JS 运行时错误、Promise 异常、Vue 组件错误
- 📊 **活跃统计**：统计 PV/UV、页面停留时长、热门页面排行
- 📈 **数据概览**：实时展示今日 PV/UV、错误总数、Top 错误列表
- 🔐 **安全认证**：AppID + HMAC 签名验证，时间戳防过期，Nonce 防重放；Dashboard 支持 JWT 登录
- 🚦 **限速保护**：IP 维度限速，防止恶意刷数据
- 🗂️ **错误去重**：相同错误合并记录，统计出现次数
          - 多平台二进制构建（Linux, macOS）
- 🎨 **内嵌 Dashboard**：前端资源打包到后端，单个二进制文件即可运行
- 🌙 **现代化 UI**：基于 Nuxt UI，支持明暗色模式、响应式布局
- 🔄 **多应用支持**：支持多应用配置，可在 Dashboard 中切换查看
- 🧹 **数据清理**：自动定期清理历史活跃数据，错误数据永久保留

---

## 快速开始

### 方案一：Docker Compose 部署（推荐）

**一键部署：**

```bash
# 1. 生成配置文件（在 Docker 容器中执行，无需本地 Go 环境）
docker-compose run --rm tracely ./scripts/gen-config.sh

# 2. 启动服务
docker-compose up -d

# 3. 访问 Dashboard
# http://localhost:3001
# 用户名：admin
# 密码：你在脚本运行时设置的密码（默认：admin123）
```

**配置说明：**
- `gen-config.sh` 脚本会自动生成 JWT Secret、App Secret 和密码哈希
- 配置文件保存在 `config.yaml`
- 数据持久化到 `./data` 目录
- **无需本地 Go 环境**：所有操作都在 Docker 容器中执行

### 方案二：手动配置

复制配置模板并修改：

```bash
cp config.example.yaml config.yaml
```

编辑 `config.yaml`：

```yaml
port: "3001"
dbPath: "./data/tracely.db"
rateLimit: 60
nonceTTL: 300
timestampTTL: 300

# 数据清理配置
activeLogRetentionDays: 90  # 活跃日志保留天数（0=不清理）

# JWT 配置（Dashboard 登录）
jwt:
  secret: "your-jwt-secret-please-change-this-to-32-chars"
  expireHours: 24

# 多应用配置（SDK 上报）
apps:
  - appId: "my-app-id"
    appSecret: "my-app-secret-please-change-this-to-32-chars"

# 多用户配置（Dashboard 登录）
users:
  - username: "admin"
    passwordHash: "$2a$10$..."  # 使用工具生成
```

生成密码哈希：

```bash
go run main.go -hashpwd -password "yourpassword"
```

将生成的哈希值复制到 `config.yaml` 的 `users[].passwordHash` 字段。

生成随机 Secret：

```bash
go run main.go -generate-secret -secret-length 32
```

### 方案三：使用环境变量（快速测试）

不想创建 `config.yaml`？可以直接在 `docker-compose.yml` 中使用环境变量：

```yaml
environment:
  - JWT_SECRET=your-secret-here
  - USERS=admin:$2a$10$...
  - APPS=my-app-id:我的应用:my-secret-here
```

取消 `docker-compose.yml` 中 environment 配置的注释并修改值即可。

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
| 后端 | Go + Gin + GORM（支持 Linux、macOS）|
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

#### POST `/report/active` 上报活跃

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
    "github.com/hanxi/tracely/sdk/go"
    tracely_gin "github.com/hanxi/tracely/sdk/go/middleware/gin"
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

### SDK 特性

- **异步上报**：内置缓冲队列，上报失败不影响主业务
- **自动重试**：上报失败自动重试，最多重试 3 次
- **Gin 集成**：提供 Recovery 和 Tracker 中间件，一行代码接入

---

## Dashboard 面板页面

### 📊 概览页 `/`
- **数据卡片**：今日 PV、今日 UV、错误总数、今日新增错误
- **Top 5 错误**：展示出现次数最多的错误列表（类型、消息、次数）
- 快速跳转到错误列表页

### 🐛 错误列表页 `/errors`
- 表格展示所有错误，字段：错误类型、错误信息、出现次数、最近出现
- 支持按错误类型筛选（全部 / jsError / promiseError / vueError）
- 支持分页（每页 20 条）
- 点击"详情"按钮查看完整错误信息（类型、消息、堆栈、URL、首次/最近出现时间）
- 支持多应用切换查看

### 📈 活跃统计页 `/stats`
- 每日统计卡片展示（7/14/30 天可选）
- 表格展示热门页面排行（页面路径、PV、平均停留时长）
- 支持切换统计天数（7 天 / 14 天 / 30 天）
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

- **活跃日志**：每天凌晨 3 点自动清理 N 天前的数据（N 由 `activeLogRetentionDays` 配置，默认 90 天）
- **错误日志**：永久保留（不清理），方便历史问题排查和趋势分析

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
