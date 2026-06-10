# 系统架构

## 架构总览

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
│         │ 上报错误/事件   │ 上报错误/事件        │ 查询数据    │
└─────────┼────────────────┼─────────────────────┼────────────┘
          │                │                     │
          ▼                ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│                      Tracely Server                         │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────┐  │
│  │                   Gin Router                         │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │  /report/* (HMAC + 限速)  │  /api/* (JWT)            │  │
│  │  /auth/login (登录)       │  /* (SPA 静态资源)        │  │
│  └──────────────────────────────────────────────────────┘  │
│         │                    │                              │
│         ▼                    ▼                              │
│  ┌─────────────┐      ┌─────────────┐                      │
│  │  Handlers   │      │ Middleware  │                      │
│  │  - error    │      │ - SignAuth  │                      │
│  │  - event    │      │ - JWTAuth   │                      │
│  │  - overview │      │ - RateLimit │                      │
│  │  - auth     │      │ - CORS      │                      │
│  └──────┬──────┘      └─────────────┘                      │
│         │                                                   │
│         ▼                                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                  GORM + SQLite                      │   │
│  │  - error_logs (错误表，永久保留)                     │   │
│  │  - events (事件表，按配置清理)                       │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Background Tasks                       │   │
│  │  - Nonce 清理 (每 5 分钟)                             │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 技术栈

| 模块 | 技术 |
|------|------|
| 后端 | Go + Gin + GORM + SQLite（纯 Go 实现，无需 CGO） |
| 前端 Dashboard | Vue 3 + Nuxt UI + Vite + Pinia + TypeScript |
| SDK | Go SDK (`sdk/go/tracely/`) + TypeScript SDK (`sdk/ts/`) |
| 包管理 | 前端使用 bun |

## 核心模块

### 后端 (`internal/`)

```
internal/
├── config/      # 配置加载（Viper，支持环境变量覆盖）
├── middleware/   # Gin 中间件（SignAuth, JWTAuth, RateLimit）
├── handler/     # HTTP 请求处理（error, event, overview, auth）
└── model/       # GORM 模型 + 数据库操作
```

**Handler 设计原则：**
- 所有 handler 接收 `*gorm.DB` 参数，直接操作数据库
- 返回统一 JSON 格式：成功 `{"message": "..."}` / 失败 `{"error": "..."}`

### Dashboard (`dashboard/`)

```
dashboard/src/
├── main.ts          # 入口（路由定义 + 守卫）
├── App.vue          # 根组件
├── pages/           # 页面：index, errors, events, app-stats, login
├── components/      # 组件：AppSwitcher, ColorModeToggle, UserMenu
├── api/             # API 封装（Axios + JWT 拦截器）
├── stores/          # Pinia 状态（auth, app，持久化到 localStorage）
└── layouts/         # 布局（侧边栏 + 主内容区）
```

前端构建产物通过 `//go:embed dist` 嵌入 Go 二进制，构建顺序：先前端再后端。

## 数据流

### 错误上报流程

1. SDK 生成签名：HMAC-SHA256(appId + timestamp + nonce, secret)
2. POST `/report/error`（携带签名头）
3. SignAuth 中间件验证签名 → RateLimit 限速
4. Handler 生成指纹 MD5(appId + type + message)
5. 已存在：count+1，更新 last_seen；不存在：插入新记录
6. 返回上报成功

### Dashboard 查询流程

1. 用户 POST `/auth/login` 获取 JWT Token
2. Token 存入 localStorage
3. 查询请求携带 `Authorization: Bearer <token>`
4. JWTAuth 中间件验证 → Handler 读取数据库 → 返回 JSON

## 设计模式

### 中间件链

```
Request → CORS → RateLimit → SignAuth/JWTAuth → Handler → Response
```

职责分离，每个中间件只负责单一功能。

### 异步队列（SDK）

```
Producer (业务代码) → Channel (Buffer=100) → Consumer (Worker Goroutine) → HTTP Request
```

解耦生产和消费，队列满时快速失败（丢弃），失败自动重试 3 次。

### 依赖注入

```go
db, _ := model.InitDB(cfg.DBPath)
r.GET("/api/errors", handler.ErrorList(db))
```

Handler 无状态，依赖显式传递，避免全局变量。
