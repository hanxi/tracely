# 开发指南

## 环境搭建

### 后端开发

```bash
# 克隆项目
git clone https://github.com/hanxi/tracely.git
cd tracely

# 安装 Go 依赖
go mod download

# 配置
cp config.example.yaml config.yaml
# 编辑 config.yaml，设置密码哈希等

# 生成密码哈希
go run . -hashpwd -password yourpassword

# 开发模式运行
make dev
```

### Dashboard 开发

```bash
cd dashboard
bun install
bun dev
# 访问 http://localhost:5173
```

Vite 开发代理配置（`vite.config.ts`）：

```typescript
server: {
  proxy: {
    '/api': { target: 'http://localhost:3001' },
    '/auth': { target: 'http://localhost:3001' },
  }
}
```

### Go SDK 开发

```bash
cd sdk/go/tracely
go test -v ./...

# 本地引用测试：在测试项目的 go.mod 中添加
# replace github.com/hanxi/tracely/sdk/go => ../sdk/go
```

### TS SDK 开发

```bash
cd sdk/ts
npm install
npm run dev    # 监听模式构建
npm run build  # 生产构建
```

## 扩展指南

### 新增 API 接口

1. **创建 Handler**（`internal/handler/xxx.go`）：

```go
func HandleXXX(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "success"})
    }
}
```

2. **注册路由**（`main.go`）：

```go
api.GET("/xxx", handler.HandleXXX(db))
```

3. **前端 API 封装**（`dashboard/src/api/xxx.ts`）：

```typescript
export function getXxx() {
  return api.get<XxxResponse>('/api/xxx')
}
```

### 新增数据表

1. **定义模型**（`internal/model/xxx.go`）：

```go
type XxxLog struct {
    ID        uint      `gorm:"primaryKey"`
    AppID     string    `gorm:"index"`
    Data      string
    CreatedAt time.Time `gorm:"index"`
}
```

2. **数据库迁移**（在 `model/db.go` 的 `InitDB` 中添加）：

```go
db.AutoMigrate(&XxxLog{})
```

### 新增 Dashboard 页面

1. **创建页面组件**（`dashboard/src/pages/xxx.vue`）
2. **注册路由**（`dashboard/src/main.ts`）：

```typescript
import Xxx from './pages/xxx.vue'
// routes 数组中添加
{ path: '/xxx', component: Xxx }
```

3. **添加导航菜单**（`dashboard/src/layouts/default.vue`）

### 新增中间件

在 `internal/middleware/` 创建新文件，实现 `gin.HandlerFunc`，然后在 `main.go` 中注册。

### 新增 SDK 语言

参考 Go SDK 结构，核心实现：
- HMAC-SHA256 签名
- 异步队列 + 后台发送
- `ReportError()` 和 `ReportEvent()` 方法

## 测试策略

### 后端测试

```go
// 签名验证测试
func TestGenerateSignature(t *testing.T) {
    sig := generateSignature("app1", "secret1", "1234567890", "nonce1")
    // 验证签名正确性
}

// Handler 集成测试
func TestReportError(t *testing.T) {
    db := setupTestDB()
    req := httptest.NewRequest("POST", "/report/error", body)
    w := httptest.NewRecorder()
    handler := ReportError(db)
    handler.ServeHTTP(w, req)
    assert.Equal(t, 200, w.Code)
}
```

### 测试数据生成

TS SDK 提供了测试脚本：

```bash
cd sdk/ts
# 编辑 test-data.ts 中的 appId, appSecret, host 配置
bun run test-data.ts
```

生成 25 条活跃记录 + 15 条错误记录，覆盖多种页面和错误类型。

## 常用命令

```bash
make build              # 一键构建前端 + 后端
make build-frontend     # 仅构建 Dashboard
make build-backend      # 编译后端二进制
make dev                # 后端开发模式
make docker             # 构建 Docker 镜像
cd dashboard && bun dev # 前端开发模式
cd dashboard && bun run lint   # ESLint 检查
```
