import { signedFetch } from './request'
import type { TracelyConfig } from './error'

/**
 * 从 localStorage 读取或生成用户唯一 userId
 */
const USER_ID_KEY = '_tracely_uid'

function getUserId(): string {
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

/**
 * 页面进入时间和当前路径
 */
let pageEnterTime = Date.now()
let currentPage = window.location.pathname

/**
 * 模块级配置（由 initTracker 初始化）
 */
let trackerConfig: TracelyConfig | null = null

/**
 * 上报活跃数据
 */
function reportActive(page: string, duration: number): void {
  if (!trackerConfig) {
    console.warn('[Tracely] Tracker not initialized')
    return
  }

  const payload = {
    appId: trackerConfig.appId,
    userId: getUserId(),
    page,
    duration: Math.floor(duration / 1000), // 转换为秒
  }

  signedFetch(trackerConfig.host, '/report/active', payload, trackerConfig.appId, trackerConfig.appSecret)
}

/**
 * 上报当前页面停留时长
 */
function reportCurrentPageStay(): void {
  if (!trackerConfig) {
    console.warn('[Tracely] Tracker not initialized')
    return
  }

  const now = Date.now()
  const duration = now - pageEnterTime
  if (duration > 0) {
    reportActive(currentPage, duration)
  }
}

/**
 * 初始化活跃统计
 */
export function initTracker(config: TracelyConfig): void {
  trackerConfig = config

  // 监听 beforeunload 上报最后一个页面的停留时长
  window.addEventListener('beforeunload', () => {
    reportCurrentPageStay()
  })

  // 监听页面可见性变化（切到后台时上报）
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'hidden') {
      reportCurrentPageStay()
      pageEnterTime = Date.now() // 重置计时
    }
  })
}

/**
 * 路由切换时调用
 */
export function onRouteChange(newPath: string): void {
  if (!trackerConfig) {
    console.warn('[Tracely] Tracker not initialized')
    return
  }

  // 上报上一个页面的停留时长
  reportCurrentPageStay()

  // 更新当前路径和进入时间
  currentPage = newPath
  pageEnterTime = Date.now()
}