import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api'

export interface AppInfo {
  appId: string
  appName: string
}

export const useAppStore = defineStore('app', () => {
  const currentAppId = ref('')
  const apps = ref<AppInfo[]>([])
  const hasMultipleApps = computed(() => apps.value.length > 1)

  async function fetchApps() {
    try {
      const res = await api.get('/api/apps')
      const appList: AppInfo[] = res.data.apps || []
      // 直接更新，不保留旧数据
      apps.value = appList
      if (appList.length > 0) {
        // 优先从 localStorage 恢复之前选择的应用 ID
        const savedAppId = localStorage.getItem('_tracely_current_app')
        if (savedAppId && appList.some((app: AppInfo) => app.appId === savedAppId)) {
          currentAppId.value = savedAppId
        } else {
          // 如果没有保存的或保存的无效，使用第一个应用
          currentAppId.value = appList[0].appId
        }
      }
    } catch (error) {
      console.error('Failed to fetch apps:', error)
    }
  }

  function setCurrentApp(appId: string) {
    currentAppId.value = appId
    localStorage.setItem('_tracely_current_app', appId)
  }

  return { currentAppId, apps, hasMultipleApps, fetchApps, setCurrentApp }
})
