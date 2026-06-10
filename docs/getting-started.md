# 快速开始

## Docker Compose 部署（推荐）

```bash
# 1. 下载配置
mkdir tracely && cd tracely
curl -o docker-compose.yaml https://raw.githubusercontent.com/hanxi/tracely/main/docker-compose.yaml

# 2. 初始化配置
docker compose run --rm tracely ./scripts/gen-config.sh init

# 3. 启动服务
docker compose up -d

# 4. 访问 Dashboard
# http://localhost:3001
```

## 配置管理

`gen-config.sh` 是交互式配置管理工具，支持以下操作：

```bash
# 交互式菜单
docker compose run --rm tracely ./scripts/gen-config.sh

# 常用子命令
docker compose run --rm tracely ./scripts/gen-config.sh show              # 查看配置摘要
docker compose run --rm tracely ./scripts/gen-config.sh set-jwt-secret    # 重新生成 JWT Secret
docker compose run --rm tracely ./scripts/gen-config.sh set-password      # 修改用户密码
docker compose run --rm tracely ./scripts/gen-config.sh app list          # 查看应用列表
docker compose run --rm tracely ./scripts/gen-config.sh app add           # 添加新应用
docker compose run --rm tracely ./scripts/gen-config.sh app set-secret    # 重新生成应用密钥
docker compose run --rm tracely ./scripts/gen-config.sh app set-id        # 修改应用 ID
docker compose run --rm tracely ./scripts/gen-config.sh app remove        # 删除应用
```

## 本地构建

```bash
# 一键构建全部（前端 + 后端）
make build

# 或分步构建
make build-frontend  # 构建 Dashboard（需要 bun）
make build-backend   # 编译后端二进制（需要 Go）

# Docker 构建
make docker
```

## 运行

```bash
# 本地运行
./tracely

# Docker 运行
docker run -d -p 3001:3001 -v $(pwd)/data:/app/data hanxi/tracely:latest
```

访问 Dashboard：http://localhost:3001

## 在线体验

- 体验地址：https://tracely.hanxi.cc/
- 用户名：`admin`
- 密码：`admin123`

## 接入 SDK

### TypeScript SDK（前端应用）

```bash
npm install @imhanxi/tracely-sdk
```

```typescript
import { Tracely } from '@imhanxi/tracely-sdk'

const tracely = new Tracely({
  appId: 'my-app-id',
  appSecret: 'my-app-secret',
  host: 'https://your-server:3001',
})

tracely.init()
```

详细文档：[TypeScript SDK](./sdk-typescript.md)

### Go SDK（后端服务）

```bash
go get github.com/hanxi/tracely/sdk/go/tracely
```

```go
client := tracely.New(tracely.Config{
    AppID:     "my-app-id",
    AppSecret: "my-app-secret",
    Host:      "https://your-server:3001",
})
```

详细文档：[Go SDK](./sdk-go.md)
