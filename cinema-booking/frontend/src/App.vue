<template>
  <div id="cinema-app">
    <NavBar />
    <main class="main-content">
      <RouterView />
    </main>
    <ToastContainer />
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useI18nStore } from '@/stores/i18n'
import NavBar from '@/components/NavBar.vue'
import ToastContainer from '@/components/ToastContainer.vue'

const auth = useAuthStore()
useI18nStore() // init
onMounted(() => auth.fetchMe())
</script>

<style scoped>
#cinema-app { min-height: 100vh; display: flex; flex-direction: column; }
.main-content { flex: 1; }
</style>
