import { signedFetch } from './request'

/**
 * 错误上报数据结构
 */
export interface ErrorData {
  type: string
  message: string
  stack?: string
  url: string
}

/**
 * 节流缓存：fingerprint -> 上次上报时间戳
 */
const reportedCache = new Map<string, number>()
const THROTTLE_MS = 60 * 1000 // 1 分钟

/**
 * 生成错误指纹（简化版，不引入额外依赖）
 */
function genFingerprint(type: string, message: string): string {
  return `${type}:${message}`.slice(0, 200)
}

/**
 * 判断是否应该上报（节流控制）
 */
function shouldReport(fingerprint: string): boolean {
  const lastTime = reportedCache.get(fingerprint)
  const now = Date.now()
  if (lastTime && now - lastTime < THROTTLE_MS) {
    return false // 1 分钟内已上报过，跳过
  }
  reportedCache.set(fingerprint, now)
  return true
}

/**
 * TracelyConfig 配置接口
 */
export interface TracelyConfig {
  appId: string
  appSecret: string
  host: string
}

/**
 * 上报错误（内部函数）
 */
function reportError(config: TracelyConfig, errorData: ErrorData): void {
  // 节流控制
  const fingerprint = genFingerprint(errorData.type, errorData.message)
  if (!shouldReport(fingerprint)) {
    return
  }

  signedFetch(config.host, '/report/error', errorData, config.appId, config.appSecret)
}

/**
 * 初始化错误捕获
 */
export function initErrorCapture(config: TracelyConfig): void {
  // 监听 window.error 事件，捕获 JS 运行时错误
  window.addEventListener('error', (event) => {
    // 过滤资源加载错误（避免与 JS 错误混淆）
    if (event.target instanceof HTMLElement) {
      return
    }

    const errorData: ErrorData = {
      type: 'jsError',
      message: event.message,
      stack: event.error?.stack,
      url: window.location.href,
    }

    reportError(config, errorData)
  })

  // 监听 window.unhandledrejection 事件，捕获 Promise 异常
  window.addEventListener('unhandledrejection', (event) => {
    const reason = event.reason
    const errorData: ErrorData = {
      type: 'promiseError',
      message: reason?.message || String(reason),
      stack: reason?.stack,
      url: window.location.href,
    }

    reportError(config, errorData)
  })
}

/**
 * 手动上报错误
 */
export function captureError(config: TracelyConfig, error: Error, info?: string): void {
  const errorData: ErrorData = {
    type: 'manualError',
    message: info ? `${error.message}: ${info}` : error.message,
    stack: error.stack,
    url: window.location.href,
  }

  reportError(config, errorData)
}