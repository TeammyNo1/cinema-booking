<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">{{ i18n.lang==='th'?'จัดการรอบหนัง':'Manage Showtimes' }}</h1>
      <button class="btn-primary" @click="openCreate">
        + {{ i18n.lang==='th'?'เพิ่มรอบหนัง':'Add Showtime' }}
      </button>
    </div>

    <!-- Showtimes Table -->
    <div class="table-card">
      <div v-if="loading" class="loading-spinner"><div class="spinner-ring"></div></div>
      <table v-else>
        <thead>
          <tr>
            <th>{{ i18n.lang==='th'?'หนัง':'Movie' }}</th>
            <th>{{ i18n.lang==='th'?'โรง':'Hall' }}</th>
            <th>{{ i18n.lang==='th'?'ประเภท':'Genre' }}</th>
            <th>{{ i18n.lang==='th'?'เวลาเริ่ม':'Start' }}</th>
            <th>{{ i18n.lang==='th'?'เวลาจบ':'End' }}</th>
            <th>{{ i18n.lang==='th'?'จัดการ':'Actions' }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="st in showtimes" :key="st.id">
            <td>
              <div class="movie-cell">
                <span class="movie-emoji">{{ st.poster_emoji || '🎬' }}</span>
                <div>
                  <p class="movie-name-cell">{{ st.movie_name }}</p>
                  <p class="movie-sub">{{ st.rating }} · {{ st.duration }}{{ i18n.lang==='th'?'นาที':'min' }}</p>
                </div>
              </div>
            </td>
            <td><span class="hall-badge">{{ st.hall }}</span></td>
            <td><span class="genre-badge">{{ st.genre }}</span></td>
            <td>{{ formatDateTime(st.start_time) }}</td>
            <td>{{ formatDateTime(st.end_time) }}</td>
            <td>
              <div class="action-btns">
                <button class="btn-edit" @click="openEdit(st)">✏️</button>
                <button class="btn-delete" @click="confirmDelete(st)">🗑️</button>
              </div>
            </td>
          </tr>
          <tr v-if="showtimes.length === 0">
            <td colspan="6" class="empty-row">{{ i18n.lang==='th'?'ไม่มีรอบหนัง':'No showtimes found' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal: Create / Edit -->
    <Teleport to="body">
      <div class="modal-overlay" v-if="showModal" @click.self="closeModal">
        <div class="modal">
          <div class="modal-header">
            <h2 class="modal-title">{{ isEditing ? (i18n.lang==='th'?'แก้ไขรอบหนัง':'Edit Showtime') : (i18n.lang==='th'?'เพิ่มรอบหนัง':'Add Showtime') }}</h2>
            <button class="modal-close" @click="closeModal">✕</button>
          </div>

          <div class="modal-body">
            <div class="form-grid">
              <!-- Movie Name -->
              <div class="form-group full">
                <label>{{ i18n.lang==='th'?'ชื่อหนัง':'Movie Name' }} *</label>
                <input v-model="form.movie_name" placeholder="e.g. Avengers: Secret Wars" />
              </div>

              <!-- Poster Emoji + Genre -->
              <div class="form-group">
                <label>{{ i18n.lang==='th'?'Emoji โปสเตอร์':'Poster Emoji' }}</label>
                <div class="emoji-picker-row">
                  <input v-model="form.poster_emoji" style="width:80px;text-align:center;font-size:1.3rem" maxlength="2" />
                  <div class="emoji-presets">
                    <button v-for="e in emojiPresets" :key="e" class="emoji-btn" @click="form.poster_emoji=e">{{ e }}</button>
                  </div>
                </div>
              </div>

              <div class="form-group">
                <label>{{ i18n.lang==='th'?'ประเภท':'Genre' }}</label>
                <select v-model="form.genre">
                  <option v-for="g in genres" :key="g">{{ g }}</option>
                </select>
              </div>

              <!-- Hall + Rating -->
              <div class="form-group">
                <label>{{ i18n.lang==='th'?'โรง':'Hall' }} *</label>
                <select v-model="form.hall">
                  <option v-for="h in halls" :key="h">{{ h }}</option>
                </select>
              </div>

              <div class="form-group">
                <label>{{ i18n.lang==='th'?'เรทติ้ง':'Rating' }}</label>
                <select v-model="form.rating">
                  <option v-for="r in ratings" :key="r">{{ r }}</option>
                </select>
              </div>

              <!-- Duration -->
              <div class="form-group">
                <label>{{ i18n.lang==='th'?'ระยะเวลา (นาที)':'Duration (min)' }}</label>
                <input v-model.number="form.duration" type="number" min="60" max="300" />
              </div>

              <!-- Start / End time -->
              <div class="form-group">
                <label>{{ i18n.lang==='th'?'เวลาเริ่ม':'Start Time' }} *</label>
                <input v-model="form.start_time_local" type="datetime-local" />
              </div>

              <div class="form-group">
                <label>{{ i18n.lang==='th'?'เวลาจบ':'End Time' }} *</label>
                <input v-model="form.end_time_local" type="datetime-local" />
              </div>

              <!-- Seats config (only on create) -->
              <template v-if="!isEditing">
                <div class="form-group">
                  <label>{{ i18n.lang==='th'?'จำนวนแถว':'Rows Count' }}</label>
                  <input v-model.number="form.rows_count" type="number" min="2" max="26" />
                </div>
                <div class="form-group">
                  <label>{{ i18n.lang==='th'?'ที่นั่งต่อแถว':'Seats per Row' }}</label>
                  <input v-model.number="form.seats_per_row" type="number" min="5" max="30" />
                </div>
                <div class="form-group">
                  <label>{{ i18n.lang==='th'?'ราคาปกติ (฿)':'Normal Price (฿)' }}</label>
                  <input v-model.number="form.price_normal" type="number" min="100" />
                </div>
                <div class="form-group">
                  <label>{{ i18n.lang==='th'?'ราคา VIP (฿)':'VIP Price (฿)' }}</label>
                  <input v-model.number="form.price_vip" type="number" min="100" />
                </div>

                <!-- Preview -->
                <div class="form-group full">
                  <label>{{ i18n.lang==='th'?'ตัวอย่างโรง':'Seat Preview' }}</label>
                  <div class="seat-preview-mini">
                    <div class="preview-screen">SCREEN</div>
                    <div v-for="r in previewRows" :key="r.row" class="preview-row">
                      <span class="preview-label">{{ r.row }}</span>
                      <div v-for="n in form.seats_per_row" :key="n" :class="`preview-seat ${r.isVIP?'pvip':'pnorm'}`"></div>
                    </div>
                    <p class="preview-info">{{ previewRows.length }} {{ i18n.lang==='th'?'แถว':'rows' }} × {{ form.seats_per_row }} = {{ previewRows.length * form.seats_per_row }} {{ i18n.lang==='th'?'ที่นั่ง':'seats' }}</p>
                  </div>
                </div>
              </template>
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn-ghost" @click="closeModal">{{ i18n.lang==='th'?'ยกเลิก':'Cancel' }}</button>
            <button class="btn-primary" :disabled="saving" @click="save">
              <span v-if="saving" class="spinner-ring" style="width:14px;height:14px;border-width:2px"></span>
              {{ saving ? (i18n.lang==='th'?'กำลังบันทึก...':'Saving...') : (i18n.lang==='th'?'บันทึก':'Save') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Confirm Delete Dialog -->
    <Teleport to="body">
      <div class="modal-overlay" v-if="deleteTarget" @click.self="deleteTarget=null">
        <div class="modal modal-sm">
          <div class="modal-header">
            <h2 class="modal-title">{{ i18n.lang==='th'?'ยืนยันการลบ':'Confirm Delete' }}</h2>
          </div>
          <div class="modal-body">
            <p class="delete-msg">
              {{ i18n.lang==='th'?'ต้องการลบรอบ':'Delete showtime for' }}
              <strong>{{ deleteTarget?.movie_name }}</strong>
              {{ i18n.lang==='th'?'ใช่หรือไม่?':'?' }}
            </p>
            <p class="delete-warn">⚠️ {{ i18n.lang==='th'?'ที่นั่งทั้งหมดในรอบนี้จะถูกลบด้วย':'All seats in this showtime will be deleted.' }}</p>
          </div>
          <div class="modal-footer">
            <button class="btn-ghost" @click="deleteTarget=null">{{ i18n.lang==='th'?'ยกเลิก':'Cancel' }}</button>
            <button class="btn-danger" :disabled="deleting" @click="doDelete">
              {{ deleting ? '...' : (i18n.lang==='th'?'ลบเลย':'Delete') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18nStore } from '@/stores/i18n'
import { useToastStore } from '@/stores/toast'
import api from '@/composables/useApi'

const i18n = useI18nStore()
const toast = useToastStore()

const showtimes = ref([])
const loading = ref(true)
const showModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)
const deleteTarget = ref(null)
const deleting = ref(false)
const editingId = ref(null)

const emojiPresets = ['⚡','🏜️','🕷️','🌀','🦇','🚀','🤫','💥','🎭','🌊','👻','🐉','🎪','🌙','🔥']
const genres  = ['Action','Sci-Fi','Drama','Horror','Animation','Romance','Thriller','Comedy','Documentary']
const halls   = ['Hall A','Hall B','Hall C','Hall D','Hall E','IMAX','VIP Cinema','4DX']
const ratings = ['G','PG','PG-13','R','NC-17','U13+','U15+','U18+']

const defaultForm = () => ({
  movie_name: '', hall: 'Hall A', poster_emoji: '🎬',
  genre: 'Action', rating: 'PG-13', duration: 120,
  start_time_local: '', end_time_local: '',
  rows_count: 8, seats_per_row: 12,
  price_normal: 300, price_vip: 400,
})

const form = ref(defaultForm())

const previewRows = computed(() => {
  const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  return Array.from({ length: Math.min(form.value.rows_count, 10) }, (_, i) => ({
    row: letters[i], isVIP: i < 2
  }))
})

onMounted(loadShowtimes)

async function loadShowtimes() {
  loading.value = true
  try {
    const res = await api.get('/api/admin/showtimes')
    showtimes.value = res.data || []
  } finally { loading.value = false }
}

function openCreate() {
  isEditing.value = false
  editingId.value = null
  form.value = defaultForm()
  // Default start = 2 hours from now
  const d = new Date(Date.now() + 2*3600*1000)
  form.value.start_time_local = toLocalInput(d)
  const e = new Date(d.getTime() + 120*60*1000)
  form.value.end_time_local = toLocalInput(e)
  showModal.value = true
}

function openEdit(st) {
  isEditing.value = true
  editingId.value = st.id
  form.value = {
    ...defaultForm(),
    movie_name: st.movie_name,
    hall: st.hall,
    poster_emoji: st.poster_emoji || '🎬',
    genre: st.genre || 'Action',
    rating: st.rating || 'PG-13',
    duration: st.duration || 120,
    start_time_local: toLocalInput(new Date(st.start_time)),
    end_time_local: toLocalInput(new Date(st.end_time)),
  }
  showModal.value = true
}

function closeModal() { showModal.value = false }

async function save() {
  if (!form.value.movie_name || !form.value.hall || !form.value.start_time_local || !form.value.end_time_local) {
    toast.error(i18n.lang==='th'?'กรุณากรอกข้อมูลที่จำเป็นให้ครบ':'Please fill in all required fields')
    return
  }
  saving.value = true
  try {
    const payload = {
      movie_name:   form.value.movie_name,
      hall:         form.value.hall,
      poster_emoji: form.value.poster_emoji,
      genre:        form.value.genre,
      rating:       form.value.rating,
      duration:     form.value.duration,
      start_time:   new Date(form.value.start_time_local).toISOString(),
      end_time:     new Date(form.value.end_time_local).toISOString(),
      rows_count:   form.value.rows_count,
      seats_per_row:form.value.seats_per_row,
      price_normal: form.value.price_normal,
      price_vip:    form.value.price_vip,
    }
    if (isEditing.value) {
      await api.put(`/api/admin/showtimes/${editingId.value}`, payload)
      toast.success(i18n.lang==='th'?'แก้ไขรอบหนังสำเร็จ!':'Showtime updated!')
    } else {
      await api.post('/api/admin/showtimes', payload)
      toast.success(i18n.lang==='th'?`เพิ่มรอบหนัง "${form.value.movie_name}" สำเร็จ!`:`Showtime "${form.value.movie_name}" created!`)
    }
    closeModal()
    await loadShowtimes()
  } catch (err) {
    toast.error(err.response?.data?.error || 'Error saving showtime')
  } finally { saving.value = false }
}

function confirmDelete(st) { deleteTarget.value = st }

async function doDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await api.delete(`/api/admin/showtimes/${deleteTarget.value.id}`)
    toast.success(i18n.lang==='th'?'ลบรอบหนังสำเร็จ':'Showtime deleted')
    deleteTarget.value = null
    await loadShowtimes()
  } catch (err) {
    toast.error(err.response?.data?.error || 'Delete failed')
  } finally { deleting.value = false }
}

function toLocalInput(d) {
  const pad = n => String(n).padStart(2,'0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function formatDateTime(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString(i18n.lang==='th'?'th-TH':'en-US', { month:'short', day:'numeric', hour:'2-digit', minute:'2-digit' })
}
</script>

<style scoped>
.table-card { background:var(--bg-card);border:1px solid var(--border);border-radius:14px;overflow:hidden;margin-top:1rem; }
.movie-cell { display:flex;align-items:center;gap:0.75rem; }
.movie-emoji { font-size:1.5rem; }
.movie-name-cell { font-weight:700;font-size:0.9rem;color:var(--text-primary);margin-bottom:0.15rem; }
.movie-sub { font-size:0.72rem;color:var(--text-muted); }
.hall-badge  { background:rgba(52,152,219,0.1);border:1px solid rgba(52,152,219,0.25);color:#3498db;padding:0.2rem 0.6rem;border-radius:6px;font-size:0.75rem;font-weight:700; }
.genre-badge { background:rgba(245,166,35,0.1);border:1px solid rgba(245,166,35,0.25);color:var(--accent);padding:0.2rem 0.6rem;border-radius:6px;font-size:0.75rem;font-weight:700; }
.action-btns { display:flex;gap:0.4rem; }
.btn-edit, .btn-delete { width:32px;height:32px;border-radius:8px;display:flex;align-items:center;justify-content:center;font-size:0.9rem;border:1px solid var(--border);background:var(--bg-secondary);cursor:pointer;transition:all 0.15s; }
.btn-edit:hover   { background:rgba(52,152,219,0.1);border-color:rgba(52,152,219,0.3); }
.btn-delete:hover { background:rgba(231,76,60,0.1);border-color:rgba(231,76,60,0.3); }
.empty-row { text-align:center;color:var(--text-muted);padding:3rem !important; }

/* Modal */
.modal-overlay {
  position:fixed;inset:0;background:rgba(0,0,0,0.75);
  backdrop-filter:blur(4px);z-index:1000;
  display:flex;align-items:center;justify-content:center;padding:1rem;
}
.modal {
  background:var(--bg-card);border:1px solid var(--border2);
  border-radius:20px;width:100%;max-width:680px;max-height:90vh;
  display:flex;flex-direction:column;overflow:hidden;
  animation:modalIn 0.25s cubic-bezier(0.34,1.56,0.64,1);
}
.modal-sm { max-width:420px; }
@keyframes modalIn { from{transform:scale(0.9);opacity:0} to{transform:scale(1);opacity:1} }

.modal-header {
  display:flex;align-items:center;justify-content:space-between;
  padding:1.25rem 1.5rem;border-bottom:1px solid var(--border);
}
.modal-title { font-family:'Syne',sans-serif;font-size:1.1rem;font-weight:800; }
.modal-close {
  width:32px;height:32px;border-radius:8px;border:1px solid var(--border);
  background:var(--bg-secondary);color:var(--text-secondary);cursor:pointer;font-size:0.9rem;
  display:flex;align-items:center;justify-content:center;
}
.modal-close:hover { background:var(--bg-hover);color:var(--text-primary); }

.modal-body { padding:1.5rem;overflow-y:auto;flex:1; }
.modal-footer {
  padding:1rem 1.5rem;border-top:1px solid var(--border);
  display:flex;justify-content:flex-end;gap:0.75rem;
}

/* Form */
.form-grid { display:grid;grid-template-columns:1fr 1fr;gap:1rem; }
.form-group { display:flex;flex-direction:column;gap:0.4rem; }
.form-group.full { grid-column:1/-1; }
.form-group label { font-size:0.72rem;color:var(--text-muted);font-weight:700;letter-spacing:0.06em;text-transform:uppercase; }
.form-group input, .form-group select { width:100%; }

.emoji-picker-row { display:flex;gap:0.5rem;align-items:center;flex-wrap:wrap; }
.emoji-presets { display:flex;flex-wrap:wrap;gap:0.25rem;flex:1; }
.emoji-btn { width:28px;height:28px;border-radius:6px;border:1px solid var(--border);background:var(--bg-secondary);cursor:pointer;font-size:1rem;display:flex;align-items:center;justify-content:center;transition:all 0.1s; }
.emoji-btn:hover { background:var(--bg-hover);border-color:var(--accent); }

/* Seat preview */
.seat-preview-mini {
  background:var(--bg-secondary);border:1px solid var(--border);
  border-radius:10px;padding:1rem;overflow-x:auto;
}
.preview-screen {
  background:rgba(255,255,255,0.05);border:1px solid rgba(255,255,255,0.1);
  text-align:center;padding:0.25rem;font-size:0.55rem;letter-spacing:0.25em;color:rgba(255,255,255,0.3);
  border-radius:4px;margin-bottom:0.5rem;max-width:300px;margin-left:auto;margin-right:auto;
}
.preview-row { display:flex;align-items:center;gap:2px;margin-bottom:2px;justify-content:center; }
.preview-label { font-size:0.55rem;color:var(--text-muted);width:12px;text-align:center;flex-shrink:0; }
.preview-seat { width:10px;height:8px;border-radius:2px 2px 1px 1px;flex-shrink:0; }
.pnorm { background:rgba(40,90,50,0.6);border:1px solid rgba(60,150,80,0.4); }
.pvip  { background:rgba(80,60,10,0.6);border:1px solid rgba(180,140,20,0.4); }
.preview-info { text-align:center;font-size:0.72rem;color:var(--text-muted);margin-top:0.5rem; }

/* Delete dialog */
.delete-msg { font-size:0.95rem;color:var(--text-secondary);margin-bottom:0.75rem; }
.delete-msg strong { color:var(--text-primary); }
.delete-warn { font-size:0.82rem;color:var(--warning);background:rgba(243,156,18,0.08);border:1px solid rgba(243,156,18,0.2);border-radius:8px;padding:0.6rem 0.85rem; }
</style>
