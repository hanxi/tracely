import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const api = axios.create({
  baseURL: '',
})

api.interceptors.request.use((config) => {
  // 从 store 获取 token 和 appId
  const authStore = useAuthStore()
  const appStore = useAppStore()
  
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  
  if (appStore.currentAppId) {
    config.params = { ...config.params, appID: appStore.currentAppId }
  }
  
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      // 使用 store 的 logout 方法，会自动清除持久化状态
      const authStore = useAuthStore()
      authStore.logout()
      // Hash 模式下使用 hash 路由跳转
      window.location.hash = '#/login'
    }
    return Promise.reject(err)
  }
)

export default api
