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
  accessToken: '',
  updateTokens: vi.fn(),
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
    authStoreMock.updateTokens.mockReset()
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
})
