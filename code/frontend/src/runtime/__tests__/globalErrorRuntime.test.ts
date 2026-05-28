import axios from 'axios'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const redirectMocks = vi.hoisted(() => ({
  redirectToErrorStatusPage: vi.fn(),
  shouldRedirectToErrorStatusPage: vi.fn(() => true),
}))

vi.mock('@/utils/errorStatusPage', () => redirectMocks)

import { request } from '@/api/request'
import {
  createGlobalRouterErrorHandler,
  createGlobalVueErrorHandler,
  handleGlobalSessionExpired,
  installGlobalHttpErrorHandling,
  setupGlobalErrorRuntime,
} from '@/runtime/globalErrorRuntime'
import { useAuthStore } from '@/stores/auth'

describe('globalErrorRuntime', () => {
  const axiosInstance = axios.create()
  const originalAdapter = axiosInstance.defaults.adapter

  beforeEach(async () => {
    setActivePinia(createPinia())
    redirectMocks.redirectToErrorStatusPage.mockReset()
    redirectMocks.shouldRedirectToErrorStatusPage.mockReset()
    redirectMocks.shouldRedirectToErrorStatusPage.mockReturnValue(true)
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'user-1',
      username: 'tester',
      role: 'student',
      name: 'Tester',
    })

    const { getAxiosInstance } = await import('@/api/request')
    axiosInstance.defaults.adapter = getAxiosInstance().defaults.adapter
  })

  afterEach(() => {
    axiosInstance.defaults.adapter = originalAdapter
  })

  it('仅对允许的 401 会话失效执行全局登出与状态页跳转', () => {
    const authStore = useAuthStore()

    const handled = handleGlobalSessionExpired({ requestUrl: '/profile' })

    expect(handled).toBe(true)
    expect(authStore.user).toBeNull()
    expect(redirectMocks.redirectToErrorStatusPage).toHaveBeenCalledWith(401, '/profile')
  })

  it('不会对登录入口请求触发全局 401 跳转', () => {
    redirectMocks.shouldRedirectToErrorStatusPage.mockReturnValue(false)
    const authStore = useAuthStore()

    const handled = handleGlobalSessionExpired({ requestUrl: '/auth/login' })

    expect(handled).toBe(false)
    expect(authStore.user?.id).toBe('user-1')
    expect(redirectMocks.redirectToErrorStatusPage).not.toHaveBeenCalled()
  })

  it('运行时安装的 HTTP handler 会接管 401 全局跳转，而不是 request transport 层自己跳页', async () => {
    const { getAxiosInstance } = await import('@/api/request')
    const pinia = createPinia()
    setActivePinia(pinia)
    useAuthStore().setAuth({
      id: 'user-1',
      username: 'tester',
      role: 'student',
      name: 'Tester',
    })

    installGlobalHttpErrorHandling(pinia)
    getAxiosInstance().defaults.adapter = vi.fn().mockRejectedValue({
      config: {
        method: 'get',
        url: '/profile',
      },
      response: {
        status: 401,
        data: {
          code: 0,
          message: '',
          request_id: 'req-unauthorized',
        },
      },
    })

    await expect(
      request({
        method: 'GET',
        url: '/profile',
      })
    ).rejects.toMatchObject({
      status: 401,
      requestUrl: '/profile',
    })

    expect(useAuthStore().user).toBeNull()
    expect(redirectMocks.redirectToErrorStatusPage).toHaveBeenCalledWith(401, '/profile')
  })

  it('Vue runtime handler 仅对非 ApiError 的崩溃跳转 /500', () => {
    const handler = createGlobalVueErrorHandler()

    handler(new Error('boom'), null, 'render')
    expect(redirectMocks.redirectToErrorStatusPage).toHaveBeenCalledWith(500)

    redirectMocks.redirectToErrorStatusPage.mockReset()
    handler(new (class ApiErrorLike extends Error {})('api-like'), null, 'render')
    expect(redirectMocks.redirectToErrorStatusPage).toHaveBeenCalledWith(500)
  })

  it('router runtime handler 统一跳转 /500', () => {
    const handler = createGlobalRouterErrorHandler()

    handler(new Error('router boom'))

    expect(redirectMocks.redirectToErrorStatusPage).toHaveBeenCalledWith(500)
  })

  it('bootstrap runtime 安装应成为唯一的 router runtime error 注册点', () => {
    const app = {
      config: {},
    } as { config: { errorHandler?: unknown } }
    const router = {
      onError: vi.fn(),
    } as unknown as { onError: ReturnType<typeof vi.fn> }
    const pinia = createPinia()

    setupGlobalErrorRuntime(app as never, router as never, pinia)

    expect(typeof app.config.errorHandler).toBe('function')
    expect(router.onError).toHaveBeenCalledTimes(1)
    expect(typeof router.onError.mock.calls[0]?.[0]).toBe('function')
  })
})
