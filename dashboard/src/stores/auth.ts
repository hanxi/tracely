import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref('')
  const username = ref('')
  const isLoggedIn = computed(() => !!token.value)

  function setAuth(newToken: string, newUsername: string) {
    token.value = newToken
    username.value = newUsername
    localStorage.setItem('_tracely_token', newToken)
    localStorage.setItem('_tracely_user', newUsername)
  }

  function logout() {
    token.value = ''
    username.value = ''
    localStorage.removeItem('_tracely_token')
    localStorage.removeItem('_tracely_user')
  }

  return { token, username, isLoggedIn, setAuth, logout }
}, { persist: true })
