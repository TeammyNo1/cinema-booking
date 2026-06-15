<template>
  <div class="home-page">

    <!-- Hero Slider -->
    <section class="hero-slider">
      <div class="slide" :style="{ background: slides[activeSlide].bg }">
        <div class="slide-overlay"></div>
        <div class="slide-content">
          <div class="slide-meta">
            <span class="slide-tag">{{ i18n.lang==='th'?'ฉายอยู่ตอนนี้':'Now Showing' }}</span>
          </div>
          <h1 class="slide-title">{{ slides[activeSlide].title }}</h1>
          <p class="slide-sub">{{ slides[activeSlide].sub }}</p>
          <div class="slide-actions">
            <template v-if="auth.isLoggedIn">
              <RouterLink to="/showtimes" class="btn-primary btn-hero">
                🎟 {{ i18n.lang==='th'?'จองตั๋วเลย':'Book Now' }}
              </RouterLink>
            </template>
            <template v-else>
              <a :href="loginUrl" class="btn-primary btn-hero">
                <span class="g-icon">G</span>
                {{ i18n.lang==='th'?'เข้าสู่ระบบด้วย Google':'Sign in with Google' }}
              </a>
            </template>
          </div>
        </div>
        <div class="slide-poster">{{ slides[activeSlide].emoji }}</div>
      </div>

      <div class="slide-dots">
        <button v-for="(_, i) in slides" :key="i"
          :class="['dot', activeSlide===i && 'active']"
          @click="activeSlide=i; resetTimer()"></button>
      </div>
    </section>

    <!-- Now Showing quick row -->
    <section class="now-showing-section">
      <div class="section-header">
        <h2 class="section-title">{{ i18n.lang==='th'?'ภาพยนตร์แนะนำ':'Featured Movies' }}</h2>
        <RouterLink v-if="auth.isLoggedIn" to="/showtimes" class="see-all">
          {{ i18n.lang==='th'?'ดูทั้งหมด →':'See all →' }}
        </RouterLink>
      </div>

      <div v-if="loadingMovies" class="loading-spinner"><div class="spinner-ring"></div></div>

      <!-- Not logged in placeholder -->
      <div v-else-if="!auth.isLoggedIn" class="login-prompt">
        <p>{{ i18n.lang==='th'?'เข้าสู่ระบบเพื่อดูรายการหนังทั้งหมด':'Sign in to see all movies and book tickets' }}</p>
        <a :href="loginUrl" class="btn-primary" style="display:inline-block;margin-top:1rem">
          {{ i18n.lang==='th'?'เข้าสู่ระบบ':'Sign in' }}
        </a>
      </div>

      <div v-else-if="featuredMovies.length === 0" class="empty-movies">
        <p>{{ i18n.lang==='th'?'ยังไม่มีหนัง':'No movies yet' }}</p>
      </div>

      <div v-else class="movies-row">
        <RouterLink
          v-for="movie in featuredMovies"
          :key="movie.id"
          :to="`/showtimes/${movie.id}`"
          class="movie-card-mini"
        >
          <div class="mini-poster" :style="{ background: posterBg(movie.movie_name) }">
            <span class="mini-emoji">{{ movie.poster_emoji || '🎬' }}</span>
            <div class="mini-overlay">
              <span class="mini-rating">{{ movie.rating || 'PG' }}</span>
            </div>
          </div>
          <div class="mini-info">
            <p class="mini-title">{{ movie.movie_name }}</p>
            <p class="mini-genre">{{ movie.genre }}</p>
            <p class="mini-time">{{ formatTime(movie.start_time) }}</p>
          </div>
        </RouterLink>
      </div>
    </section>

    <!-- Features -->
    <section class="features-section">
      <div class="feature-item" v-for="f in features" :key="f.title">
        <div class="feature-emoji">{{ f.emoji }}</div>
        <h3>{{ f.title }}</h3>
        <p>{{ f.desc }}</p>
      </div>
    </section>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useI18nStore } from '@/stores/i18n'
import api from '@/composables/useApi'

const auth = useAuthStore()
const i18n = useI18nStore()
const loginUrl = computed(() => `${window.location.origin}/api/auth/google`)

const activeSlide = ref(0)
const featuredMovies = ref([])
const loadingMovies = ref(false)

const slides = [
  { title: 'Avengers: Secret Wars', sub: i18n.lang==='th'?'ซูเปอร์ฮีโร่รวมพลังครั้งสุดท้าย':'The ultimate superhero crossover event', emoji:'⚡', bg:'linear-gradient(135deg,#0d0221 0%,#3a0a6e 50%,#1a0033 100%)' },
  { title: 'Dune: Messiah',         sub: i18n.lang==='th'?'เส้นทางชะตากรรมที่ไม่อาจหลีกเลี่ยง':'The path of destiny continues',        emoji:'🏜️', bg:'linear-gradient(135deg,#1a0a00 0%,#6e3a00 50%,#2a1500 100%)' },
  { title: 'Spider-Man: Beyond',    sub: i18n.lang==='th'?'ผงาดข้ามจักรวาลสไปเดอร์แมน':'Across every universe',                  emoji:'🕷️', bg:'linear-gradient(135deg,#001020 0%,#003060 50%,#000820 100%)' },
  { title: 'Inception 2',           sub: i18n.lang==='th'?'ดำดิ่งสู่ความฝันอีกครั้ง':'Dive deeper into the dream',               emoji:'🌀', bg:'linear-gradient(135deg,#0a0a1a 0%,#1a1a5e 50%,#050510 100%)' },
]

let slideTimer = null

function resetTimer() {
  clearInterval(slideTimer)
  slideTimer = setInterval(() => {
    activeSlide.value = (activeSlide.value + 1) % slides.length
  }, 5000)
}

onMounted(async () => {
  resetTimer()

  // Only fetch movies if logged in (to avoid 401 spam)
  if (auth.isLoggedIn) {
    loadingMovies.value = true
    try {
      const res = await api.get('/api/showtimes')
      const all = res.data || []
      const seen = new Set()
      featuredMovies.value = all.filter(s => {
        if (seen.has(s.movie_name)) return false
        seen.add(s.movie_name)
        return true
      }).slice(0, 8)
    } catch {
      // silently fail
    } finally {
      loadingMovies.value = false
    }
  }
})

onUnmounted(() => clearInterval(slideTimer))

const gradients = [
  'linear-gradient(135deg,#0d0221,#3a0a6e)',
  'linear-gradient(135deg,#021520,#0a4a7f)',
  'linear-gradient(135deg,#1a0800,#6e2a00)',
  'linear-gradient(135deg,#001a0a,#006e2a)',
  'linear-gradient(135deg,#1a0015,#6e003a)',
  'linear-gradient(135deg,#0a0a1a,#2a2a6e)',
]
function posterBg(name) {
  let h=0; for(const c of name) h=(h*31+c.charCodeAt(0))&0xffff
  return gradients[h%gradients.length]
}
function formatTime(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleTimeString(i18n.lang==='th'?'th-TH':'en-US',{hour:'2-digit',minute:'2-digit',hour12:false})
}

const features = computed(() => [
  { emoji:'⚡', title: i18n.lang==='th'?'Real-time Seat Map':'Real-time Seat Map',     desc: i18n.lang==='th'?'เห็นสถานะที่นั่งอัปเดตสดๆ':'Live seat availability updates' },
  { emoji:'🔐', title: i18n.lang==='th'?'ระบบล็อคที่นั่ง':'Distributed Lock',         desc: i18n.lang==='th'?'Redis lock ป้องกันการจองซ้ำ':'Redis prevents double bookings' },
  { emoji:'🎟', title: i18n.lang==='th'?'ยืนยันทันที':'Instant Confirmation',         desc: i18n.lang==='th'?'จ่ายและยืนยันในไม่กี่วินาที':'Pay and confirm in seconds' },
  { emoji:'📱', title: i18n.lang==='th'?'รองรับทุกอุปกรณ์':'All Devices',             desc: i18n.lang==='th'?'ใช้งานได้บนมือถือและคอมพิวเตอร์':'Mobile and desktop friendly' },
])
</script>

<style scoped>
.home-page { overflow-x:hidden; }

.hero-slider { position:relative; }
.slide {
  min-height:500px; display:flex; align-items:center;
  position:relative; overflow:hidden; padding:4rem 4rem 5rem;
}
.slide-overlay {
  position:absolute; inset:0;
  background:linear-gradient(to right, rgba(0,0,0,0.8) 0%, rgba(0,0,0,0.3) 60%, transparent 100%);
}
.slide-content { position:relative; z-index:2; max-width:540px; }
.slide-meta { margin-bottom:1rem; }
.slide-tag {
  background:rgba(245,166,35,0.2); border:1px solid rgba(245,166,35,0.4);
  color:var(--accent); font-size:0.72rem; font-weight:800;
  letter-spacing:0.12em; text-transform:uppercase;
  padding:0.25rem 0.85rem; border-radius:4px;
}
.slide-title {
  font-family:'Syne',sans-serif; font-size:clamp(2rem,4vw,3.5rem);
  font-weight:800; color:white; line-height:1.1; margin-bottom:0.75rem;
  text-shadow:0 2px 20px rgba(0,0,0,0.5);
}
.slide-sub { color:rgba(255,255,255,0.7); font-size:1rem; margin-bottom:2rem; line-height:1.6; }
.slide-actions { display:flex; gap:1rem; }
.btn-hero {
  padding:0.85rem 2rem; font-size:1rem; font-weight:700;
  display:inline-flex; align-items:center; gap:0.5rem; border-radius:12px;
}
.g-icon {
  width:20px; height:20px; background:white; color:#4285f4;
  border-radius:3px; display:inline-flex; align-items:center; justify-content:center;
  font-size:0.78rem; font-weight:900;
}
.slide-poster {
  position:absolute; right:8%; top:50%; transform:translateY(-50%);
  font-size:clamp(6rem,12vw,12rem);
  filter:drop-shadow(0 8px 40px rgba(0,0,0,0.6));
  animation:float 4s ease-in-out infinite; z-index:1; pointer-events:none;
}
@keyframes float{0%,100%{transform:translateY(-50%);}50%{transform:translateY(calc(-50% - 14px));}}

.slide-dots {
  position:absolute; bottom:1.5rem; left:50%; transform:translateX(-50%);
  display:flex; gap:0.5rem; z-index:3;
}
.dot {
  width:8px; height:8px; border-radius:50%; background:rgba(255,255,255,0.3);
  border:none; cursor:pointer; transition:all 0.25s; padding:0;
}
.dot.active { width:24px; border-radius:4px; background:var(--accent); }

/* Now showing */
.now-showing-section { max-width:1200px; margin:0 auto; padding:2.5rem 2rem; }
.section-header { display:flex; align-items:center; justify-content:space-between; margin-bottom:1.5rem; }
.section-title { font-family:'Syne',sans-serif; font-size:1.4rem; font-weight:800; color:var(--text-primary); }
.see-all { font-size:0.85rem; font-weight:600; color:var(--accent); }

.login-prompt {
  text-align:center; padding:3rem 2rem;
  background:var(--bg-card); border:1px solid var(--border); border-radius:16px;
  color:var(--text-secondary);
}

.empty-movies { text-align:center; padding:2rem; color:var(--text-muted); }

.movies-row {
  display:grid;
  grid-template-columns:repeat(auto-fill, minmax(145px,1fr));
  gap:1rem;
}
.movie-card-mini {
  background:var(--bg-card); border:1px solid var(--border);
  border-radius:12px; overflow:hidden; text-decoration:none; color:inherit;
  transition:all 0.2s;
}
.movie-card-mini:hover { transform:translateY(-4px); border-color:rgba(245,166,35,0.3); box-shadow:0 8px 24px rgba(0,0,0,0.3); }
.mini-poster {
  height:120px; display:flex; align-items:center; justify-content:center; position:relative;
}
.mini-emoji { font-size:2.5rem; filter:drop-shadow(0 2px 8px rgba(0,0,0,0.5)); }
.mini-overlay { position:absolute; bottom:0.4rem; left:0.4rem; }
.mini-rating {
  background:rgba(0,0,0,0.7); color:white;
  font-size:0.6rem; font-weight:800; padding:0.1rem 0.4rem; border-radius:3px;
}
.mini-info { padding:0.65rem 0.75rem; }
.mini-title {
  font-weight:700; font-size:0.78rem; color:var(--text-primary); line-height:1.3; margin-bottom:0.2rem;
  display:-webkit-box; -webkit-line-clamp:2; -webkit-box-orient:vertical; overflow:hidden;
}
.mini-genre { font-size:0.68rem; color:var(--accent); font-weight:600; margin-bottom:0.15rem; }
.mini-time  { font-size:0.7rem; color:var(--text-muted); }

/* Features */
.features-section {
  display:grid; grid-template-columns:repeat(4,1fr);
  border-top:1px solid var(--border); background:var(--bg-secondary);
}
.feature-item {
  padding:2rem; text-align:center; border-right:1px solid var(--border); transition:background 0.2s;
}
.feature-item:last-child { border-right:none; }
.feature-item:hover { background:var(--bg-hover); }
.feature-emoji { font-size:2rem; margin-bottom:0.75rem; }
.feature-item h3 { font-weight:700; font-size:0.9rem; color:var(--text-primary); margin-bottom:0.4rem; }
.feature-item p  { font-size:0.8rem; color:var(--text-secondary); line-height:1.5; }

@media(max-width:768px) {
  .slide { padding:2rem; min-height:400px; }
  .slide-poster { display:none; }
  .features-section { grid-template-columns:repeat(2,1fr); }
  .feature-item:nth-child(2) { border-right:none; }
}
</style>
