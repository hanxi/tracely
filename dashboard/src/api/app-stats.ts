import api from './index'

export interface MetadataDistItem {
  value: string
  count: number
}

export interface DailyItem {
  date: string
  count: number
}

export interface AppStatsResponse {
  installTotal: number
  installToday: number
  installUV: number
  upgradeTotal: number
  upgradeToday: number
  upgradeUV: number
  dailyInstalls: DailyItem[]
  dailyUpgrades: DailyItem[]
  versionDist: MetadataDistItem[]
  platformDist: MetadataDistItem[]
}

export function getAppStats(days: number = 7) {
  return api.get<AppStatsResponse>('/api/app-stats', { params: { days } })
}
