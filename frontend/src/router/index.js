import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/login', component: () => import('../views/Login.vue') },
  { path: '/', redirect: '/dashboard' },
  { path: '/dashboard', component: () => import('../views/Dashboard.vue'), meta: { requiresAuth: true } },
  { path: '/employee', component: () => import('../views/Employee.vue'), meta: { requiresAuth: true } },
  { path: '/notice', component: () => import('../views/Notice.vue'), meta: { requiresAuth: true } },
  { path: '/log', component: () => import('../views/OperationLog.vue'), meta: { requiresAuth: true } },
  { path: '/meeting-room', component: () => import('../views/MeetingRoom.vue'), meta: { requiresAuth: true } },
  { path: '/event-booking', component: () => import('../views/EventBooking.vue'), meta: { requiresAuth: true } },
  { path: '/leave-request', component: () => import('../views/LeaveRequest.vue'), meta: { requiresAuth: true } },
  { path: '/resignation', component: () => import('../views/Resignation.vue'), meta: { requiresAuth: true } },
  { path: '/onboarding', component: () => import('../views/Onboarding.vue'), meta: { requiresAuth: true } },
  { path: '/workflow', component: () => import('../views/Workflow.vue'), meta: { requiresAuth: true } },
  { path: '/menu', component: () => import('../views/Menu.vue'), meta: { requiresAuth: true } },
  { path: '/schedule', component: () => import('../views/Schedule.vue'), meta: { requiresAuth: true } },
  { path: '/role', component: () => import('../views/Role.vue'), meta: { requiresAuth: true } }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
