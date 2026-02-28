<template>
  <div>
    <USelect
      v-if="store.hasMultipleApps"
      :model-value="selectedAppId"
      :items="appOptions"
      placeholder="选择应用"
      class="w-full"
      @update:model-value="onAppChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

// 使用 computed 自动响应 store.apps 的变化
const appOptions = computed(() => 
  store.apps.map(app => ({
    label: app.appName,
    value: app.appId
  }))
)

// 使用 computed 确保 selectedAppId 始终与 store.currentAppId 同步
const selectedAppId = computed({
  get: () => store.currentAppId,
  set: (value) => {
    store.setCurrentApp(value)
  }
})

function onAppChange(value: string) {
  // 确保已保存到 store
  store.setCurrentApp(value)
  // 等待下一个 tick 让状态更新完成，然后刷新页面
  setTimeout(() => {
    window.location.reload()
  }, 0)
}
</script>