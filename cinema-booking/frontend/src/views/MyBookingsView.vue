<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">{{ t.myBookings.title }}</h1>
      <span class="ticket-count" v-if="bookings.length > 0">{{ bookings.length }} {{ i18n.lang === 'th' ? 'ใบ' : 'tickets' }}</span>
    </div>

    <div v-if="loading" class="loading-spinner"><div class="spinner-ring"></div></div>

    <div v-else-if="bookings.length === 0" class="empty-state">
      <div class="empty-icon">🎟</div>
      <p>{{ t.myBookings.noBookings }}</p>
      <RouterLink to="/showtimes" class="btn-primary" style="margin-top:1.5rem;display:inline-block">{{ t.myBookings.browse }}</RouterLink>
    </div>

    <div v-else class="bookings-grid">
      <div v-for="b in bookings" :key="b.id" class="ticket-card" :class="`ticket-${b.status.toLowerCase()}`">
        <!-- Left stripe -->
        <div class="ticket-stripe"></div>

        <!-- Content -->
        <div class="ticket-content">
          <div class="ticket-top">
            <div class="seat-display">
              <span class="seat-label">{{ i18n.lang === 'th' ? 'ที่นั่ง' : 'SEAT' }}</span>
              <span class="seat-code">{{ b.seat_code }}</span>
            </div>
            <span :class="statusBadge(b.status)">{{ t.status[b.status] || b.status }}</span>
          </div>

          <div class="ticket-divider">
            <div class="divider-circle left"></div>
            <div class="divider-line"></div>
            <div class="divider-circle right"></div>
          </div>

          <div class="ticket-details">
            <div class="detail-row">
              <span class="detail-icon">🎫</span>
              <span class="detail-label">{{ t.myBookings.bookingId }}</span>
              <span class="detail-val mono">{{ b.id?.substring(0,16) }}...</span>
            </div>
            <div class="detail-row">
              <span class="detail-icon">💰</span>
              <span class="detail-label">{{ t.myBookings.price }}</span>
              <span class="detail-val price">฿{{ b.total_price }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-icon">📅</span>
              <span class="detail-label">{{ t.myBookings.created }}</span>
              <span class="detail-val">{{ formatDate(b.created_at) }}</span>
            </div>
            <div class="detail-row" v-if="b.expires_at && b.status === 'LOCKED'">
              <span class="detail-icon">⏰</span>
              <span class="detail-label">{{ t.myBookings.expires }}</span>
              <span class="detail-val warning">{{ formatDate(b.expires_at) }}</span>
            </div>
          </div>
        </div>

        <!-- Barcode decoration -->
        <div class="ticket-barcode">
          <div v-for="i in 18" :key="i" class="bar" :style="{ height: (30 + Math.sin(i*1.5)*20) + 'px' }"></div>
        </div>
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

onMounted(async () => {
  try {
    const res = await api.get('/api/bookings/me')
    bookings.value = res.data || []
  } finally {
    loading.value = false
  }
})

function statusBadge(s) {
  return { CONFIRMED:'badge badge-success', LOCKED:'badge badge-warning', TIMEOUT:'badge badge-error', CANCELLED:'badge badge-error' }[s] || 'badge badge-info'
}
function formatDate(iso) {
  if (!iso) return '-'
  return new Date(iso).toLocaleString(i18n.lang === 'th' ? 'th-TH' : 'en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.page { max-width: 900px; margin: 0 auto; padding: 2rem; }

.ticket-count {
  background: var(--accent-dim);
  color: var(--accent);
  border: 1px solid rgba(245,166,35,0.25);
  border-radius: 999px;
  padding: 0.3rem 1rem;
  font-size: 0.82rem;
  font-weight: 700;
}

.bookings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 1.25rem;
}

.ticket-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  gap: 0;
  position: relative;
  transition: transform 0.2s, box-shadow 0.2s;
}
.ticket-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 12px 36px rgba(0,0,0,0.3);
}

.ticket-stripe {
  width: 6px;
  flex-shrink: 0;
}
.ticket-confirmed .ticket-stripe { background: linear-gradient(180deg, var(--success), #27ae60); }
.ticket-locked    .ticket-stripe { background: linear-gradient(180deg, var(--warning), #e67e22); }
.ticket-timeout   .ticket-stripe { background: linear-gradient(180deg, var(--error), #c0392b); }
.ticket-cancelled .ticket-stripe { background: linear-gradient(180deg, var(--text-muted), #3a3a5a); }

.ticket-content { flex: 1; padding: 1.25rem; }

.ticket-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.seat-display { display: flex; flex-direction: column; gap: 0.1rem; }
.seat-label { font-size: 0.65rem; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.1em; }
.seat-code {
  font-family: 'Syne', sans-serif;
  font-size: 2.5rem;
  font-weight: 800;
  line-height: 1;
  color: var(--accent);
}

.ticket-divider {
  display: flex;
  align-items: center;
  gap: 0;
  margin: 0.75rem -1.25rem;
  position: relative;
}
.divider-circle {
  width: 16px; height: 16px;
  border-radius: 50%;
  background: var(--bg-primary);
  flex-shrink: 0;
}
.divider-line {
  flex: 1;
  border-top: 2px dashed var(--border2);
}

.ticket-details { display: flex; flex-direction: column; gap: 0.5rem; }
.detail-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.82rem;
}
.detail-icon { font-size: 0.9rem; flex-shrink: 0; }
.detail-label { color: var(--text-muted); min-width: 80px; }
.detail-val { color: var(--text-secondary); font-weight: 500; margin-left: auto; }
.detail-val.price { color: var(--success); font-weight: 700; font-size: 0.95rem; }
.detail-val.warning { color: var(--warning); }
.mono { font-family: monospace; font-size: 0.75rem; }

.ticket-barcode {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 1rem 0.75rem;
  border-left: 1px dashed var(--border);
  opacity: 0.3;
}
.bar {
  width: 2px;
  background: var(--text-secondary);
  border-radius: 1px;
  flex-shrink: 0;
}

.empty-state {
  text-align: center;
  padding: 5rem 2rem;
  color: var(--text-secondary);
}
.empty-icon { font-size: 4rem; margin-bottom: 1rem; }
</style>
