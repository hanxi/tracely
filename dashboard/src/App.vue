<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import DefaultLayout from '@/layouts/default.vue'

const route = useRoute()
const showLayout = computed(() => route.path !== '/login')
</script>

<template>
  <UApp :disable-transitions="true">
    <DefaultLayout v-if="showLayout">
      <RouterView />
    </DefaultLayout>
    <RouterView v-else />
  </UApp>
</template>