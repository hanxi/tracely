import api from './index'

export interface ErrorLog {
  ID: number
  Fingerprint?: string
  Type: string
  Message: string
  Stack: string
  URL: string
  AppID?: string
  UserAgent?: string
  Count: number
  FirstSeen?: string
  LastSeen?: string
}

export interface ErrorListResponse {
  list: ErrorLog[]
  total: number
}

export interface ErrorListParams {
  page?: number
  pageSize?: number
  type?: string
}

export function getErrorList(params?: ErrorListParams) {
  return api.get<ErrorListResponse>('/api/errors', { params })
}
