import api from './index'

export interface LoginResponse {
  token: string
  username: string
}

export function login(username: string, password: string) {
  return api.post<LoginResponse>('/auth/login', { username, password })
}
