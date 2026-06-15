<template>
  <div class="seatmap-page">

    <!-- Movie Hero Banner -->
    <div class="movie-hero" v-if="showtime">
      <div class="hero-gradient"></div>
      <div class="hero-content">
        <div class="poster-icon">{{ showtime.poster_emoji || '🎬' }}</div>
        <div class="hero-info">
          <div class="hero-tags">
            <span class="tag-rating">{{ showtime.rating || 'G' }}</span>
            <span class="tag-genre">{{ showtime.genre || 'Action' }}</span>
            <span class="tag-duration">🕐 {{ showtime.duration || 120 }} {{ i18n.lang==='th'?'นาที':'min' }}</span>
          </div>
          <h1 class="hero-title">{{ showtime.movie_name }}</h1>
          <p class="hero-meta">
            🎭 {{ showtime.hall }}
            &nbsp;·&nbsp;
            📅 {{ formatDate(showtime.start_time) }}
            &nbsp;·&nbsp;
            🕐 {{ formatTime(showtime.start_time) }} – {{ formatTime(showtime.end_time) }}
          </p>
        </div>
        <div class="ws-badge" :class="{ connected: wsConnected }">
          <span class="ws-dot"></span>
          {{ wsConnected ? (i18n.lang==='th'?'สด':'Live') : (i18n.lang==='th'?'กำลังเชื่อม...':'Reconnecting...') }}
        </div>
      </div>
    </div>

    <!-- Step bar -->
    <div class="step-bar">
      <div class="step" :class="{ active: !activeBookings.length, done: activeBookings.length }">
        <span class="step-num">{{ activeBookings.length ? '✓' : '1' }}</span>
        <span>{{ i18n.lang==='th' ? 'เลือกที่นั่ง' : 'Select Seats' }}</span>
      </div>
      <div class="step-line"></div>
      <div class="step" :class="{ active: activeBookings.length }">
        <span class="step-num">2</span>
        <span>{{ i18n.lang==='th' ? 'ชำระเงิน' : 'Payment' }}</span>
      </div>
      <div class="step-line"></div>
      <div class="step">
        <span class="step-num">3</span>
        <span>{{ i18n.lang==='th' ? 'ยืนยัน' : 'Confirm' }}</span>
      </div>
    </div>

    <div class="main-layout">
      <!-- Left: Seat Map -->
      <div class="seatmap-container">
        <div v-if="loading" class="loading-spinner"><div class="spinner-ring"></div></div>
        <template v-else>

          <!-- Screen -->
          <div class="screen-section">
            <div class="screen-box">
              <span>{{ i18n.lang==='th'?'จอภาพยนตร์':'SCREEN' }}</span>
            </div>
            <div class="screen-glow"></div>
          </div>

          <!-- Legend -->
          <div class="legend">
            <div class="legend-item" v-for="l in legends" :key="l.key">
              <div :class="`legend-seat ls-${l.key}`"><span class="seat-icon">🪑</span></div>
              <span>{{ l.label }}</span>
            </div>
          </div>

          <!-- Price legend -->
          <div class="price-legend">
            <div class="price-item">
              <span class="price-dot normal"></span>
              Normal · ฿{{ priceNormal }}
            </div>
            <div class="price-item">
              <span class="price-dot vip"></span>
              VIP · ฿{{ priceVIP }}
            </div>
          </div>

          <!-- Seat grid -->
          <div class="seat-grid">
            <div v-for="row in seatRows" :key="row.row" class="seat-row">
              <span class="row-label">{{ row.row }}</span>
              <div class="seats-wrap">
                <!-- Aisle gap in middle -->
                <template v-for="(seat, idx) in row.seats" :key="seat.id">
                  <div class="aisle-gap" v-if="idx === Math.floor(row.seats.length / 2)"></div>
                  <button
                    :class="getSeatClass(seat)"
                    :disabled="isSeatDisabled(seat)"
                    :title="`${seat.seat_code} · ${seat.type} · ฿${seat.price}`"
                    @click="toggleSeat(seat)"
                  >
                    <span class="seat-chair">🪑</span>
                    <span class="seat-num">{{ seat.number }}</span>
                  </button>
                </template>
              </div>
              <span class="row-label">{{ row.row }}</span>
            </div>
          </div>

        </template>
      </div>

      <!-- Right: Summary Panel -->
      <div class="summary-panel">
        <div class="summary-card">
          <h3 class="summary-title">{{ i18n.lang==='th'?'สรุปการจอง':'Summary' }}</h3>

          <!-- Movie info -->
          <div class="summary-movie" v-if="showtime">
            <div class="summary-poster">{{ showtime.poster_emoji || '🎬' }}</div>
            <div>
              <p class="summary-movie-name">{{ showtime.movie_name }}</p>
              <p class="summary-movie-meta">{{ showtime.hall }}</p>
              <p class="summary-movie-meta">{{ formatTime(showtime.start_time) }}</p>
            </div>
          </div>

          <div class="summary-divider"></div>

          <!-- Selected seats -->
          <div class="summary-section">
            <p class="summary-label">{{ i18n.lang==='th'?'ที่นั่งที่เลือก':'Selected Seats' }}</p>
            <div v-if="displaySeats.length === 0" class="summary-empty">
              {{ i18n.lang==='th'?'ยังไม่ได้เลือกที่นั่ง':'No seats selected' }}
            </div>
            <div v-else class="summary-seats">
              <div v-for="s in displaySeats" :key="s.id" class="summary-seat-row">
                <span class="summary-seat-code">{{ s.seat_code }}</span>
                <span class="summary-seat-type" :class="s.type?.toLowerCase()">{{ s.type }}</span>
                <span class="summary-seat-price">฿{{ s.price }}</span>
              </div>
            </div>
          </div>

          <!-- Countdown -->
          <div class="summary-countdown" v-if="activeBookings.length && lockExpiresAt">
            <div class="countdown-header">
              <span>⏱ {{ i18n.lang==='th'?'เหลือเวลา':'Time remaining' }}</span>
              <span class="countdown-time" :class="countdownPct < 30 ? 'urgent' : ''">{{ countdownText }}</span>
            </div>
            <div class="countdown-track">
              <div class="countdown-fill" :style="{ width: countdownPct + '%' }" :class="countdownPct < 30 ? 'urgent' : ''"></div>
            </div>
          </div>

          <div class="summary-divider"></div>

          <!-- Total -->
          <div class="summary-total-row">
            <span>{{ i18n.lang==='th'?'ราคารวม':'Total' }}</span>
            <span class="summary-total-price">฿{{ totalPrice }}</span>
          </div>

          <!-- Action buttons -->
          <div class="summary-actions">
            <template v-if="activeBookings.length === 0">
              <button class="btn-book" :disabled="selectedSeats.length === 0 || locking" @click="lockSeats">
                <span v-if="locking" class="spinner-ring" style="width:14px;height:14px;border-width:2px;border-color:rgba(255,255,255,0.3);border-top-color:white"></span>
                {{ locking ? (i18n.lang==='th'?'กำลังจอง...':'Locking...') : (i18n.lang==='th'?`🔒 จองที่นั่ง (${selectedSeats.length})`:`🔒 Hold Seats (${selectedSeats.length})`) }}
              </button>
              <button class="btn-clear" v-if="selectedSeats.length" @click="selectedSeats=[]">
                {{ i18n.lang==='th'?'ล้างการเลือก':'Clear' }}
              </button>
            </template>
            <template v-else>
              <button class="btn-confirm" :disabled="confirming" @click="confirmAll">
                <span v-if="confirming" class="spinner-ring" style="width:14px;height:14px;border-width:2px;border-color:rgba(255,255,255,0.3);border-top-color:white"></span>
                {{ confirming ? (i18n.lang==='th'?'กำลังยืนยัน...':'Confirming...') : (i18n.lang==='th'?'✅ ยืนยันและชำระเงิน':'✅ Confirm & Pay') }}
              </button>
              <button class="btn-cancel" @click="cancelAll">
                {{ i18n.lang==='th'?'ยกเลิก':'Cancel' }}
              </button>
            </template>
          </div>

          <button class="btn-back" @click="$router.push('/showtimes')">
            ← {{ i18n.lang==='th'?'เลือกรอบอื่น':'Change showtime' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/composables/useApi'
import { useWebSocket } from '@/composables/useWebSocket'
import { useToastStore } from '@/stores/toast'
import { useI18nStore } from '@/stores/i18n'

const route = useRoute()
const toast = useToastStore()
const i18n = useI18nStore()
const showtimeId = route.params.id

const loading = ref(true)
const showtime = ref(null)
const seats = ref([])
const selectedSeats = ref([])
const activeBookings = ref([])
const locking = ref(false)
const confirming = ref(false)
const lockExpiresAt = ref(null)
const countdown = ref(300)
let countdownInterval = null

const { connected: wsConnected, connect, onMessage } = useWebSocket(showtimeId)
onMessage((msg) => {
  const idx = seats.value.findIndex(s => s.seat_code === msg.seat_code)
  if (idx !== -1) seats.value[idx] = { ...seats.value[idx], status: msg.status }
})

onMounted(async () => {
  connect()
  try {
    const [stRes, seatsRes] = await Promise.all([
      api.get('/api/showtimes'),
      api.get(`/api/showtimes/${showtimeId}/seats`)
    ])
    showtime.value = (stRes.data || []).find(s => s.id === showtimeId)
    seats.value = seatsRes.data || []
  } finally { loading.value = false }
})
onUnmounted(() => clearInterval(countdownInterval))

const priceNormal = computed(() => seats.value.find(s => s.type === 'NORMAL')?.price || 300)
const priceVIP    = computed(() => seats.value.find(s => s.type === 'VIP')?.price || 400)

const seatRows = computed(() => {
  const rows = {}
  for (const s of seats.value) {
    if (!rows[s.row]) rows[s.row] = { row: s.row, seats: [] }
    rows[s.row].seats.push(s)
  }
  return Object.values(rows).sort((a, b) => a.row.localeCompare(b.row))
})

const displaySeats = computed(() =>
  activeBookings.value.length
    ? activeBookings.value.map(b => ({ id: b.seat_id, seat_code: b.seat_code, type: b.type, price: b.price }))
    : selectedSeats.value
)

const totalPrice = computed(() => displaySeats.value.reduce((sum, s) => sum + (s.price || 0), 0))

const countdownPct = computed(() => Math.max(0, (countdown.value / 300) * 100))
const countdownText = computed(() => {
  const m = Math.floor(countdown.value / 60)
  const s = countdown.value % 60
  return `${m}:${s.toString().padStart(2, '0')}`
})

const legends = computed(() => [
  { key: 'available', label: i18n.lang==='th'?'ว่าง':'Available' },
  { key: 'selected',  label: i18n.lang==='th'?'เลือกแล้ว':'Selected' },
  { key: 'locked',    label: i18n.lang==='th'?'ถูกจอง':'Reserved' },
  { key: 'booked',    label: i18n.lang==='th'?'ขายแล้ว':'Sold' },
])

function getSeatClass(seat) {
  const isSelected = selectedSeats.value.some(s => s.id === seat.id)
  const isActive   = activeBookings.value.some(b => b.seat_id === seat.id)
  const isVIP      = seat.type === 'VIP'
  const base = `seat ${isVIP ? 'seat-vip' : 'seat-normal'}`
  if (isActive)                 return base + ' seat-active'
  if (isSelected)               return base + ' seat-selected'
  if (seat.status === 'LOCKED') return base + ' seat-locked'
  if (seat.status === 'BOOKED') return base + ' seat-booked'
  return base + ' seat-available'
}

function isSeatDisabled(seat) {
  if (activeBookings.value.length) return true
  return seat.status === 'LOCKED' || seat.status === 'BOOKED'
}

function toggleSeat(seat) {
  if (isSeatDisabled(seat)) return
  const idx = selectedSeats.value.findIndex(s => s.id === seat.id)
  if (idx === -1) selectedSeats.value.push(seat)
  else selectedSeats.value.splice(idx, 1)
}

async function lockSeats() {
  if (!selectedSeats.value.length) return
  locking.value = true
  const results = []
  for (const seat of selectedSeats.value) {
    try {
      const res = await api.post('/api/bookings/lock', { showtime_id: showtimeId, seat_id: seat.id })
      results.push({ booking_id: res.data.booking_id, seat_id: seat.id, seat_code: seat.seat_code, type: seat.type, price: seat.price, expires_at: res.data.expires_at })
    } catch (err) {
      toast.error(`${seat.seat_code}: ${err.response?.data?.error || 'Failed'}`)
    }
  }
  if (results.length) {
    activeBookings.value = results
    lockExpiresAt.value = new Date(results[0].expires_at)
    selectedSeats.value = []
    toast.success(i18n.lang==='th' ? `🔒 จองสำเร็จ ${results.length} ที่นั่ง` : `🔒 ${results.length} seats locked!`)
    startCountdown(lockExpiresAt.value)
  }
  locking.value = false
}

async function confirmAll() {
  confirming.value = true
  let ok = 0
  for (const b of activeBookings.value) {
    try {
      await api.post(`/api/bookings/${b.booking_id}/confirm`)
      ok++
      const idx = seats.value.findIndex(s => s.id === b.seat_id)
      if (idx !== -1) seats.value[idx] = { ...seats.value[idx], status: 'BOOKED' }
    } catch (err) {
      toast.error(`${b.seat_code}: ${err.response?.data?.error || 'Failed'}`)
    }
  }
  if (ok > 0) {
    toast.success(i18n.lang==='th' ? `🎉 ยืนยัน ${ok} ที่นั่งสำเร็จ!` : `🎉 ${ok} seats confirmed!`)
    activeBookings.value = []
    lockExpiresAt.value = null
    clearInterval(countdownInterval)
  }
  confirming.value = false
}

async function cancelAll() {
  for (const b of activeBookings.value) {
    try {
      await api.delete(`/api/bookings/${b.booking_id}`)
      const idx = seats.value.findIndex(s => s.id === b.seat_id)
      if (idx !== -1) seats.value[idx] = { ...seats.value[idx], status: 'AVAILABLE' }
    } catch {}
  }
  activeBookings.value = []
  lockExpiresAt.value = null
  clearInterval(countdownInterval)
  toast.info(i18n.lang==='th'?'ยกเลิกที่นั่งแล้ว':'Seats released.')
}

function startCountdown(exp) {
  clearInterval(countdownInterval)
  const tick = () => {
    countdown.value = Math.max(0, Math.floor((exp - Date.now()) / 1000))
    if (countdown.value === 0) {
      clearInterval(countdownInterval)
      toast.error(i18n.lang==='th'?'หมดเวลา! เลือกที่นั่งใหม่ได้เลย':'Lock expired!')
      activeBookings.value.forEach(b => {
        const idx = seats.value.findIndex(s => s.id === b.seat_id)
        if (idx !== -1) seats.value[idx] = { ...seats.value[idx], status: 'AVAILABLE' }
      })
      activeBookings.value = []
      lockExpiresAt.value = null
    }
  }
  tick()
  countdownInterval = setInterval(tick, 1000)
}

function formatTime(iso) { return iso ? new Date(iso).toLocaleTimeString(i18n.lang==='th'?'th-TH':'en-US',{hour:'2-digit',minute:'2-digit'}) : '' }
function formatDate(iso) { return iso ? new Date(iso).toLocaleDateString(i18n.lang==='th'?'th-TH':'en-US',{weekday:'short',month:'short',day:'numeric'}) : '' }
</script>

<style scoped>
.seatmap-page { min-height: 100vh; background: var(--bg-primary); }

/* Hero */
.movie-hero {
  position: relative;
  background: linear-gradient(135deg, #0a0015 0%, #150025 50%, #080010 100%);
  padding: 2rem;
  border-bottom: 1px solid var(--border);
  overflow: hidden;
}
.hero-gradient {
  position: absolute; inset: 0;
  background: radial-gradient(ellipse at 30% 50%, rgba(224,64,251,0.08) 0%, transparent 60%),
              radial-gradient(ellipse at 70% 50%, rgba(245,166,35,0.06) 0%, transparent 60%);
}
.hero-content {
  position: relative;
  max-width: 1400px; margin: 0 auto;
  display: flex; align-items: center; gap: 1.5rem;
}
.poster-icon {
  font-size: 3.5rem;
  background: var(--bg-card2);
  border: 1px solid var(--border2);
  width: 80px; height: 80px;
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.hero-info { flex: 1; }
.hero-tags { display: flex; gap: 0.5rem; margin-bottom: 0.5rem; flex-wrap: wrap; }
.tag-rating, .tag-genre, .tag-duration {
  font-size: 0.7rem; font-weight: 700; letter-spacing: 0.06em;
  padding: 0.2rem 0.6rem; border-radius: 4px;
}
.tag-rating  { background: rgba(231,76,60,0.2); color: #e74c3c; border: 1px solid rgba(231,76,60,0.3); }
.tag-genre   { background: rgba(52,152,219,0.2); color: #3498db; border: 1px solid rgba(52,152,219,0.3); }
.tag-duration{ background: rgba(245,166,35,0.15); color: var(--accent); border: 1px solid rgba(245,166,35,0.25); }
.hero-title  { font-family:'Syne',sans-serif; font-size:1.6rem; font-weight:800; color:var(--text-primary); margin-bottom:0.3rem; }
.hero-meta   { font-size:0.82rem; color:var(--text-secondary); }

.ws-badge {
  display:flex; align-items:center; gap:0.4rem;
  background:var(--bg-card); border:1px solid var(--border);
  border-radius:999px; padding:0.3rem 0.85rem;
  font-size:0.72rem; font-weight:700; color:var(--text-muted);
  flex-shrink:0;
}
.ws-dot { width:7px;height:7px;border-radius:50%;background:var(--text-muted); }
.ws-badge.connected { border-color:rgba(46,204,113,0.3);color:var(--success); }
.ws-badge.connected .ws-dot { background:var(--success);box-shadow:0 0 6px var(--success);animation:blink 1.5s infinite; }
@keyframes blink{0%,100%{opacity:1;}50%{opacity:0.3;}}

/* Step bar */
.step-bar {
  display:flex; align-items:center; justify-content:center; gap:0;
  background:var(--bg-secondary); border-bottom:1px solid var(--border);
  padding:0.85rem 2rem;
}
.step { display:flex; align-items:center; gap:0.5rem; font-size:0.82rem; font-weight:600; color:var(--text-muted); }
.step.active { color:var(--accent); }
.step.done   { color:var(--success); }
.step-num {
  width:24px;height:24px;border-radius:50%;
  background:var(--bg-card); border:2px solid var(--text-muted);
  display:flex;align-items:center;justify-content:center;
  font-size:0.72rem; font-weight:800;
}
.step.active .step-num { background:var(--accent);border-color:var(--accent);color:#080810; }
.step.done   .step-num { background:var(--success);border-color:var(--success);color:#080810; }
.step-line { width:60px; height:2px; background:var(--border); margin:0 0.75rem; }

/* Main layout */
.main-layout {
  display:grid; grid-template-columns:1fr 320px;
  gap:0; max-width:1400px; margin:0 auto;
  min-height:calc(100vh - 200px);
}

/* Seat map container */
.seatmap-container {
  padding:2rem;
  border-right:1px solid var(--border);
  overflow-x:auto;
}

/* Screen */
.screen-section { margin-bottom:1rem; text-align:center; }
.screen-box {
  display:inline-block;
  background:linear-gradient(180deg, rgba(255,255,255,0.08) 0%, rgba(255,255,255,0.03) 100%);
  border:1px solid rgba(255,255,255,0.15);
  border-bottom:3px solid rgba(255,255,255,0.2);
  color:rgba(255,255,255,0.5);
  font-size:0.65rem; font-weight:800; letter-spacing:0.3em;
  padding:0.5rem 5rem;
  border-radius:4px 4px 0 0;
  box-shadow:0 4px 20px rgba(255,255,255,0.05);
}
.screen-glow {
  width:100%; max-width:600px; height:20px; margin:0 auto;
  background:radial-gradient(ellipse, rgba(255,255,255,0.06) 0%, transparent 70%);
}

/* Legend */
.legend { display:flex;gap:1.5rem;justify-content:center;margin-bottom:0.5rem;flex-wrap:wrap; }
.legend-item { display:flex;align-items:center;gap:0.4rem;font-size:0.75rem;color:var(--text-secondary); }
.legend-seat {
  width:28px;height:24px;border-radius:5px 5px 2px 2px;
  display:flex;align-items:center;justify-content:center;font-size:0.75rem;
}
.ls-available { background:rgba(80,120,80,0.3);border:1px solid rgba(80,180,80,0.4); }
.ls-selected  { background:rgba(30,90,160,0.4);border:1px solid rgba(52,152,219,0.6); }
.ls-locked    { background:rgba(80,60,10,0.4);border:1px solid rgba(180,130,20,0.4); }
.ls-booked    { background:rgba(60,20,20,0.4);border:1px solid rgba(120,40,40,0.4); }

.price-legend { display:flex;gap:2rem;justify-content:center;margin-bottom:1.5rem; }
.price-item { display:flex;align-items:center;gap:0.4rem;font-size:0.75rem;color:var(--text-secondary); }
.price-dot { width:10px;height:10px;border-radius:50%; }
.price-dot.normal { background:#4a8a5a; }
.price-dot.vip    { background:#8a6a20; }

/* Seat grid */
.seat-grid { display:flex;flex-direction:column;gap:0.4rem;min-width:600px; }
.seat-row  { display:flex;align-items:center;gap:0.5rem; }
.row-label { width:20px;text-align:center;font-size:0.7rem;font-weight:700;color:var(--text-muted);flex-shrink:0; }
.seats-wrap { display:flex;gap:0.3rem;flex:1;justify-content:center;align-items:center; }
.aisle-gap { width:20px;flex-shrink:0; }

/* Seat button */
.seat {
  width:34px;height:30px;
  border-radius:5px 5px 2px 2px;
  display:flex;flex-direction:column;align-items:center;justify-content:center;
  cursor:pointer;border:1px solid transparent;
  transition:all 0.12s;
  position:relative;
  flex-shrink:0;
  padding:0;
}
.seat-icon { font-size:0.9rem;line-height:1;pointer-events:none; }
.seat-num  { font-size:0.5rem;font-weight:700;line-height:1;pointer-events:none; }

.seat-available {
  background:rgba(40,90,50,0.5);
  border-color:rgba(60,150,80,0.5);
  color:rgba(100,200,120,0.9);
}
.seat-available:hover:not(:disabled) {
  background:rgba(50,120,65,0.7);
  border-color:rgba(80,200,100,0.8);
  transform:scale(1.15) translateY(-2px);
  box-shadow:0 4px 12px rgba(60,180,80,0.3);
  z-index:2;
}

.seat-vip.seat-available {
  background:rgba(80,60,10,0.5);
  border-color:rgba(180,140,20,0.5);
  color:rgba(220,180,60,0.9);
}
.seat-vip.seat-available:hover:not(:disabled) {
  background:rgba(100,80,15,0.7);
  border-color:rgba(220,180,30,0.8);
  box-shadow:0 4px 12px rgba(200,160,20,0.3);
}

.seat-selected {
  background:rgba(20,70,140,0.7) !important;
  border-color:rgba(52,152,219,0.9) !important;
  color:#5dade2 !important;
  transform:scale(1.12) translateY(-2px);
  box-shadow:0 4px 14px rgba(52,152,219,0.4);
  z-index:2;
}

.seat-active {
  background:rgba(20,100,40,0.7) !important;
  border-color:rgba(46,204,113,0.9) !important;
  color:#6ee89a !important;
  transform:scale(1.1);
  box-shadow:0 4px 14px rgba(46,204,113,0.3);
}

.seat-locked {
  background:rgba(30,25,5,0.6);
  border-color:rgba(100,80,10,0.4);
  color:rgba(100,80,10,0.5);
  cursor:not-allowed;
}

.seat-booked {
  background:rgba(40,10,10,0.5);
  border-color:rgba(80,20,20,0.4);
  color:rgba(80,20,20,0.5);
  cursor:not-allowed;
}
.seat-booked .seat-icon { opacity:0.3; }

/* Right summary panel */
.summary-panel {
  padding:1.5rem;
  background:var(--bg-secondary);
  position:sticky; top:64px;
  height:fit-content;
  max-height:calc(100vh - 64px);
  overflow-y:auto;
}

.summary-card {
  background:var(--bg-card);
  border:1px solid var(--border);
  border-radius:16px;
  padding:1.25rem;
}

.summary-title {
  font-family:'Syne',sans-serif;
  font-size:1rem;font-weight:800;color:var(--text-primary);
  margin-bottom:1rem;
}

.summary-movie {
  display:flex;gap:0.75rem;align-items:center;
  margin-bottom:1rem;
}
.summary-poster {
  width:48px;height:48px;border-radius:8px;
  background:var(--bg-secondary);border:1px solid var(--border);
  display:flex;align-items:center;justify-content:center;
  font-size:1.5rem;flex-shrink:0;
}
.summary-movie-name { font-weight:700;font-size:0.88rem;color:var(--text-primary);line-height:1.3;margin-bottom:0.15rem; }
.summary-movie-meta { font-size:0.75rem;color:var(--text-muted); }

.summary-divider { height:1px;background:var(--border);margin:1rem 0; }

.summary-section {}
.summary-label { font-size:0.72rem;font-weight:700;color:var(--text-muted);text-transform:uppercase;letter-spacing:0.08em;margin-bottom:0.6rem; }
.summary-empty { font-size:0.82rem;color:var(--text-muted);text-align:center;padding:0.75rem;background:var(--bg-secondary);border-radius:8px; }

.summary-seats { display:flex;flex-direction:column;gap:0.4rem; }
.summary-seat-row { display:flex;align-items:center;gap:0.5rem; }
.summary-seat-code { font-family:'Syne',sans-serif;font-weight:800;font-size:1rem;color:var(--accent);min-width:40px; }
.summary-seat-type { font-size:0.65rem;font-weight:700;text-transform:uppercase;padding:0.15rem 0.5rem;border-radius:4px; }
.summary-seat-type.normal { background:rgba(46,204,113,0.1);color:var(--success); }
.summary-seat-type.vip    { background:rgba(245,166,35,0.1);color:var(--accent); }
.summary-seat-price { margin-left:auto;font-size:0.88rem;font-weight:600;color:var(--text-secondary); }

/* Countdown */
.summary-countdown { background:var(--bg-secondary);border-radius:10px;padding:0.85rem;margin:0.75rem 0; }
.countdown-header { display:flex;justify-content:space-between;align-items:center;margin-bottom:0.5rem;font-size:0.78rem;color:var(--text-secondary); }
.countdown-time { font-family:'Syne',sans-serif;font-size:1.2rem;font-weight:800;color:var(--accent); }
.countdown-time.urgent { color:var(--error);animation:pulse 1s ease-in-out infinite; }
@keyframes pulse{0%,100%{opacity:1;}50%{opacity:0.6;}}
.countdown-track { height:4px;background:var(--bg-card);border-radius:2px;overflow:hidden; }
.countdown-fill  { height:100%;border-radius:2px;background:var(--accent);transition:width 1s linear; }
.countdown-fill.urgent { background:var(--error); }

.summary-total-row { display:flex;justify-content:space-between;align-items:center;margin-bottom:1rem; font-size:0.88rem;color:var(--text-secondary); }
.summary-total-price { font-family:'Syne',sans-serif;font-size:1.6rem;font-weight:800;color:var(--success); }

/* Action buttons */
.summary-actions { display:flex;flex-direction:column;gap:0.6rem;margin-bottom:0.75rem; }
.btn-book {
  width:100%;padding:0.85rem;border-radius:10px;
  background:linear-gradient(135deg,#f5a623,#e8890a);
  color:#080810;font-weight:700;font-size:0.92rem;
  border:none;cursor:pointer;
  display:flex;align-items:center;justify-content:center;gap:0.5rem;
  transition:all 0.2s;
}
.btn-book:hover:not(:disabled) { filter:brightness(1.1);transform:translateY(-1px); }
.btn-book:disabled { opacity:0.4;cursor:not-allowed;transform:none; }

.btn-confirm {
  width:100%;padding:0.85rem;border-radius:10px;
  background:linear-gradient(135deg,#2ecc71,#27ae60);
  color:white;font-weight:700;font-size:0.92rem;
  border:none;cursor:pointer;
  display:flex;align-items:center;justify-content:center;gap:0.5rem;
  transition:all 0.2s;
}
.btn-confirm:hover:not(:disabled) { filter:brightness(1.1); }
.btn-confirm:disabled { opacity:0.4;cursor:not-allowed; }

.btn-cancel, .btn-clear {
  width:100%;padding:0.65rem;border-radius:10px;
  background:transparent;border:1px solid var(--border2);
  color:var(--text-secondary);font-size:0.85rem;cursor:pointer;
  transition:all 0.15s;
}
.btn-cancel:hover { background:rgba(231,76,60,0.08);border-color:rgba(231,76,60,0.3);color:var(--error); }
.btn-clear:hover  { background:var(--bg-hover);color:var(--text-primary); }

.btn-back {
  width:100%;padding:0.55rem;border-radius:8px;
  background:transparent;border:none;
  color:var(--text-muted);font-size:0.78rem;cursor:pointer;
  transition:color 0.15s;
}
.btn-back:hover { color:var(--text-secondary); }

@media(max-width:900px) {
  .main-layout { grid-template-columns:1fr; }
  .summary-panel { position:fixed;bottom:0;left:0;right:0;max-height:40vh;padding:1rem;border-top:1px solid var(--border2);z-index:50; }
  .seatmap-container { padding-bottom:200px; }
}
</style>
