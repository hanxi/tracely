# AGENTS.md

本文件为 AI 编程助手提供 tracely 项目的**入口信息**：项目结构、常用命令、铁律与踩坑总结。代码本身就是真实来源的内容（目录树、依赖、API 表、表结构）请直接看代码或下方链接的详细文档。

## 项目概述

Tracely 是一个轻量级前端监控平台，支持错误收集和用户活跃统计。单个二进制文件即可运行（前端资源内嵌到后端）。

## 常用命令

### 构建

```bash
make build              # 一键构建前端 + 后端
make build-frontend     # 仅构建 Dashboard (cd dashboard && bun install && bun run build)
make build-backend      # 编译后端 Go 二进制（依赖 build-frontend）
make docker             # 构建 Docker 镜像
```

### 开发

```bash
make dev                # 后端开发模式 (go run .)
cd dashboard && bun dev # 前端开发模式 (Vite dev server)
```

### TS SDK

```bash
cd sdk/ts && npm run build    # 构建 TS SDK
cd sdk/ts && npm run dev      # 监听模式构建
```

### 工具命令

```bash
./tracely -hashpwd -password <password>   # 生成密码哈希
./tracely -generate-secret                # 生成随机 Secret
./tracely -version                        # 显示版本信息
```

### Lint

```bash
cd dashboard && bun run lint   # ESLint 检查前端代码
```

## 架构

### 技术栈

- **后端**: Go + Gin + GORM + SQLite（纯 Go SQLite 实现，无需 CGO）
- **前端 Dashboard**: Vue 3 + Nuxt UI + Vite + Pinia + TypeScript
- **SDK**: Go SDK (`sdk/go/tracely/`) + TypeScript SDK (`sdk/ts/`, npm 包 `@imhanxi/tracely-sdk`)
- **包管理**: 前端使用 bun

### 核心数据流

1. **SDK 上报** → `POST /report/error` 或 `/report/event` → RateLimit 中间件 → HMAC 签名验证 (`SignAuth`) → Handler 写入 SQLite
2. **Dashboard 查询** → `GET /api/*` → JWT 验证 → Handler 读取 SQLite → JSON 响应
3. **前端路由**: Hash 模式 (`createWebHashHistory`)，后端只处理 `/` 返回 index.html

### 认证机制

- **SDK 上报**: AppID + HMAC-SHA256 签名（`appId + timestamp + nonce` 用 `appSecret` 签名），带时间戳防过期和 Nonce 防重放
- **Dashboard**: JWT Token，登录后存 localStorage，路由守卫校验

### 关键目录

- `internal/handler/` — 业务逻辑（错误上报、事件统计、概览、认证）
- `internal/middleware/` — SignAuth (HMAC)、JWTAuth、RateLimit
- `internal/model/` — GORM 模型 + 数据库初始化 + 定时清理任务
- `internal/config/` — Viper 配置加载
- `dashboard/src/` — Vue SPA（pages/、components/、stores/、api/）
- `sdk/go/tracely/` — Go SDK（异步队列上报，自动重试）
- `sdk/ts/src/` — TypeScript SDK（浏览器端错误捕获 + 事件上报）

### 前端嵌入

`dashboard/` 目录下有一个 `embed.go` 文件，使用 `//go:embed dist` 将构建产物嵌入 Go 二进制。构建顺序：先 `build-frontend` 再 `build-backend`。

### 配置

运行时读取 `config.yaml`（参考 `config.example.yaml`）。多应用通过 `apps[]` 配置，事件白名单通过 `events[]` 配置，每个事件可设置独立的 `retentionDays` 清理策略。

## Git 提交约定

- 提交信息**禁止**添加 `Co-Authored-By` 尾部标记
- 遵循 Conventional Commits 格式：`type(scope): description`

## CI/CD

- GitHub Actions 在 push tag `v*` 时自动发布 TS SDK 到 npm
- Docker 镜像通过 `docker-compose.yml` 部署，数据持久化到 `./data` 目录
