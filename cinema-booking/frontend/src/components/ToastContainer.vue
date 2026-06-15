<template>
  <div class="toast-container">
    <TransitionGroup name="toast">
      <div
        v-for="toast in toastStore.toasts"
        :key="toast.id"
        :class="`toast toast-${toast.type}`"
        @click="toastStore.remove(toast.id)"
      >
        <span class="toast-icon">{{ icons[toast.type] }}</span>
        <span>{{ toast.message }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup>
import { useToastStore } from '@/stores/toast'
const toastStore = useToastStore()
const icons = { success: '✅', error: '❌', info: 'ℹ️' }
</script>

<style scoped>
.toast-container {
  position: fixed; bottom: 2rem; right: 2rem; z-index: 9999;
  display: flex; flex-direction: column; gap: 0.6rem; max-width: 360px;
}
.toast {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.9rem 1.25rem;
  border-radius: 12px;
  font-size: 0.88rem; font-weight: 500;
  box-shadow: 0 8px 32px rgba(0,0,0,0.5);
  border-left: 3px solid;
  backdrop-filter: blur(10px);
  cursor: pointer;
}
.toast-success { background: rgba(15,40,25,0.96); border-color: var(--success); color: #6ee89a; }
.toast-error   { background: rgba(40,15,15,0.96); border-color: var(--error);   color: #f08080; }
.toast-info    { background: rgba(15,25,40,0.96); border-color: var(--info);    color: #80b8f0; }
.toast-icon { font-size: 1rem; flex-shrink: 0; }

.toast-enter-active { transition: all 0.35s cubic-bezier(0.34,1.56,0.64,1); }
.toast-leave-active { transition: all 0.2s ease; }
.toast-enter-from { transform: translateX(120%); opacity: 0; }
.toast-leave-to   { transform: translateX(120%); opacity: 0; }
</style>
