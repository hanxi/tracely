import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref('')
  const username = ref('')
  const isLoggedIn = computed(() => !!token.value)

  function setAuth(newToken: string, newUsername: string) {
    token.value = newToken
    username.value = newUsername
    // Pinia 持久化插件会自动保存，不需要手动操作 localStorage
  }

  function logout() {
    token.value = ''
    username.value = ''
    // Pinia 持久化插件会自动清除，不需要手动操作 localStorage
  }

  return { token, username, isLoggedIn, setAuth, logout }
}, {
  persist: {
    key: 'auth-store',
    storage: localStorage,
    pick: ['token', 'username'],
  },
})
