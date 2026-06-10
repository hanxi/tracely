# 安全设计

## 认证机制

Tracely 采用双层认证体系：

| 接口类型 | 认证方式 | 使用场景 |
|---------|---------|---------|
| 上报接口 (`/report/*`) | HMAC-SHA256 签名 | SDK 调用，客户端到服务端 |
| 查询接口 (`/api/*`) | JWT Token | Dashboard 调用，用户浏览器 |
| 登录接口 (`/auth/login`) | 无 | 用户首次登录 |

### HMAC 签名验证

SDK 上报请求使用 HMAC-SHA256 签名，流程如下：

```
1. 检查请求头 (X-App-Id, X-Timestamp, X-Nonce, X-Signature)
   ↓
2. 根据 AppID 查找 Secret
   ↓
3. 验证时间戳（与服务器时间差 < TimestampTTL，默认 300 秒）
   ↓
4. 验证 Nonce 是否已使用（防重放攻击）
   ↓
5. 计算签名并比对：HMAC-SHA256(appId + timestamp + nonce, secret)
   ↓
6. 验证通过，继续处理请求
```

**安全特性：**
- 时间戳验证：防止请求重放（默认 ±300 秒窗口）
- Nonce 验证：防止同一请求重复提交（内存存储，5 分钟清理）
- 签名验证：确保请求未被篡改

### JWT Token 验证

Dashboard 使用 JWT Token 认证，流程如下：

```
1. 从 Authorization 头提取 Bearer Token
   ↓
2. 解析并验证 JWT Token（HS256 签名、有效期）
   ↓
3. 将用户名写入上下文
   ↓
4. 验证通过，继续处理请求
```

**安全特性：**
- HS256 签名算法，密钥保存在服务端
- 有效期控制：默认 24 小时，过期需重新登录
- 无状态认证，支持水平扩展

## 限速保护

IP 维度限速（`middleware/ratelimit.go`）：

- 使用 `sync.Map` 存储每个 IP 的请求时间戳列表
- 滑动窗口算法：过滤掉 60 秒前的记录
- 超过限制返回 HTTP 429
- 默认限制：每分钟 60 次请求

## 数据隔离

### 多应用支持

- 每个应用独立的 AppID 和 AppSecret
- 数据表包含 `app_id` 字段，查询时按应用筛选
- Dashboard 支持切换应用查看数据

### 多用户支持

- 配置文件支持多个用户
- 密码使用 bcrypt 哈希存储
- JWT Token 包含用户名信息

## AppSecret 安全建议

**前端 SDK（TS SDK）**：AppSecret 在前端代码中可见，建议：
1. 对打包产物进行代码混淆
2. 定期更换 AppSecret
3. 配合服务端限速策略使用

**后端 SDK（Go SDK）**：建议通过环境变量管理密钥：
```go
client := tracely.New(tracely.Config{
    AppID:     os.Getenv("TRACELY_APP_ID"),
    AppSecret: os.Getenv("TRACELY_APP_SECRET"),
    Host:      os.Getenv("TRACELY_HOST"),
})
```
