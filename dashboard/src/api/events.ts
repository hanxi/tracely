import axios from './index'

// 事件统计响应
export interface EventStatsResponse {
  stats: Array<{
    eventName: string
    count: number
  }>
}

// Top 事件响应
export interface TopEventsResponse {
  events: Array<{
    eventName: string
    count: number
  }>
}

// 每日事件响应
export interface DailyEventsResponse {
  daily: Array<{
    date: string
    count: number
  }>
}

// 事件概览响应
export interface EventOverviewResponse {
  todayEventCount: number
  todayActivePV: number
  todayActiveUV: number
  topEvents: Array<{
    eventName: string
    count: number
  }>
}

/**
 * 获取事件统计
 * @param days 统计天数
 * @param appID 应用 ID（可选）
 * @param eventName 事件名称（可选）
 */
export async function getEventStats(
  days: number = 7,
  appID?: string,
  eventName?: string
): Promise<EventStatsResponse> {
  const params: Record<string, string | number> = { days }
  if (appID) params.appID = appID
  if (eventName) params.eventName = eventName

  const { data } = await axios.get('/api/events/stats', { params })
  return data
}

/**
 * 获取 Top 事件排行
 * @param days 统计天数
 * @param appID 应用 ID（可选）
 * @param limit 返回数量限制
 */
export async function getTopEvents(
  days: number = 7,
  appID?: string,
  limit: number = 10
): Promise<TopEventsResponse> {
  const params: Record<string, string | number> = { days, limit }
  if (appID) params.appID = appID

  const { data } = await axios.get('/api/events/top', { params })
  return data
}

/**
 * 获取每日事件统计
 * @param days 统计天数
 * @param appID 应用 ID（可选）
 * @param eventName 事件名称（可选）
 */
export async function getDailyEvents(
  days: number = 7,
  appID?: string,
  eventName?: string
): Promise<DailyEventsResponse> {
  const params: Record<string, string | number> = { days }
  if (appID) params.appID = appID
  if (eventName) params.eventName = eventName

  const { data } = await axios.get('/api/events/daily', { params })
  return data
}

/**
 * 获取事件概览数据
 * @param appID 应用 ID（可选）
 */
export async function getEventOverview(appID?: string): Promise<EventOverviewResponse> {
  const params: Record<string, string> = {}
  if (appID) params.appID = appID

  const { data } = await axios.get('/api/events/overview', { params })
  return data
}

// 事件详情
export interface EventDetail {
  id: number
  eventName: string
  metadata: Record<string, any> | null
  appId: string
  userId: string
  createdAt: string
}

// 事件列表响应
export interface EventListResponse {
  list: EventDetail[]
  total: number
}

// 事件统计摘要响应
export interface EventStatsSummaryResponse {
  totalCount: number
  todayCount: number
  uv: number
}

/**
 * 获取事件列表
 * @param eventName 事件名称
 * @param page 页码
 * @param pageSize 每页数量
 * @param appID 应用 ID（可选）
 */
export async function getEventList(
  eventName: string,
  page: number = 1,
  pageSize: number = 20,
  appID?: string
): Promise<EventListResponse> {
  const params: Record<string, string | number> = { eventName, page, pageSize }
  if (appID) params.appID = appID

  const { data } = await axios.get('/api/events/list', { params })
  return data
}

/**
 * 获取事件统计摘要
 * @param eventName 事件名称
 * @param days 统计天数
 * @param appID 应用 ID（可选）
 */
export async function getEventStatsSummary(
  eventName: string,
  days: number = 7,
  appID?: string
): Promise<EventStatsSummaryResponse> {
  const params: Record<string, string | number> = { eventName, days }
  if (appID) params.appID = appID

  const { data } = await axios.get('/api/events/stats/summary', { params })
  return data
}
