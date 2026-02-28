<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import AppSwitcher from '@/components/AppSwitcher.vue'
import ColorModeToggle from '@/components/ColorModeToggle.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const app = useAppStore()
const isCollapsed = ref(false)
// 移动端侧边栏状态
const isMobileMenuOpen = ref(false)
const isMobile = ref(false)

const links = [
  { label: '概览', icon: 'i-lucide-layout-dashboard', to: '/' },
  { label: '错误列表', icon: 'i-lucide-bug', to: '/errors' },
  { label: '活跃统计', icon: 'i-lucide-bar-chart', to: '/stats' },
]

function handleLogout() {
  auth.logout()
  router.push('/login')
}

// 检测是否为移动端
const checkMobile = () => {
  isMobile.value = window.innerWidth < 768
  if (isMobile.value) {
    isCollapsed.value = false
  }
}

onMounted(async () => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  
  // 只在已登录时获取应用列表
  if (auth.isLoggedIn) {
    await app.fetchApps()
  }
})

// 关闭移动端菜单
const closeMobileMenu = () => {
  isMobileMenuOpen.value = false
}
</script>

<template>
  <div class="flex h-screen bg-gray-50 dark:bg-gray-900">
    <!-- 移动端遮罩层 -->
    <div 
      v-if="isMobileMenuOpen" 
      class="fixed inset-0 bg-black/50 z-40 md:hidden"
      @click="closeMobileMenu"
    />
    
    <!-- Sidebar -->
    <div 
      class="bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col transition-all duration-300 fixed md:relative h-full z-50 md:z-auto"
      :class="[
        isMobile ? (isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full') : (isCollapsed ? 'w-16' : 'w-64'),
        isMobile ? 'w-64' : ''
      ]"
    >
      <!-- Logo & App Switcher -->
      <div class="border-b border-gray-200 dark:border-gray-700 px-4 py-3">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center">
            <img src="/logo.svg" alt="Tracely" class="w-8 h-8">
            <span v-if="!isCollapsed || isMobile" class="ml-2 text-xl font-bold text-green-500">Tracely</span>
          </div>
          <!-- 移动端关闭按钮 -->
          <button 
            v-if="isMobile" 
            @click="closeMobileMenu"
            class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            <UIcon name="i-lucide-x" class="w-5 h-5 text-gray-500" />
          </button>
        </div>
        <AppSwitcher v-if="!isCollapsed || isMobile" />
      </div>

      <!-- Navigation -->
      <nav class="flex-1 py-4 px-2 space-y-1">
        <RouterLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="flex items-center gap-3 px-3 py-2 rounded-lg transition-colors"
          :class="route.path === link.to ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700'"
          @click="isMobile && closeMobileMenu()"
        >
          <UIcon :name="link.icon" class="w-5 h-5" />
          <span v-if="!isCollapsed || isMobile">{{ link.label }}</span>
        </RouterLink>
      </nav>

      <!-- Footer -->
      <div class="border-t border-gray-200 dark:border-gray-700 p-4">
        <div class="flex items-center justify-between w-full">
          <!-- 左边：主题切换按钮 -->
          <ColorModeToggle />
          
          <!-- 右边：帐号信息 -->
          <UDropdownMenu
            :items="[[{ label: '退出登录', icon: 'i-lucide-log-out', onSelect: handleLogout }]]"
          >
            <button class="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
              <UAvatar :alt="auth.username" size="sm" />
              <span v-if="!isCollapsed || isMobile" class="text-sm text-gray-700 dark:text-gray-200">{{ auth.username }}</span>
              <UIcon v-if="!isCollapsed || isMobile" name="i-lucide-chevron-down" class="w-4 h-4 text-gray-400" />
            </button>
          </UDropdownMenu>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <main class="flex-1 overflow-auto p-4 md:p-6">
      <!-- 移动端汉堡菜单按钮 -->
      <div class="md:hidden mb-4">
        <button 
          @click="isMobileMenuOpen = true"
          class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
        >
          <UIcon name="i-lucide-menu" class="w-6 h-6 text-gray-600 dark:text-gray-300" />
        </button>
      </div>
      
      <RouterView />
    </main>
  </div>
</template>