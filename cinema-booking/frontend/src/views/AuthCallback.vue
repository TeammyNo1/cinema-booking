<template>
  <div class="callback-page">
    <div class="callback-card">
      <div class="spinner-ring"></div>
      <p>Signing you in...</p>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

onMounted(async () => {
  const token = route.query.token
  if (token) {
    auth.setToken(token)
    await auth.fetchMe()
    router.replace('/showtimes')
  } else {
    router.replace('/')
  }
})
</script>

<style scoped>
.callback-page {
  display: flex; align-items: center; justify-content: center; min-height: 60vh;
}
.callback-card {
  display: flex; flex-direction: column; align-items: center; gap: 1rem;
  color: var(--text-secondary);
}
</style>
