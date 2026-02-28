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
import Stats from './pages/stats.vue'

const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

const routes = [
  { path: '/', component: Index },
  { path: '/login', component: Login },
  { path: '/errors', component: Errors },
  { path: '/stats', component: Stats },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to) => {
  const token = localStorage.getItem('_tracely_token')
  if (to.path !== '/login' && !token) return '/login'
  if (to.path === '/login' && token) return '/'
})

const app = createApp(App)
app.use(pinia)
app.use(router)
app.use(ui)
app.mount('#app')
