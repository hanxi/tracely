<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">应用统计</h1>
        <p class="text-gray-500 dark:text-gray-400">安装与升级数据概览</p>
      </div>
      <div>
        <USelectMenu
          v-model="selectedDayOption"
          :items="dayOptions"
          value-key="value"
          class="w-32"
        />
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <UIcon name="i-lucide-loader-2" class="w-8 h-8 animate-spin text-primary" />
    </div>

    <template v-else>
      <!-- Summary Cards -->
      <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
        <UCard v-for="stat in summaryStats" :key="stat.label">
          <div class="flex items-center gap-4">
            <div
              class="w-12 h-12 rounded-lg flex items-center justify-center"
              :class="stat.bgClass"
            >
              <UIcon :name="stat.icon" class="w-6 h-6" />
            </div>
            <div>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ stat.label }}</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">
                {{ stat.value.toLocaleString() }}
              </p>
            </div>
          </div>
        </UCard>
      </div>

      <!-- Daily Trends -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Daily Installs -->
        <UCard>
          <template #header>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">每日安装趋势</h3>
          </template>
          <div v-if="data.dailyInstalls.length === 0" class="text-center py-8 text-gray-500">暂无数据</div>
          <table v-else class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-2 text-gray-500 dark:text-gray-400">日期</th>
                <th class="text-right py-2 text-gray-500 dark:text-gray-400">次数</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in data.dailyInstalls" :key="item.date" class="border-b border-gray-100 dark:border-gray-800">
                <td class="py-2 text-gray-700 dark:text-gray-300">{{ item.date }}</td>
                <td class="py-2 text-right font-medium text-gray-900 dark:text-white">{{ item.count.toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </UCard>

        <!-- Daily Upgrades -->
        <UCard>
          <template #header>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">每日升级趋势</h3>
          </template>
          <div v-if="data.dailyUpgrades.length === 0" class="text-center py-8 text-gray-500">暂无数据</div>
          <table v-else class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-2 text-gray-500 dark:text-gray-400">日期</th>
                <th class="text-right py-2 text-gray-500 dark:text-gray-400">次数</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in data.dailyUpgrades" :key="item.date" class="border-b border-gray-100 dark:border-gray-800">
                <td class="py-2 text-gray-700 dark:text-gray-300">{{ item.date }}</td>
                <td class="py-2 text-right font-medium text-gray-900 dark:text-white">{{ item.count.toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </UCard>
      </div>

      <!-- Distribution -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Version Distribution -->
        <UCard>
          <template #header>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">版本分布</h3>
          </template>
          <div v-if="data.versionDist.length === 0" class="text-center py-8 text-gray-500">暂无数据</div>
          <div v-else class="space-y-3">
            <div
              v-for="(item, index) in data.versionDist"
              :key="item.value"
              class="flex items-center gap-3"
            >
              <div
                class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0"
                :class="rankClass(index)"
              >
                {{ index + 1 }}
              </div>
              <div class="flex-1 min-w-0">
                <span class="font-mono text-sm text-gray-900 dark:text-white">{{ item.value }}</span>
              </div>
              <span class="font-medium text-gray-700 dark:text-gray-300">{{ item.count.toLocaleString() }}</span>
            </div>
          </div>
        </UCard>

        <!-- Platform Distribution -->
        <UCard>
          <template #header>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">平台分布</h3>
          </template>
          <div v-if="data.platformDist.length === 0" class="text-center py-8 text-gray-500">暂无数据</div>
          <div v-else class="space-y-3">
            <div
              v-for="(item, index) in data.platformDist"
              :key="item.value"
              class="flex items-center gap-3"
            >
              <div
                class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0"
                :class="rankClass(index)"
              >
                {{ index + 1 }}
              </div>
              <div class="flex-1 min-w-0">
                <span class="text-sm text-gray-900 dark:text-white">{{ item.value }}</span>
              </div>
              <span class="font-medium text-gray-700 dark:text-gray-300">{{ item.count.toLocaleString() }}</span>
            </div>
          </div>
        </UCard>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { getAppStats } from '@/api/app-stats'
import type { AppStatsResponse } from '@/api/app-stats'

const appStore = useAppStore()
const currentAppID = computed(() => appStore.currentAppId)

const loading = ref(false)

const dayOptions = [
  { label: '近 7 天', value: 7 },
  { label: '近 14 天', value: 14 },
  { label: '近 30 天', value: 30 }
]

const selectedDayOption = ref(dayOptions[0])

const data = ref<AppStatsResponse>({
  installTotal: 0,
  installToday: 0,
  installUV: 0,
  upgradeTotal: 0,
  upgradeToday: 0,
  upgradeUV: 0,
  dailyInstalls: [],
  dailyUpgrades: [],
  versionDist: [],
  platformDist: []
})

const colorMap: Record<string, { bg: string; text: string }> = {
  info: { bg: 'bg-blue-100 dark:bg-blue-900', text: 'text-blue-600 dark:text-blue-400' },
  violet: { bg: 'bg-violet-100 dark:bg-violet-900', text: 'text-violet-600 dark:text-violet-400' },
  success: { bg: 'bg-success-100 dark:bg-success-900', text: 'text-success-600 dark:text-success-400' },
  primary: { bg: 'bg-primary-100 dark:bg-primary-900', text: 'text-primary-600 dark:text-primary-400' },
  warning: { bg: 'bg-warning-100 dark:bg-warning-900', text: 'text-warning-600 dark:text-warning-400' },
  error: { bg: 'bg-error-100 dark:bg-error-900', text: 'text-error-600 dark:text-error-400' }
}

const summaryStats = computed(() => [
  { label: '安装总数', value: data.value.installTotal, icon: 'i-lucide-download', color: 'info' },
  { label: '今日安装', value: data.value.installToday, icon: 'i-lucide-calendar', color: 'primary' },
  { label: '安装用户数', value: data.value.installUV, icon: 'i-lucide-users', color: 'success' },
  { label: '升级总数', value: data.value.upgradeTotal, icon: 'i-lucide-arrow-up-circle', color: 'violet' },
  { label: '今日升级', value: data.value.upgradeToday, icon: 'i-lucide-calendar-check', color: 'warning' },
  { label: '升级用户数', value: data.value.upgradeUV, icon: 'i-lucide-user-check', color: 'error' }
].map(stat => ({
  ...stat,
  bgClass: `${colorMap[stat.color].bg} ${colorMap[stat.color].text}`
})))

function rankClass(index: number): string {
  if (index === 0) return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
  if (index === 1) return 'bg-gray-200 text-gray-600 dark:bg-gray-700 dark:text-gray-300'
  if (index === 2) return 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400'
  return 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getAppStats(selectedDayOption.value.value)
    data.value = res.data
  } finally {
    loading.value = false
  }
}

watch([selectedDayOption, currentAppID], () => {
  loadData()
})

onMounted(() => {
  loadData()
})
</script>
