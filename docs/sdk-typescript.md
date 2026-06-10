# Tracely TypeScript SDK

轻量级前端监控客户端，支持错误捕获、活跃统计和事件上报。

## 功能特性

- 自动错误捕获：JS 运行时错误、Promise 未处理异常、Vue 组件错误
- 活跃统计：页面 PV/UV、停留时长、路由切换追踪
- 手动上报：支持手动上报错误和自定义事件
- 安装/升级追踪：支持应用安装和升级事件上报
- 节流控制：相同错误 1 分钟内只上报一次
- 静默失败：上报失败不影响业务逻辑

## 安装

```bash
npm install @imhanxi/tracely-sdk
# 或
yarn add @imhanxi/tracely-sdk
# 或
pnpm add @imhanxi/tracely-sdk
# 或
bun add @imhanxi/tracely-sdk
```

### CDN 引入

```html
<!-- ES Module -->
<script type="module">
  import { Tracely } from 'https://unpkg.com/@imhanxi/tracely-sdk/dist/tracely-sdk.mjs'
</script>

<!-- UMD -->
<script src="https://unpkg.com/@imhanxi/tracely-sdk/dist/tracely-sdk.js"></script>
```

## 快速开始

### 基础用法

```typescript
import { Tracely } from '@imhanxi/tracely-sdk'

const tracely = new Tracely({
  appId: 'my-app-id',
  appSecret: 'my-app-secret',
  host: 'https://tracely.example.com',
})

tracely.init()
```

### Vue 项目

```typescript
import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { Tracely } from '@imhanxi/tracely-sdk'
import App from './App.vue'

const app = createApp(App)
const router = createRouter({ history: createWebHistory(), routes: [...] })

const tracely = new Tracely({
  appId: 'my-app-id',
  appSecret: 'my-app-secret',
  host: 'https://tracely.example.com',
})

// 传入 app 和 router，自动捕获 Vue 错误和路由变化
tracely.init(app, router)

app.use(router)
app.mount('#app')
```

### React 项目

```typescript
import { Tracely } from '@imhanxi/tracely-sdk'

const tracely = new Tracely({
  appId: 'my-app-id',
  appSecret: 'my-app-secret',
  host: 'https://tracely.example.com',
})

tracely.init()
```

## API 参考

### 构造函数

```typescript
new Tracely(config: TracelyConfig)
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| appId | string | 是 | 应用 ID |
| appSecret | string | 是 | 应用密钥 |
| host | string | 是 | 服务器地址 |

### init(app?, router?)

初始化 SDK：注册全局错误监听、Promise 异常监听、页面活跃统计。传入 Vue app 自动捕获组件错误，传入 router 自动追踪路由变化。

### reportError(error, info?)

手动上报错误。

```typescript
try {
  JSON.parse(invalidJson)
} catch (err) {
  tracely.reportError(err as Error, '解析配置失败')
}
```

### reportEvent(eventName, metadata?, userId?)

上报自定义事件。userId 可选，默认使用自动生成的 ID。

```typescript
tracely.reportEvent('purchase', { productId: '12345', amount: 99.99 })
```

### reportInstall(version, platform, userId?)

上报应用安装事件。

```typescript
tracely.reportInstall('1.0.0', 'web')
```

### reportUpgrade(fromVersion, toVersion, platform, userId?)

上报应用升级事件。

```typescript
tracely.reportUpgrade('1.0.0', '2.0.0', 'web')
```

### 导出函数

```typescript
import { captureError, onRouteChange, reportEvent } from '@imhanxi/tracely-sdk'
```

- `captureError(config, error, info?)` — 静态方法，无需实例
- `onRouteChange(newPath)` — 手动触发路由切换
- `reportEvent(config, eventName, metadata, userId)` — 静态方法

## 自动捕获的错误类型

| 类型 | 触发场景 |
|------|---------|
| `jsError` | JS 运行时错误（window.error） |
| `promiseError` | Promise 未处理异常（unhandledrejection） |
| `manualError` | 手动调用 reportError() |

## 活跃统计

SDK 自动追踪：

- **页面加载**：进入页面时记录时间
- **页面关闭**：离开前上报停留时长（`beforeunload`）
- **页面切后台**：切换标签页时上报（`visibilitychange`）
- **路由切换**：传入 Vue Router 后自动追踪

### 用户标识

自动生成 UUID 存储到 `localStorage`（key: `_tracely_uid`），用于 UV 统计。

## 节流控制

- 维度：错误指纹（`type:message`）
- 窗口：1 分钟
- 相同错误 1 分钟内只上报一次

## 安全说明

签名机制：`HMAC-SHA256(appId + timestamp + nonce, appSecret)`

**注意**：AppSecret 在前端代码中可见，建议：
1. 对打包产物进行代码混淆
2. 定期更换 AppSecret
3. 配合服务端限速策略使用

## 构建 SDK

```bash
cd sdk/ts
npm install
npm run build    # 生产构建
npm run dev      # 监听模式
```

构建产物：
- `dist/tracely-sdk.js` — UMD 格式
- `dist/tracely-sdk.mjs` — ES Module 格式
- `dist/index.d.ts` — TypeScript 类型定义

## 测试数据生成

```bash
cd sdk/ts
# 编辑 test-data.ts 中的 appId, appSecret, host
bun run test-data.ts
```

生成 25 条活跃记录 + 15 条错误记录（覆盖多种页面和错误类型），请求间隔 100ms。

## 常见问题

**Q: 上报失败无报错？** SDK 采用静默失败策略。打开浏览器控制台查看 `[Tracely]` 日志和 Network 面板 `/report/*` 请求。

**Q: 如何禁用 SDK？** 按环境判断是否初始化：

```typescript
if (process.env.NODE_ENV === 'production') {
  tracely.init()
}
```

**Q: 支持哪些浏览器？** 现代浏览器（Chrome/Edge/Firefox/Safari 最新版），需要 `fetch`、`localStorage`、`crypto` API。

## 许可证

MIT License
