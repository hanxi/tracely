import type { TracelyConfig } from './error'
import { initErrorCapture, captureError } from './error'
import { initTracker, onRouteChange } from './tracker'
import { reportEvent as reportEventFn } from './event'

/**
 * 扩展 Vue App 类型
 */
interface VueApp {
  config: {
    errorHandler?: (
      err: unknown,
      instance: any,
      info: string
    ) => void
  }
}

/**
 * 扩展 Vue Router 类型
 */
interface VueRouter {
  afterEach: (guard: (to: unknown, from: unknown) => void) => void
}

/**
 * Tracely SDK 主类
 */
export class Tracely {
  private config: TracelyConfig

  constructor(config: TracelyConfig) {
    this.config = config
  }

  /**
   * 初始化 SDK
   * @param app Vue app 实例（可选）
   * @param router Vue Router 实例（可选）
   */
  init(app?: VueApp, router?: VueRouter): void {
    // 初始化错误捕获
    initErrorCapture(this.config)

    // 如果传入 Vue app，注册 Vue 错误处理器
    if (app) {
      app.config.errorHandler = (err: unknown, _instance: unknown, info: string) => {
        if (err instanceof Error) {
          captureError(this.config, err, info)
        }
      }
    }

    // 初始化活跃统计
    initTracker(this.config)

    // 如果传入 Vue Router，注册路由切换钩子
    if (router) {
      router.afterEach((_to, _from) => {
        // 获取新路径
        const newPath = typeof window !== 'undefined' ? window.location.pathname : ''
        onRouteChange(newPath)
      })
    }
  }

  /**
   * 手动上报错误
   */
  reportError(error: Error, info?: string): void {
    captureError(this.config, error, info)
  }

  /**
   * 手动上报事件
   * @param eventName 事件名称
   * @param metadata 元数据（可选，可包含 page 和 duration 等字段）
   * @param userId 用户 ID（可选，默认使用自动生成的 userId）
   */
  reportEvent(
    eventName: string,
    metadata?: Record<string, unknown>,
    userId?: string
  ): void {
    const finalUserId = userId || this.getUserId()
    reportEventFn(this.config, eventName, metadata, finalUserId)
  }

  /**
   * 获取用户 ID
   */
  private getUserId(): string {
    const USER_ID_KEY = '_tracely_uid'
    let userId = localStorage.getItem(USER_ID_KEY)
    if (!userId) {
      if (typeof crypto !== 'undefined' && crypto.randomUUID) {
        userId = crypto.randomUUID()
      } else {
        userId = Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15)
      }
      localStorage.setItem(USER_ID_KEY, userId)
    }
    return userId
  }
}

export type { TracelyConfig }
export { captureError, onRouteChange, reportEventFn as reportEvent }
