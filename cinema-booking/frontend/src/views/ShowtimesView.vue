<template>
  <div class="page">
    <!-- Header -->
    <div class="page-header">
      <h1 class="page-title">{{ t.showtimes.title }}</h1>
      <div class="header-right">
        <!-- Search -->
        <div class="search-box">
          <span class="search-icon">🔍</span>
          <input v-model="search" :placeholder="i18n.lang==='th'?'ค้นหาหนัง...':'Search movies...'" />
        </div>
      </div>
    </div>

    <!-- Genre filter pills -->
    <div class="genre-bar">
      <button
        v-for="g in genres" :key="g.key"
        :class="['genre-pill', activeGenre===g.key && 'active']"
        @click="activeGenre = g.key"
      >
        <span>{{ g.emoji }}</span> {{ i18n.lang==='th' ? g.labelTh : g.labelEn }}
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-spinner"><div class="spinner-ring"></div></div>

    <!-- Empty -->
    <div v-else-if="filtered.length === 0" class="empty-state">
      <p style="font-size:3rem">🎬</p>
      <p>{{ t.showtimes.noShowtimes }}</p>
    </div>

    <!-- Movie Grid — group by movie name -->
    <div v-else>
      <div v-for="group in groupedMovies" :key="group.movieName" class="movie-section">
        <!-- Movie Banner -->
        <div class="movie-banner" :style="{ background: posterGradient(group.movieName) }">
          <div class="banner-overlay"></div>
          <div class="banner-content">
            <div class="banner-poster">{{ group.posterEmoji }}</div>
            <div class="banner-info">
              <div class="banner-tags">
                <span class="tag-rating">{{ group.rating }}</span>
                <span class="tag-genre">{{ group.genre }}</span>
                <span class="tag-duration">🕐 {{ group.duration }}{{ i18n.lang==='th'?'นาที':'min' }}</span>
              </div>
              <h2 class="banner-title">{{ group.movieName }}</h2>
            </div>
          </div>
        </div>

        <!-- Showtime chips for this movie -->
        <div class="showtime-chips">
          <RouterLink
            v-for="st in group.showtimes"
            :key="st.id"
            :to="`/showtimes/${st.id}`"
            class="showtime-chip"
          >
            <div class="chip-hall">{{ st.hall }}</div>
            <div class="chip-time">{{ formatTime(st.start_time) }}</div>
            <div class="chip-endtime">{{ formatTime(st.end_time) }}</div>
            <div class="chip-price">฿{{ priceFor(st) }}</div>
          </RouterLink>
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
const showtimes = ref([])
const loading = ref(true)
const activeGenre = ref('all')
const search = ref('')

const genres = [
  { key:'all',       emoji:'🎬', labelTh:'ทั้งหมด',   labelEn:'All' },
  { key:'Action',    emoji:'⚡', labelTh:'แอคชั่น',   labelEn:'Action' },
  { key:'Sci-Fi',    emoji:'🚀', labelTh:'ไซไฟ',      labelEn:'Sci-Fi' },
  { key:'Drama',     emoji:'🎭', labelTh:'ดราม่า',    labelEn:'Drama' },
  { key:'Horror',    emoji:'👻', labelTh:'สยองขวัญ',  labelEn:'Horror' },
  { key:'Animation', emoji:'✨', labelTh:'การ์ตูน',   labelEn:'Animation' },
  { key:'Romance',   emoji:'💕', labelTh:'โรแมนติก',  labelEn:'Romance' },
  { key:'Thriller',  emoji:'😱', labelTh:'ระทึกขวัญ', labelEn:'Thriller' },
]

const gradients = [
  'linear-gradient(135deg,#0d0221,#3a0a6e)',
  'linear-gradient(135deg,#021520,#0a4a7f)',
  'linear-gradient(135deg,#1a0800,#6e2a00)',
  'linear-gradient(135deg,#001a0a,#006e2a)',
  'linear-gradient(135deg,#1a0015,#6e003a)',
  'linear-gradient(135deg,#0a0a1a,#2a2a6e)',
  'linear-gradient(135deg,#1a0a00,#6e3a00)',
  'linear-gradient(135deg,#001015,#004a5a)',
]

function posterGradient(name) {
  let hash = 0
  for (const c of name) hash = (hash * 31 + c.charCodeAt(0)) & 0xffff
  return gradients[hash % gradients.length]
}

function priceFor(st) {
  // return default price label
  return 300
}

onMounted(async () => {
  try {
    const res = await api.get('/api/showtimes')
    showtimes.value = res.data || []
  } finally { loading.value = false }
})

const filtered = computed(() => {
  let list = showtimes.value
  if (activeGenre.value !== 'all') list = list.filter(s => s.genre === activeGenre.value)
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(s => s.movie_name.toLowerCase().includes(q))
  }
  return list
})

const groupedMovies = computed(() => {
  const map = {}
  for (const st of filtered.value) {
    if (!map[st.movie_name]) {
      map[st.movie_name] = {
        movieName: st.movie_name,
        posterEmoji: st.poster_emoji || '🎬',
        genre: st.genre || 'Action',
        rating: st.rating || 'G',
        duration: st.duration || 120,
        showtimes: [],
      }
    }
    map[st.movie_name].showtimes.push(st)
  }
  return Object.values(map)
})

function formatTime(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleTimeString(i18n.lang==='th'?'th-TH':'en-US', { hour:'2-digit', minute:'2-digit', hour12:false })
}
</script>

<style scoped>
.page { max-width:1200px; margin:0 auto; padding:2rem; }

.header-right { display:flex; gap:1rem; align-items:center; }
.search-box {
  display:flex; align-items:center; gap:0.5rem;
  background:var(--bg-card); border:1px solid var(--border2);
  border-radius:10px; padding:0.45rem 0.85rem;
}
.search-box input {
  background:transparent; border:none; color:var(--text-primary);
  font-size:0.88rem; outline:none; min-width:200px;
}
.search-icon { color:var(--text-muted); font-size:0.9rem; }

/* Genre pills */
.genre-bar {
  display:flex; gap:0.5rem; flex-wrap:wrap; margin-bottom:2rem;
  padding-bottom:1rem; border-bottom:1px solid var(--border);
}
.genre-pill {
  display:flex; align-items:center; gap:0.35rem;
  background:var(--bg-card); border:1px solid var(--border);
  color:var(--text-secondary); padding:0.4rem 1rem;
  border-radius:999px; font-size:0.82rem; font-weight:600;
  cursor:pointer; transition:all 0.15s;
}
.genre-pill:hover { border-color:var(--accent); color:var(--accent); }
.genre-pill.active { background:var(--accent-dim); border-color:var(--accent); color:var(--accent); }

/* Movie Section */
.movie-section { margin-bottom:2rem; }

.movie-banner {
  border-radius:16px 16px 0 0;
  position:relative; overflow:hidden; height:140px;
  display:flex; align-items:flex-end;
}
.banner-overlay {
  position:absolute; inset:0;
  background:linear-gradient(to right, rgba(0,0,0,0.8) 0%, rgba(0,0,0,0.2) 60%, transparent 100%);
}
.banner-content {
  position:relative; display:flex; align-items:center; gap:1rem; padding:1.25rem;
}
.banner-poster {
  font-size:3rem; width:64px; height:64px;
  background:rgba(255,255,255,0.08); border-radius:10px;
  display:flex; align-items:center; justify-content:center;
  border:1px solid rgba(255,255,255,0.1); flex-shrink:0;
}
.banner-tags { display:flex; gap:0.4rem; margin-bottom:0.4rem; flex-wrap:wrap; }
.tag-rating  { font-size:0.65rem; font-weight:800; padding:0.15rem 0.5rem; border-radius:4px; background:rgba(231,76,60,0.25); color:#e74c3c; border:1px solid rgba(231,76,60,0.3); }
.tag-genre   { font-size:0.65rem; font-weight:800; padding:0.15rem 0.5rem; border-radius:4px; background:rgba(52,152,219,0.2); color:#3498db; border:1px solid rgba(52,152,219,0.3); }
.tag-duration{ font-size:0.65rem; font-weight:700; padding:0.15rem 0.5rem; border-radius:4px; background:rgba(245,166,35,0.15); color:var(--accent); border:1px solid rgba(245,166,35,0.25); }
.banner-title {
  font-family:'Syne',sans-serif; font-size:1.3rem; font-weight:800;
  color:white; text-shadow:0 2px 8px rgba(0,0,0,0.8); line-height:1.2;
}

/* Showtime chips */
.showtime-chips {
  display:flex; flex-wrap:wrap; gap:0.75rem;
  background:var(--bg-card); border:1px solid var(--border);
  border-top:none; border-radius:0 0 16px 16px;
  padding:1rem 1.25rem;
}
.showtime-chip {
  display:flex; flex-direction:column; gap:0.15rem;
  background:var(--bg-secondary); border:1px solid var(--border2);
  border-radius:10px; padding:0.65rem 1rem;
  text-decoration:none; color:inherit;
  transition:all 0.15s; min-width:100px; text-align:center;
}
.showtime-chip:hover {
  border-color:var(--accent); background:var(--accent-dim);
  transform:translateY(-2px);
}
.chip-hall    { font-size:0.68rem; color:var(--text-muted); font-weight:600; letter-spacing:0.04em; }
.chip-time    { font-family:'Syne',sans-serif; font-size:1.2rem; font-weight:800; color:var(--accent); line-height:1; }
.chip-endtime { font-size:0.72rem; color:var(--text-secondary); }
.chip-price   { font-size:0.75rem; font-weight:700; color:var(--success); margin-top:0.2rem; }

.empty-state { text-align:center; padding:4rem 2rem; color:var(--text-muted); }
</style>
