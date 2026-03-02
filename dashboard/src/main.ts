import './assets/css/main.css'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import { createRouter, createWebHashHistory } from 'vue-router'
import ui from '@nuxt/ui/vue-plugin'
import App from './App.vue'

// 导入路由配置
import Index from './pages/index.vue'
import Login from './pages/login.vue'
import Errors from './pages/errors.vue'
import Events from './pages/events.vue'

// 导入 Store
import { useAuthStore } from './stores/auth'

const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

const routes = [
  { path: '/', component: Index },
  { path: '/login', component: Login },
  { path: '/errors', component: Errors },
  { path: '/events', component: Events },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to) => {
  // 从 store 获取 token（Pinia 持久化插件会自动恢复）
  const authStore = useAuthStore()
  if (to.path !== '/login' && !authStore.token) return '/login'
  if (to.path === '/login' && authStore.token) return '/'
})

const app = createApp(App)
app.use(pinia)
app.use(router)
app.use(ui)
app.mount('#app')
