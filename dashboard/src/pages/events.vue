<script setup lang="ts">
import { ref, onMounted, computed, watch, h, resolveComponent } from 'vue'
import type { TableColumn } from '@nuxt/ui'
import { useAppStore } from '@/stores/app'
import { getEventStats, getTopEvents, getDailyEvents, getEventList, getEventStatsSummary } from '@/api/events'
import type { EventDetail, EventStatsSummaryResponse } from '@/api/events'

const appStore = useAppStore()

// 当前选择的应用 ID
const currentAppID = computed(() => appStore.currentAppId)

// 统计天数选项
const daysOptions = [
  { label: '7 天', value: 7 },
  { label: '14 天', value: 14 },
  { label: '30 天', value: 30 },
]
const selectedDayOption = ref<{ label: string; value: number } | undefined>(daysOptions[0])

// 事件类型筛选
const selectedEventOption = ref<{ label: string; value: string } | undefined>()
const eventNameOptions = ref<Array<{ label: string; value: string }>>([
  { label: '全部事件', value: '' },
])

// 数据状态
const loading = ref(false)
const eventStats = ref<Array<{ eventName: string; count: number }>>([])
const topEvents = ref<Array<{ eventName: string; count: number }>>([])
const dailyEvents = ref<Array<{ date: string; count: number }>>([])

// 事件详情弹窗状态
const showDetailModal = ref(false)
const currentEventName = ref('')
const detailLoading = ref(false)
const eventList = ref<EventDetail[]>([])
const eventTotal = ref(0)
const eventPage = ref(1)
const eventPageSize = ref(10)
const eventStatsSummary = ref<EventStatsSummaryResponse | null>(null)

// 获取当前选中的天数
const getSelectedDays = () => selectedDayOption.value?.value ?? 7

// 获取当前选中的事件名称
const getSelectedEventName = () => selectedEventOption.value?.value ?? ''

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    // 并行加载所有数据
    const [statsRes, topRes, dailyRes] = await Promise.all([
      getEventStats(getSelectedDays(), currentAppID.value, getSelectedEventName()),
      getTopEvents(getSelectedDays(), currentAppID.value, 10),
      getDailyEvents(getSelectedDays(), currentAppID.value, getSelectedEventName()),
    ])

    eventStats.value = statsRes.stats
    topEvents.value = topRes.events
    dailyEvents.value = dailyRes.daily

    // 更新事件类型选项（用于筛选）
    if (getSelectedEventName() === '') {
      const eventNames = statsRes.stats.map((s) => ({
        label: s.eventName,
        value: s.eventName,
      }))
      eventNameOptions.value = [{ label: '全部事件', value: '' }, ...eventNames]
    }
  } catch (error) {
    console.error('Failed to load event data:', error)
  } finally {
    loading.value = false
  }
}

// 查看事件详情
const handleViewDetail = async (eventName: string) => {
  currentEventName.value = eventName
  showDetailModal.value = true
  eventPage.value = 1
  await loadEventDetail()
}

// 加载事件详情数据
const loadEventDetail = async () => {
  detailLoading.value = true
  try {
    const [listRes, summaryRes] = await Promise.all([
      getEventList(currentEventName.value, eventPage.value, eventPageSize.value, currentAppID.value),
      getEventStatsSummary(currentEventName.value, getSelectedDays(), currentAppID.value),
    ])
    eventList.value = listRes.list
    eventTotal.value = listRes.total
    eventStatsSummary.value = summaryRes
  } catch (error) {
    console.error('Failed to load event detail:', error)
  } finally {
    detailLoading.value = false
  }
}

// 事件详情表格列定义
const detailColumns: TableColumn<EventDetail>[] = [
  {
    accessorKey: 'userId',
    header: '用户 ID',
    cell: ({ row }) => h('span', { class: 'font-mono text-sm' }, row.original.userId)
  },
  {
    accessorKey: 'metadataKeys',
    header: 'Metadata 字段',
    cell: ({ row }) => {
      const metadata = row.original.metadata
      const keys = metadata && typeof metadata === 'object' ? Object.keys(metadata) : []
      const keyText = keys.length > 0 ? keys.join(', ') : '无'
      return h('span', { class: 'text-sm text-gray-600 dark:text-gray-400 truncate max-w-md block' }, keyText)
    }
  },
  {
    accessorKey: 'createdAt',
    header: '时间',
    cell: ({ row }) => h('span', { class: 'text-sm' }, new Date(row.original.createdAt).toLocaleString('zh-CN'))
  },
  {
    id: 'metadata',
    header: '详情',
    cell: ({ row }) => {
      const hasMetadata = row.original.metadata && Object.keys(row.original.metadata).length > 0
      return h(resolveComponent('UButton'), {
        label: hasMetadata ? '查看' : '-',
        color: 'neutral',
        variant: 'ghost',
        size: 'sm',
        disabled: !hasMetadata,
        onClick: () => hasMetadata && handleViewMetadata(row.original)
      })
    }
  }
]

// Metadata 详情弹窗
const showMetadataModal = ref(false)
const currentMetadata = ref<Record<string, any> | null>(null)

// 查看 Metadata
const handleViewMetadata = (event: EventDetail) => {
  currentMetadata.value = event.metadata
  showMetadataModal.value = true
}

// 监听筛选条件变化，自动加载数据
watch([selectedDayOption, selectedEventOption], () => {
  loadData()
})

// 初始化
onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">事件统计</h1>
      
      <!-- 筛选器 -->
      <div class="flex items-center gap-4">
        <!-- 天数选择 -->
        <USelectMenu
          v-model="selectedDayOption"
          :items="daysOptions"
          placeholder="选择天数"
        />

        <!-- 事件类型筛选 -->
        <USelectMenu
          v-model="selectedEventOption"
          :items="eventNameOptions"
          placeholder="选择事件类型"
        />
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="flex justify-center py-12">
      <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 animate-spin" />
    </div>

    <!-- 数据展示 -->
    <div v-else class="space-y-6">
      <!-- 事件类型分布卡片 -->
      <UCard>
        <template #header>
          <h2 class="text-lg font-semibold">事件类型分布</h2>
        </template>

        <div v-if="eventStats.length === 0" class="text-center py-8 text-gray-500">
          暂无数据
        </div>

        <div v-else class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          <div
            v-for="stat in eventStats"
            :key="stat.eventName"
            class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-primary-500 dark:hover:border-primary-500 cursor-pointer transition-colors"
            @click="handleViewDetail(stat.eventName)"
          >
            <div class="text-sm text-gray-600 dark:text-gray-400 mb-1">
              {{ stat.eventName }}
            </div>
            <div class="flex items-center justify-between">
              <div class="text-2xl font-bold">{{ stat.count.toLocaleString() }}</div>
              <UIcon name="i-lucide-arrow-right" class="w-5 h-5 text-gray-400" />
            </div>
          </div>
        </div>
      </UCard>

      <!-- 每日事件趋势 -->
      <UCard>
        <template #header>
          <h2 class="text-lg font-semibold">每日事件趋势</h2>
        </template>

        <div v-if="dailyEvents.length === 0" class="text-center py-8 text-gray-500">
          暂无数据
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-3 px-4 font-semibold">日期</th>
                <th class="text-right py-3 px-4 font-semibold">事件数</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="daily in dailyEvents"
                :key="daily.date"
                class="border-b border-gray-100 dark:border-gray-800"
              >
                <td class="py-3 px-4">{{ daily.date }}</td>
                <td class="py-3 px-4 text-right font-mono">
                  {{ daily.count.toLocaleString() }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </UCard>

      <!-- Top 事件排行 -->
      <UCard>
        <template #header>
          <h2 class="text-lg font-semibold">Top 10 事件排行</h2>
        </template>

        <div v-if="topEvents.length === 0" class="text-center py-8 text-gray-500">
          暂无数据
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="(event, index) in topEvents"
            :key="event.eventName"
            class="flex items-center justify-between p-3 rounded-lg bg-gray-50 dark:bg-gray-800"
          >
            <div class="flex items-center gap-3">
              <div
                class="w-8 h-8 rounded-full flex items-center justify-center font-bold"
                :class="{
                  'bg-yellow-500 text-white': index === 0,
                  'bg-gray-400 text-white': index === 1,
                  'bg-orange-500 text-white': index === 2,
                  'bg-gray-200 dark:bg-gray-700': index > 2,
                }"
              >
                {{ index + 1 }}
              </div>
              <div class="font-medium">{{ event.eventName }}</div>
            </div>
            <div class="flex items-center gap-3">
              <div class="text-lg font-bold">{{ event.count.toLocaleString() }}</div>
              <UButton
                label="详情"
                color="neutral"
                variant="ghost"
                size="sm"
                @click="handleViewDetail(event.eventName)"
              />
            </div>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 事件详情弹窗 -->
    <UModal
    fullscreen
      v-model:open="showDetailModal"
      :overlay="true"
      :modal="true"
      :title="`事件详情 - ${currentEventName}`"
      :close="{ color: 'neutral', variant: 'ghost' }"
    >
      <template #body>
        <div v-if="detailLoading" class="flex justify-center py-12">
          <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 animate-spin" />
        </div>

        <div v-show="!detailLoading" class="space-y-6">
          <!-- 统计摘要 -->
          <div v-if="eventStatsSummary" class="grid grid-cols-3 gap-4">
            <div class="p-4 rounded-lg bg-gray-50 dark:bg-gray-800">
              <div class="text-sm text-gray-600 dark:text-gray-400 mb-1">总次数</div>
              <div class="text-2xl font-bold">{{ eventStatsSummary.totalCount.toLocaleString() }}</div>
            </div>
            <div class="p-4 rounded-lg bg-gray-50 dark:bg-gray-800">
              <div class="text-sm text-gray-600 dark:text-gray-400 mb-1">今日次数</div>
              <div class="text-2xl font-bold">{{ eventStatsSummary.todayCount.toLocaleString() }}</div>
            </div>
            <div class="p-4 rounded-lg bg-gray-50 dark:bg-gray-800">
              <div class="text-sm text-gray-600 dark:text-gray-400 mb-1">独立用户</div>
              <div class="text-2xl font-bold">{{ eventStatsSummary.uv.toLocaleString() }}</div>
            </div>
          </div>

          <!-- 事件列表 -->
          <div>
            <h3 class="text-lg font-semibold mb-4">事件列表</h3>
            <UTable
              :data="eventList"
              :columns="detailColumns"
              :loading="detailLoading"
            />

            <!-- 分页 -->
            <div v-if="eventTotal > eventPageSize" class="mt-4 flex justify-center">
              <UPagination
                :page="eventPage"
                :total="eventTotal"
                :items-per-page="eventPageSize"
                @update:page="(val) => { eventPage = val; loadEventDetail() }"
              />
            </div>

            <UAlert
              v-if="!detailLoading && eventList.length === 0"
              color="neutral"
              variant="soft"
              title="暂无数据"
              class="mt-4"
            />
          </div>
        </div>
      </template>
    </UModal>

    <!-- Metadata 详情弹窗 -->
    <UModal
      v-model:open="showMetadataModal"
      :overlay="true"
      :modal="true"
      title="Metadata 详情"
      :close="{ color: 'neutral', variant: 'ghost' }"
    >
      <template #body>
        <div v-if="currentMetadata">
          <pre class="text-xs bg-gray-900 text-gray-100 dark:bg-gray-800 dark:text-gray-200 p-4 rounded overflow-auto max-h-96">{{ JSON.stringify(currentMetadata, null, 2) }}</pre>
        </div>
      </template>
    </UModal>
  </div>
</template>
