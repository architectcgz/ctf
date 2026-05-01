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
