import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'

const { refreshTokenMock } = vi.hoisted(() => ({
  refreshTokenMock: vi.fn(),
}))

vi.mock('@/api/auth', () => ({
  refreshToken: refreshTokenMock,
}))

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    refreshTokenMock.mockReset()
  })

  it('不应再把 access token 持久化到 localStorage', () => {
    const authStore = useAuthStore()

    authStore.setAuth({ id: '1', username: 'alice', role: 'student' }, 'token-1')
    expect(localStorage.getItem('ctf_access_token')).toBeNull()

    authStore.updateTokens('token-2')
    expect(localStorage.getItem('ctf_access_token')).toBeNull()
  })

  it('应通过 refresh cookie 静默恢复 access token，并清理旧 localStorage 键', async () => {
    localStorage.setItem('ctf_access_token', 'legacy-access-token')
    localStorage.setItem('ctf_refresh_token', 'legacy-refresh-token')
    refreshTokenMock.mockResolvedValue({ access_token: 'fresh-token' })

    const authStore = useAuthStore()
    await authStore.restore()

    expect(refreshTokenMock).toHaveBeenCalledWith({ suppressErrorToast: true })
    expect(authStore.accessToken).toBe('fresh-token')
    expect(authStore.sessionRestored).toBe(true)
    expect(localStorage.getItem('ctf_access_token')).toBeNull()
    expect(localStorage.getItem('ctf_refresh_token')).toBeNull()
  })

  it('恢复失败时应静默降级到未登录内存态', async () => {
    refreshTokenMock.mockRejectedValue(new Error('expired'))

    const authStore = useAuthStore()
    await authStore.restore()

    expect(authStore.accessToken).toBe('')
    expect(authStore.sessionRestored).toBe(true)
  })

  it('并发恢复时应复用同一条静默恢复请求', async () => {
    let resolveRefresh: ((value: { access_token: string }) => void) | undefined
    refreshTokenMock.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveRefresh = resolve
        })
    )

    const authStore = useAuthStore()
    const firstRestore = authStore.restore()
    const secondRestore = authStore.restore()

    expect(refreshTokenMock).toHaveBeenCalledTimes(1)

    resolveRefresh?.({ access_token: 'shared-token' })
    await Promise.all([firstRestore, secondRestore])

    expect(authStore.accessToken).toBe('shared-token')
    expect(authStore.sessionRestored).toBe(true)
  })
})
