<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900">概览</h1>
      <p class="text-gray-500">实时监控应用状态</p>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <UCard v-for="stat in stats" :key="stat.label">
        <div class="flex items-center gap-4">
          <div 
            class="w-12 h-12 rounded-lg flex items-center justify-center"
            :class="`bg-${stat.color}-100 text-${stat.color}-600`"
          >
            <UIcon :name="stat.icon" class="w-6 h-6" />
          </div>
          <div>
            <p class="text-sm text-gray-500">
              {{ stat.label }}
            </p>
            <p class="text-2xl font-bold text-gray-900">
              {{ stat.value().toLocaleString() }}
            </p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- Top Errors Table -->
    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">Top 5 错误</h3>
          <RouterLink to="/errors" class="text-green-600 hover:text-green-700 flex items-center gap-1">
            查看全部
            <UIcon name="i-lucide-arrow-right" class="w-4 h-4" />
          </RouterLink>
        </div>
      </template>

      <UTable
        :data="overview.topErrors"
        :columns="columns"
        :loading="loading"
      >
        <template #type-data="{ row }">
          <UBadge :color="(row as unknown as TopError).type === 'jsError' ? 'error' : 'primary'" variant="subtle">
            {{ (row as unknown as TopError).type }}
          </UBadge>
        </template>
      </UTable>

      <UAlert
        v-if="!loading && overview.topErrors?.length === 0"
        color="success"
        variant="soft"
        title="暂无错误记录"
        icon="i-lucide-check-circle"
        class="mt-4"
      />
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h, resolveComponent } from 'vue'
import type { TableColumn } from '@nuxt/ui'
import type { TopError } from '@/api/overview'
import { getOverview } from '@/api/overview'

const loading = ref(false)

const overview = ref({
  todayPV: 0,
  todayUV: 0,
  totalErrors: 0,
  todayErrors: 0,
  topErrors: [] as TopError[]
})

const stats = [
  { label: '今日 PV', value: () => overview.value.todayPV, icon: 'i-lucide-eye', color: 'blue' },
  { label: '今日 UV', value: () => overview.value.todayUV, icon: 'i-lucide-users', color: 'green' },
  { label: '错误总数', value: () => overview.value.totalErrors, icon: 'i-lucide-alert-triangle', color: 'red' },
  { label: '今日新增', value: () => overview.value.todayErrors, icon: 'i-lucide-trending-up', color: 'orange' }
]

const columns: TableColumn<TopError>[] = [
  {
    accessorKey: 'type',
    header: '类型',
    cell: ({ row }) => h(resolveComponent('UBadge'), {
      color: row.original.type === 'jsError' ? 'error' : 'primary',
      variant: 'subtle'
    }, () => row.original.type)
  },
  {
    accessorKey: 'message',
    header: '错误信息'
  },
  {
    accessorKey: 'count',
    header: '次数',
    meta: {
      class: {
        th: 'text-right',
        td: 'text-right font-medium'
      }
    }
  }
]

onMounted(async () => {
  loading.value = true
  try {
    const res = await getOverview()
    overview.value = res.data
  } finally {
    loading.value = false
  }
})
</script>