# Tracely SDK 测试数据生成器

用于生成测试数据填充到 Tracely 监控系统中。

## 前置要求

- 已安装 [Bun](https://bun.sh/) (版本 1.3.6+)
- Tracely 服务器正在运行（默认 http://localhost:3001）

## 配置

编辑 `test-data.ts` 文件中的配置部分：

```typescript
const config = {
  appId: 'test-app-id',      // 替换为你的 AppID
  appSecret: 'test-app-secret', // 替换为你的 AppSecret
  host: 'http://localhost:3001', // Tracely 服务器地址
}
```

## 运行

在项目根目录执行：

```bash
# 方式 1：使用 npm 脚本
npm run test:data

# 方式 2：直接使用 Bun
bun run test-data.ts
```

## 生成的数据

测试脚本会生成以下数据：

### 活跃数据
- **25 条** 活跃记录
- 覆盖 5 个页面：`/dashboard`, `/dashboard/overview`, `/dashboard/errors`, `/dashboard/active`, `/dashboard/settings`
- 每个页面 5 条记录
- 随机用户 ID 和停留时长（10-310 秒）

### 错误数据
- **15 条** 错误记录
- 5 种错误类型：
  - JS 运行时错误（TypeError、ReferenceError）
  - Promise 未处理错误（网络超时、API 500）
  - 手动上报错误
- 每种类型 3 条记录
- 随机分布在各个页面

## 注意事项

1. 确保 Tracely 服务器正在运行，否则请求会失败
2. 测试数据使用随机生成的用户 ID，不会与真实数据冲突
3. 请求之间有 100ms 延迟，避免请求过快
4. 测试完成后可以在 Dashboard 查看上报的数据
