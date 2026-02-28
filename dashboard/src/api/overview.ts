import api from './index'

export interface TopError {
  type: string
  message: string
  count: number
}

export interface ErrorTrend {
  date: string
  count: number
}

export interface OverviewResponse {
  todayPV: number
  todayUV: number
  totalErrors: number
  todayErrors: number
  topErrors: TopError[]
  errorTrend: ErrorTrend[]
}

export function getOverview() {
  return api.get<OverviewResponse>('/api/overview')
}
