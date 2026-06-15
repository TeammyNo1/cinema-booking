<template>
  <div class="admin-layout">
    <aside class="admin-sidebar">
      <div class="sidebar-brand">
        <span>⚙️</span>
        <span>Admin Panel</span>
      </div>
      <nav class="sidebar-nav">
        <RouterLink to="/admin" exact-active-class="active" class="nav-item">
          <span class="nav-icon">📊</span>
          <span>{{ t.admin.dashboard }}</span>
        </RouterLink>
        <RouterLink to="/admin/showtimes" active-class="active" class="nav-item">
          <span class="nav-icon">🎬</span>
          <span>{{ i18n.lang==='th'?'จัดการรอบหนัง':'Showtimes' }}</span>
        </RouterLink>
        <RouterLink to="/admin/audit" active-class="active" class="nav-item">
          <span class="nav-icon">📋</span>
          <span>{{ t.admin.auditLogs }}</span>
        </RouterLink>
      </nav>
      <div class="sidebar-footer">
        <RouterLink to="/showtimes" class="nav-item back-site">
          <span class="nav-icon">🎭</span>
          <span>{{ t.nav.movies }}</span>
        </RouterLink>
      </div>
    </aside>
    <div class="admin-content">
      <RouterView />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18nStore } from '@/stores/i18n'
const i18n = useI18nStore()
const t = computed(() => i18n.t)
</script>

<style scoped>
.admin-layout { display:flex; min-height:calc(100vh - 64px); }
.admin-sidebar {
  width:220px; background:var(--bg-secondary);
  border-right:1px solid var(--border);
  display:flex; flex-direction:column; flex-shrink:0;
  padding:1.25rem 0.75rem; gap:0.2rem;
}
.sidebar-brand {
  display:flex; align-items:center; gap:0.5rem;
  font-size:0.7rem; font-weight:800; letter-spacing:0.12em;
  text-transform:uppercase; color:var(--text-muted);
  padding:0 0.5rem 1rem; border-bottom:1px solid var(--border); margin-bottom:0.5rem;
}
.sidebar-nav { display:flex; flex-direction:column; gap:0.2rem; flex:1; }
.nav-item {
  display:flex; align-items:center; gap:0.65rem;
  padding:0.65rem 0.85rem; border-radius:10px;
  font-size:0.86rem; font-weight:500; color:var(--text-secondary);
  text-decoration:none; transition:all 0.15s;
}
.nav-item:hover { background:var(--bg-hover); color:var(--text-primary); }
.nav-item.active { background:var(--accent-dim); color:var(--accent); border:1px solid rgba(245,166,35,0.2); }
.nav-icon { font-size:1rem; flex-shrink:0; }
.sidebar-footer { border-top:1px solid var(--border); padding-top:0.75rem; margin-top:0.5rem; }
.back-site { color:var(--text-muted) !important; font-size:0.82rem; }
.back-site:hover { color:var(--text-secondary) !important; }
.admin-content { flex:1; padding:2rem; overflow:auto; }
</style>
