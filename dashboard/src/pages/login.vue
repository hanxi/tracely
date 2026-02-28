<template>
  <div class="min-h-screen flex items-center justify-center bg-linear-to-br from-green-400 to-green-600 dark:from-gray-900 dark:to-gray-800 p-4">
    <UCard class="w-full max-w-md">
      <!-- 主题切换按钮 - 移到 UCard 右上角 -->
      <template #header>
        <div class="relative">
          <div class="absolute top-0 right-0">
            <ColorModeToggle />
          </div>
          <div class="text-center">
            <img src="/logo.svg" alt="Tracely Logo" class="w-12 h-12 mx-auto mb-4">
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
              Tracely
            </h1>
            <p class="text-gray-500 dark:text-gray-400 mt-1">
              错误监控系统
            </p>
          </div>
        </div>
      </template>

      <UForm :state="form" @submit="onSubmit">
        <UFormField label="用户名" name="username" class="w-full">
          <UInput
            v-model="form.username"
            icon="i-lucide-user"
            placeholder="请输入用户名"
            size="lg"
            class="w-full"
          />
        </UFormField>

        <UFormField label="密码" name="password" class="mt-4 w-full">
          <UInput
            v-model="form.password"
            type="password"
            icon="i-lucide-lock"
            placeholder="请输入密码"
            size="lg"
            class="w-full"
          />
        </UFormField>

        <UAlert
          v-if="error"
          color="error"
          variant="soft"
          class="mt-4"
          :title="error"
        />

        <UButton
          type="submit"
          color="success"
          size="lg"
          block
          :loading="loading"
          class="mt-6"
        >
          {{ loading ? '登录中...' : '登录' }}
        </UButton>
      </UForm>
    </UCard>

    <!-- 移除原来的固定位置主题切换按钮 -->
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { login } from '@/api/auth'
import ColorModeToggle from '@/components/ColorModeToggle.vue'

const router = useRouter()
const auth = useAuthStore()

const form = ref({ username: 'admin', password: '' })
const loading = ref(false)
const error = ref('')

async function onSubmit() {
  loading.value = true
  error.value = ''
  
  try {
    const res = await login(form.value.username, form.value.password)
    auth.setAuth(res.data.token, res.data.username)
    router.push('/')
  } catch (err: unknown) {
    if (err && typeof err === 'object' && 'response' in err) {
      const errorObj = err as { response?: { data?: { error?: string } } }
      error.value = errorObj.response?.data?.error || '登录失败'
    } else {
      error.value = '登录失败'
    }
  } finally {
    loading.value = false
  }
}
</script>