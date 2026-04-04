import type { RouteLocationNormalized, Router } from 'vue-router'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'

import { useAuthStore } from '@/stores/auth'
import { APP_TITLE_PREFIX } from '@/utils/constants'
import { useToast } from '@/composables/useToast'
import { getProfile } from '@/api/auth'
import type { UserRole } from '@/utils/constants'
import { resolveRouteTitle } from '@/utils/routeTitle'
import { redirectToErrorStatusPage } from '@/utils/errorStatusPage'

NProgress.configure({ showSpinner: false })

function isPublicRoute(to: RouteLocationNormalized): boolean {
  return to.path === '/login' || to.path === '/register'
}

function isAuthLandingRoute(to: RouteLocationNormalized): boolean {
  return to.path === '/login' || to.path === '/register'
}

export function sanitizeRedirectPath(input: unknown): string {
  if (typeof input !== 'string') return '/'
  if (/^\s*$/.test(input)) return '/'
  if (/^(?:[a-z][a-z0-9+\-.]*:)?\/\//i.test(input) || input.startsWith('/\\')) return '/'
  // 移除所有前导斜杠，只保留一个
  const normalized = '/' + input.replace(/^\/+/, '')
  return normalized
}

export function hasRequiredRole(requiredRoles: RouteLocationNormalized['meta']['roles'], currentRole: UserRole | undefined): boolean {
  if (!requiredRoles || requiredRoles.length === 0) return true
  if (!currentRole) return false
  return requiredRoles.includes(currentRole)
}

async function ensureProfileLoaded(): Promise<void> {
  const authStore = useAuthStore()
  if (!authStore.accessToken) return
  if (authStore.user) return

  const profile = await getProfile()
  authStore.setAuth(profile, authStore.accessToken)
}

function updatePageTitle(to: RouteLocationNormalized): void {
  const routeTitle = resolveRouteTitle(to)
  const title = routeTitle ? `${APP_TITLE_PREFIX} - ${routeTitle}` : APP_TITLE_PREFIX
  document.title = title
}

export function setupRouterGuards(router: Router): void {
  router.beforeEach(async (to, _from, next) => {
    NProgress.start()

    const authStore = useAuthStore()
    const toast = useToast()

    try {
      if (isPublicRoute(to)) {
        if (isAuthLandingRoute(to) && authStore.isLoggedIn) {
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
      if (!hasRequiredRole(to.meta?.roles, userRole)) {
        toast.warning('无权限访问该页面')
        next('/403')
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
    redirectToErrorStatusPage(500)
  })
}
