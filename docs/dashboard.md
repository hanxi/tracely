# Tracely Dashboard

> 实时监控应用状态，快速定位和解决问题

## 技术栈

| 类别 | 技术 |
|------|------|
| 框架 | Vue 3 (Composition API) |
| 构建工具 | Vite |
| UI 组件库 | @nuxt/ui |
| 状态管理 | Pinia + 持久化插件 |
| 路由 | Vue Router (Hash 模式) |
| HTTP 客户端 | Axios |
| 样式 | Tailwind CSS |

## 页面功能

### 概览页 `/`

- **数据卡片**：今日 PV、今日 UV、错误总数、今日新增错误、今日安装、今日升级
- **Top 5 错误**：展示出现次数最多的错误列表（类型、消息、次数）
- 快速跳转到错误列表页

### 错误列表页 `/errors`

- 表格展示所有错误，字段：错误类型、错误信息、出现次数、最近出现
- 支持按错误类型筛选（全部 / jsError / promiseError / vueError）
- 支持分页（每页 20 条）
- 点击"详情"查看完整错误信息（类型、消息、堆栈、URL、首次/最近出现时间）
- 支持多应用切换查看

### 事件统计页 `/events`

- **事件类型分布**：展示所有事件类型及其数量
- **每日事件趋势**：表格展示每日事件统计数据
- **Top 10 事件排行**：展示最热门的事件（支持筛选事件类型）
- 支持切换统计天数（7/14/30 天）
- 支持按事件类型筛选
- 支持多应用切换查看

### 应用统计页 `/app-stats`

- **摘要卡片**：安装总数、今日安装、安装用户数、升级总数、今日升级、升级用户数
- **每日安装/升级趋势**：表格展示每日数据
- **版本分布**：按安装版本排名
- **平台分布**：按平台（android/ios/web 等）排名
- 支持切换统计天数（7/14/30 天）

### 登录页 `/login`

- 用户名 + 密码登录
- JWT Token 认证
- 登录状态持久化（localStorage）
- 路由守卫保护

## 通用功能

- **明暗色模式**：基于 Nuxt UI 自动适配，支持手动切换
- **响应式布局**：桌面端侧边栏布局，移动端抽屉式菜单
- **Hash 路由**：使用 `createWebHashHistory`
- **用户菜单**：显示当前用户，支持退出登录
- **应用切换**：多应用配置时显示切换下拉框

## 开发

```bash
cd dashboard
bun install          # 安装依赖
bun dev              # 开发模式（http://localhost:5173）
bun run build        # 生产构建
bun run lint         # ESLint 检查
```

开发环境需要后端服务运行在 `localhost:3001`，API 代理在 `vite.config.ts` 中配置。

## 数据存储

- 登录凭证：localStorage（`_tracely_token`、`_tracely_user`）
- 应用选择状态：localStorage（`_tracely_current_app`）
- Pinia 状态通过持久化插件自动保存

## 浏览器兼容性

Chrome >= 90, Firefox >= 88, Safari >= 14, Edge >= 90
