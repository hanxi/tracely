# Tracely TypeScript SDK

Tracely TypeScript SDK 是一个轻量级的前端监控客户端，支持错误捕获和活跃统计。

## 功能特性

- 🐛 **自动错误捕获**：JS 运行时错误、Promise 未处理异常、Vue 组件错误
- 📊 **活跃统计**：页面 PV/UV、停留时长、路由切换追踪
- 📝 **手动上报**：支持手动上报错误和自定义事件
- ⏱️ **节流控制**：相同错误 1 分钟内只上报一次，避免重复数据
- 🔇 **静默失败**：上报失败不影响业务逻辑

## 安装

### npm

```bash
npm install @imhanxi/tracely-sdk
```

### yarn

```bash
yarn add @imhanxi/tracely-sdk
```

### pnpm

```bash
pnpm add @imhanxi/tracely-sdk
```

### bun

```bash
bun add @imhanxi/tracely-sdk
```

### CDN 引入

```html
<script src="https://unpkg.com/@imhanxi/tracely-sdk/dist/index.js"></script>
```

## 配置

在使用 SDK 前，需要先在 Tracely 服务器配置中获取以下信息：

1. **AppID**：应用唯一标识（从 `config.yaml` 的 `apps[].appId` 获取）
2. **AppSecret**：应用密钥（从 `config.yaml` 的 `apps[].appSecret` 获取）
3. **Host**：Tracely 服务器地址（如：`https://tracely.example.com`）

## 快速开始

### 基础用法（纯 JS/TS 项目）

```typescript
import { Tracely } from '@imhanxi/tracely-sdk'

// 创建 SDK 实例
const tracely = new Tracely({
  appId: 'my-app-id',
  appSecret: 'my-app-secret',
  host: 'https://tracely.example.com',
})

// 初始化 SDK（自动捕获全局错误和页面活跃统计）
tracely.init()
```

### Vue 项目接入

```typescript
// main.ts
import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { Tracely } from '@imhanxi/tracely-sdk'
import App from './App.vue'

const app = createApp(App)

const router = createRouter({
  history: createWebHistory(),
  routes: [...],
})

// 创建并初始化 SDK
const tracely = new Tracely({
  appId: 'my-app-id',
  appSecret: 'my-app-secret',
  host: 'https://tracely.example.com',
})

// 传入 app 和 router 实例，自动捕获 Vue 错误和路由变化
tracely.init(app, router)

app.use(router)
app.mount('#app')
```

### React 项目接入

```typescript
// index.tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import { Tracely } from '@imhanxi/tracely-sdk'
import App from './App'

// 创建并初始化 SDK
const tracely = new Tracely({
  appId: 'my-app-id',
  appSecret: 'my-app-secret',
  host: 'https://tracely.example.com',
})

// 初始化 SDK（自动捕获全局错误）
tracely.init()

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
```

## API 参考

### Tracely 类

#### 构造函数

```typescript
new Tracely(config: TracelyConfig)
```

**配置参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| appId | string | ✅ | 应用 ID |
| appSecret | string | ✅ | 应用密钥 |
| host | string | ✅ | Tracely 服务器地址 |

#### init() 方法

```typescript
tracely.init(app?: VueApp, router?: VueRouter): void
```

**参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| app | Vue App 实例 | ❌ | Vue 应用实例，传入后自动捕获 Vue 组件错误 |
| router | Vue Router 实例 | ❌ | Vue Router 实例，传入后自动追踪路由变化 |

**功能：**
- 注册全局错误监听器（`window.error`）
- 注册 Promise 未处理异常监听器（`window.unhandledrejection`）
- 初始化页面活跃统计（PV/UV、停留时长）
- 如果传入 `app`，注册 Vue 错误处理器
- 如果传入 `router`，注册路由切换钩子

#### reportError() 方法

```typescript
tracely.reportError(error: Error, info?: string): void
```

**参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| error | Error | ✅ | 错误对象 |
| info | string | ❌ | 附加信息（如：错误发生时的上下文） |

**示例：**
```typescript
try {
  JSON.parse(invalidJson)
} catch (err) {
  tracely.reportError(err as Error, '解析用户配置失败')
}
```

### 导出函数

#### captureError()

```typescript
import { captureError } from '@imhanxi/tracely-sdk'

captureError(config: TracelyConfig, error: Error, info?: string): void
```

手动上报错误的静态方法，无需创建 SDK 实例。

#### onRouteChange()

```typescript
import { onRouteChange } from '@imhanxi/tracely-sdk'

onRouteChange(newPath: string): void
```

手动触发路由切换事件，用于 SPA 应用。

## 自动捕获的错误类型

| 错误类型 | 触发场景 | 示例 |
|---------|---------|------|
| `jsError` | JS 运行时错误 | `Cannot read property 'name' of undefined` |
| `promiseError` | Promise 未处理异常 | `Promise rejected: Network timeout` |
| `manualError` | 手动上报的错误 | `tracely.reportError()` 调用 |

## 活跃统计

### 自动追踪

SDK 会自动追踪以下行为：

- **页面加载**：页面打开时自动记录进入时间
- **页面关闭**：离开页面前上报停留时长
- **页面切到后台**：用户切换标签页时上报
- **路由切换**：如果传入 Vue Router，自动追踪路由变化

### 用户标识

SDK 会自动生成并存储用户唯一标识到 `localStorage`：

- Key：`_tracely_uid`
- 格式：UUID v4 或随机字符串（降级方案）
- 持久化：长期保存，用于统计 UV

### 上报数据

每次活跃上报包含：

```typescript
{
  eventName: string,    // 事件名称（如 "_active"）
  metadata: {           // 元数据
    page: string,       // 页面路径
    duration: number,   // 停留时长（秒）
    // ...其他自定义字段
  },
  appId: string,        // 应用 ID
  userId: string,       // 用户唯一标识
}
```

## 节流控制

为避免相同错误重复上报，SDK 实现了节流机制：

- **节流维度**：错误指纹（`type:message`）
- **节流时间**：1 分钟
- **行为**：相同错误 1 分钟内只上报一次

## 安全说明

### 签名机制

SDK 使用 HMAC-SHA256 签名确保请求安全：

```typescript
signature = HMAC-SHA256(appId + timestamp + nonce, appSecret)
```

**请求头：**
- `X-App-Id`：应用 ID
- `X-Timestamp`：当前时间戳（毫秒）
- `X-Nonce`：随机字符串（防重放）
- `X-Signature`：签名值

### AppSecret 安全性

**注意**：AppSecret 在前端代码中是可见的，建议：

1. 对打包产物进行代码混淆
2. 使用构建工具压缩代码
3. 定期更换 AppSecret
4. 配合服务端的限速策略使用

## 构建 SDK

如果需要修改 SDK 源码并重新构建：

```bash
# 进入 sdk/ts 目录
cd sdk/ts

# 安装依赖
npm install

# 开发模式（监听变化）
npm run dev

# 生产构建
npm run build
```

构建产物：
- `dist/index.js` - CommonJS 格式
- `dist/index.mjs` - ES Module 格式
- `dist/index.d.ts` - TypeScript 类型定义

## 测试数据生成

SDK 提供了测试脚本用于生成模拟数据：

```bash
# 编辑 test-data.ts 配置
# 修改 appId, appSecret, host

# 运行测试脚本（需要安装 Bun）
bun run test-data.ts
```

生成的数据：
- 活跃数据：覆盖多个页面，随机用户和停留时长
- 错误数据：多种错误类型，模拟真实场景

## 常见问题

### Q1: 为什么上报失败没有报错？

SDK 采用静默失败策略，上报错误不会影响业务逻辑。可以在浏览器控制台查看警告信息：`[Tracely] Failed to send report:`

### Q2: 如何调试 SDK？

打开浏览器开发者工具：
- Console 中查看 `[Tracely]` 开头的日志
- Network 面板查看 `/report/*` 请求

### Q3: 如何禁用 SDK？

根据环境条件判断是否初始化：

```typescript
if (process.env.NODE_ENV === 'production') {
  tracely.init()
}
```

### Q4: 支持哪些浏览器？

支持现代浏览器：
- Chrome / Edge（最新版）
- Firefox（最新版）
- Safari（最新版）

需要 `fetch`、`localStorage`、`crypto` API 支持。

## 许可证

MIT License
