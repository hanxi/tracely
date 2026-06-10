# 故障排查

## 签名验证失败

**可能原因：**

1. 时间戳过期（与服务器时间差 > 300 秒）
2. Nonce 已使用（重复请求）
3. 签名算法错误（确保使用 HMAC-SHA256）
4. AppSecret 配置错误

**排查步骤：**

1. 检查客户端时间是否同步
2. 检查是否重复使用 Nonce
3. 验证签名算法：`HMAC-SHA256(appId + timestamp + nonce, appSecret)`
4. 检查 `config.yaml` 中的 AppSecret 与 SDK 配置是否一致

## 数据库锁等待

**现象：** `database is locked` 错误

**原因：** SQLite 并发写入锁竞争

**解决方案：**

1. 确保 `SetMaxOpenConns(1)`（代码中已默认配置）
2. 确认 WAL 模式已开启
3. 减少批量写入，使用异步队列
4. 如流量持续增长，考虑升级到 PostgreSQL

## Dashboard 401 错误

**可能原因：**

1. JWT Token 过期（默认 24 小时）
2. Token 未正确注入请求头
3. JWT Secret 配置变更导致旧 Token 失效

**排查步骤：**

1. 检查浏览器 localStorage 中 Token 是否存在
2. 检查 Network 面板中请求是否携带 `Authorization` 头
3. 重新登录获取新 Token

## 事件上报被 403 拒绝

**原因：** 事件名未在 `config.yaml` 的 `events[]` 白名单中

**解决方案：** 在配置文件的 `events` 数组中添加对应的事件名条目，然后重启服务。

## SDK 上报失败无报错

SDK 采用**静默失败**策略，上报失败不会影响业务逻辑。

**排查方法：**

- **TS SDK**：打开浏览器控制台，查看 `[Tracely]` 开头的日志和 Network 面板的 `/report/*` 请求
- **Go SDK**：检查 Tracely 服务端日志，确认网络连接和认证配置
