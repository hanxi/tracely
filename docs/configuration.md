# 配置详解

运行时读取 `config.yaml`（参考 `config.example.yaml`）。

## 配置文件结构

```yaml
# 服务配置
port: "3001"                    # 监听端口
dbPath: "./data/tracely.db"     # SQLite 数据库路径
rateLimit: 60                   # 每分钟每 IP 最大请求数
nonceTTL: 300                   # Nonce 过期时间（秒）
timestampTTL: 300               # 时间戳过期时间（秒）

# JWT 配置（Dashboard 登录）
jwt:
  secret: "your-jwt-secret"     # JWT 签名密钥（建议 32 字符以上）
  expireHours: 24               # Token 有效期（小时）

# 多应用配置（SDK 上报）
apps:
  - appId: "my-app-id"          # 应用唯一标识
    appName: "我的应用"           # 应用显示名称
    appSecret: "my-app-secret"  # HMAC 签名密钥

# 多用户配置（Dashboard 登录）
users:
  - username: "admin"
    passwordHash: "$2a$10$..."  # bcrypt 哈希，用 ./tracely -hashpwd 生成

# 自定义事件配置（白名单）
events:
  - eventName: "_active"        # 用户活跃事件（内置）
    description: "用户活跃事件"
    retentionDays: 90
  - eventName: "_app_install"   # 应用安装事件（内置）
    description: "应用安装事件"
    retentionDays: 365
  - eventName: "_app_upgrade"   # 应用升级事件（内置）
    description: "应用升级事件"
    retentionDays: 365
  - eventName: "click_button"   # 自定义事件示例
    description: "按钮点击事件"
    retentionDays: 30
```

## 事件白名单

所有通过 SDK 上报的事件必须在 `events[]` 中注册，未注册的事件名将被 HTTP 403 拒绝。

内置事件类型（以下划线开头）：

| 事件名 | 用途 | 保留建议 |
|--------|------|---------|
| `_active` | 用户活跃统计（PV/UV） | 90 天 |
| `_app_install` | 应用安装追踪 | 365 天 |
| `_app_upgrade` | 应用升级追踪 | 365 天 |

`retentionDays` 为 0 表示永久保留。

## 工具命令

```bash
./tracely -hashpwd -password <password>   # 生成密码的 bcrypt 哈希
./tracely -generate-secret                # 生成随机 Secret
./tracely -version                        # 显示版本信息
```

## 数据库设计

### 错误表 `error_logs`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键 |
| fingerprint | TEXT | 错误唯一指纹（唯一索引），MD5(appId + type + message) |
| type | TEXT | 错误类型：jsError / promiseError / vueError |
| message | TEXT | 错误信息 |
| stack | TEXT | 错误堆栈 |
| url | TEXT | 发生错误的页面地址 |
| app_id | TEXT | 应用 ID |
| user_agent | TEXT | 浏览器 UA |
| count | INTEGER | 出现次数，默认 1 |
| first_seen | DATETIME | 首次出现时间 |
| last_seen | DATETIME | 最近出现时间 |

### 事件表 `events`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键 |
| event_name | TEXT | 事件名称 |
| metadata | TEXT | 元数据（JSON 格式） |
| app_id | TEXT | 应用 ID |
| user_id | TEXT | 用户唯一标识 |
| created_at | DATETIME | 创建时间 |

### 数据清理策略

- **事件数据**：根据每个事件的 `retentionDays` 配置自动清理
- **错误日志**：永久保留，方便历史问题排查
