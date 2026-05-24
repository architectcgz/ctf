import axios from 'axios'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
  dismiss: vi.fn(),
}))

const authStoreMock = vi.hoisted(() => ({
  logout: vi.fn(),
}))

const redirectMocks = vi.hoisted(() => ({
  redirectToErrorStatusPage: vi.fn(),
  shouldRedirectToErrorStatusPage: vi.fn(() => false),
}))

vi.mock('nprogress', () => ({
  default: {
    start: vi.fn(),
    done: vi.fn(),
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreMock,
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

vi.mock('@/utils/errorStatusPage', () => redirectMocks)

import { getAxiosInstance, request } from '@/api/request'

describe('request cancel handling', () => {
  const axiosInstance = getAxiosInstance()
  const originalAdapter = axiosInstance.defaults.adapter

  beforeEach(() => {
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()
    toastMocks.info.mockReset()
    toastMocks.dismiss.mockReset()
    authStoreMock.logout.mockReset()
    redirectMocks.redirectToErrorStatusPage.mockReset()
    redirectMocks.shouldRedirectToErrorStatusPage.mockReset()
    redirectMocks.shouldRedirectToErrorStatusPage.mockReturnValue(false)
  })

  afterEach(() => {
    axiosInstance.defaults.adapter = originalAdapter
  })

  it('取消请求时不应弹出网络错误提示或跳转错误页', async () => {
    axiosInstance.defaults.adapter = vi.fn().mockRejectedValue(new axios.CanceledError('canceled'))

    await expect(
      request({
        method: 'GET',
        url: '/cancel-me',
      })
    ).rejects.toMatchObject({
      code: 'ERR_CANCELED',
    })

    expect(toastMocks.error).not.toHaveBeenCalled()
    expect(redirectMocks.redirectToErrorStatusPage).not.toHaveBeenCalled()
  })

  it('业务错误应由调用方决定如何展示，请求层只返回标准化异常', async () => {
    axiosInstance.defaults.adapter = vi.fn().mockRejectedValue({
      config: {
        method: 'post',
        url: '/local-error',
      },
      response: {
        status: 409,
        data: {
          code: 14099,
          message: '普通冲突',
          request_id: 'req-local',
        },
      },
    })

    await expect(
      request({
        method: 'POST',
        url: '/local-error',
      })
    ).rejects.toMatchObject({
      message: '普通冲突',
      code: 14099,
      status: 409,
    })

    expect(toastMocks.error).not.toHaveBeenCalled()
  })

  it('429 限流错误不应再由请求层强制跳转错误页', async () => {
    axiosInstance.defaults.adapter = vi.fn().mockRejectedValue({
      config: {
        method: 'get',
        url: '/rate-limited',
      },
      response: {
        status: 429,
        headers: {
          'retry-after': '30',
        },
        data: {
          code: 0,
          message: '',
          request_id: 'req-rate',
        },
      },
    })

    await expect(
      request({
        method: 'GET',
        url: '/rate-limited',
      })
    ).rejects.toMatchObject({
      message: '请求过于频繁，请 30 秒后重试',
      status: 429,
      requestId: 'req-rate',
    })

    expect(redirectMocks.redirectToErrorStatusPage).not.toHaveBeenCalled()
    expect(authStoreMock.logout).not.toHaveBeenCalled()
  })

  it('普通服务端错误应返回给页面 owner，而不是在请求层统一跳转', async () => {
    axiosInstance.defaults.adapter = vi.fn().mockRejectedValue({
      config: {
        method: 'get',
        url: '/server-error',
      },
      response: {
        status: 500,
        data: {
          code: 0,
          message: '',
          request_id: 'req-server',
        },
      },
    })

    await expect(
      request({
        method: 'GET',
        url: '/server-error',
      })
    ).rejects.toMatchObject({
      message: '服务暂时不可用，请稍后重试',
      status: 500,
      requestId: 'req-server',
    })

    expect(redirectMocks.redirectToErrorStatusPage).not.toHaveBeenCalled()
    expect(authStoreMock.logout).not.toHaveBeenCalled()
  })

  it('401 全局会话失效仍保留登出并跳转错误页', async () => {
    redirectMocks.shouldRedirectToErrorStatusPage.mockReturnValue(true)
    axiosInstance.defaults.adapter = vi.fn().mockRejectedValue({
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
      message: '登录状态已失效，请重新登录',
      status: 401,
      requestId: 'req-unauthorized',
    })

    expect(authStoreMock.logout).toHaveBeenCalledTimes(1)
    expect(redirectMocks.redirectToErrorStatusPage).toHaveBeenCalledWith(401, '/profile')
    expect(redirectMocks.shouldRedirectToErrorStatusPage).toHaveBeenCalledWith(401, '/profile')
  })

  it('网络错误也不应由请求层直接弹错', async () => {
    axiosInstance.defaults.adapter = vi.fn().mockRejectedValue({
      config: {
        method: 'get',
        url: '/silent-error',
      },
    })

    await expect(
      request({
        method: 'GET',
        url: '/silent-error',
      })
    ).rejects.toMatchObject({})

    expect(toastMocks.error).not.toHaveBeenCalled()
  })
})
