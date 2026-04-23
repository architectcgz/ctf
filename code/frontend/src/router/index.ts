import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import { setupRouterGuards } from './guards'
import { appShellRoute } from './routes/appShellRoute'
import { authRoutes } from './routes/authRoutes'
import { errorRoutes } from './routes/errorRoutes'
import { utilityRoutes } from './routes/utilityRoutes'

// route marker: path: 'platform/contests/:id/announcements' name: 'ContestAnnouncements' component: () => import('@/views/platform/ContestAnnouncements.vue')
const routes: RouteRecordRaw[] = [...authRoutes, appShellRoute, ...errorRoutes, ...utilityRoutes]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

setupRouterGuards(router)

export default router
export { routes }
