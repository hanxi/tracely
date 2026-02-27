# Tracely 实现计划文档

## 一、整体开发顺序

```
server（后端服务）
    ↓
sdk/go（Go SDK）
    ↓
sdk/ts（前端 TypeScript SDK）
    ↓
dashboard（可视化面板）
```

---

## 二、server 后端服务

### 2.1 初始化项目

- [ ] `go mod init github.com/yourname/tracely/server`
- [ ] 安装依赖
  - `github.com/gin-gonic/gin` Web 框架
  - `gorm.io/gorm` ORM
  - `gorm.io/driver/sqlite` SQLite 驱动
  - `gopkg.in/yaml.v3` 读取配置文件

### 2.2 config/config.go

- [ ] 定义 `App` 结构体
  ```go
  type App struct {
      AppID     string `yaml:"appId"`
      AppSecret string `yaml:"appSecret"`
  }
  ```
- [ ] 定义 `Config` 结构体
  ```go
  type Config struct {
      Port         string `yaml:"port"`
      DBPath       string `yaml:"dbPath"`
      RateLimit    int    `yaml:"rateLimit"`
      NonceTTL     int    `yaml:"nonceTTL"`
      TimestampTTL int    `yaml:"timestampTTL"`
      Apps         []App  `yaml:"apps"`
  }
  ```
- [ ] 实现 `Load()` 方法，读取 `config.yaml`，不存在则使用环境变量默认值
- [ ] 实现 `GetSecret(appID string) (string, bool)` 方法，根据 AppID 查找对应 Secret

### 2.3 model/error_log.go

- [ ] 定义 `ErrorLog` 结构体，字段参考数据库设计章节
- [ ] 实现 `GenFingerprint(appID, errType, message string) string`
  - 规则：`MD5(appId + type + message)`

### 2.4 model/active_log.go

- [ ] 定义 `ActiveLog` 结构体，字段参考数据库设计章节

### 2.5 middleware/auth.go

- [ ] 实现 `SignAuth(config *config.Config) gin.HandlerFunc`
- [ ] 按以下顺序验证，任意一步失败直接返回 `401`
  - [ ] 检查请求头 `X-App-Id`、`X-Timestamp`、`X-Nonce`、`X-Signature` 是否存在
  - [ ] 根据 `X-App-Id` 查找对应 Secret，不存在则拒绝
  - [ ] 验证 `X-Timestamp` 与服务器当前时间差是否在 `TimestampTTL` 秒内
  - [ ] 验证 `X-Nonce` 是否已使用，使用 `sync.Map` 存储已用 Nonce
  - [ ] 计算签名 `HMAC-SHA256(appId + timestamp + nonce, secret)` 并与 `X-Signature` 对比
- [ ] 启动定时 goroutine，每 5 分钟清理过期 Nonce

### 2.6 middleware/ratelimit.go

- [ ] 实现 `RateLimit(maxPerMin int) gin.HandlerFunc`
- [ ] 使用 `sync.Map` 存储每个 IP 的请求时间戳列表
- [ ] 每次请求时过滤掉 60 秒之前的记录，判断剩余数量是否超过限制
- [ ] 超过限制返回 `429 Too Many Requests`

### 2.7 handler/error.go

- [ ] 实现 `POST /api/error`
  - [ ] 绑定并校验请求体（type、message 为必填）
  - [ ] 生成错误指纹
  - [ ] 查询数据库是否存在相同指纹
    - 存在：更新 `count + 1`、`last_seen`、`stack`、`url`
    - 不存在：插入新记录，设置 `first_seen`、`last_seen`、`count = 1`
  - [ ] 返回 `{ "message": "上报成功" }`

- [ ] 实现 `GET /api/errors`
  - [ ] 接收 Query 参数：`page`（默认1）、`pageSize`（默认20）、`type`（可选）
  - [ ] 按 `count` 降序查询，支持分页
  - [ ] 返回 `{ "total": 100, "list": [...] }`

### 2.8 handler/active.go

- [ ] 实现 `POST /api/active`
  - [ ] 绑定并校验请求体（userId、page 为必填）
  - [ ] 从请求头读取 `User-Agent` 写入记录
  - [ ] 插入数据库
  - [ ] 返回 `{ "message": "ok" }`

- [ ] 实现 `GET /api/stats`
  - [ ] 接收 Query 参数：`days`（默认7）
  - [ ] 查询每日 PV/UV 数据（按日期分组）
  - [ ] 查询热门页面排行（按 PV 降序，取前10）
  - [ ] 返回 `{ "daily": [...], "topPages": [...] }`

### 2.9 main.go

- [ ] 加载配置
- [ ] 初始化 SQLite 数据库，执行 `AutoMigrate`
- [ ] 创建 Gin 实例，注册全局中间件
  - `gin.Recovery()`
  - 跨域中间件（允许所有来源，允许自定义请求头）
- [ ] 注册路由组 `/api`，按顺序挂载中间件
  - `RateLimit`
  - `SignAuth`
- [ ] 注册所有接口路由
- [ ] 启动服务

### 2.10 Dockerfile

- [ ] 使用多阶段构建
  - 第一阶段：`golang:1.22-alpine` 编译二进制
  - 第二阶段：`alpine:latest` 只复制二进制文件
- [ ] 暴露端口 `3001`

---

## 三、sdk/go Go SDK

### 3.1 初始化项目

- [ ] `go mod init github.com/yourname/tracely-go`
- [ ] 无外部依赖，只使用标准库

### 3.2 payload.go

- [ ] 定义 `ErrorPayload` 结构体
  ```go
  type ErrorPayload struct {
      Type    string `json:"type"`
      Message string `json:"message"`
      Stack   string `json:"stack"`
      URL     string `json:"url"`
      AppID   string `json:"appId"`
  }
  ```
- [ ] 定义 `ActivePayload` 结构体
  ```go
  type ActivePayload struct {
      AppID    string `json:"appId"`
      UserID   string `json:"userId"`
      Page     string `json:"page"`
      Duration int    `json:"duration"`
  }
  ```

### 3.3 sign.go

- [ ] 实现 `generateNonce() string`
  - 使用 `crypto/rand` 生成 16 字节随机数，转十六进制字符串
- [ ] 实现 `generateSignature(appID, appSecret, timestamp, nonce string) string`
  - 算法：`HMAC-SHA256(appId + timestamp + nonce, appSecret)`
  - 返回十六进制字符串
- [ ] 实现 `buildHeaders(appID, appSecret string) map[string]string`
  - 生成 timestamp、nonce、signature
  - 返回包含四个认证请求头的 map

### 3.4 queue.go

- [ ] 定义 `reportTask` 结构体，包含 `url`、`body`、`retryCount` 字段
- [ ] 实现异步队列
  - 使用带缓冲的 channel（容量 100）存储上报任务
  - 启动后台 goroutine 消费队列
  - 上报失败自动重试，最多重试 3 次，每次重试间隔 1 秒
  - 队列满时直接丢弃，不阻塞业务

### 3.5 client.go

- [ ] 定义 `Config` 结构体
  ```go
  type Config struct {
      AppID     string
      AppSecret string
      Host      string
      Timeout   time.Duration // 默认 5s
  }
  ```
- [ ] 定义 `Client` 结构体，持有 `Config` 和 `queue`
- [ ] 实现 `New(config Config) *Client`
  - 初始化 `http.Client`（带超时）
  - 启动异步队列 goroutine
- [ ] 实现 `ReportError(payload ErrorPayload)`
  - 自动填充 `payload.AppID`
  - 将任务投入异步队列
- [ ] 实现 `ReportActive(payload ActivePayload)`
  - 自动填充 `payload.AppID`
  - 将任务投入异步队列
- [ ] 实现内部 `send(url string, body interface{}) error`
  - 序列化 body 为 JSON
  - 调用 `buildHeaders` 生成认证头
  - 发送 HTTP POST 请求

### 3.6 middleware/gin/recovery.go

- [ ] 实现 `Recovery(client *tracely.Client) gin.HandlerFunc`
- [ ] 使用 `defer recover()` 捕获 panic
- [ ] 捕获到 panic 后调用 `client.ReportError`，上报字段：
  - `Type`: `"panicError"`
  - `Message`: panic 内容转字符串
  - `Stack`: `string(debug.Stack())`
  - `URL`: `c.FullPath()`
- [ ] 上报完成后返回 `500` 响应，并调用 `c.Abort()`

### 3.7 middleware/gin/tracker.go

- [ ] 实现 `Tracker(client *tracely.Client) gin.HandlerFunc`
- [ ] 请求开始时记录时间
- [ ] 请求结束后（`defer`）调用 `client.ReportActive`，上报字段：
  - `Page`: `c.FullPath()`
  - `Duration`: 请求耗时（毫秒）
  - `UserID`: 从请求头 `X-User-Id` 读取，不存在则为空字符串

---

## 四、sdk/ts 前端 TypeScript SDK

### 4.1 初始化项目

- [ ] `npm init`，配置 `package.json`，包名 `tracely-sdk`
- [ ] 安装依赖
  - `crypto-js` HMAC 签名
  - `@types/crypto-js`
- [ ] 配置 `tsconfig.json`，输出 ESM + CJS 双格式

### 4.2 src/request.ts

- [ ] 实现 `generateNonce(): string`
  - 使用 `crypto.randomUUID()` 生成，去掉横线
- [ ] 实现 `generateSignature(appID, appSecret, timestamp, nonce): string`
  - 使用 `crypto-js` 的 `HmacSHA256`
  - 算法：`HMAC-SHA256(appId + timestamp + nonce, appSecret)`
- [ ] 实现 `signedFetch(host, path, body, config): Promise<void>`
  - 生成 timestamp、nonce、signature
  - 发送带认证头的 POST 请求
  - 使用 `keepalive: true`，确保页面关闭时请求不丢失
  - 捕获所有异常，上报失败不影响业务

### 4.3 src/error.ts

- [ ] 实现 `initErrorCapture(config, reportFn): void`
- [ ] 监听 `window.addEventListener('error')`
  - 过滤资源加载错误（`e.target` 为 `HTMLElement` 时跳过，避免与 JS 错误混淆）
  - 上报 `type: "jsError"`
- [ ] 监听 `window.addEventListener('unhandledrejection')`
  - 上报 `type: "promiseError"`
  - `message` 取 `e.reason?.message || String(e.reason)`
- [ ] 导出 `captureError(err, info)` 方法，用于手动上报

### 4.4 src/tracker.ts

- [ ] 实现 `getUserId(): string`
  - 从 `localStorage` 读取 `_tracely_uid`
  - 不存在则用 `crypto.randomUUID()` 生成并存储
- [ ] 实现 `initTracker(config, reportFn)`
  - 记录页面进入时间和当前路径
  - 监听 `beforeunload` 上报最后一个页面的停留时长
  - 导出 `onRouteChange(newPath)` 方法，供 Vue Router 调用
    - 上报上一个页面的停留时长
    - 更新当前路径和进入时间

### 4.5 src/index.ts

- [ ] 定义 `TracelyConfig` 接口
  ```ts
  interface TracelyConfig {
      appId: string
      appSecret: string
      host: string
  }
  ```
- [ ] 实现 `Tracely` 类
  - 构造函数接收 `TracelyConfig`
  - 实现 `init(app: App, router: Router): void`
    - 调用 `initErrorCapture`
    - 注册 `app.config.errorHandler`，上报 Vue 内部错误
    - 调用 `initTracker`
    - 注册 `router.afterEach`，调用 `onRouteChange`
  - 实现 `captureError(err: Error, info?: string): void`，手动上报错误

---

## 五、dashboard 可视化面板

### 5.1 初始化项目

- [ ] `npm create vue@latest dashboard`
- [ ] 安装依赖
  - `vue-router` 路由
  - `axios` 请求
  - `echarts` 图表
  - `vue-echarts` ECharts 的 Vue 封装

### 5.2 路由配置

- [ ] `/` 重定向到 `/errors`
- [ ] `/errors` 错误列表页
- [ ] `/stats` 活跃统计页

### 5.3 views/ErrorList.vue

- [ ] 页面顶部筛选栏
  - 错误类型下拉选择（全部 / jsError / promiseError / vueError / serverError / panicError）
  - 排序方式选择（出现次数 / 最近出现时间）
- [ ] 错误列表表格，列定义：
  - 错误类型（badge 样式区分颜色）
  - 错误信息（超长截断，最多显示 100 字符）
  - 出现次数
  - 首次出现时间
  - 最近出现时间
- [ ] 点击某行展开详情抽屉/弹窗，显示：
  - 完整错误信息
  - 完整 Stack Trace（等宽字体，保留换行）
  - 发生错误的页面 URL
  - 浏览器 UA
- [ ] 分页组件，支持切换每页条数（20 / 50 / 100）
- [ ] 数据加载中显示骨架屏
- [ ] 请求失败显示错误提示

### 5.4 views/ActiveStats.vue

- [ ] 顶部统计天数切换按钮组（7天 / 14天 / 30天）
- [ ] PV/UV 折线图
  - X 轴：日期
  - Y 轴：数量
  - 双折线：PV（蓝色）、UV（绿色）
  - 鼠标悬停显示当日具体数值
- [ ] 热门页面排行表格，列定义：
  - 排名
  - 页面路径
  - PV（访问次数）
  - 平均停留时长（格式化为 xx秒 / xx分钟）
- [ ] 数据加载中显示骨架屏

### 5.5 Dockerfile

- [ ] 第一阶段：`node:20-alpine` 构建 Vue 产物
- [ ] 第二阶段：`nginx:alpine` 托管静态文件
- [ ] 配置 Nginx，所有路由指向 `index.html`（支持 Vue Router history 模式）

---

## 六、docker-compose.yml

- [ ] 定义 `server` 服务
  - 构建镜像：`./server`
  - 映射端口：`3001:3001`
  - 挂载数据目录：`./data:/app/data`（SQLite 持久化）
  - 挂载配置文件：`./config.yaml:/app/config.yaml`
  - 环境变量：`PORT`、`DB_PATH`
  - 重启策略：`unless-stopped`
- [ ] 定义 `dashboard` 服务
  - 构建镜像：`./dashboard`
  - 映射端口：`8080:80`
  - 依赖 `server` 服务
  - 重启策略：`unless-stopped`

---

## 七、配置文件模板 config.yaml

- [ ] 提供 `config.example.yaml`，内容如下：

```yaml
port: "3001"
dbPath: "./data/tracely.db"
rateLimit: 60
nonceTTL: 300
timestampTTL: 300
apps:
  - appId: "my-app-id"
    appSecret: "my-app-secret-please-change-this"
```

---

## 八、README.md 补充内容

- [ ] 快速开始章节，步骤说明：
  1. 克隆项目
  2. 复制 `config.example.yaml` 为 `config.yaml` 并修改配置
  3. 执行 `docker-compose up -d`
  4. 访问 `http://服务器IP:8080` 查看面板
  5. 在项目中接入 SDK

---

## 九、各模块交付标准

### server

| 接口 | 验收标准 |
|------|---------|
| POST /api/error | 签名错误返回 401；相同错误合并记录；count 正确累加 |
| GET /api/errors | 分页正确；type 筛选生效；按 count 降序 |
| POST /api/active | 正确记录 userId、page、duration |
| GET /api/stats | daily 数据按日期升序；topPages 按 PV 降序取前10 |
| 限速 | 同一 IP 第 61 次请求返回 429 |
| 防重放 | 同一 Nonce 第二次请求返回 401 |
| 时间戳过期 | 超过 300 秒的请求返回 401 |

### sdk/go

| 功能 | 验收标准 |
|------|---------|
| ReportError | 异步上报，不阻塞调用方；失败重试3次 |
| ReportActive | 异步上报，不阻塞调用方；失败重试3次 |
| Gin Recovery | panic 被捕获后上报，并正常返回 500 |
| Gin Tracker | 每次请求后上报路径和耗时 |
| 队列满 | 直接丢弃，不阻塞 |

### sdk/ts

| 功能 | 验收标准 |
|------|---------|
| JS 错误捕获 | window.error 触发后自动上报 |
| Promise 错误捕获 | unhandledrejection 触发后自动上报 |
| Vue 错误捕获 | Vue 组件内部报错自动上报 |
| 活跃上报 | 路由切换时上报上一个页面的停留时长 |
| 页面关闭 | beforeunload 时上报最后一个页面的停留时长 |
| 上报失败 | 不影响业务，静默失败 |

### dashboard

| 页面 | 验收标准 |
|------|---------|
| 错误列表 | 分页、筛选、排序正常；点击展开 Stack Trace 正常显示 |
| 活跃统计 | 折线图正常渲染；切换天数后数据刷新；热门页面排行正确 |

---

## 十、开发注意事项

| 事项 | 说明 |
|------|------|
| 时区 | 服务端统一使用 UTC 时间存储，前端展示时转换为本地时间 |
| 跨域 | server 需允许所有来源，因为前端 SDK 会从不同域上报 |
| SQLite 并发 | 开启 WAL 模式（`PRAGMA journal_mode=WAL`），提升并发写入性能 |
| AppSecret 安全 | 文档中明确提示前端 AppSecret 可见，建议配合代码混淆使用 |
| SDK 体积 | ts SDK 打包后建议控制在 10KB 以内（gzip后），避免影响页面加载 |
| 错误风暴 | 同一错误短时间内大量上报时，前端 SDK 应做本地节流（相同 fingerprint 1分钟内只上报一次） |

---

## 十一、前端 SDK 错误节流设计

> 防止同一个错误短时间内大量上报，浪费服务器资源。

### 节流规则

- 相同 `fingerprint`（`type + message` 的 MD5）在 **1 分钟内只上报一次**
- 使用内存 Map 存储已上报的 fingerprint 和上报时间
- 页面刷新后重置，不做持久化

### 实现位置

在 `sdk/ts/src/error.ts` 中实现：

```ts
// 节流缓存：fingerprint -> 上次上报时间戳
const reportedCache = new Map<string, number>();
const THROTTLE_MS = 60 * 1000; // 1 分钟

function shouldReport(fingerprint: string): boolean {
  const lastTime = reportedCache.get(fingerprint);
  const now = Date.now();
  if (lastTime && now - lastTime < THROTTLE_MS) {
    return false; // 1 分钟内已上报过，跳过
  }
  reportedCache.set(fingerprint, now);
  return true;
}

function genFingerprint(type: string, message: string): string {
  // 简单实现，不引入额外依赖
  return `${type}:${message}`.slice(0, 200);
}
```

---

## 十二、SQLite 性能优化

在 `server` 初始化数据库连接后，执行以下 PRAGMA 配置：

```go
// model/db.go
func InitDB(path string) (*gorm.DB, error) {
    db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // 获取底层 sql.DB
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // WAL 模式：提升并发写入性能
    sqlDB.Exec("PRAGMA journal_mode=WAL;")
    // 同步模式：NORMAL 在 WAL 模式下足够安全，性能更好
    sqlDB.Exec("PRAGMA synchronous=NORMAL;")
    // 缓存大小：64MB
    sqlDB.Exec("PRAGMA cache_size=-65536;")
    // 临时表存内存
    sqlDB.Exec("PRAGMA temp_store=MEMORY;")

    // 连接池配置
    sqlDB.SetMaxOpenConns(1)  // SQLite 只支持单写，避免锁竞争
    sqlDB.SetMaxIdleConns(1)

    // 自动迁移
    db.AutoMigrate(&ErrorLog{}, &ActiveLog{})

    return db, nil
}
```

---

## 十三、数据库索引设计

在 `model` 层的结构体 tag 中声明索引：

### error_log.go

```go
type ErrorLog struct {
    ID          uint      `gorm:"primaryKey"`
    Fingerprint string    `gorm:"uniqueIndex"`           // 去重查询
    Type        string    `gorm:"index"`                 // 按类型筛选
    Message     string
    Stack       string
    URL         string
    AppID       string    `gorm:"index"`                 // 按应用筛选
    UserAgent   string
    Count       int       `gorm:"default:1"`
    FirstSeen   time.Time
    LastSeen    time.Time `gorm:"index"`                 // 按最近出现排序
}
```

### active_log.go

```go
type ActiveLog struct {
    ID        uint      `gorm:"primaryKey"`
    AppID     string    `gorm:"index"`
    UserID    string    `gorm:"index:idx_user_page"`     // UV 统计去重
    Page      string    `gorm:"index:idx_user_page"`     // 热门页面统计
    Duration  int
    UserAgent string
    CreatedAt time.Time `gorm:"index"`                   // 按日期分组
}
```

---

## 十四、接口错误码规范

server 统一返回格式：

```json
// 成功
{ "message": "ok" }

// 失败
{ "error": "错误描述" }
```

| HTTP 状态码 | 含义 | 触发场景 |
|------------|------|---------|
| 200 | 成功 | 正常处理完成 |
| 400 | 请求参数错误 | 请求体格式错误、必填字段缺失 |
| 401 | 认证失败 | 签名错误、时间戳过期、Nonce 重放、AppID 不存在 |
| 429 | 请求过于频繁 | 触发 IP 限速 |
| 500 | 服务器内部错误 | 数据库操作失败等 |

---

## 十五、日志规范

server 统一使用结构化日志输出，方便排查问题：

```go
// 启动日志
[Tracely] Server started on port 3001
[Tracely] Database initialized: ./data/tracely.db
[Tracely] Loaded 2 apps from config

// 请求日志（由 Gin 默认日志中间件输出）
[GIN] 2024/01/01 - 12:00:00 | 200 | 1.2ms | 127.0.0.1 | POST /api/error

// 认证失败日志
[Tracely] Auth failed: invalid signature, appId=my-app-id, ip=127.0.0.1

// 限速日志
[Tracely] Rate limit exceeded, ip=127.0.0.1

// 错误日志
[Tracely] DB error: ...
```

---

## 十六、测试用例

### server 接口测试

使用 `curl` 手动验证（正式开发可用 Go 自带 `testing` 包编写单元测试）：

```bash
# 生成签名（以下为 Python 示例，方便快速验证）
python3 -c "
import hmac, hashlib, time, uuid
app_id = 'my-app-id'
app_secret = 'my-app-secret-please-change-this'
timestamp = str(int(time.time()))
nonce = uuid.uuid4().hex
raw = app_id + timestamp + nonce
sig = hmac.new(app_secret.encode(), raw.encode(), hashlib.sha256).hexdigest()
print(f'timestamp={timestamp}')
print(f'nonce={nonce}')
print(f'signature={sig}')
"

# 上报错误
curl -X POST http://localhost:3001/api/error \
  -H "Content-Type: application/json" \
  -H "X-App-Id: my-app-id" \
  -H "X-Timestamp: ${timestamp}" \
  -H "X-Nonce: ${nonce}" \
  -H "X-Signature: ${signature}" \
  -d '{
    "type": "jsError",
    "message": "Cannot read properties of undefined",
    "stack": "TypeError: Cannot read...\n    at main.js:10:5",
    "url": "https://example.com/home",
    "appId": "my-app-id"
  }'

# 查询错误列表
curl "http://localhost:3001/api/errors?page=1&pageSize=20" \
  -H "X-App-Id: my-app-id" \
  -H "X-Timestamp: ${timestamp}" \
  -H "X-Nonce: ${nonce}" \
  -H "X-Signature: ${signature}"

# 上报活跃
curl -X POST http://localhost:3001/api/active \
  -H "Content-Type: application/json" \
  -H "X-App-Id: my-app-id" \
  -H "X-Timestamp: ${timestamp}" \
  -H "X-Nonce: ${nonce}" \
  -H "X-Signature: ${signature}" \
  -d '{
    "appId": "my-app-id",
    "userId": "user-uuid-123",
    "page": "/home",
    "duration": 30
  }'

# 查询统计数据
curl "http://localhost:3001/api/stats?days=7" \
  -H "X-App-Id: my-app-id" \
  -H "X-Timestamp: ${timestamp}" \
  -H "X-Nonce: ${nonce}" \
  -H "X-Signature: ${signature}"
```

### 认证异常测试

```bash
# 测试签名错误
curl -X POST http://localhost:3001/api/error \
  -H "X-App-Id: my-app-id" \
  -H "X-Timestamp: ${timestamp}" \
  -H "X-Nonce: ${nonce}" \
  -H "X-Signature: invalidsignature"
# 预期返回：401 { "error": "签名错误" }

# 测试时间戳过期（使用 300 秒前的时间戳）
# 预期返回：401 { "error": "请求已过期" }

# 测试 Nonce 重放（使用相同 Nonce 发送两次）
# 预期返回：401 { "error": "重放攻击" }

# 测试非法 AppID
curl -X POST http://localhost:3001/api/error \
  -H "X-App-Id: fake-app-id" \
  ...
# 预期返回：401 { "error": "非法 AppID" }
```

### sdk/go 测试

```go
// sdk/go/client_test.go
func TestReportError(t *testing.T) {
    client := New(Config{
        AppID:     "my-app-id",
        AppSecret: "my-app-secret",
        Host:      "http://localhost:3001",
    })

    // 正常上报
    client.ReportError(ErrorPayload{
        Type:    "serverError",
        Message: "test error",
        Stack:   "goroutine 1 ...",
        URL:     "/api/test",
    })
    // 不应 panic，不阻塞
}

func TestQueueFull(t *testing.T) {
    // 模拟队列满的情况，验证不阻塞
}

func TestSign(t *testing.T) {
    // 验证签名算法与服务端一致
    sig := generateSignature("my-app-id", "my-secret", "1704067200", "abc123")
    // 与 Python/服务端计算结果对比
}
```

---

## 十七、里程碑计划

| 阶段 | 内容 | 产出 |
|------|------|------|
| M1 | server 基础框架搭建、数据库初始化、配置加载 | 服务可启动 |
| M2 | server 认证中间件、限速中间件 | 安全层完成 |
| M3 | server 错误收集接口（含去重）、活跃统计接口 | 核心接口完成 |
| M4 | server 查询接口、Docker 镜像 | server 完成 |
| M5 | sdk/go 签名、队列、Client | Go SDK 核心完成 |
| M6 | sdk/go Gin 中间件 | Go SDK 完成 |
| M7 | sdk/ts 签名请求、错误捕获、活跃统计 | TS SDK 完成 |
| M8 | dashboard 错误列表页 | 面板基础功能完成 |
| M9 | dashboard 活跃统计页 | 面板完成 |
| M10 | docker-compose 联调、README 完善 | 项目完成 ✅ |

---

## 十八、目录文件清单

> 实现完成后，完整的文件列表应如下，可用于验收检查。

```
tracely/
├── config.example.yaml
├── docker-compose.yml
├── README.md
│
├── server/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── config/
│   │   └── config.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── ratelimit.go
│   ├── handler/
│   │   ├── error.go
│   │   └── active.go
│   └── model/
│       ├── db.go
│       ├── error_log.go
│       └── active_log.go
│
├── sdk/
│   ├── ts/
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   └── src/
│   │       ├── index.ts
│   │       ├── error.ts
│   │       ├── tracker.ts
│   │       └── request.ts
│   │
│   └── go/
│       ├── go.mod
│       ├── go.sum
│       ├── README.md
│       ├── client.go
│       ├── sign.go
│       ├── payload.go
│       ├── queue.go
│       ├── client_test.go
│       └── middleware/
│           └── gin/
│               ├── recovery.go
│               └── tracker.go
│
└── dashboard/
    ├── Dockerfile
    ├── nginx.conf
    ├── package.json
    ├── tsconfig.json
    ├── vite.config.ts
    └── src/
        ├── main.ts
        ├── App.vue
        ├── router/
        │   └── index.ts
        ├── api/
        │   ├── error.ts        # 封装错误相关接口请求
        │   └── stats.ts        # 封装统计相关接口请求
        ├── views/
        │   ├── ErrorList.vue
        │   └── ActiveStats.vue
        └── components/
            ├── ErrorDetail.vue  # 错误详情抽屉组件
            └── PageChart.vue    # ECharts 折线图封装组件
```

---

## 十九、依赖清单

### server/go.mod

```
github.com/gin-gonic/gin
gorm.io/gorm
gorm.io/driver/sqlite
gopkg.in/yaml.v3
```

### sdk/go/go.mod

```
# 无外部依赖，仅使用标准库
# crypto/hmac
# crypto/sha256
# crypto/rand
# runtime/debug
# encoding/json
# net/http
# sync
# time
```

### sdk/ts/package.json

```json
{
  "name": "tracely-sdk",
  "dependencies": {
    "crypto-js": "^4.2.0"
  },
  "devDependencies": {
    "@types/crypto-js": "^4.2.0",
    "typescript": "^5.0.0",
    "vite": "^5.0.0"
  }
}
```

### dashboard/package.json

```json
{
  "dependencies": {
    "vue": "^3.4.0",
    "vue-router": "^4.3.0",
    "axios": "^1.6.0",
    "echarts": "^5.4.0",
    "vue-echarts": "^6.6.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "typescript": "^5.0.0",
    "vite": "^5.0.0"
  }
}
```

---

## 二十、Nginx 配置（dashboard）

```nginx
# dashboard/nginx.conf
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    # Vue Router history 模式支持
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 反向代理 API 请求，避免跨域（可选）
    # 如果 dashboard 和 server 部署在同一台机器
    location /api/ {
        proxy_pass http://server:3001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 静态资源缓存
    location /assets/ {
        expires 7d;
        add_header Cache-Control "public, immutable";
    }
}
```

---

## 二十一、环境变量与配置优先级

server 配置读取优先级：`环境变量 > config.yaml > 默认值`

| 配置项 | 环境变量 | config.yaml | 默认值 |
|--------|---------|-------------|--------|
| 端口 | `PORT` | `port` | `3001` |
| 数据库路径 | `DB_PATH` | `dbPath` | `./tracely.db` |
| 限速（次/分钟） | `RATE_LIMIT` | `rateLimit` | `60` |
| Nonce 有效期（秒） | `NONCE_TTL` | `nonceTTL` | `300` |
| 时间戳有效期（秒） | `TIMESTAMP_TTL` | `timestampTTL` | `300` |

> apps 配置（AppID / AppSecret）仅支持 `config.yaml`，不支持环境变量，避免多个 App 配置难以通过环境变量表达。

---

## 二十二、sdk/go README.md 内容要求

> `sdk/go/README.md` 作为独立文档，需包含以下内容：

- [ ] 安装方式
- [ ] 快速开始（最简示例）
- [ ] `Config` 参数说明表格
- [ ] `ReportError` 方法说明及示例
- [ ] `ReportActive` 方法说明及示例
- [ ] Gin 中间件接入示例（Recovery + Tracker）
- [ ] 签名算法说明（方便对接其他语言参考）
- [ ] 注意事项（异步上报、队列满丢弃策略）

---

## 二十三、已知限制与后续可扩展方向

### 已知限制

| 限制 | 说明 |
|------|------|
| SQLite 单写 | 并发写入有锁，适合日上报量 10 万以内的场景 |
| Nonce 内存存储 | 服务重启后 Nonce 记录丢失，极短时间内存在重放风险 |
| AppSecret 前端可见 | 无法从根本上防止伪造上报，只能提高攻击成本 |
| 无告警功能 | 新增高频错误时不会主动通知 |
| 无 Source Map 还原 | Stack Trace 为编译后代码，不易定位源码位置 |

### 后续可扩展方向

| 方向 | 说明 |
|------|------|
| 告警通知 | 新错误出现或错误频率超过阈值时，发送 Webhook 到企业微信 / 钉钉 / Slack |
| Source Map 上传 | 构建时上传 Source Map，服务端还原 Stack Trace 为源码位置 |
| 多应用切换 | Dashboard 支持多 AppID 切换查看 |
| 数据清理策略 | 定时清理 N 天前的 active_logs，避免数据库无限增长 |
| 接入更多框架 | sdk/go 增加 Echo、Fiber 等框架的中间件支持 |
| 接入更多语言 | 增加 sdk/python、sdk/java 等 |
| 替换为 PostgreSQL | 数据量增大后可替换存储层，server 代码无需改动（GORM 屏蔽差异） |


---

## 二十四、数据清理策略

> active_logs 表数据量增长较快，需要定期清理历史数据。

在 `server/main.go` 启动时，开启一个后台 goroutine 定期清理：

```go
// 每天凌晨 3 点清理 90 天前的活跃日志
func startCleanupJob(db *gorm.DB) {
    go func() {
        for {
            now := time.Now()
            // 计算距离下一个凌晨 3 点的时间
            next := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, now.Location())
            time.Sleep(time.Until(next))

            db.Where("created_at < ?", time.Now().AddDate(0, 0, -90)).
                Delete(&ActiveLog{})

            // error_logs 不清理，保留所有历史报错记录
        }
    }()
}
```

> 清理天数建议通过配置文件控制，在 `config.yaml` 中增加 `activeLogRetentionDays: 90`

---

## 二十五、SDK 版本管理

### sdk/ts 版本规范

遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)：

| 版本变更 | 场景 |
|---------|------|
| Patch（1.0.x） | Bug 修复，不影响接口 |
| Minor（1.x.0） | 新增功能，向后兼容 |
| Major（x.0.0） | 接口破坏性变更 |

`package.json` 初始版本设为 `1.0.0`

### sdk/go 版本规范

- 使用 Git Tag 管理版本，格式 `sdk/go/v1.0.0`
- `go.mod` module 路径：`github.com/yourname/tracely/sdk/go`
- 重大变更时升级 major 版本，module 路径改为 `.../sdk/go/v2`

---

## 二十六、Git 仓库规范

### 分支策略

```
main          ← 稳定版本，打 Tag 发布
dev           ← 日常开发合并分支
feature/xxx   ← 新功能开发分支
fix/xxx       ← Bug 修复分支
```

### .gitignore

```gitignore
# server
server/tracely.db
server/data/

# 配置文件（含密钥，不提交）
config.yaml

# 前端构建产物
dashboard/dist/
sdk/ts/dist/

# 依赖
node_modules/

# IDE
.idea/
.vscode/
*.DS_Store
```

### Commit Message 规范

```
feat(server): 新增错误上报去重逻辑
fix(sdk-ts): 修复 beforeunload 时 fetch 请求丢失问题
docs(readme): 补充 Go SDK 使用示例
refactor(middleware): 优化限速中间件内存清理逻辑
chore(docker): 更新基础镜像版本
```

---

## 二十七、安全检查清单

在项目完成后，逐项确认以下安全事项：

### server 端

- [ ] `config.yaml` 已加入 `.gitignore`，确保密钥不提交到仓库
- [ ] AppSecret 长度不低于 32 位随机字符串
- [ ] 所有接口均经过签名验证中间件
- [ ] 限速中间件已启用
- [ ] SQLite 文件路径不在 Web 可访问目录下
- [ ] 生产环境关闭 Gin debug 模式（`gin.SetMode(gin.ReleaseMode)`）
- [ ] 跨域配置已限制允许的请求头，不使用 `*` 通配符允许所有请求头

### sdk/ts 端

- [ ] 文档中已提示 AppSecret 在前端可见的风险
- [ ] 上报失败已做静默处理，不向用户暴露错误
- [ ] `keepalive: true` 已设置，防止页面关闭时请求丢失
- [ ] 错误节流已实现，防止同一错误短时间内大量上报

### sdk/go 端

- [ ] 异步队列满时丢弃处理，不阻塞业务
- [ ] HTTP 请求已设置超时，不因上报阻塞业务
- [ ] Gin Recovery 中间件在上报后正确调用 `c.Abort()`

### 部署端

- [ ] 生产环境已配置 HTTPS
- [ ] server 端口 `3001` 不直接对外暴露，通过 Nginx 代理
- [ ] Dashboard 端口 `8080` 建议添加基础 HTTP 认证（Nginx basic auth），避免数据泄露
- [ ] 数据目录 `./data` 已定期备份

---

## 二十八、FAQ

**Q: 前端 SDK 上报请求会影响页面性能吗？**

> 不会。上报使用 `fetch` 异步发送，不阻塞主线程。页面关闭时使用 `keepalive: true` 保证请求不丢失。

---

**Q: 同一个错误在不同页面发生，会合并吗？**

> 会。指纹只根据 `appId + type + message` 生成，与 URL 无关。但每次更新记录时会同步更新 `url` 字段为最新发生的页面。

---

**Q: 用户没有登录体系，userId 怎么处理？**

> sdk/ts 会自动在 `localStorage` 生成并持久化一个 UUID 作为 userId。清除浏览器存储后会重新生成，视为新用户。

---

**Q: 如果上报量很大，SQLite 会成为瓶颈吗？**

> SQLite 在 WAL 模式下，写入性能约为每秒 1000~5000 次，适合日上报量 10 万以内的场景。超过此规模建议替换为 PostgreSQL，只需修改 GORM 驱动，业务代码无需改动。

---

**Q: Go SDK 的异步队列如果进程崩溃，未上报的数据会丢失吗？**

> 会丢失。队列存储在内存中，进程崩溃时未消费的任务会丢失。对于 Go 服务端的 panic，`Recovery` 中间件会在崩溃前同步上报，因此不受影响。普通的异步上报丢失影响较小，可接受。

---

**Q: dashboard 需要登录认证吗？**

> 当前版本不内置登录，建议通过 Nginx basic auth 保护访问入口。后续版本可考虑内置简单的用户名密码登录。

---

## 二十九、CHANGELOG 模板

```markdown
# Changelog

## [Unreleased]

## [1.0.0] - 2024-xx-xx

### Added
- server：错误收集接口，支持去重和次数统计
- server：活跃统计接口，支持 PV/UV 查询
- server：HMAC 签名认证中间件
- server：IP 限速中间件
- sdk/go：核心客户端，支持异步上报和失败重试
- sdk/go：Gin Recovery 中间件
- sdk/go：Gin Tracker 中间件
- sdk/ts：错误自动捕获（JS错误、Promise错误、Vue错误）
- sdk/ts：用户活跃统计，支持 Vue Router 集成
- dashboard：错误列表页，支持筛选、分页、详情查看
- dashboard：活跃统计页，PV/UV 折线图和热门页面排行
- docker-compose：一键部署配置
```

---

## 三十、最终交付检查清单

在认为项目完成时，逐项确认：

### 功能完整性
- [ ] `POST /api/error` 上报、去重、累计次数正常
- [ ] `GET /api/errors` 分页、筛选、排序正常
- [ ] `POST /api/active` 上报正常
- [ ] `GET /api/stats` PV/UV 和热门页面数据正常
- [ ] sdk/ts 在 Vue 项目中一行代码完成初始化
- [ ] sdk/go 在 Gin 项目中两行中间件完成接入
- [ ] dashboard 错误列表和统计页面正常展示数据

### 工程质量
- [ ] 所有模块可独立构建，无编译错误
- [ ] `docker-compose up -d` 一键启动，服务正常运行
- [ ] `.gitignore` 已正确配置，无敏感文件提交
- [ ] `config.example.yaml` 已提供
- [ ] `README.md` 包含完整的部署和接入说明
- [ ] `sdk/go/README.md` 包含完整的使用说明

### 安全性
- [ ] 二十七章安全检查清单全部通过

### 文档
- [ ] README.md 与实际实现一致
- [ ] CHANGELOG.md 已记录 1.0.0 版本内容

---

## 三十一、Dashboard 登录设计

### 整体方案

采用 **用户名 + 密码 → JWT Token** 的方式，简单轻量，不引入额外依赖。

```
用户输入账号密码
      ↓
POST /auth/login
      ↓
server 验证账号密码（对比配置文件中的 bcrypt 哈希）
      ↓
验证通过，返回 JWT Token
      ↓
dashboard 存储 Token 到 localStorage
      ↓
后续所有 /api 请求携带 Authorization: Bearer <token>
      ↓
server 新增 JWT 验证中间件，验证 Token 合法性
```

---

### 配置文件新增用户配置

`config.yaml` 新增 `users` 字段：

```yaml
# config.yaml
jwt:
  secret: "your-jwt-secret-please-change-this"  # JWT 签名密钥
  expireHours: 24                                # Token 有效期（小时）

users:
  - username: "admin"
    # 使用 bcrypt 哈希存储密码，不明文保存
    # 生成方式见下方说明
    passwordHash: "$2a$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

**生成 bcrypt 密码哈希：**

```bash
# 方式一：使用项目自带工具
go run ./server/cmd/hashpwd/main.go yourpassword

# 方式二：使用 Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'yourpassword', bcrypt.gensalt()).decode())"
```

---

### server 端新增内容

#### config/config.go 新增字段

```go
type JWT struct {
    Secret      string `yaml:"secret"`
    ExpireHours int    `yaml:"expireHours"`
}

type User struct {
    Username     string `yaml:"username"`
    PasswordHash string `yaml:"passwordHash"`
}

type Config struct {
    // ...原有字段...
    JWT   JWT    `yaml:"jwt"`
    Users []User `yaml:"users"`
}

// 根据用户名查找用户
func (c *Config) GetUser(username string) (User, bool) {
    for _, u := range c.Users {
        if u.Username == username {
            return u, true
        }
    }
    return User{}, false
}
```

#### 安装新依赖

```
golang.org/x/crypto          # bcrypt 密码验证
github.com/golang-jwt/jwt/v5 # JWT 生成与验证
```

#### handler/auth.go

```go
package handler

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

func Login(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req LoginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
            return
        }

        // 查找用户
        user, ok := cfg.GetUser(req.Username)
        if !ok {
            // 故意不区分"用户不存在"和"密码错误"，防止用户名枚举
            c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
            return
        }

        // 验证密码
        if err := bcrypt.CompareHashAndPassword(
            []byte(user.PasswordHash),
            []byte(req.Password),
        ); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
            return
        }

        // 生成 JWT
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "username": user.Username,
            "exp":      time.Now().Add(time.Duration(cfg.JWT.ExpireHours) * time.Hour).Unix(),
            "iat":      time.Now().Unix(),
        })

        tokenStr, err := token.SignedString([]byte(cfg.JWT.Secret))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Token 失败"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "token":    tokenStr,
            "username": user.Username,
        })
    }
}
```

#### middleware/jwt.go

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

func JWTAuth(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 Authorization 头获取 Token
        auth := c.GetHeader("Authorization")
        if !strings.HasPrefix(auth, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
            return
        }
        tokenStr := strings.TrimPrefix(auth, "Bearer ")

        // 解析并验证 Token
        token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("非法签名方式")
            }
            return []byte(secret), nil
        })

        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
            return
        }

        // 将用户名写入上下文，后续 handler 可读取
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("username", claims["username"])
        }

        c.Next()
    }
}
```

#### main.go 路由调整

```go
// 登录接口不需要任何验证
r.POST("/auth/login", handler.Login(cfg))

// API 接口：需要 JWT 验证（dashboard 调用）
dashboard := r.Group("/api")
dashboard.Use(middleware.JWTAuth(cfg.JWT.Secret))
{
    dashboard.GET("/errors", handler.ErrorList)
    dashboard.GET("/stats", handler.Stats)
}

// 上报接口：需要签名验证（SDK 调用）
report := r.Group("/report")
report.Use(middleware.RateLimit(cfg.RateLimit))
report.Use(middleware.SignAuth(cfg))
{
    report.POST("/error", handler.ReportError)
    report.POST("/active", handler.ReportActive)
}
```

> 注意：上报接口和查询接口分开路由组，各自使用不同的验证方式

---

### dashboard 端新增内容

#### 新增登录页 views/Login.vue

```
- 用户名输入框
- 密码输入框
- 登录按钮
- 登录失败显示错误提示
- 登录成功后跳转到 /errors
```

#### api/auth.ts

```ts
import axios from "axios";

export function login(username: string, password: string) {
  return axios.post("/auth/login", { username, password });
}
```

#### stores/auth.ts（使用 Pinia）

```ts
import { defineStore } from "pinia";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    token: localStorage.getItem("_tracely_token") || "",
    username: localStorage.getItem("_tracely_user") || "",
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
  },

  actions: {
    setAuth(token: string, username: string) {
      this.token = token;
      this.username = username;
      localStorage.setItem("_tracely_token", token);
      localStorage.setItem("_tracely_user", username);
    },
    logout() {
      this.token = "";
      this.username = "";
      localStorage.removeItem("_tracely_token");
      localStorage.removeItem("_tracely_user");
    },
  },
});
```

#### axios 拦截器配置

```ts
// src/api/index.ts
import axios from "axios";
import { useAuthStore } from "@/stores/auth";
import router from "@/router";

// 请求拦截：自动注入 Token
axios.interceptors.request.use((config) => {
  const auth = useAuthStore();
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`;
  }
  return config;
});

// 响应拦截：Token 过期自动跳转登录页
axios.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      const auth = useAuthStore();
      auth.logout();
      router.push("/login");
    }
    return Promise.reject(err);
  }
);
```

#### 路由守卫

```ts
// src/router/index.ts
import { useAuthStore } from "@/stores/auth";

router.beforeEach((to) => {
  const auth = useAuthStore();
  if (to.path !== "/login" && !auth.isLoggedIn) {
    return "/login";
  }
  if (to.path === "/login" && auth.isLoggedIn) {
    return "/errors";
  }
});
```

---

### 新增文件清单

```
server/
├── handler/
│   └── auth.go          # 新增：登录接口
├── middleware/
│   └── jwt.go           # 新增：JWT 验证中间件
└── cmd/
    └── hashpwd/
        └── main.go      # 新增：生成密码哈希的命令行工具

dashboard/
└── src/
    ├── stores/
    │   └── auth.ts      # 新增：登录状态管理
    ├── api/
    │   └── auth.ts      # 新增：登录接口请求
    └── views/
        └── Login.vue    # 新增：登录页
```

---

### 新增依赖

**server go.mod：**
```
golang.org/x/crypto
github.com/golang-jwt/jwt/v5
```

**dashboard package.json：**
```
pinia
```

---

### 安全注意事项

| 事项 | 说明 |
|------|------|
| 密码存储 | 使用 bcrypt 哈希，cost 建议 10 以上，不明文存储 |
| JWT Secret | 建议 32 位以上随机字符串，与 AppSecret 不同 |
| Token 有效期 | 默认 24 小时，过期后需重新登录 |
| 登录失败提示 | 不区分用户名不存在和密码错误，防止用户名枚举 |
| HTTPS | 生产环境必须使用 HTTPS，防止 Token 被中间人截获 |
| Token 存储 | 存储在 localStorage，页面关闭后不丢失，注意 XSS 风险 |
