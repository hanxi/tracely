# API 参考

## 上报接口（SDK 调用）

所有上报接口需要 HMAC-SHA256 签名认证：

| 请求头 | 说明 |
|--------|------|
| X-App-Id | 应用 ID |
| X-Timestamp | 当前 Unix 时间戳（秒） |
| X-Nonce | 随机字符串 |
| X-Signature | HMAC-SHA256(appId + timestamp + nonce, appSecret) |

安全规则：时间戳与服务器时间差超过 300 秒则拒绝；同一 Nonce 只能使用一次；同一 IP 每分钟最多 60 次。

### POST `/report/error` 上报错误

```json
// 请求
{
  "type": "jsError",
  "message": "Cannot read properties of undefined",
  "stack": "TypeError: Cannot read...\n    at xxx.js:10:5",
  "url": "https://example.com/home",
  "appId": "my-app-id"
}

// 响应
{ "message": "上报成功" }
```

相同指纹（MD5(appId + type + message)）的错误会合并，更新 count 和 last_seen。

### POST `/report/event` 上报事件

```json
// 请求
{
  "eventName": "_active",
  "metadata": { "page": "/home", "duration": 30 },
  "appId": "my-app-id",
  "userId": "550e8400-e29b-41d4-a716-446655440000"
}

// 响应
{ "message": "上报成功" }
```

`eventName` 必须在 `config.yaml` 事件白名单中，否则返回 403。

---

## Dashboard 接口（JWT 认证）

所有接口需要请求头：`Authorization: Bearer <JWT_TOKEN>`

### POST `/auth/login` 登录

```json
// 请求
{ "username": "admin", "password": "yourpassword" }

// 响应
{ "token": "eyJhbGciOi...", "username": "admin" }
```

### GET `/api/apps` 获取应用列表

```json
// 响应
{
  "apps": [
    { "appId": "my-app-id", "appName": "我的应用" }
  ]
}
```

### GET `/api/overview` 获取概览数据

**参数：** `appID` - 应用 ID 筛选

```json
// 响应
{
  "todayPV": 1500,
  "todayUV": 420,
  "totalErrors": 85,
  "todayErrors": 12,
  "todayInstalls": 30,
  "todayUpgrades": 15,
  "topErrors": [
    { "type": "jsError", "message": "Cannot read...", "count": 25 }
  ],
  "errorTrend": [
    { "date": "01/01", "count": 5 }
  ]
}
```

### GET `/api/errors` 获取错误列表

**参数：** `page`(默认 1), `pageSize`(默认 20), `type`(错误类型筛选), `appID`

```json
// 响应
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

### GET `/api/events/stats` 事件统计

**参数：** `days`(默认 7), `appID`, `eventName`

```json
// 响应
{ "stats": [{ "eventName": "_active", "count": 1500 }] }
```

### GET `/api/events/top` Top 事件排行

**参数：** `days`(默认 7), `appID`, `limit`(默认 10)

```json
// 响应
{ "events": [{ "eventName": "_active", "count": 1500 }] }
```

### GET `/api/events/daily` 每日事件统计

**参数：** `days`(默认 7), `appID`, `eventName`

```json
// 响应
{ "daily": [{ "date": "2024-01-01", "count": 200 }] }
```

### GET `/api/events/overview` 事件概览

**参数：** `appID`

```json
// 响应
{
  "todayEventCount": 500,
  "todayActivePV": 300,
  "todayActiveUV": 120,
  "topEvents": [{ "eventName": "_active", "count": 300 }]
}
```

### GET `/api/events/list` 事件列表

**参数：** `appID`, `eventName`, `page`(默认 1), `pageSize`(默认 20)

```json
// 响应
{
  "total": 500,
  "list": [
    {
      "id": 1,
      "eventName": "_active",
      "metadata": { "page": "/home", "duration": 30 },
      "appId": "my-app-id",
      "userId": "uuid",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### GET `/api/events/stats/summary` 事件统计摘要

**参数：** `appID`, `eventName`, `days`(默认 7)

```json
// 响应
{ "totalCount": 1500, "todayCount": 200, "uv": 120 }
```

### GET `/api/app-stats` 应用安装升级统计

**参数：** `days`(默认 7), `appID`

```json
// 响应
{
  "installTotal": 500,
  "installToday": 30,
  "installUV": 450,
  "upgradeTotal": 200,
  "upgradeToday": 15,
  "upgradeUV": 180,
  "dailyInstalls": [{ "date": "2024-01-01", "count": 50 }],
  "dailyUpgrades": [{ "date": "2024-01-01", "count": 20 }],
  "versionDist": [{ "value": "2.0.0", "count": 300 }],
  "platformDist": [{ "value": "android", "count": 250 }]
}
```
