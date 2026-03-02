import { signedFetch } from './request'
import type { TracelyConfig } from './error'

/**
 * 事件上报数据结构
 */
export interface EventPayload {
  eventName: string
  metadata?: Record<string, unknown>
  appId: string
  userId: string
}

/**
 * 上报事件
 * @param config SDK 配置
 * @param eventName 事件名称
 * @param metadata 元数据（可选，可包含 page 和 duration 等字段）
 * @param userId 用户 ID
 */
export function reportEvent(
  config: TracelyConfig,
  eventName: string,
  metadata: Record<string, unknown> | undefined,
  userId: string
): void {
  const payload: EventPayload = {
    eventName,
    metadata,
    appId: config.appId,
    userId,
  }

  signedFetch(config.host, '/report/event', payload, config.appId, config.appSecret)
}
