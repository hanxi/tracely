# Tracely 架构文档

> 本文档主要面向 AI 开发者，详细描述 Tracely 的架构设计、核心模块、数据流和扩展点。

---

## 一、系统架构总览

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Side                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Browser JS  │  │  Go Client  │  │   Dashboard (Vue)   │ │
│  │   + SDK     │  │   + SDK     │  │                     │ │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
│         │                │                     │            │
│         │ HMAC 签名       │ HMAC 签名            │ JWT Token   │
│         │ 上报错误/活跃   │ 上报错误/活跃        │ 查询数据    │
└─────────┼────────────────┼─────────────────────┼────────────┘
          │                │                     │
          ▼                ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│                      Tracely Server                         │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────┐  │
│  │                   Gin Router                        │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │  /report/* (HMAC 签名验证 + 限速)  │  /api/* (JWT)   │  │
│  │  /auth/login (登录)                │  /* (SPA)       │  │
│  └──────────────────────────────────────────────────────┘  │
│         │                    │                              │
│         ▼                    ▼                              │
│  ┌─────────────┐      ┌─────────────┐                      │
│  │  Handlers   │      │ Middleware  │                      │
│  │  - error    │      │ - SignAuth  │                      │
│  │  - active   │      │ - JWTAuth   │                      │
│  │  - overview │      │ - RateLimit │                      │
│  │  - stats    │      │ - CORS      │                      │
│  │  - auth     │      │             │                      │
│  └──────┬──────┘      └─────────────┘                      │
│         │                                                   │
│         ▼                                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                  GORM + SQLite                      │   │
│  │  - error_logs (错误表，永久保留)                     │   │
│  │  - active_logs (活跃表，定期清理)                    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Background Tasks                       │   │
│  │  - Nonce 清理 (每 5 分钟)                             │   │
│  │  - 活跃日志清理 (每天凌晨 3 点)                        │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 二、核心模块详解

### 2.1 后端架构 (`internal/`)

#### 目录结构

```
internal/
├── config/          # 配置加载与管理
├── middleware/      # Gin 中间件
├── handler/         # HTTP 请求处理
└── model/           # 数据模型 + 数据库操作
```

#### 2.1.1 配置层 (`config/config.go`)

**职责：** 加载和管理配置，支持环境变量覆盖

**核心结构：**
```go
type Config struct {
    Port                   string
    DBPath                 string
    RateLimit              int
    NonceTTL               int
    TimestampTTL           int
    ActiveLogRetentionDays int
    JWT                    JWT
    Apps                   []App
    Users                  []User
}
```

**关键方法：**
- `Load()` - 加载配置（优先级：环境变量 > config.yaml > 默认值）
- `GetSecret(appID)` - 根据 AppID 获取密钥
- `GetUser(username)` - 根据用户名获取用户配置

**扩展点：**
- 新增配置项：在 `Config` 结构体添加字段，在 `Load()` 中支持环境变量
- 新增配置源：修改 `Load()` 方法，支持从数据库、配置中心等加载

#### 2.1.2 中间件层 (`middleware/`)

**文件清单：**
- `auth.go` - HMAC 签名验证中间件
- `jwt.go` - JWT Token 验证中间件
- `ratelimit.go` - IP 限速中间件

**`SignAuth` 中间件流程：**
```
1. 检查请求头 (X-App-Id, X-Timestamp, X-Nonce, X-Signature)
   ↓
2. 根据 AppID 查找 Secret
   ↓
3. 验证时间戳（与服务器时间差 < TimestampTTL）
   ↓
4. 验证 Nonce 是否已使用（防重放攻击）
   ↓
5. 计算签名并比对 HMAC-SHA256(appId+timestamp+nonce, secret)
   ↓
6. 验证通过，继续处理请求
```

**`JWTAuth` 中间件流程：**
```
1. 从 Authorization 头提取 Bearer Token
   ↓
2. 解析并验证 JWT Token（签名、有效期）
   ↓
3. 将用户名写入上下文 (c.Set("username", ...))
   ↓
4. 验证通过，继续处理请求
```

**`RateLimit` 中间件流程：**
```
1. 获取客户端 IP
   ↓
2. 从 sync.Map 获取该 IP 的请求时间戳列表
   ↓
3. 过滤掉 60 秒前的记录（滑动窗口）
   ↓
4. 统计剩余有效请求数
   ↓
5. 超过限制返回 429，否则添加当前时间戳并继续
```

**扩展点：**
- 新增中间件：在 `middleware/` 创建新文件，实现 `gin.HandlerFunc`
- 修改认证逻辑：修改对应中间件的验证流程

#### 2.1.3 处理器层 (`handler/`)

**文件清单：**
- `error.go` - 错误上报和查询接口
- `active.go` - 活跃上报和统计接口
- `overview.go` - 概览数据接口
- `auth.go` - 登录认证接口

**接口设计原则：**
- 所有 handler 接收 `*gorm.DB` 参数，直接操作数据库
- 返回统一的 JSON 格式：成功 `{ "message": "..." }`，失败 `{ "error": "..." }`
- 错误处理：参数错误返回 400，认证失败返回 401，服务器错误返回 500

**核心接口：**

| 接口 | 方法 | 路径 | 认证方式 | 说明 |
|------|------|------|----------|------|
| 获取应用列表 | GET | `/api/apps` | JWT Token | Dashboard 获取应用列表 |
| 上报错误 | POST | `/report/error` | HMAC 签名 | SDK 调用 |
| 上报活跃 | POST | `/report/active` | HMAC 签名 | SDK 调用 |
| 获取错误列表 | GET | `/api/errors` | JWT Token | Dashboard 调用 |
| 获取统计数据 | GET | `/api/stats` | JWT Token | Dashboard 调用 |
| 获取概览数据 | GET | `/api/overview` | JWT Token | Dashboard 调用 |
| 登录 | POST | `/auth/login` | 无 | Dashboard 调用 |

**扩展点：**
- 新增接口：在 `handler/` 创建新文件，实现 handler 函数，在 `main.go` 注册路由
- 修改接口逻辑：直接修改对应 handler 文件

#### 2.1.4 模型层 (`model/`)

**文件清单：**
- `error_log.go` - 错误日志模型
- `active_log.go` - 活跃日志模型
- `db.go` - 数据库初始化（如果有独立文件）

**数据模型：**

```go
// 错误日志
type ErrorLog struct {
    ID          uint   `gorm:"primaryKey"`
    Fingerprint string `gorm:"uniqueIndex"` // 去重查询
    Type        string `gorm:"index"`       // 按类型筛选
    Message     string
    Stack       string
    URL         string
    AppID       string `gorm:"index"` // 按应用筛选
    UserAgent   string
    Count       int    `gorm:"default:1"`
    FirstSeen   time.Time
    LastSeen    time.Time `gorm:"index"` // 按最近出现排序
}

// 活跃日志
type ActiveLog struct {
    ID        uint   `gorm:"primaryKey"`
    AppID     string `gorm:"index"`
    UserID    string `gorm:"index:idx_user_page"`
    Page      string `gorm:"index:idx_user_page"`
    Duration  int
    UserAgent string
    CreatedAt time.Time `gorm:"index"`
}
```

**关键函数：**
- `GenFingerprint(appID, errType, message)` - 生成错误指纹（MD5）
- `InitDB(dbPath)` - 初始化数据库连接，执行 AutoMigrate
- `StartActiveLogCleanup(db, retentionDays)` - 启动定时清理任务

**数据库优化：**
```go
// WAL 模式提升并发性能
db.Exec("PRAGMA journal_mode=WAL;")
db.Exec("PRAGMA synchronous=NORMAL;")
db.Exec("PRAGMA cache_size=-65536;")
db.SetMaxOpenConns(1) // SQLite 只支持单写
```

**扩展点：**
- 新增表：在 `model/` 创建新结构体，添加 `InitDB` 中 AutoMigrate
- 修改表结构：修改结构体字段，GORM 会自动迁移
- 新增查询方法：在对应 model 文件中添加函数

---

### 2.2 Dashboard 架构 (`dashboard/`)

#### 技术栈

```
Vue 3 (Composition API)
├── Nuxt UI (组件库)
├── Vue Router (路由，Hash 模式)
├── Pinia (状态管理，带持久化插件)
├── Axios (HTTP 客户端)
└── Tailwind CSS (样式)
```

#### 目录结构

```
dashboard/
├── src/
│   ├── main.ts              # 应用入口（含路由守卫）
│   ├── App.vue              # 根组件
│   ├── pages/               # 页面组件
│   │   ├── index.vue        # 概览页
│   │   ├── errors.vue       # 错误列表页
│   │   ├── stats.vue        # 活跃统计页
│   │   └── login.vue        # 登录页
│   ├── components/          # 可复用组件
│   │   ├── AppSwitcher.vue  # 应用切换
│   │   └── UserMenu.vue     # 用户菜单
│   ├── api/                 # API 请求封装
│   │   ├── apps.ts          # 应用列表接口
│   │   ├── auth.ts          # 认证接口
│   │   ├── errors.ts        # 错误接口
│   │   ├── index.ts         # Axios 实例配置
│   │   ├── overview.ts      # 概览接口
│   │   └── stats.ts         # 统计接口
│   ├── stores/              # Pinia 状态管理
│   │   ├── auth.ts          # 认证状态（持久化）
│   │   └── app.ts           # 应用状态（持久化）
│   └── layouts/             # 布局组件
│       └── default.vue
├── dist/                    # 构建产物（嵌入后端）
└── vite.config.ts           # Vite 配置（含 API 代理）
```

### 🎨 通用功能
- **明暗色模式**：基于 Nuxt UI 自动适配
- **响应式布局**：基于 Tailwind CSS
- **Hash 路由**：使用 `createWebHashHistory`
- **用户菜单**：显示当前用户，支持退出登录
- **应用切换**：多应用配置时显示切换下拉框（从 `/api/apps` 接口加载）

---

### 2.3 Go SDK 架构 (`sdk/go/`)

#### 目录结构

```
sdk/go/
├── client.go          # 客户端核心
├── payload.go         # 请求体结构
├── sign.go            # 签名工具
├── queue.go           # 异步队列
└── middleware/
    └── gin/
        ├── recovery.go  # Panic 捕获中间件
        └── tracker.go   # 请求追踪中间件
```

#### 核心设计

**2.3.1 异步队列设计**

```go
type reportTask struct {
    url     string
    body    interface{}
    headers map[string]string
}

type Client struct {
    config     Config
    httpClient *http.Client
    queue      chan *reportTask  // 缓冲容量 100
}

func (c *Client) startQueueWorker() {
    go func() {
        for task := range c.queue {
            // 发送请求，失败重试 3 次
            c.sendWithRetry(task)
        }
    }()
}
```

**特点：**
- 缓冲 channel 容量 100，队列满时丢弃（不阻塞业务）
- 后台 goroutine 消费队列，异步上报
- 失败自动重试 3 次，间隔 1 秒

**2.3.2 Gin 中间件**

```go
// middleware/gin/recovery.go
func Recovery(client *Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                client.ReportError(ErrorPayload{
                    Type:    "panicError",
                    Message: fmt.Sprintf("%v", err),
                    Stack:   string(debug.Stack()),
                    URL:     c.FullPath(),
                })
                c.AbortWithStatus(500)
            }
        }()
        c.Next()
    }
}

// middleware/gin/tracker.go
func Tracker(client *Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        defer func() {
            client.ReportActive(ActivePayload{
                Page:     c.FullPath(),
                Duration: int(time.Since(start).Milliseconds()),
            })
        }()
        c.Next()
    }
}
```

**扩展点：**
- 新增中间件：在 `middleware/` 创建新框架的中间件（如 Echo、Fiber）
- 修改上报逻辑：修改 `client.go` 中的 `ReportError`/`ReportActive`

---

## 三、数据流详解

### 3.1 错误上报流程

```
┌──────────────┐
│  Client SDK  │
│  (Go/TS)     │
└──────┬───────┘
       │ 1. 生成签名
       │    - timestamp
       │    - nonce
       │    - signature = HMAC-SHA256(appId+ts+nonce, secret)
       │
       │ 2. POST /report/error
       │    Headers: X-App-Id, X-Timestamp, X-Nonce, X-Signature
       │    Body: { type, message, stack, url, appId }
       │
       ▼
┌──────────────────────────────────────┐
│           Tracely Server             │
│                                      │
│  3. SignAuth 中间件验证              │
│     - 检查请求头完整性               │
│     - 根据 AppID 查找 Secret           │
│     - 验证时间戳（±300 秒）            │
│     - 验证 Nonce 未使用（防重放）      │
│     - 验证签名正确性                 │
│                                      │
│  4. RateLimit 中间件限速             │
│     - IP 维度，60 次/分钟              │
│                                      │
│  5. handler.ReportError              │
│     - 生成指纹：MD5(appId+type+message) │
│     - 查询数据库是否存在             │
│        - 存在：count+1, 更新 last_seen │
│        - 不存在：插入新记录           │
│                                      │
│  6. 返回 { "message": "上报成功" }    │
└──────────────────────────────────────┘
       │
       │ 7. 响应
       │
       ▼
┌──────────────┐
│  Client SDK  │
│  异步队列消费│
│  失败重试 3 次 │
└──────────────┘
```

### 3.2 Dashboard 查询流程

```
┌──────────────┐
│   Browser    │
│   (Vue App)  │
└──────┬───────┘
       │ 1. 用户登录
       │    POST /auth/login
       │    Body: { username, password }
       │
       ▼
┌──────────────────────────────────────┐
│           Tracely Server             │
│                                      │
│  2. handler.Login                    │
│     - 查找用户配置                   │
│     - bcrypt 验证密码                 │
│     - 生成 JWT Token（24 小时有效期）   │
│                                      │
│  3. 返回 { token, username }         │
└──────────────────────────────────────┘
       │
       │ 4. 存储 Token 到 localStorage
       │
       │ 5. 查询数据（携带 Token）
       │    GET /api/overview
       │    Headers: Authorization: Bearer <token>
       │
       ▼
┌──────────────────────────────────────┐
│           Tracely Server             │
│                                      │
│  6. JWTAuth 中间件验证               │
│     - 解析 Token                      │
│     - 验证签名和有效期               │
│     - 将用户名写入上下文             │
│                                      │
│  7. handler.Overview                 │
│     - 查询今日 PV/UV                 │
│     - 查询错误统计                   │
│     - 查询 Top 错误                   │
│                                      │
│  8. 返回 JSON 数据                    │
└──────────────────────────────────────┘
       │
       │ 9. 渲染页面
       │
       ▼
┌──────────────┐
│   Browser    │
│   展示数据   │
└──────────────┘
```

### 3.3 定时任务流程

```
┌──────────────────────────────────────┐
│         Background Goroutines        │
│                                      │
│  1. Nonce 清理 (每 5 分钟)              │
│     middleware.StartNonceCleaner()   │
│     - 遍历 nonceStore                │
│     - 删除超过 TTL 的 Nonce             │
│                                      │
│  2. 活跃日志清理 (每天凌晨 3 点)        │
│     model.StartActiveLogCleanup()    │
│     - 计算下一个凌晨 3 点              │
│     - 等待到时间                     │
│     - DELETE FROM active_logs        │
│       WHERE created_at < NOW - N 天   │
│     - 循环执行                       │
└──────────────────────────────────────┘
```

---

## 四、安全设计

### 4.1 认证机制

**双层认证体系：**

| 接口类型 | 认证方式 | 使用场景 |
|---------|---------|---------|
| 上报接口 (`/report/*`) | HMAC 签名 | SDK 调用，服务端到服务端 |
| 查询接口 (`/api/*`) | JWT Token | Dashboard 调用，用户浏览器 |
| 登录接口 (`/auth/login`) | 无 | 用户首次登录 |

**HMAC 签名安全特性：**
- 时间戳验证：防止请求重放（±300 秒）
- Nonce 验证：防止同一请求重复提交（内存存储，5 分钟清理）
- 签名验证：确保请求未被篡改

**JWT Token 安全特性：**
- 签名验证：HS256 算法，密钥保存在服务端
- 有效期控制：默认 24 小时，过期需重新登录
- 不存储状态：无状态认证，支持水平扩展

### 4.2 限速保护

**IP 维度限速：**
- 使用 `sync.Map` 存储每个 IP 的请求时间戳列表
- 每次请求时过滤掉 60 秒前的记录
- 剩余请求数超过限制返回 429

**扩展点：**
- 修改限速策略：修改 `middleware/ratelimit.go`
- 新增限速维度：如按 AppID 限速

### 4.3 数据隔离

**多应用支持：**
- 每个应用独立的 AppID 和 AppSecret
- 数据表包含 `app_id` 字段，查询时支持筛选
- Dashboard 支持切换应用查看数据

**多用户支持：**
- 配置文件支持多个用户
- 密码使用 bcrypt 哈希存储
- JWT Token 包含用户名信息

---

## 五、性能优化

### 5.1 数据库优化

**SQLite 配置：**
```go
// WAL 模式：提升并发写入性能
PRAGMA journal_mode=WAL;

// 同步模式：NORMAL 在 WAL 模式下足够安全
PRAGMA synchronous=NORMAL;

// 缓存大小：64MB
PRAGMA cache_size=-65536;

// 临时表存内存
PRAGMA temp_store=MEMORY;

// 连接池：SQLite 只支持单写
SetMaxOpenConns(1)
SetMaxIdleConns(1)
```

**索引设计：**
- `error_logs.fingerprint` - 唯一索引，去重查询
- `error_logs.type` - 按类型筛选
- `error_logs.app_id` - 按应用筛选
- `error_logs.last_seen` - 按最近出现排序
- `active_logs.app_id` - 按应用筛选
- `active_logs.user_id + page` - 复合索引，UV 统计
- `active_logs.created_at` - 按日期分组

### 5.2 内存管理

**Nonce 存储：**
- 使用 `sync.Map` 存储已使用的 Nonce
- 每 5 分钟清理过期记录
- 服务重启后清空（可接受的安全风险）

**SDK 队列：**
- 缓冲容量 100，队列满时丢弃
- 避免内存无限增长
- 上报失败不影响主业务

### 5.3 前端优化

**构建优化：**
```typescript
// vite.config.ts
build: {
  rollupOptions: {
    output: {
      manualChunks: {
        'vue-vendor': ['vue', 'vue-router', 'pinia'],
      },
    },
  },
}
```

**按需加载：**
- ECharts 组件按需引入
- 路由懒加载
- 组件按需导入

---

## 六、扩展指南

### 6.1 新增 API 接口

**步骤：**

1. **创建 Handler** (`internal/handler/xxx.go`)
```go
func HandleXXX(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 处理逻辑
        c.JSON(http.StatusOK, gin.H{"message": "success"})
    }
}
```

2. **注册路由** (`cmd/server/main.go`)
```go
api := r.Group("/api")
api.Use(middleware.JWTAuth(cfg.JWT.Secret))
{
    api.GET("/xxx", handler.HandleXXX(db))
}
```

3. **Dashboard 调用** (`dashboard/src/api/xxx.ts`)
```typescript
export async function getXxx(): Promise<XxxResponse> {
  const { data } = await axios.get('/api/xxx')
  return data
}
```

### 6.2 新增数据表

**步骤：**

1. **定义模型** (`internal/model/xxx.go`)
```go
type XxxLog struct {
    ID        uint   `gorm:"primaryKey"`
    AppID     string `gorm:"index"`
    Data      string
    CreatedAt time.Time `gorm:"index"`
}
```

2. **数据库迁移** (`internal/model/db.go` 或 `main.go`)
```go
db.AutoMigrate(&XxxLog{})
```

3. **创建 Handler** - 参考 6.1

### 6.3 新增 Dashboard 页面

**步骤：**

1. **创建页面组件** (`dashboard/src/pages/xxx.vue`)
```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getXxx } from '@/api/xxx'

const data = ref()
onMounted(async () => {
  data.value = await getXxx()
})
</script>

<template>
  <div>页面内容</div>
</template>
```

2. **注册路由** (`dashboard/src/main.ts` 或 `router/index.ts`)
```typescript
{
  path: '/xxx',
  name: 'Xxx',
  component: () => import('@/pages/xxx.vue'),
  meta: { requiresAuth: true }
}
```

3. **添加导航菜单** (`dashboard/src/layouts/default.vue`)

### 6.4 新增 SDK 语言支持

**参考 Go SDK 结构：**

```
sdk/python/
├── tracely/
│   ├── __init__.py
│   ├── client.py        # 客户端核心
│   ├── payload.py       # 请求体结构
│   ├── sign.py          # 签名工具
│   └── queue.py         # 异步队列
├── setup.py
└── README.md
```

**核心实现：**
- 实现 HMAC 签名（使用 `hmac` 和 `hashlib` 标准库）
- 实现异步队列（使用 `queue.Queue` + 后台线程）
- 实现 `report_error()` 和 `report_active()` 方法

---

## 七、常见问题排查

### 7.1 签名验证失败

**可能原因：**
1. 时间戳过期（与服务器时间差 > 300 秒）
2. Nonce 已使用（重复请求）
3. 签名算法错误（确保使用 HMAC-SHA256）
4. AppSecret 配置错误

**排查步骤：**
1. 检查客户端时间是否同步
2. 检查是否重复使用 Nonce
3. 使用 Python 脚本验证签名算法
4. 检查配置文件中的 AppSecret

### 7.2 数据库锁等待

**现象：** `database is locked` 错误

**原因：** SQLite 并发写入锁竞争

**解决方案：**
1. 确保 `SetMaxOpenConns(1)`
2. 开启 WAL 模式
3. 减少批量写入，使用异步队列
4. 考虑升级到 PostgreSQL

### 7.3 Dashboard 401 错误

**可能原因：**
1. JWT Token 过期
2. Token 未正确注入请求头
3. JWT Secret 配置变更

**排查步骤：**
1. 检查 localStorage 中 Token 是否存在
2. 检查 Axios 拦截器是否正确注入
3. 重新登录获取新 Token

---

## 八、开发环境搭建

### 8.1 后端开发

```bash
# 1. 克隆项目
git clone <repo>
cd tracely

# 2. 安装依赖
go mod download

# 3. 配置
cp config.example.yaml config.yaml
# 编辑 config.yaml

# 4. 生成密码哈希
go run ./cmd/hashpwd/main.go yourpassword

# 5. 运行
go run ./cmd/server/main.go
```

### 8.2 Dashboard 开发

```bash
# 1. 进入目录
cd dashboard

# 2. 安装依赖
bun install  # 或 npm install

# 3. 开发模式运行
bun run dev

# 4. 构建
bun run build
```

**代理配置：**
```typescript
// vite.config.ts
server: {
  proxy: {
    '/api': { target: 'http://localhost:3001' },
    '/auth': { target: 'http://localhost:3001' },
  }
}
```

### 8.3 Go SDK 开发

```bash
# 1. 进入目录
cd sdk/go

# 2. 运行测试
go test ./...

# 3. 本地引用测试
# 在测试项目的 go.mod 中添加
replace github.com/hanxi/tracely/sdk/go => ../sdk/go
```

---

## 九、部署架构

### 9.1 单机部署

```
┌─────────────────────┐
│   Nginx (反向代理)   │
│   - HTTPS 终止        │
│   - 静态资源缓存     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   Tracely Server    │
│   - 单二进制文件     │
│   - 内嵌 Dashboard   │
│   - SQLite 数据库    │
└─────────────────────┘
```

**Docker 部署：**
```bash
docker run -d \
  -p 3001:3001 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config.yaml:/app/config.yaml \
  hanxi/tracely:latest
```

### 9.2 高可用部署（未来扩展）

```
┌─────────────────────┐
│      Nginx LB       │
└──────────┬──────────┘
           │
     ┌─────┴─────┐
     │           │
     ▼           ▼
┌─────────┐ ┌─────────┐
│ Server  │ │ Server  │
│ (Node1) │ │ (Node2) │
└────┬────┘ └────┬────┘
     │           │
     └─────┬─────┘
           │
           ▼
┌─────────────────────┐
│   PostgreSQL        │
│   (主从复制)         │
└─────────────────────┘
```

**改造点：**
1. 数据库替换为 PostgreSQL
2. Nonce 存储改为 Redis（多节点共享）
3. 配置文件改为从配置中心加载

---

## 十、技术决策记录

### 10.1 为什么选择 SQLite？

**优点：**
- 零配置，单文件
- 嵌入二进制，无外部依赖
- 适合中小流量（日上报 < 10 万）
- GORM 支持良好

**缺点：**
- 并发写入有锁
- 不适合大规模部署

**未来扩展：**
- 支持 PostgreSQL（GORM 屏蔽差异）
- 通过配置选择数据库类型

### 10.2 为什么使用 HMAC 签名而非 API Key？

**原因：**
- 防止请求篡改
- 支持时间戳验证（防重放）
- 支持 Nonce 验证（防重复提交）
- 标准协议，多语言支持良好

### 10.3 为什么 Dashboard 嵌入后端？

**优点：**
- 单二进制文件部署
- 避免跨域问题
- 简化部署流程
- 减少运维成本

**缺点：**
- 前端构建后才能编译后端
- 前后端耦合

### 10.4 为什么使用异步队列上报？

**原因：**
- 不阻塞主业务
- 失败重试不影响服务
- 削峰填谷，保护数据库
- 队列满丢弃，避免内存爆炸

---

## 十一、待实现功能

### 11.1 短期计划

- [ ] **告警通知**：错误频率超过阈值时发送 Webhook（钉钉/企业微信/Slack）
- [ ] **Source Map 支持**：上传 Source Map，服务端还原堆栈为源码位置
- [ ] **多应用动态配置**：通过 API 管理应用配置，而非配置文件
- [ ] **数据导出**：导出错误列表为 CSV/JSON

### 11.2 长期计划

- [ ] **多语言 SDK**：Python、Java、Node.js
- [ ] **更多框架中间件**：Echo、Fiber、Gin 以外的框架
- [ ] **分布式部署**：支持 PostgreSQL + Redis
- [ ] **用户权限管理**：RBAC 权限控制
- [ ] **错误分组**：智能聚合相似错误
- [ ] **趋势分析**：错误趋势预测、异常检测

---

## 十二、关键设计模式

### 12.1 中间件链

```
Request → CORS → RateLimit → SignAuth/JWTAuth → Handler → Response
```

**特点：**
- 职责分离，每个中间件只负责单一功能
- 可组合，可复用
- 易于测试和替换

### 12.2 异步队列

```
Producer (Handler) → Channel (Buffer=100) → Consumer (Worker Goroutine) → HTTP Request
```

**特点：**
- 解耦生产和消费
- 削峰填谷
- 失败重试
- 队列满丢弃（快速失败）

### 12.3 依赖注入

```go
// main.go
db, _ := model.InitDB(cfg.DBPath)
r.GET("/api/errors", handler.ErrorList(db))
```

**特点：**
- Handler 无状态，易于测试
- 依赖显式传递，清晰明了
- 避免全局变量

---

## 十三、测试策略

### 13.1 后端测试

**单元测试：**
```go
// sdk/go/sign_test.go
func TestGenerateSignature(t *testing.T) {
    sig := generateSignature("app1", "secret1", "1234567890", "nonce1")
    expected := "..." // 预计算的正确值
    if sig != expected {
        t.Errorf("签名错误：got %s, want %s", sig, expected)
    }
}
```

**集成测试：**
```go
// internal/handler/error_test.go
func TestReportError(t *testing.T) {
    // 创建测试数据库
    db := setupTestDB()
    
    // 创建测试请求
    req := httptest.NewRequest("POST", "/report/error", body)
    w := httptest.NewRecorder()
    
    // 调用 handler
    handler := ReportError(db)
    handler.ServeHTTP(w, req)
    
    // 验证响应
    assert.Equal(t, 200, w.Code)
}
```

### 13.2 Dashboard 测试

**组件测试：**
```typescript
// 使用 Vitest + @vue/test-utils
import { mount } from '@vue/test-utils'
import Overview from '@/pages/index.vue'

test('显示今日 PV', async () => {
  const wrapper = mount(Overview, {
    global: {
      mocks: { $api: { getOverview: () => ({ todayPV: 100 }) } }
    }
  })
  expect(wrapper.text()).toContain('100')
})
```

### 13.3 E2E 测试（未来）

使用 Playwright 或 Cypress 进行端到端测试：
- 登录流程
- 错误上报和展示
- 统计数据查询

---

## 十四、性能基准

### 14.1 预期性能指标

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 上报接口延迟 | < 10ms | P99 |
| 查询接口延迟 | < 100ms | P99 |
| 并发上报 | 1000 QPS | 单实例 |
| 数据库大小 | < 10GB | 一年数据 |
| 内存占用 | < 200MB | 空闲状态 |

### 14.2 性能测试方法

**上报压力测试：**
```bash
# 使用 wrk 进行压测
wrk -t12 -c400 -d30s http://localhost:3001/report/error
```

**数据库查询测试：**
```sql
-- 查询错误列表（带分页）
EXPLAIN QUERY PLAN
SELECT * FROM error_logs
ORDER BY last_seen DESC
LIMIT 20 OFFSET 0;
```

---

## 十五、监控与日志

### 15.1 日志规范

```go
// 启动日志
log.Println("[Tracely] Server started on port 3001")

// 认证失败
log.Printf("[Tracely] Auth failed: invalid signature, appId=%s, ip=%s", appID, ip)

// 限速触发
log.Printf("[Tracely] Rate limit exceeded, ip=%s", ip)
```

**日志级别：**
- INFO：启动、关闭、配置加载
- WARN：配置缺失、非关键错误
- ERROR：数据库错误、关键功能失败

### 15.2 自我监控（未来）

Tracely 可以使用自身 SDK 监控自己：
```go
// cmd/server/main.go
client := tracely.New(tracely.Config{...})
r.Use(tracely_gin.Recovery(client))
r.Use(tracely_gin.Tracker(client))
```

---

## 十六、总结

Tracely 是一个设计简洁、易于扩展的监控平台，核心特点：

1. **轻量级**：Go + SQLite，单二进制部署
2. **安全性**：双层认证（HMAC + JWT），限速保护
3. **高性能**：异步队列，数据库优化
4. **易扩展**：模块化设计，清晰的职责划分
5. **友好体验**：现代化 Dashboard，明暗色模式

本文档详细描述了架构设计的各个方面，希望帮助 AI 开发者快速理解和扩展 Tracely。

---

**文档维护：**
- 新增功能时同步更新本文档
- 架构变更时更新相关章节
- 保持代码与文档一致
