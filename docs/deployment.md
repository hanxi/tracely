# 部署指南

## Docker 部署

```bash
docker run -d \
  -p 3001:3001 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config.yaml:/app/config.yaml \
  hanxi/tracely:latest
```

## Docker Compose 部署

参考根目录的 `docker-compose.yml`，数据持久化到 `./data` 目录，配置文件挂载到容器内。

详细步骤见 [快速开始](./getting-started.md)。

## 生产环境建议

- 在前面挂 Nginx 做反向代理并配置 HTTPS
- 定期备份 `data/tracely.db` 数据库文件
- AppSecret 在前端可见，建议对打包产物进行代码混淆
- SQLite 适合中小流量，日上报量建议不超过 10 万条
- Dashboard 构建产物已嵌入后端二进制文件，无需单独部署前端

## 单机部署架构

```
┌─────────────────────┐
│   Nginx (反向代理)   │
│   - HTTPS 终止       │
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

## 高可用部署（未来扩展）

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
1. 数据库替换为 PostgreSQL（GORM 抽象层支持）
2. Nonce 存储改为 Redis（多节点共享）
3. 配置文件改为从配置中心加载

## 性能优化

### 数据库配置

```sql
PRAGMA journal_mode=WAL;       -- WAL 模式提升并发写入
PRAGMA synchronous=NORMAL;     -- NORMAL 在 WAL 模式下足够安全
PRAGMA cache_size=-65536;      -- 缓存 64MB
PRAGMA temp_store=MEMORY;      -- 临时表存内存
```

连接池：`MaxOpenConns=1`，`MaxIdleConns=1`（SQLite 只支持单写）

### 性能基准

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 上报接口延迟 | < 10ms | P99 |
| 查询接口延迟 | < 100ms | P99 |
| 并发上报 | 1000 QPS | 单实例 |
| 数据库大小 | < 10GB | 一年数据 |
| 内存占用 | < 200MB | 空闲状态 |
