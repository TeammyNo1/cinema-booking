<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">{{ t.admin.dashboard }}</h1>
      <button class="btn-secondary" @click="loadAll" style="font-size:0.82rem">🔄 Refresh</button>
    </div>

    <!-- Stats -->
    <div class="stats-grid">
      <div v-for="s in statCards" :key="s.label" class="stat-card" :class="`stat-${s.color}`">
        <div class="stat-icon">{{ s.icon }}</div>
        <div class="stat-value">{{ s.value }}</div>
        <div class="stat-label">{{ s.label }}</div>
        <div class="stat-bar" :style="{ width: s.pct + '%' }"></div>
      </div>
    </div>

    <!-- Filters -->
    <div class="filter-row">
      <div class="filter-group">
        <label>{{ t.admin.status }}</label>
        <select v-model="filters.status" @change="page=1; loadBookings()">
          <option value="">{{ t.admin.allStatuses }}</option>
          <option value="CONFIRMED">{{ t.status.CONFIRMED }}</option>
          <option value="LOCKED">{{ t.status.LOCKED }}</option>
          <option value="TIMEOUT">{{ t.status.TIMEOUT }}</option>
          <option value="CANCELLED">{{ t.status.CANCELLED }}</option>
        </select>
      </div>
      <div class="filter-group">
        <label>Showtime ID</label>
        <input v-model="filters.showtime_id" :placeholder="t.admin.filterShowtime" @input="page=1; loadBookings()" />
      </div>
      <button class="btn-ghost" style="align-self:flex-end" @click="resetFilters">✕ {{ t.admin.clearFilters }}</button>
    </div>

    <!-- Table -->
    <div class="table-card">
      <div v-if="loading" class="loading-spinner"><div class="spinner-ring"></div></div>
      <table v-else>
        <thead>
          <tr>
            <th>{{ t.admin.bookingId }}</th>
            <th>{{ t.admin.seat }}</th>
            <th>{{ t.admin.userId }}</th>
            <th>{{ t.admin.status }}</th>
            <th>{{ t.admin.price }}</th>
            <th>{{ t.admin.created }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in bookings" :key="b.id">
            <td><span class="mono text-muted">{{ shortId(b.id) }}</span></td>
            <td><span class="seat-badge">{{ b.seat_code }}</span></td>
            <td><span class="mono text-muted">{{ shortId(b.user_id) }}</span></td>
            <td><span :class="statusBadge(b.status)">{{ t.status[b.status] || b.status }}</span></td>
            <td><span class="price-text">฿{{ b.total_price }}</span></td>
            <td>{{ formatDate(b.created_at) }}</td>
          </tr>
          <tr v-if="bookings.length === 0">
            <td colspan="6" class="empty-row">{{ t.admin.noBookings }}</td>
          </tr>
        </tbody>
      </table>
      <div class="pagination" v-if="total > 0">
        <button class="btn-ghost pag-btn" :disabled="page===1" @click="page--;loadBookings()">‹</button>
        <span class="pag-info">{{ page }} / {{ totalPages }}</span>
        <button class="btn-ghost pag-btn" :disabled="page>=totalPages" @click="page++;loadBookings()">›</button>
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

const bookings = ref([])
const loading = ref(true)
const total = ref(0)
const page = ref(1)
const limit = 20
const stats = ref(null)
const filters = ref({ status: '', showtime_id: '' })
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit)))

const statCards = computed(() => {
  if (!stats.value) return []
  const tot = stats.value.total_bookings || 1
  return [
    { icon: '🎟', label: t.value.admin.totalBookings, value: stats.value.total_bookings, color: 'blue', pct: 100 },
    { icon: '✅', label: t.value.admin.confirmed,     value: stats.value.confirmed_bookings, color: 'green', pct: Math.round(stats.value.confirmed_bookings/tot*100) },
    { icon: '🔒', label: t.value.admin.locked,        value: stats.value.locked_bookings, color: 'yellow', pct: Math.round(stats.value.locked_bookings/tot*100) },
    { icon: '⏰', label: t.value.admin.timedOut,      value: stats.value.timeout_bookings, color: 'red', pct: Math.round(stats.value.timeout_bookings/tot*100) },
  ]
})

onMounted(loadAll)
async function loadAll() { await Promise.all([loadBookings(), loadStats()]) }

async function loadBookings() {
  loading.value = true
  try {
    const params = { page: page.value, limit }
    if (filters.value.status) params.status = filters.value.status
    if (filters.value.showtime_id) params.showtime_id = filters.value.showtime_id
    const res = await api.get('/api/admin/bookings', { params })
    bookings.value = res.data.data || []
    total.value = res.data.total || 0
  } finally { loading.value = false }
}

async function loadStats() {
  try { stats.value = (await api.get('/api/admin/stats')).data } catch {}
}

function resetFilters() {
  filters.value = { status: '', showtime_id: '' }
  page.value = 1
  loadBookings()
}

function statusBadge(s) {
  return { CONFIRMED:'badge badge-success', LOCKED:'badge badge-warning', TIMEOUT:'badge badge-error', CANCELLED:'badge badge-error' }[s] || 'badge badge-info'
}
function shortId(id) { return id ? id.substring(0,8)+'...' : '-' }
function formatDate(iso) {
  if (!iso) return '-'
  return new Date(iso).toLocaleString(i18n.lang==='th'?'th-TH':'en-US',{month:'short',day:'numeric',hour:'2-digit',minute:'2-digit'})
}
</script>

<style scoped>
.stats-grid {
  display: grid; grid-template-columns: repeat(4,1fr); gap: 1rem; margin-bottom: 1.5rem;
}
.stat-card {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: 14px; padding: 1.25rem; position: relative; overflow: hidden;
}
.stat-icon { font-size: 1.6rem; margin-bottom: 0.5rem; }
.stat-value {
  font-family:'Syne',sans-serif; font-size:2.2rem; font-weight:800; line-height:1;
  margin-bottom:0.3rem;
}
.stat-label { font-size:0.75rem; color:var(--text-muted); font-weight:600; }
.stat-bar {
  position:absolute; bottom:0; left:0; height:3px;
  border-radius:0 2px 0 0; transition:width 0.5s;
}
.stat-blue .stat-value { color: var(--info); }
.stat-blue .stat-bar { background: var(--info); }
.stat-green .stat-value { color: var(--success); }
.stat-green .stat-bar { background: var(--success); }
.stat-yellow .stat-value { color: var(--warning); }
.stat-yellow .stat-bar { background: var(--warning); }
.stat-red .stat-value { color: var(--error); }
.stat-red .stat-bar { background: var(--error); }

.filter-row {
  display: flex; align-items: flex-end; gap: 1rem; flex-wrap: wrap;
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: 14px; padding: 1.25rem; margin-bottom: 1.25rem;
}
.filter-group { display:flex; flex-direction:column; gap:0.4rem; }
.filter-group label { font-size:0.72rem; color:var(--text-muted); font-weight:700; letter-spacing:0.06em; text-transform:uppercase; }
.filter-group input { min-width:200px; }

.table-card {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: 14px; overflow: hidden;
}
.seat-badge {
  font-family:'Syne',sans-serif; font-weight:800; font-size:1rem; color:var(--accent);
}
.price-text { color:var(--success); font-weight:600; }
.mono { font-family:monospace; }
.text-muted { color:var(--text-muted); font-size:0.8rem; }
.empty-row { text-align:center; color:var(--text-muted); padding:3rem !important; }

.pagination {
  display:flex; align-items:center; gap:0.75rem;
  padding:1rem 1rem; border-top:1px solid var(--border);
}
.pag-btn { width:32px; height:32px; padding:0; font-size:1.1rem; display:flex; align-items:center; justify-content:center; }
.pag-info { font-size:0.85rem; color:var(--text-secondary); min-width:60px; text-align:center; }
.total-count { font-size:0.78rem; color:var(--text-muted); margin-left:auto; }

@media(max-width:768px){ .stats-grid{grid-template-columns:repeat(2,1fr);} }
</style>
