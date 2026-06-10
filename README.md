[![GitHub License](https://img.shields.io/github/license/hanxi/tracely)](https://github.com/hanxi/tracely)
[![Docker Image Version](https://img.shields.io/docker/v/hanxi/tracely?sort=semver&label=docker%20image)](https://hub.docker.com/r/hanxi/tracely)
[![Docker Pulls](https://img.shields.io/docker/pulls/hanxi/tracely)](https://hub.docker.com/r/hanxi/tracely)
[![GitHub Release](https://img.shields.io/github/v/release/hanxi/tracely)](https://github.com/hanxi/tracely/releases)
[![Visitors](https://api.visitorbadge.io/api/daily?path=hanxi%2Ftracely&label=daily%20visitor&countColor=%232ccce4&style=flat)](https://visitorbadge.io/status?path=hanxi%2Ftracely)
[![Visitors](https://api.visitorbadge.io/api/visitors?path=hanxi%2Ftracely&label=total%20visitor&countColor=%232ccce4&style=flat)](https://visitorbadge.io/status?path=hanxi%2Ftracely)

# Tracely

一个轻量级的前端监控平台，支持 **错误收集**、**用户活跃统计** 和 **应用安装/升级追踪**，可自托管部署。

## 功能特性

- **错误收集**：自动捕获 JS 运行时错误、Promise 异常、Vue 组件错误
- **自定义事件**：支持自定义事件上报和统计，灵活的元数据支持
- **应用追踪**：安装/升级次数统计，版本和平台分布分析
- **数据概览**：实时展示 PV/UV、错误总数、安装/升级、Top 事件排行
- **安全认证**：HMAC 签名验证 + JWT 登录，限速保护
- **错误去重**：相同错误合并记录，统计出现次数
- **事件白名单**：配置文件控制允许上报的事件类型
- **内嵌 Dashboard**：前端资源打包到后端，单个二进制文件即可运行
- **现代化 UI**：基于 Nuxt UI，支持明暗色模式、响应式布局
- **多应用支持**：支持多应用配置，Dashboard 中切换查看

**在线体验：** https://tracely.hanxi.cc/ （用户名：`admin` / 密码：`admin123`）

---

## 快速开始

```bash
# Docker Compose 一键部署
mkdir tracely && cd tracely
curl -o docker-compose.yaml https://raw.githubusercontent.com/hanxi/tracely/main/docker-compose.yaml
docker compose run --rm tracely ./scripts/gen-config.sh init
docker compose up -d
# 访问 http://localhost:3001
```

```bash
# 本地构建
make build    # 一键构建前端 + 后端
./tracely     # 运行
```

详细部署说明见 [快速开始文档](./docs/getting-started.md)。

---

## 项目结构

```
tracely/
├── internal/            # 后端核心（config, middleware, handler, model）
├── sdk/
│   ├── go/tracely/      # Go SDK
│   └── ts/              # TypeScript SDK (@imhanxi/tracely-sdk)
├── dashboard/           # Vue 3 + Nuxt UI Dashboard
├── docs/                # 完整文档
├── config.example.yaml  # 配置模板
├── Makefile             # 构建脚本
└── Dockerfile           # Docker 镜像
```

---

## 文档

| 文档 | 说明 |
|------|------|
| [快速开始](./docs/getting-started.md) | 安装、部署和首次使用 |
| [配置详解](./docs/configuration.md) | 配置文件说明和数据库设计 |
| [Dashboard](./docs/dashboard.md) | 面板功能介绍 |
| [TypeScript SDK](./docs/sdk-typescript.md) | 前端错误捕获和事件上报 |
| [Go SDK](./docs/sdk-go.md) | 后端错误上报和事件追踪 |
| [API 参考](./docs/api-reference.md) | HTTP API 接口文档 |
| [系统架构](./docs/architecture.md) | 架构总览和核心模块 |
| [安全设计](./docs/security.md) | 认证、限速和数据隔离 |
| [开发指南](./docs/development.md) | 环境搭建和扩展指南 |
| [部署指南](./docs/deployment.md) | Docker 和生产环境部署 |
| [故障排查](./docs/troubleshooting.md) | 常见问题和排查方法 |
| [技术决策](./docs/design-decisions.md) | 技术选型和路线图 |

---

## 致谢

感谢使用 Tracely！如有问题或建议，欢迎提交 Issue 或 PR。
