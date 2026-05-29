import { beforeEach, describe, expect, it, vi } from 'vitest'

const apiMocks = vi.hoisted(() => ({
  login: vi.fn(),
  register: vi.fn(),
  logout: vi.fn(),
}))

const authStoreMocks = vi.hoisted(() => ({
  setAuth: vi.fn(),
  logout: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  info: vi.fn(),
}))

vi.mock('@/api/auth', () => ({
  login: apiMocks.login,
  register: apiMocks.register,
  logout: apiMocks.logout,
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreMocks,
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

import { useAuth } from './useAuth'

describe('useAuth', () => {
  beforeEach(() => {
    apiMocks.login.mockReset()
    apiMocks.register.mockReset()
    apiMocks.logout.mockReset()
    authStoreMocks.setAuth.mockReset()
    authStoreMocks.logout.mockReset()
    toastMocks.success.mockReset()
    toastMocks.info.mockReset()
  })

  it('应在登录后写入 auth store 并返回用户', async () => {
    const user = { id: '1', username: 'alice', role: 'teacher' }
    apiMocks.login.mockResolvedValue({ user })

    const { login } = useAuth()

    await expect(login({ username: 'alice', password: 'secure-pass' })).resolves.toEqual(user)
    expect(authStoreMocks.setAuth).toHaveBeenCalledWith(user)
    expect(toastMocks.success).toHaveBeenCalledWith('登录成功')
  })

  it('应在注册后写入 auth store 并返回用户', async () => {
    const user = { id: '2', username: 'bob', role: 'student' }
    apiMocks.register.mockResolvedValue({ user })

    const { register } = useAuth()

    await expect(
      register({ username: 'bob', password: 'secure-pass', class_name: 'CTF-1' })
    ).resolves.toEqual(user)
    expect(authStoreMocks.setAuth).toHaveBeenCalledWith(user)
    expect(toastMocks.success).toHaveBeenCalledWith('注册成功')
  })

  it('退出登录接口失败时仍应清理本地会话', async () => {
    apiMocks.logout.mockRejectedValue(new Error('network error'))

    const { logout } = useAuth()

    await expect(logout()).resolves.toBeUndefined()
    expect(authStoreMocks.logout).toHaveBeenCalledTimes(1)
    expect(toastMocks.info).toHaveBeenCalledWith('已退出登录')
  })
})
