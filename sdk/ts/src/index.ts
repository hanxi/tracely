import type { TracelyConfig } from './error'
import { initErrorCapture, captureError } from './error'
import { initTracker, onRouteChange } from './tracker'

/**
 * 扩展 Vue App 类型
 */
interface VueApp {
  config: {
    errorHandler?: (err: unknown, instance: unknown, info: string) => void
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
}

export type { TracelyConfig }
export { captureError, onRouteChange }
