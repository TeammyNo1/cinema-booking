<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">{{ t.admin.auditLogs }}</h1>
      <button class="btn-secondary" @click="loadLogs" style="font-size:0.82rem">🔄 Refresh</button>
    </div>

    <div class="filter-row">
      <div class="filter-group">
        <label>{{ t.admin.event }}</label>
        <select v-model="filters.event_type" @change="page=1; loadLogs()">
          <option value="">{{ i18n.lang === 'th' ? 'ทุกเหตุการณ์' : 'All Events' }}</option>
          <option value="BOOKING_SUCCESS">BOOKING_SUCCESS</option>
          <option value="BOOKING_TIMEOUT">BOOKING_TIMEOUT</option>
          <option value="SEAT_RELEASED">SEAT_RELEASED</option>
          <option value="SYSTEM_ERROR">SYSTEM_ERROR</option>
          <option value="BOOKING_LOCKED">BOOKING_LOCKED</option>
        </select>
      </div>
      <button class="btn-ghost" style="align-self:flex-end" @click="filters.event_type=''; page=1; loadLogs()">✕ {{ t.admin.clearFilters }}</button>
    </div>

    <div class="table-card">
      <div v-if="loading" class="loading-spinner"><div class="spinner-ring"></div></div>
      <table v-else>
        <thead>
          <tr>
            <th>{{ t.admin.event }}</th>
            <th>{{ t.admin.seat }}</th>
            <th>{{ t.admin.userId }}</th>
            <th>{{ t.admin.bookingId }}</th>
            <th>{{ t.admin.created }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id">
            <td><span :class="eventBadge(log.event_type)">{{ log.event_type }}</span></td>
            <td><span class="seat-badge">{{ log.seat_code || '—' }}</span></td>
            <td><span class="mono muted">{{ shortId(log.user_id) }}</span></td>
            <td><span class="mono muted">{{ shortId(log.booking_id) }}</span></td>
            <td>{{ formatDate(log.created_at) }}</td>
          </tr>
          <tr v-if="logs.length === 0">
            <td colspan="5" class="empty-row">{{ t.admin.noLogs }}</td>
          </tr>
        </tbody>
      </table>
      <div class="pagination" v-if="total > 0">
        <button class="btn-ghost pag-btn" :disabled="page===1" @click="page--;loadLogs()">‹</button>
        <span class="pag-info">{{ page }} / {{ totalPages }}</span>
        <button class="btn-ghost pag-btn" :disabled="page>=totalPages" @click="page++;loadLogs()">›</button>
        <span class="total-count">{{ total }} {{ i18n.lang === 'th' ? 'รายการ' : 'records' }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18nStore } from '@/stores/i18n'
import api from '@/composables/useApi'

const i18n = useI18nStore()
const t = computed(() => i18n.t)
const logs = ref([])
const loading = ref(true)
const total = ref(0)
const page = ref(1)
const limit = 50
const filters = ref({ event_type: '' })
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit)))

onMounted(loadLogs)

async function loadLogs() {
  loading.value = true
  try {
    const params = { page: page.value, limit }
    if (filters.value.event_type) params.event_type = filters.value.event_type
    const res = await api.get('/api/admin/audit-logs', { params })
    logs.value = res.data.data || []
    total.value = res.data.total || 0
  } finally { loading.value = false }
}

function eventBadge(e) {
  return {
    BOOKING_SUCCESS: 'badge badge-success',
    BOOKING_TIMEOUT: 'badge badge-error',
    SEAT_RELEASED:   'badge badge-warning',
    SYSTEM_ERROR:    'badge badge-error',
    BOOKING_LOCKED:  'badge badge-info',
  }[e] || 'badge badge-accent'
}
function shortId(id) { return id ? id.substring(0,8)+'...' : '—' }
function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString(i18n.lang === 'th' ? 'th-TH' : 'en-US', { month:'short', day:'numeric', hour:'2-digit', minute:'2-digit', second:'2-digit' })
}
</script>

<style scoped>
.filter-row {
  display: flex; align-items: flex-end; gap: 1rem; flex-wrap: wrap;
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: 14px; padding: 1.25rem; margin-bottom: 1.25rem;
}
.filter-group { display:flex; flex-direction:column; gap:0.4rem; }
.filter-group label { font-size:0.72rem; color:var(--text-muted); font-weight:700; letter-spacing:0.06em; text-transform:uppercase; }

.table-card {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: 14px; overflow: hidden;
}
.seat-badge { font-family:'Syne',sans-serif; font-weight:800; font-size:1rem; color:var(--accent); }
.mono { font-family:monospace; font-size:0.78rem; }
.muted { color:var(--text-muted); }
.empty-row { text-align:center; color:var(--text-muted); padding:3rem !important; }

.pagination {
  display:flex; align-items:center; gap:0.75rem;
  padding:1rem; border-top:1px solid var(--border);
}
.pag-btn { width:32px; height:32px; padding:0; font-size:1.1rem; display:flex; align-items:center; justify-content:center; }
.pag-info { font-size:0.85rem; color:var(--text-secondary); min-width:60px; text-align:center; }
.total-count { font-size:0.78rem; color:var(--text-muted); margin-left:auto; }
</style>
