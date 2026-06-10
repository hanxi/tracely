# Tracely TypeScript SDK

轻量级前端监控客户端，支持错误捕获和活跃统计。

完整文档请参阅 [docs/sdk-typescript.md](../../docs/sdk-typescript.md)。

## 安装

```bash
npm install @imhanxi/tracely-sdk
```

## 快速开始

```typescript
import { Tracely } from '@imhanxi/tracely-sdk'

const tracely = new Tracely({
  appId: 'my-app-id',
  appSecret: 'my-app-secret',
  host: 'https://tracely.example.com',
})

tracely.init()
```
