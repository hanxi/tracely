<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900">错误列表</h1>
      <p class="text-gray-500">查看和管理应用错误</p>
    </div>

    <!-- Filters -->
    <div class="flex gap-4 mb-6">
      <USelect
        v-model="filters.type"
        :items="typeOptions"
        placeholder="筛选类型"
        class="w-40"
        @update:model-value="loadData"
      />
    </div>

    <!-- Table -->
    <UCard>
      <UTable
        :data="errors"
        :columns="columns"
        :loading="loading"
      />

      <!-- Pagination -->
      <div class="mt-4 flex justify-center">
        <UPagination
          v-model="page"
          :total="total"
          :page-count="pageSize"
          @update:model-value="loadData"
        />
      </div>
    </UCard>

    <!-- Detail Modal -->
    <UModal
      v-model:open="showDetail"
      :overlay="true"
      :modal="true"
      title="错误详情"
      description="查看错误的详细信息"
      :close="{ color: 'neutral', variant: 'ghost' }"
    >
      <template #body>
        <div v-if="currentError" class="space-y-4">
          <div>
            <label class="text-sm text-gray-500">类型</label>
            <UBadge :color="currentError.Type === 'jsError' ? 'error' : 'primary'">
              {{ currentError.Type }}
            </UBadge>
          </div>
          
          <div>
            <label class="text-sm text-gray-500">消息</label>
            <p class="text-gray-900 bg-gray-50 p-3 rounded">
              {{ currentError.Message }}
            </p>
          </div>
          
          <div>
            <label class="text-sm text-gray-500">堆栈</label>
            <pre class="text-xs bg-gray-900 text-gray-100 p-3 rounded overflow-auto max-h-40">{{ currentError.Stack }}</pre>
          </div>
          
          <div>
            <label class="text-sm text-gray-500">URL</label>
            <p class="text-gray-900">{{ currentError.URL }}</p>
          </div>
          
          <div>
            <label class="text-sm text-gray-500">首次出现</label>
            <p class="text-gray-900">{{ currentError.FirstSeen }}</p>
          </div>
          
          <div>
            <label class="text-sm text-gray-500">最近出现</label>
            <p class="text-gray-900">{{ currentError.LastSeen }}</p>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h, resolveComponent } from 'vue'
import { getErrorList } from '@/api/error'
import type { ErrorLog } from '@/api/error'
import type { TableColumn } from '@nuxt/ui'

const loading = ref(false)
const errors = ref<ErrorLog[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const showDetail = ref(false)
const currentError = ref<ErrorLog | null>(null)

const filters = ref({ type: 'all', sortBy: 'count' })

const typeOptions = [
  { label: '全部类型', value: 'all' },
  { label: 'JS Error', value: 'jsError' },
  { label: 'Promise Error', value: 'promiseError' },
  { label: 'Vue Error', value: 'vueError' },
]

const columns: TableColumn<ErrorLog>[] = [
  { 
    accessorKey: 'Type', 
    header: '类型',
    cell: ({ row }) => h(resolveComponent('UBadge'), {
      color: row.original.Type === 'jsError' ? 'error' : 'primary',
      variant: 'subtle'
    }, () => row.original.Type)
  },
  { 
    accessorKey: 'Message', 
    header: '错误信息',
    cell: ({ row }) => h('span', {
      class: 'truncate max-w-xs block'
    }, row.original.Message)
  },
  { 
    accessorKey: 'Count', 
    header: '次数',
    cell: ({ row }) => h(resolveComponent('UBadge'), {
      color: 'neutral',
      variant: 'solid'
    }, () => row.original.Count.toString())
  },
  { 
    accessorKey: 'LastSeen', 
    header: '最近出现',
    cell: ({ row }) => h('span', {}, row.original.LastSeen)
  },
  {
    id: 'actions',
    header: '操作',
    cell: ({ row }) => h(resolveComponent('UButton'), {
      label: '详情',
      color: 'neutral',
      variant: 'ghost',
      size: 'sm',
      onClick: () => handleViewDetail(row)
    })
  }
]

// 查看详情
function handleViewDetail(row: { original: ErrorLog }) {
  currentError.value = row.original
  showDetail.value = true
}

async function loadData() {
  loading.value = true
  try {
    const res = await getErrorList({ 
      page: page.value, 
      pageSize: pageSize.value,
      type: filters.value.type === 'all' ? undefined : filters.value.type
    })
    errors.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>