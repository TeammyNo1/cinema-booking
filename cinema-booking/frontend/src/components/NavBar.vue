<template>
  <header class="navbar">
    <RouterLink to="/" class="logo">
      <span class="logo-icon">🎬</span>
      <span class="logo-text">CineBook</span>
    </RouterLink>

    <nav class="nav-links">
      <template v-if="auth.isLoggedIn">
        <RouterLink to="/showtimes" class="nav-link">{{ t.nav.movies }}</RouterLink>
        <RouterLink to="/my-bookings" class="nav-link">{{ t.nav.myTickets }}</RouterLink>
        <RouterLink v-if="auth.isAdmin" to="/admin" class="nav-link admin-link">{{ t.nav.admin }}</RouterLink>
      </template>

      <!-- Language Toggle -->
      <button class="lang-toggle" @click="toggleLang" :title="i18n.lang === 'th' ? 'Switch to English' : 'เปลี่ยนเป็นภาษาไทย'">
        <span class="lang-flag">{{ i18n.lang === 'th' ? '🇹🇭' : '🇬🇧' }}</span>
        <span class="lang-label">{{ i18n.lang === 'th' ? 'TH' : 'EN' }}</span>
      </button>

      <template v-if="auth.isLoggedIn">
        <div class="user-chip">
          <span class="user-avatar">{{ userInitial }}</span>
          <span class="user-email">{{ shortEmail }}</span>
          <span v-if="auth.isAdmin" class="role-badge">ADMIN</span>
        </div>
        <button class="btn-ghost btn-sm" @click="doLogout">{{ t.nav.logout }}</button>
      </template>
      <template v-else>
        <a :href="loginUrl" class="btn-primary">{{ t.nav.login }}</a>
      </template>
    </nav>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useI18nStore } from '@/stores/i18n'

const auth = useAuthStore()
const i18n = useI18nStore()
const router = useRouter()
const t = computed(() => i18n.t)
const loginUrl = computed(() => `${window.location.origin}/api/auth/google`)

const userInitial = computed(() => {
  const name = auth.user?.name || auth.user?.email || '?'
  return name.charAt(0).toUpperCase()
})
const shortEmail = computed(() => {
  const e = auth.user?.email || ''
  return e.length > 18 ? e.substring(0, 16) + '...' : e
})

function toggleLang() {
  i18n.setLang(i18n.lang === 'th' ? 'en' : 'th')
}
function doLogout() {
  auth.logout()
  router.push('/')
}
</script>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2rem;
  height: 64px;
  background: rgba(8,8,16,0.85);
  backdrop-filter: blur(16px);
  border-bottom: 1px solid var(--border);
  position: sticky;
  top: 0;
  z-index: 100;
}

.logo {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  text-decoration: none;
}
.logo-icon { font-size: 1.4rem; }
.logo-text {
  font-family: 'Syne', sans-serif;
  font-size: 1.35rem;
  font-weight: 800;
  background: linear-gradient(135deg, #f5a623, #f0c060);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.nav-link {
  padding: 0.4rem 0.85rem;
  border-radius: 8px;
  font-size: 0.88rem;
  font-weight: 500;
  color: var(--text-secondary);
  text-decoration: none;
  transition: all 0.15s;
}
.nav-link:hover { color: var(--text-primary); background: var(--bg-hover); }
.nav-link.router-link-active { color: var(--text-primary); background: rgba(255,255,255,0.07); }
.admin-link { color: var(--accent) !important; }
.admin-link:hover { background: var(--accent-dim) !important; }

.lang-toggle {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  background: var(--bg-card2);
  border: 1px solid var(--border2);
  color: var(--text-secondary);
  padding: 0.35rem 0.75rem;
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  margin: 0 0.25rem;
}
.lang-toggle:hover { background: var(--bg-hover); color: var(--text-primary); border-color: var(--accent); }
.lang-flag { font-size: 1rem; }

.user-chip {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--bg-card2);
  border: 1px solid var(--border);
  border-radius: 999px;
  padding: 0.3rem 0.75rem 0.3rem 0.3rem;
}
.user-avatar {
  width: 26px; height: 26px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), #e040fb);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.75rem; font-weight: 700; color: #080810;
}
.user-email { font-size: 0.8rem; color: var(--text-secondary); }
.role-badge {
  background: var(--accent-dim);
  color: var(--accent);
  font-size: 0.6rem;
  font-weight: 800;
  padding: 0.1rem 0.45rem;
  border-radius: 4px;
  letter-spacing: 0.08em;
}
.btn-sm { padding: 0.38rem 0.85rem; font-size: 0.82rem; }
</style>
