import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'

const { getProfileMock } = vi.hoisted(() => ({
  getProfileMock: vi.fn(),
}))

vi.mock('@/api/auth', () => ({
  getProfile: getProfileMock,
}))

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    getProfileMock.mockReset()
  })

  it('不应再把认证态持久化到 localStorage', () => {
    const authStore = useAuthStore()

    authStore.setAuth({ id: '1', username: 'alice', role: 'student' })

    expect(localStorage.getItem('ctf_access_token')).toBeNull()
    expect(localStorage.getItem('ctf_refresh_token')).toBeNull()
  })

  it('应通过 session cookie 静默恢复用户资料，并清理旧 localStorage 键', async () => {
    localStorage.setItem('ctf_access_token', 'legacy-access-token')
    localStorage.setItem('ctf_refresh_token', 'legacy-refresh-token')
    getProfileMock.mockResolvedValue({ id: '1', username: 'alice', role: 'teacher' })

    const authStore = useAuthStore()
    await authStore.restore()

    expect(getProfileMock).toHaveBeenCalledWith({ suppressErrorToast: true })
    expect(authStore.user).toEqual({ id: '1', username: 'alice', role: 'teacher' })
    expect(authStore.isLoggedIn).toBe(true)
    expect(authStore.sessionRestored).toBe(true)
    expect(localStorage.getItem('ctf_access_token')).toBeNull()
    expect(localStorage.getItem('ctf_refresh_token')).toBeNull()
  })

  it('恢复失败时应静默降级到未登录态', async () => {
    getProfileMock.mockRejectedValue(new Error('unauthorized'))

    const authStore = useAuthStore()
    await authStore.restore()

    expect(authStore.user).toBeNull()
    expect(authStore.isLoggedIn).toBe(false)
    expect(authStore.sessionRestored).toBe(true)
  })

  it('并发恢复时应复用同一条资料恢复请求', async () => {
    let resolveProfile: ((value: { id: string; username: string; role: 'student' }) => void) | undefined
    getProfileMock.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveProfile = resolve
        })
    )

    const authStore = useAuthStore()
    const firstRestore = authStore.restore()
    const secondRestore = authStore.restore()

    expect(getProfileMock).toHaveBeenCalledTimes(1)

    resolveProfile?.({ id: '1', username: 'alice', role: 'student' })
    await Promise.all([firstRestore, secondRestore])

    expect(authStore.user).toEqual({ id: '1', username: 'alice', role: 'student' })
    expect(authStore.sessionRestored).toBe(true)
  })
})
