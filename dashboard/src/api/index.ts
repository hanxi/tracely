import axios from 'axios'

const api = axios.create({
  baseURL: '',
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('_tracely_token')
  const appId = localStorage.getItem('_tracely_current_app')
  
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  
  if (appId) {
    config.params = { ...config.params, appID: appId }
  }
  
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('_tracely_token')
      localStorage.removeItem('_tracely_user')
      // Hash 模式下使用 hash 路由跳转
      window.location.hash = '#/login'
    }
    return Promise.reject(err)
  }
)

export default api
