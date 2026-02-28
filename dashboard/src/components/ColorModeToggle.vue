<script setup lang="ts">
import { computed } from 'vue'
import { useColorMode } from '@vueuse/core'

const colorMode = useColorMode<'light' | 'dark' | 'auto'>()

const mode = computed({
  get() {
    return colorMode.value
  },
  set(newMode: 'light' | 'dark' | 'auto') {
    colorMode.value = newMode
  }
})

const modeOptions = [
  { label: '浅色', value: 'light', icon: 'i-lucide-sun' },
  { label: '深色', value: 'dark', icon: 'i-lucide-moon' },
  { label: '系统', value: 'auto', icon: 'i-lucide-monitor' }
]

const currentMode = computed(() => {
  const option = modeOptions.find(opt => opt.value === (mode.value === 'auto' ? 'auto' : mode.value))
  return option || modeOptions[2]
})
</script>

<template>
  <UDropdownMenu
    :items="[modeOptions.map(opt => ({
      label: opt.label,
      icon: opt.icon,
      onSelect: () => mode = opt.value as 'light' | 'dark' | 'auto'
    }))]"
  >
    <UButton
      :icon="currentMode.icon"
      variant="ghost"
      size="sm"
      :aria-label="`当前主题：${currentMode.label}`"
      class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
    />
  </UDropdownMenu>
</template>
