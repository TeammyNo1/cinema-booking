import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/',             component: () => import('@/views/HomeView.vue') },
    { path: '/auth/callback',component: () => import('@/views/AuthCallback.vue') },
    { path: '/showtimes',    component: () => import('@/views/ShowtimesView.vue'), meta:{ requiresAuth:true } },
    { path: '/showtimes/:id',component: () => import('@/views/SeatMapView.vue'),  meta:{ requiresAuth:true } },
    { path: '/my-bookings',  component: () => import('@/views/MyBookingsView.vue'), meta:{ requiresAuth:true } },
    {
      path: '/admin',
      component: () => import('@/views/admin/AdminLayout.vue'),
      meta: { requiresAdmin: true },
      children: [
        { path: '',           component: () => import('@/views/admin/DashboardView.vue') },
        { path: 'showtimes',  component: () => import('@/views/admin/ShowtimeManageView.vue') },
        { path: 'audit',      component: () => import('@/views/admin/AuditLogsView.vue') },
      ]
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ]
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) return '/'
  if (to.meta.requiresAdmin && !auth.isAdmin) return '/'
})

export default router
