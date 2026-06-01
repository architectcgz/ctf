import type { RouteRecordRaw } from 'vue-router'

import { platformRoutes } from './platformRoutes'
import { studentRoutes } from './studentRoutes'
import { teacherRoutes } from './teacherRoutes'

export const appShellRoute: RouteRecordRaw = {
  path: '/',
  component: () => import('@/pages/AppShellRoutePage.vue'),
  redirect: '/student/dashboard',
  meta: { requiresAuth: true },
  children: [...studentRoutes, ...teacherRoutes, ...platformRoutes],
}
