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
        // 验证 currentAppId 是否有效
        const isValid = appList.some((app: AppInfo) => app.appId === currentAppId.value)
        
        // 如果当前 appId 无效或为空，使用第一个应用
        if (!isValid || !currentAppId.value) {
          currentAppId.value = appList[0].appId
        }
        // 如果 currentAppId 有效，保持不变（Pinia 持久化插件会自动恢复之前的值）
      }
    } catch (error) {
      console.error('Failed to fetch apps:', error)
    }
  }

  function setCurrentApp(appId: string) {
    currentAppId.value = appId
    // Pinia 持久化插件会自动保存到 localStorage，不需要手动操作
  }

  return { currentAppId, apps, hasMultipleApps, fetchApps, setCurrentApp }
}, {
  persist: {
    key: 'app-store',
    storage: localStorage,
    pick: ['currentAppId'],
  },
})
