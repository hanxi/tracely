<template>
  <UDropdownMenu :items="[[{ label: '退出登录', icon: 'i-lucide-log-out', onSelect: logout }]]">
    <UButton variant="ghost" color="neutral">
      <UAvatar :alt="auth.username" size="sm" />
      <span class="ml-2">{{ auth.username }}</span>
    </UButton>
  </UDropdownMenu>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

function logout() {
  auth.logout()
  router.push('/login')
}
</script>