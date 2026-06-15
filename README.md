# 🎬 CineBook — Cinema Ticket Booking System

> ระบบจองตั๋วโรงภาพยนตร์แบบ Real-time พร้อม Distributed Locking, WebSocket, Multi-seat selection, Admin CRUD และรองรับ 2 ภาษา (ไทย/English)

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         User Browser                                  │
│              Vue 3 + Pinia + Vue Router + WebSocket                   │
│         (Home / Showtimes / Seat Map / My Tickets / Admin)           │
└────────────────────────────┬──────────────────────────────────────────┘
                              │ HTTP / WebSocket
                     ┌────────▼────────┐
                     │   Nginx (:80)   │  ← Serve SPA + Reverse Proxy
                     └────────┬────────┘
                              │
                     ┌────────▼────────┐
                     │  Go/Gin Backend │  :8080
                     │                 │
                     │  ┌───────────┐  │
                     │  │ Auth      │  │  Google OAuth 2.0 + JWT (Role: USER/ADMIN)
                     │  │ Showtime  │  │  Public list + Admin CRUD (Create/Edit/Delete)
                     │  │ Booking   │  │  Multi-seat lock / confirm / cancel
                     │  │ Admin     │  │  Dashboard, Stats, Audit Logs
                     │  │ WebSocket │  │  Realtime seat map broadcast hub
                     │  └───────────┘  │
                     └──┬──────┬───┬───┘
                        │      │   │
           ┌────────────▼─┐ ┌──▼──┐ ┌▼──────────┐
           │   MongoDB    │ │Redis│ │ RabbitMQ  │
           │  (Data Store)│ │(Lock│ │  (Events) │
           │              │ │+TTL)│ │           │
           │  users       │ │     │ │ booking_  │
           │  showtimes   │ │seat_│ │ events    │
           │  seats       │ │lock:│ │ audit_    │
           │  bookings    │ │{id}:│ │ logs      │
           │  audit_logs  │ │{cod}│ │ notifs    │
           └──────────────┘ └─────┘ └───────────┘
```

---

## Tech Stack

| Layer         | Technology                                  |
|---------------|----------------------------------------------|
| Backend       | Go 1.21, Gin framework                        |
| Frontend      | Vue 3, Vite, Pinia, Vue Router                |
| Database      | MongoDB 7                                     |
| Cache / Lock  | Redis 7 (Distributed Lock via `SET NX EX`)    |
| Realtime      | WebSocket (gorilla/websocket)                 |
| Message Queue | RabbitMQ 3.13                                 |
| Auth          | Google OAuth 2.0 + JWT (HS256)                |
| i18n          | Custom Pinia store (Thai 🇹🇭 / English 🇬🇧)    |
| Deployment    | Docker + docker-compose                       |
| Web Server    | Nginx (frontend SPA + reverse proxy)          |

---

## ✨ Features Overview

### 👤 User
- เข้าสู่ระบบด้วย Google OAuth 2.0 → ได้ `user_id` + JWT token
- หน้าแรก — Hero slider + featured movies (เข้าได้โดยไม่ต้อง login)
- เลือกดูหนัง — จัดกลุ่มตามชื่อหนัง พร้อม genre filter และค้นหา
- **หน้าเลือกที่นั่งสไตล์โรงหนังจริง** — เลือกได้หลายที่นั่งพร้อมกัน (ไม่จำกัด 1 ที่นั่ง)
  - ที่นั่ง 2 ระดับ: **Normal** และ **VIP** (ราคาต่างกัน)
  - แสดงสถานะ Real-time: AVAILABLE / SELECTED / LOCKED / BOOKED
  - Summary panel พร้อม countdown 5 นาที และราคารวม
- ดูตั๋วของตัวเองในหน้า "My Tickets" (ดีไซน์เป็นบัตรตั๋วจริง)
- สลับภาษาไทย/อังกฤษได้ทันทีทุกหน้า (ปุ่ม TH/EN ที่ navbar)

### 🧑‍💼 Admin
- **Dashboard** — สถิติการจอง (Total / Confirmed / Locked / Timeout) + filter ตาม status และ showtime
- **จัดการรอบหนัง (Showtime CRUD)**
  - เพิ่ม / แก้ไข / ลบ รอบหนัง
  - ตั้งค่า: ชื่อหนัง, emoji โปสเตอร์, genre, rating, ระยะเวลา, โรง, เวลาเริ่ม-จบ
  - กำหนดจำนวนแถว/ที่นั่งต่อแถว และราคา Normal/VIP (พร้อม seat preview ก่อนสร้าง)
  - ลบรอบจะลบที่นั่งทั้งหมดที่เกี่ยวข้องด้วย (ป้องกันลบถ้ามี booking ที่ LOCKED อยู่)
- **Audit Logs** — ดู event ทั้งหมด (BOOKING_SUCCESS, BOOKING_TIMEOUT, SEAT_RELEASED, SYSTEM_ERROR, BOOKING_LOCKED)

---

## Booking Flow (Multi-seat)

```
ผู้ใช้เลือกหลายที่นั่งพร้อมกัน (คลิกสะสม)
      │
      ▼
กด "Hold Seats (N)" → วน loop เรียก POST /api/bookings/lock ทีละที่นั่ง
      │
      ├─► ตรวจสอบสถานะที่นั่งใน MongoDB → ต้องเป็น AVAILABLE
      │
      ├─► Redis SET NX EX 300   ◄─── Atomic Distributed Lock
      │     key:   seat_lock:{showtimeID}:{seatCode}
      │     value: userID
      │     TTL:   300s (5 นาที)
      │
      ├─► Lock สำเร็จ:
      │     - อัปเดตที่นั่ง → LOCKED ใน MongoDB
      │     - สร้าง Booking record (status=LOCKED, expires_at=now+5min)
      │     - Broadcast ผ่าน WebSocket → ทุกคนที่ดูรอบนี้เห็นที่นั่งเป็น LOCKED
      │     - Spawn goroutine สำหรับ auto-release หลัง TTL หมด
      │
      └─► Lock ไม่สำเร็จ (ที่นั่งถูกจองไปแล้ว):
            - แสดง toast error เฉพาะที่นั่งนั้น แต่ที่นั่งอื่นที่ lock สำเร็จยังดำเนินต่อ

            ┌──────────────────────────────────┐
            │  Countdown 5 นาที เริ่มทำงาน      │
            │  แสดงใน Summary Panel             │
            └──────────────────────────────────┘

ผู้ใช้กด "Confirm & Pay"
      │
      ▼
วน loop เรียก POST /api/bookings/:id/confirm ทุก booking ที่ active
      │
      ├─► ตรวจสอบว่า booking เป็นของ user คนนี้
      ├─► ตรวจสอบ Redis lock ยังถูกต้อง (TTL > 0, owner == userID)
      ├─► อัปเดต Booking → CONFIRMED, ที่นั่ง → BOOKED
      ├─► ลบ Redis lock
      ├─► Broadcast ผ่าน WebSocket → BOOKED
      ├─► Publish ไป RabbitMQ: booking_events (BOOKING_CONFIRMED)
      ├─► Publish ไป RabbitMQ: notifications (mock email)
      └─► เขียน AuditLog: BOOKING_SUCCESS

หาก 5 นาทีผ่านไปโดยไม่กดยืนยัน:
      │
      ├─► Goroutine ที่ spawn ไว้ตอน lock ตื่นขึ้นมา
      ├─► เช็คว่า booking ยังเป็น LOCKED อยู่ → ใช่
      ├─► อัปเดต Booking → TIMEOUT, ที่นั่ง → AVAILABLE
      ├─► Broadcast ผ่าน WebSocket → AVAILABLE
      ├─► Publish ไป RabbitMQ: BOOKING_TIMEOUT
      └─► เขียน AuditLog: BOOKING_TIMEOUT + SEAT_RELEASED
```

---

## Redis Lock Strategy

### คำสั่งหลัก

```
SET seat_lock:{showtimeID}:{seatCode} {userID} NX EX 300
```

- **NX** (Not eXists) — set ได้เมื่อ key ยังไม่มีเท่านั้น เป็น atomic operation ระดับเดียว แม้มี 1,000 คนคลิกที่นั่งเดียวกันพร้อมกัน มีคนเดียวเท่านั้นที่ได้ `OK`
- **EX 300** — TTL อัตโนมัติ 300 วินาที ถ้า server crash lock จะปล่อยตัวเองโดยไม่มี orphaned lock
- **Value = userID** — ฝัง ownership ไว้ใน lock เลย ใช้ตรวจสอบสิทธิ์ confirm/cancel ได้ทันที

### ทำไมไม่ใช้ SETNX + EXPIRE แยกกัน?

`SETNX` แล้วตามด้วย `EXPIRE` เป็น 2 operation แยกกัน ไม่ atomic — ถ้า crash ระหว่างกลางอาจเหลือ lock ที่ไม่มี TTL ค้างอยู่ การใช้ `SET key value NX EX` (Redis ≥ 2.6.12) รวมเป็น 1 คำสั่ง atomic เสมอ

### Lock Lifecycle

```
AVAILABLE ──[user selects + lock]──► LOCKED (Redis key สร้างขึ้น, TTL=300s)
    │                                       │
    │              ┌────────────────────────┤
    │              ▼                        ▼
    │        User confirms              TTL หมดอายุ
    │              │                     (goroutine)
    │        ลบ Redis key                ลบ Redis key อัตโนมัติด้วย TTL
    │              │                          │
    └──◄───────────┘                    seat → AVAILABLE
         BOOKED                          broadcast WS
```

### ป้องกัน Double Booking

MongoDB seat status + Redis lock ทำงานร่วมกัน:
1. Redis `NX` ป้องกันสองคนล็อกที่นั่งเดียวกันในเวลาเดียวกัน
2. ถ้า Redis ล้มเหลว MongoDB seat status check ก็จะปฏิเสธที่นั่งที่ไม่ใช่ AVAILABLE ก่อนพยายาม lock อยู่ดี

---

## Message Queue Usage (RabbitMQ)

ระบบ declare 3 queues ตอน startup:

| Queue            | Producer                  | Consumer (mock)        | ใช้ทำอะไร                          |
|------------------|---------------------------|--------------------------|------------------------------------|
| `booking_events` | Confirm / Timeout         | Log consumer            | Audit trail, analytics pipeline    |
| `audit_logs`     | ทุก booking action         | Log consumer            | บันทึก audit log แบบ async         |
| `notifications`  | Booking confirm           | Notification consumer   | Mock แจ้งเตือนอีเมล/Line           |

**เหตุผลที่ใช้ async:** การส่ง notification ไม่ critical และอาจช้า — publish ไป RabbitMQ แบบ fire-and-forget ทำให้ HTTP response กลับไปที่ user ทันที โดย consumer จัดการ delivery ใน background

**สิ่งที่ publish:**
- `BOOKING_CONFIRMED` → booking_events + notifications
- `BOOKING_TIMEOUT`   → booking_events
- ทุก audit event      → audit_logs (async, ไม่ block response)

---

## วิธีรันระบบ

### Prerequisites
- Docker ≥ 24
- Docker Compose v2
- Docker Desktop ต้องเปิดอยู่ (Windows/Mac)

### 1. Config environment

```bash
cp .env.example .env
```

แก้ไข `.env`:
```env
GOOGLE_CLIENT_ID=xxxxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-xxxxxxxxxxxx
GOOGLE_REDIRECT_URL=http://localhost/api/auth/google/callback

JWT_SECRET=your-strong-random-secret

ADMIN_EMAILS=youremail@gmail.com

SEAT_LOCK_TTL=300
```

> 📌 ดูวิธีตั้งค่า Google OAuth: ไปที่ [Google Cloud Console](https://console.cloud.google.com/apis/credentials) → สร้าง OAuth Client ID → เพิ่ม Authorized redirect URI: `http://localhost/api/auth/google/callback`

### 2. รันทุกอย่างด้วยคำสั่งเดียว

```bash
docker compose up --build
```

รอจนเห็น:
```
cinema_backend   | 🎬 Cinema Booking API running on :8080
cinema_frontend  | /docker-entrypoint.sh: Configuration complete; ready for start up
```

### 3. ตรวจสอบสถานะ

```bash
docker compose ps
```
ทุก service ต้องเป็น `running (healthy)`

### 4. เข้าใช้งาน

| Service          | URL                                   |
|-------------------|---------------------------------------|
| Frontend          | http://localhost                      |
| Backend API       | http://localhost/api                  |
| RabbitMQ UI       | http://localhost:15672 (guest/guest)  |

### 5. เป็น Admin

ใส่ Gmail ของคุณใน `ADMIN_EMAILS` ใน `.env` ก่อน build → login ด้วย Google email นั้น จะเห็นเมนู **Admin** ใน navbar

### 6. ล้างข้อมูลและเริ่มใหม่

```bash
docker compose down
docker volume rm cinema-booking_mongo_data   # ล้าง seed data เดิม เพื่อ regenerate
docker compose up --build
```

---

## API Reference

### Public (ไม่ต้อง login)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/api/auth/google` | Redirect ไป Google login |
| GET | `/api/auth/google/callback` | OAuth callback |
| GET | `/api/showtimes` | รายการรอบหนังทั้งหมด |
| GET | `/api/showtimes/:id/seats` | Seat map พร้อม lock status |
| WS  | `/ws/showtimes/:id` | WebSocket realtime seat updates |

### User (ต้อง login — Bearer token)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/auth/me` | ข้อมูล user ปัจจุบัน |
| POST | `/api/bookings/lock` | ล็อกที่นั่ง (5 นาที) |
| POST | `/api/bookings/:id/confirm` | ยืนยันการจ่ายเงิน |
| DELETE | `/api/bookings/:id` | ยกเลิก/ปล่อยที่นั่ง |
| GET | `/api/bookings/me` | ตั๋วของฉัน |

### Admin (ต้อง login + role=ADMIN)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/bookings` | รายการ booking ทั้งหมด (filter ได้) |
| GET | `/api/admin/audit-logs` | Audit log viewer |
| GET | `/api/admin/stats` | สถิติการจอง |
| GET | `/api/admin/showtimes` | รายการรอบหนัง (admin view) |
| GET | `/api/admin/showtimes/:id` | ดูรอบหนังตาม id |
| POST | `/api/admin/showtimes` | เพิ่มรอบหนังใหม่ (+ สร้างที่นั่งอัตโนมัติ) |
| PUT | `/api/admin/showtimes/:id` | แก้ไขรอบหนัง |
| DELETE | `/api/admin/showtimes/:id` | ลบรอบหนัง (+ ลบที่นั่งทั้งหมด) |

---

## Project Structure

```
cinema-booking/
├── backend/
│   ├── cmd/server/main.go          # Entry point, route setup (public vs protected routes)
│   ├── config/config.go            # Env-based configuration
│   └── internal/
│       ├── models/models.go        # Domain types (Seat มี Type: NORMAL/VIP, Showtime มี Genre/Rating/Duration)
│       ├── repository/
│       │   ├── mongodb.go          # MongoDB ops + Showtime/Seat CRUD + Seed (10 หนัง × 2 รอบ = 20 showtimes)
│       │   └── redis.go            # Redis distributed lock operations
│       ├── auth/handler.go         # Google OAuth + JWT
│       ├── middleware/auth.go      # JWT + role middleware
│       ├── booking/handler.go      # Multi-seat lock / confirm / cancel logic
│       ├── showtime/handler.go     # Showtime CRUD (admin) + public list
│       ├── admin/handler.go        # Admin dashboard, stats, audit logs API
│       ├── websocket/hub.go        # WS broadcast hub (per-showtime rooms)
│       └── queue/rabbitmq.go       # MQ publisher + mock consumers
├── frontend/
│   ├── src/
│   │   ├── App.vue
│   │   ├── main.js
│   │   ├── router/index.js          # รวม route admin/showtimes (CRUD)
│   │   ├── stores/
│   │   │   ├── auth.js              # Pinia auth store
│   │   │   ├── toast.js             # Toast notifications
│   │   │   └── i18n.js              # ไทย/English translations + toggle
│   │   ├── composables/
│   │   │   ├── useApi.js            # Axios + auth interceptor (public route aware)
│   │   │   └── useWebSocket.js      # WS with auto-reconnect
│   │   ├── views/
│   │   │   ├── HomeView.vue         # Hero slider + featured movies
│   │   │   ├── ShowtimesView.vue    # จัดกลุ่มหนัง + genre filter + search
│   │   │   ├── SeatMapView.vue      # หน้าเลือกที่นั่งสไตล์โรงหนัง (multi-select, VIP/Normal)
│   │   │   ├── MyBookingsView.vue   # ตั๋วของฉัน (ticket card design)
│   │   │   └── admin/
│   │   │       ├── AdminLayout.vue
│   │   │       ├── DashboardView.vue        # สถิติ + filter bookings
│   │   │       ├── ShowtimeManageView.vue   # CRUD รอบหนัง + seat preview
│   │   │       └── AuditLogsView.vue
│   │   └── components/
│   │       ├── NavBar.vue           # รวมปุ่มสลับภาษา TH/EN
│   │       └── ToastContainer.vue
│   ├── nginx.conf
│   └── Dockerfile
├── docker-compose.yml
├── .env.example
├── postman_collection.json
└── README.md
```

---

## Security

- **Role separation**: `USER` และ `ADMIN` แยกผ่าน JWT claims + middleware ทุก admin endpoint เรียก `AdminRequired()` ปฏิเสธ token ที่ไม่ใช่ admin ด้วย 403
- **Public vs Protected routes**: `/api/showtimes` และ `/api/showtimes/:id/seats` เป็น public (read-only) เพื่อให้แสดงหน้าแรกได้โดยไม่ต้อง login แต่ booking actions ทั้งหมดต้อง auth
- **Booking ownership**: Confirm/cancel ตรวจสอบ `booking.UserID == requestUserID` เสมอ
- **Seat lock ownership**: Redis lock value เก็บ `userID` — confirm จะเช็คว่า Redis owner ตรงกับ user ที่ request
- **JWT**: HS256, อายุ 24 ชม., secret มาจาก env (`JWT_SECRET`)
- **ไม่มี hardcoded secret**: ทุก credential ผ่าน `.env`
- **CORS**: จำกัดเฉพาะ `FRONTEND_URL` ที่ตั้งค่าไว้

---

## Assumptions & Trade-offs

| การตัดสินใจ | เหตุผล |
|---|---|
| Redis `SET NX EX` แทน Redlock | Redis เดี่ยวเพียงพอสำหรับ scale นี้ Redlock (multi-node) เพิ่มความซับซ้อนโดยไม่จำเป็น เว้นแต่ทำ Redis HA |
| Goroutine-based lock expiry | ง่ายและเพียงพอสำหรับ monolith ถ้า scale หลาย instance ควรเปลี่ยนเป็น Redis keyspace notification subscriber |
| RabbitMQ แทน Kafka | ops overhead ต่ำกว่าสำหรับ use case นี้ Kafka เหมาะกับ 10k+ events/sec หรือต้อง replay log |
| Mock payment | ไม่ผูก payment gateway จริง endpoint confirm จำลองการจ่ายเงินสำเร็จ |
| Seed อัตโนมัติตอน startup | 10 หนัง × 2 รอบ = 20 showtimes, แต่ละรอบ 140 ที่นั่ง (10 แถว × 14 ที่นั่ง, 2 แถวแรกเป็น VIP) |
| Multi-seat lock เป็น loop ของ single-seat lock | ใช้ atomic operation เดิมต่อที่นั่ง รองรับกรณีบางที่นั่ง lock สำเร็จแต่บางที่นั่งไม่สำเร็จ (ผู้ใช้เห็น error เฉพาะที่นั่งที่ fail) |
| i18n เป็น client-side store | ไม่ต้องเรียก backend เพิ่ม, สลับภาษาทันทีไม่ reload หน้า, เก็บค่าใน localStorage |
| WebSocket per-showtime rooms | client subscribe เฉพาะ showtime ที่กำลังดู, seat update broadcast เฉพาะคนที่ดูรอบนั้น ไม่ส่ง noise ไปทุกคน |
