import type { NavigationGuardNext, RouteLocationNormalized, Router } from 'vue-router'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'

import { useAuthStore } from '@/stores/auth'
import { APP_TITLE_PREFIX, type UserRole } from '@/utils/constants'
import { useToast } from '@/composables/useToast'
import { getProfile } from '@/api/auth'

NProgress.configure({ showSpinner: false })

function isPublicRoute(to: RouteLocationNormalized): boolean {
  return to.path === '/login' || to.path === '/register'
}

function sanitizeRedirectPath(input: unknown): string {
  if (typeof input !== 'string') return '/'
  // 移除所有前导斜杠，只保留一个
  const normalized = '/' + input.replace(/^\/+/, '')
  // 检查是否包含协议或双斜杠
  if (/^\/\/|^\/\\|:\/\//.test(normalized)) return '/'
  return normalized
}

async function ensureProfileLoaded(): Promise<void> {
  const authStore = useAuthStore()
  if (!authStore.accessToken) return
  if (authStore.user) return

  const profile = await getProfile()
  authStore.setAuth(profile, authStore.accessToken)
}

function hasRole(requiredRoles: UserRole[] | undefined, userRole: UserRole | undefined): boolean {
  if (!requiredRoles || requiredRoles.length === 0) return true
  if (!userRole) return false
  return requiredRoles.includes(userRole)
}

function updatePageTitle(to: RouteLocationNormalized): void {
  const title = to.meta?.title ? `${APP_TITLE_PREFIX} - ${to.meta.title}` : APP_TITLE_PREFIX
  document.title = title
}

let currentAbortController: AbortController | null = null

export function setupRouterGuards(router: Router): void {
  router.beforeEach(async (to, _from, next) => {
    NProgress.start()

    currentAbortController?.abort()
    currentAbortController = new AbortController()

    const authStore = useAuthStore()
    const toast = useToast()

    try {
      if (isPublicRoute(to)) {
        if (authStore.isLoggedIn) {
          const redirectTo = sanitizeRedirectPath(to.query.redirect)
          next(redirectTo)
          return
        }
        next()
        return
      }

      if (to.meta?.requiresAuth && !authStore.isLoggedIn) {
        next({ path: '/login', query: { redirect: to.fullPath } })
        return
      }

      if (to.meta?.requiresAuth) {
        await ensureProfileLoaded()
      }

      const userRole = authStore.user?.role
      const requiredRoles = to.meta?.roles
      if (!hasRole(requiredRoles, userRole)) {
        toast.warning('无权限访问该页面')
        next('/dashboard')
        return
      }

      next()
    } catch (err) {
      // Avoid exposing raw backend message.
      console.error('Router guard error:', err)
      toast.error('登录状态异常，请重新登录')
      authStore.logout()
      next({ path: '/login', query: { redirect: to.fullPath } })
    }
  })

  router.afterEach((to) => {
    updatePageTitle(to)
    NProgress.done()
  })

  router.onError((error) => {
    console.error('Router error:', error)
  })
}

