import api from './index'

export interface DailyStats {
  date: string
  pv: number
  uv: number
}

export interface TopPage {
  page: string
  pv: number
  avgDuration: number
}

export interface StatsResponse {
  daily: DailyStats[]
  topPages: TopPage[]
}

export interface StatsParams {
  days?: number
  appID?: string
}

export function getStats(params?: StatsParams) {
  return api.get<StatsResponse>('/api/stats', { params })
}
