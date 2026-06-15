import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/composables/useApi'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('cinema_token') || null)
  const user = ref(JSON.parse(localStorage.getItem('cinema_user') || 'null'))

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'ADMIN')

  function setToken(t) {
    token.value = t
    localStorage.setItem('cinema_token', t)
  }

  function setUser(u) {
    user.value = u
    localStorage.setItem('cinema_user', JSON.stringify(u))
  }

  async function fetchMe() {
    if (!token.value) return
    try {
      const res = await api.get('/api/auth/me')
      setUser(res.data)
    } catch {
      logout()
    }
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('cinema_token')
    localStorage.removeItem('cinema_user')
  }

  return { token, user, isLoggedIn, isAdmin, setToken, setUser, fetchMe, logout }
})
