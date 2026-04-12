import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import type { NavigationGuardNext, RouteLocationNormalized, Router } from 'vue-router'

import { setupRouterGuards, sanitizeRedirectPath, hasRequiredRole } from '@/router/guards'
import { useAuthStore } from '@/stores/auth'
import type { AuthUser } from '@/stores/auth'

const { errorMock, getProfileMock, warningMock } = vi.hoisted(() => ({
  warningMock: vi.fn(),
  errorMock: vi.fn(),
  getProfileMock: vi.fn(),
}))

vi.mock('nprogress', () => ({
  default: {
    configure: vi.fn(),
    start: vi.fn(),
    done: vi.fn(),
  },
}))

vi.mock('@/api/auth', () => ({
  getProfile: getProfileMock,
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => ({
    success: vi.fn(),
    info: vi.fn(),
    dismiss: vi.fn(),
    warning: warningMock,
    error: errorMock,
  }),
}))

function createRoute(
  overrides: Partial<RouteLocationNormalized> & {
    meta?: RouteLocationNormalized['meta']
    query?: RouteLocationNormalized['query']
  } = {}
): RouteLocationNormalized {
  return {
    path: '/',
    fullPath: '/',
    query: {},
    meta: {},
    name: undefined,
    hash: '',
    matched: [],
    params: {},
    redirectedFrom: undefined,
    href: '/',
    ...overrides,
  } as RouteLocationNormalized
}

function createRouterMock() {
  let beforeEachHandler:
    | ((
        to: RouteLocationNormalized,
        from: RouteLocationNormalized,
        next: NavigationGuardNext
      ) => Promise<void>)
    | undefined

  const router = {
    beforeEach: vi.fn((handler) => {
      beforeEachHandler = handler
    }),
    afterEach: vi.fn(),
    onError: vi.fn(),
  } as unknown as Router

  setupRouterGuards(router)

  return {
    router,
    runBeforeEach: async (to: RouteLocationNormalized, next = vi.fn()) => {
      if (!beforeEachHandler) {
        throw new Error('beforeEach guard was not registered')
      }
      await beforeEachHandler(to, createRoute(), next)
      return next
    },
  }
}

function buildUser(role: AuthUser['role']): AuthUser {
  return {
    id: '1',
    username: 'tester',
    role,
    name: 'Tester',
    class_name: 'Class-A',
  }
}

describe('router guards', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    warningMock.mockReset()
    errorMock.mockReset()
    getProfileMock.mockReset()
  })

  it('应该阻止未登录用户访问受保护路由', async () => {
    const { runBeforeEach } = createRouterMock()

    const next = await runBeforeEach(
      createRoute({
        path: '/platform/overview',
        fullPath: '/platform/overview',
        meta: { requiresAuth: true },
      })
    )

    expect(next).toHaveBeenCalledWith({
      path: '/login',
      query: { redirect: '/platform/overview' },
    })
  })

  it('应该阻止无权限用户访问受限路由', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(buildUser('student'), 'token')

    const { runBeforeEach } = createRouterMock()
    const next = await runBeforeEach(
      createRoute({
        path: '/platform/overview',
        fullPath: '/platform/overview',
        meta: { requiresAuth: true, roles: ['admin'] },
      })
    )

    expect(warningMock).toHaveBeenCalledWith('无权限访问该页面')
    expect(next).toHaveBeenCalledWith('/403')
  })

  it('应该在仅有 token 时自动拉取用户资料并放行 AWD 复盘入口', async () => {
    const authStore = useAuthStore()
    authStore.updateTokens('token')
    getProfileMock.mockResolvedValue(buildUser('teacher'))

    const { runBeforeEach } = createRouterMock()
    const next = await runBeforeEach(
      createRoute({
        path: '/academy/awd-reviews',
        fullPath: '/academy/awd-reviews',
        meta: { requiresAuth: true, roles: ['teacher', 'admin'] },
      })
    )

    expect(getProfileMock).toHaveBeenCalledTimes(1)
    expect(authStore.user?.role).toBe('teacher')
    expect(next).toHaveBeenCalledWith()
  })

  it('已登录用户访问登录页且没有 redirect 时应返回角色工作台', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(buildUser('teacher'), 'token')

    const { runBeforeEach } = createRouterMock()
    const next = await runBeforeEach(
      createRoute({
        path: '/login',
        fullPath: '/login',
      })
    )

    expect(next).toHaveBeenCalledWith('/academy/overview')
  })
})

describe('guard helpers', () => {
  it('应该清洗非法 redirect 路径', () => {
    expect(sanitizeRedirectPath('//evil.com')).toBe('/')
    expect(sanitizeRedirectPath('https://evil.com')).toBe('/')
    expect(sanitizeRedirectPath('/dashboard')).toBe('/dashboard')
  })

  it('应该正确判断角色权限', () => {
    expect(hasRequiredRole(undefined, 'student')).toBe(true)
    expect(hasRequiredRole(['admin'], undefined)).toBe(false)
    expect(hasRequiredRole(['teacher', 'admin'], 'teacher')).toBe(true)
    expect(hasRequiredRole(['admin'], 'student')).toBe(false)
  })
})
