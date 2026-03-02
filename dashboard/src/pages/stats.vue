<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">活跃统计</h1>
      <p class="text-gray-500 dark:text-gray-400">查看用户访问数据</p>
    </div>

    <!-- Days Selector -->
    <URadioGroup
      v-model="days"
      :options="dayOptions"
      class="mb-6"
      @update:model-value="loadData"
    />

    <!-- Daily Stats -->
    <UCard class="mb-6">
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">每日统计</h3>
      </template>

      <div v-if="stats.daily?.length" class="grid grid-cols-7 gap-4">
        <div 
          v-for="d in stats.daily" 
          :key="d.date"
          class="text-center p-4 bg-gray-50 dark:bg-gray-800 rounded-lg"
        >
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ d.date }}
          </p>
          <p class="text-lg font-bold text-success-600 dark:text-success-400">
            {{ d.pv }}
          </p>
          <p class="text-xs text-gray-400 dark:text-gray-500">
            PV
          </p>
        </div>
      </div>
      
      <UAlert
        v-else
        color="primary"
        variant="soft"
        title="暂无数据"
        icon="i-lucide-info"
      />
    </UCard>

    <!-- Top Pages -->
    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">热门页面</h3>
      </template>

      <UTable
        :data="stats.topPages"
        :columns="columns"
        :loading="loading"
      >
        <template #page-data="{ row }">
          <span class="truncate max-w-md block">{{ (row as unknown as TopPage).page }}</span>
        </template>
        
        <template #avgDuration-data="{ row }">
          {{ Math.floor((row as unknown as TopPage).avgDuration / 60) }}m {{ (row as unknown as TopPage).avgDuration % 60 }}s
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getStats } from '@/api/stats'
import { useAppStore } from '@/stores/app'
import type { DailyStats, TopPage } from '@/api/stats'

const appStore = useAppStore()
const loading = ref(false)
const days = ref(7)
const stats = ref({ daily: [] as DailyStats[], topPages: [] as TopPage[] })

const dayOptions = [
  { label: '7 天', value: 7 },
  { label: '14 天', value: 14 },
  { label: '30 天', value: 30 },
]

const columns = [
  { accessorKey: 'page', header: '页面路径' },
  { accessorKey: 'pv', header: 'PV' },
  { accessorKey: 'avgDuration', header: '平均停留' }
]

async function loadData() {
  loading.value = true
  try {
    const appId = appStore.currentAppId
    const res = await getStats({ 
      days: days.value,
      appID: appId
    })
    stats.value = {
      daily: res.data.daily || [],
      topPages: res.data.topPages || []
    }
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>