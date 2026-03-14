import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { useCASAuth } from '@/composables/useCASAuth'
import { useAuthStore } from '@/stores/auth'

const authApiMocks = vi.hoisted(() => ({
  getCASStatus: vi.fn(),
  getCASLogin: vi.fn(),
  completeCASLogin: vi.fn(),
}))

const routerMocks = vi.hoisted(() => ({
  replace: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
  dismiss: vi.fn(),
}))

const browserMocks = vi.hoisted(() => ({
  redirectTo: vi.fn(),
}))

vi.mock('@/api/auth', () => authApiMocks)
vi.mock('vue-router', () => ({
  useRouter: () => routerMocks,
}))
vi.mock('@/utils/browser', () => browserMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

describe('useCASAuth', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    window.sessionStorage.clear()
    authApiMocks.getCASStatus.mockReset()
    authApiMocks.getCASLogin.mockReset()
    authApiMocks.completeCASLogin.mockReset()
    routerMocks.replace.mockReset()
    toastMocks.success.mockReset()
    browserMocks.redirectTo.mockReset()
  })

  it('应该拉取 CAS 状态', async () => {
    authApiMocks.getCASStatus.mockResolvedValue({
      provider: 'cas',
      enabled: true,
      configured: true,
      auto_provision: false,
      login_path: '/api/v1/auth/cas/login',
      callback_path: '/api/v1/auth/cas/callback',
    })

    const cas = useCASAuth()
    await cas.fetchCASStatus()

    expect(cas.casStatus.value?.enabled).toBe(true)
    expect(cas.casReady.value).toBe(true)
  })

  it('应该保存 redirect 并跳转到 CAS 登录地址', async () => {
    authApiMocks.getCASLogin.mockResolvedValue({
      provider: 'cas',
      redirect_url: 'https://cas.example.edu/login',
      callback_url: 'https://ctf.example.edu/login/cas/callback',
    })

    const cas = useCASAuth()
    await cas.beginCASLogin('/admin/dashboard')

    expect(window.sessionStorage.getItem('ctf_cas_post_login_redirect')).toBe('/admin/dashboard')
    expect(browserMocks.redirectTo).toHaveBeenCalledWith('https://cas.example.edu/login')
  })

  it('应该在回调完成后写入登录态并恢复 redirect', async () => {
    window.sessionStorage.setItem('ctf_cas_post_login_redirect', '/teacher/reports')
    authApiMocks.completeCASLogin.mockResolvedValue({
      access_token: 'cas-token',
      token_type: 'Bearer',
      expires_in: 7200,
      user: {
        id: '7',
        username: 'cas_user',
        role: 'teacher',
        name: 'Cas User',
      },
    })

    const cas = useCASAuth()
    await cas.finishCASLogin('ST-1')

    const authStore = useAuthStore()
    expect(authStore.user?.username).toBe('cas_user')
    expect(authStore.accessToken).toBe('cas-token')
    expect(routerMocks.replace).toHaveBeenCalledWith('/teacher/reports')
    expect(toastMocks.success).toHaveBeenCalledWith('CAS 登录成功')
  })
})
