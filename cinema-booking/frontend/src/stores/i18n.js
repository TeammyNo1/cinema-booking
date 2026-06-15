import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

const translations = {
  th: {
    nav: {
      movies: 'หนัง', myTickets: 'ตั๋วของฉัน', admin: 'แอดมิน', login: 'เข้าสู่ระบบด้วย Google', logout: 'ออกจากระบบ'
    },
    home: {
      eyebrow: '🎬 ระบบจองตั๋วโรงหนัง',
      title1: 'ที่นั่งของคุณ.',
      title2: 'หนังของคุณ.',
      title3: 'ช่วงเวลาของคุณ.',
      subtitle: 'จองที่นั่งโรงหนังแบบ Real-time ไม่มีการจองซ้ำ ด้วยเทคโนโลยี Distributed Lock',
      cta: 'เลือกดูหนังเลย →',
      ctaLogin: 'เข้าสู่ระบบด้วย Google',
      ctaNote: 'ฟรี · ไม่ต้องใช้บัตรเครดิต',
      feat1Title: 'แผนที่ที่นั่ง Real-time',
      feat1Desc: 'เห็นสถานะที่นั่งอัปเดตสดๆ ไม่ต้อง refresh หน้า',
      feat2Title: 'ล็อคที่นั่งด้วย Redis',
      feat2Desc: 'ระบบล็อคแบบ Distributed ป้องกันการจองซ้ำ 5 นาที',
      feat3Title: 'ยืนยันทันที',
      feat3Desc: 'จ่ายและยืนยันในไม่กี่วินาที ตั๋วของคุณรับประกัน',
    },
    showtimes: {
      title: 'หนังที่กำลังฉาย',
      selectSeats: 'เลือกที่นั่ง →',
      hall: 'โรง',
      noShowtimes: 'ไม่มีรอบฉายในขณะนี้',
    },
    seatmap: {
      screen: 'จอภาพยนตร์',
      available: 'ว่าง',
      locked: 'ถูกจอง (5 นาที)',
      booked: 'จองแล้ว',
      selected: 'ที่เลือก',
      holdSeats: '🔒 จองที่นั่ง',
      confirmPay: '✅ ยืนยันและชำระเงิน',
      release: 'ยกเลิก',
      deselect: 'ยกเลิกการเลือก',
      remaining: 'เหลือเวลา',
      toPay: 'นาทีในการชำระเงิน',
      seats: 'ที่นั่ง',
      total: 'รวม',
      locking: 'กำลังจอง...',
      confirming: 'กำลังยืนยัน...',
      back: '← กลับ',
      live: 'สด',
      reconnecting: 'กำลังเชื่อมต่อ...',
      selectHint: 'เลือกได้หลายที่นั่งพร้อมกัน',
    },
    myBookings: {
      title: 'ตั๋วของฉัน',
      noBookings: '🎟 ยังไม่มีการจอง',
      browse: 'เลือกดูหนัง',
      bookingId: 'รหัสการจอง',
      price: 'ราคา',
      created: 'วันที่จอง',
      expires: 'หมดอายุ',
    },
    admin: {
      dashboard: 'แดชบอร์ด',
      auditLogs: 'Audit Logs',
      totalBookings: 'การจองทั้งหมด',
      confirmed: 'ยืนยันแล้ว',
      locked: 'กำลังล็อค',
      timedOut: 'หมดเวลา',
      allStatuses: 'ทุกสถานะ',
      clearFilters: 'ล้างตัวกรอง',
      bookingId: 'รหัสการจอง',
      seat: 'ที่นั่ง',
      userId: 'รหัสผู้ใช้',
      status: 'สถานะ',
      price: 'ราคา',
      created: 'วันที่',
      filterShowtime: 'กรองตาม Showtime ID...',
      noBookings: 'ไม่พบข้อมูลการจอง',
      event: 'เหตุการณ์',
      noLogs: 'ไม่พบ Audit Logs',
    },
    status: {
      CONFIRMED: 'ยืนยันแล้ว', LOCKED: 'กำลังล็อค',
      TIMEOUT: 'หมดเวลา', CANCELLED: 'ยกเลิกแล้ว'
    }
  },
  en: {
    nav: {
      movies: 'Movies', myTickets: 'My Tickets', admin: 'Admin', login: 'Sign in with Google', logout: 'Logout'
    },
    home: {
      eyebrow: '🎬 Cinema Ticket Booking',
      title1: 'Your Seat.',
      title2: 'Your Movie.',
      title3: 'Your Moment.',
      subtitle: 'Book cinema seats in real-time. No double bookings, powered by distributed locking technology.',
      cta: 'Browse Movies →',
      ctaLogin: 'Sign in with Google',
      ctaNote: 'Free to sign in · No credit card required',
      feat1Title: 'Real-time Seat Map',
      feat1Desc: 'See seat availability update live — no page refresh needed.',
      feat2Title: 'Distributed Locking',
      feat2Desc: 'Redis locks prevent double bookings with 5-minute hold windows.',
      feat3Title: 'Instant Confirmation',
      feat3Desc: 'Pay and confirm in seconds. Your ticket is guaranteed.',
    },
    showtimes: {
      title: 'Now Showing',
      selectSeats: 'Select Seats →',
      hall: 'Hall',
      noShowtimes: 'No showtimes available.',
    },
    seatmap: {
      screen: 'SCREEN',
      available: 'Available',
      locked: 'Locked (5 min)',
      booked: 'Booked',
      selected: 'Selected',
      holdSeats: '🔒 Hold Seats',
      confirmPay: '✅ Confirm & Pay',
      release: 'Release',
      deselect: 'Deselect All',
      remaining: 'Remaining',
      toPay: 'min to pay',
      seats: 'seats',
      total: 'Total',
      locking: 'Locking...',
      confirming: 'Confirming...',
      back: '← Back',
      live: 'Live',
      reconnecting: 'Reconnecting...',
      selectHint: 'Select multiple seats at once',
    },
    myBookings: {
      title: 'My Tickets',
      noBookings: '🎟 No bookings yet.',
      browse: 'Browse Movies',
      bookingId: 'Booking ID',
      price: 'Price',
      created: 'Created',
      expires: 'Expires',
    },
    admin: {
      dashboard: 'Dashboard',
      auditLogs: 'Audit Logs',
      totalBookings: 'Total Bookings',
      confirmed: 'Confirmed',
      locked: 'Locked (Active)',
      timedOut: 'Timed Out',
      allStatuses: 'All Statuses',
      clearFilters: 'Clear Filters',
      bookingId: 'Booking ID',
      seat: 'Seat',
      userId: 'User ID',
      status: 'Status',
      price: 'Price',
      created: 'Created',
      filterShowtime: 'Filter by showtime ID...',
      noBookings: 'No bookings found',
      event: 'Event',
      noLogs: 'No audit logs found',
    },
    status: {
      CONFIRMED: 'Confirmed', LOCKED: 'Locked',
      TIMEOUT: 'Timed Out', CANCELLED: 'Cancelled'
    }
  }
}

export const useI18nStore = defineStore('i18n', () => {
  const lang = ref(localStorage.getItem('cinema_lang') || 'th')

  function setLang(l) {
    lang.value = l
    localStorage.setItem('cinema_lang', l)
  }

  const t = computed(() => translations[lang.value])

  return { lang, setLang, t }
})
