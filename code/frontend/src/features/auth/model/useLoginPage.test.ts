import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'

const authMocks = vi.hoisted(() => ({
  login: vi.fn(),
}))

const redirectState = vi.hoisted(() => ({
  value: '/',
}))

const probeMocks = vi.hoisted(() => ({
  track: vi.fn(),
}))

vi.mock('./useAuth', () => ({
  useAuth: () => authMocks,
}))

vi.mock('./useLoginViewPage', () => ({
  useLoginViewPage: () => ({
    redirectTo: ref(redirectState.value),
  }),
}))

vi.mock('@/composables/useProbeEasterEggs', () => ({
  useProbeEasterEggs: () => probeMocks,
}))

import { useLoginPage } from './useLoginPage'

describe('useLoginPage', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  beforeEach(() => {
    authMocks.login.mockReset()
    redirectState.value = '/'
    probeMocks.track.mockReset()
    probeMocks.track.mockReturnValue({
      unlocked: false,
      activated: false,
      count: 1,
    })
    vi.spyOn(console, 'log').mockImplementation(() => {})
  })

  it('应提交登录并在根路径重定向时传 undefined', async () => {
    authMocks.login.mockResolvedValue(undefined)
    const page = useLoginPage()
    page.form.username = '  alice  '
    page.form.password = 'pass'

    await page.onSubmit()

    expect(authMocks.login).toHaveBeenCalledWith({ username: 'alice', password: 'pass' }, undefined)
  })

  it('应支持使用输入框回填值提交', async () => {
    authMocks.login.mockResolvedValue(undefined)
    const page = useLoginPage()

    await page.onSubmit({
      username: 'alice',
      password: 'browser-saved-password',
    })

    expect(authMocks.login).toHaveBeenCalledWith(
      { username: 'alice', password: 'browser-saved-password' },
      undefined
    )
  })

  it('应回填登录错误并关闭加载态', async () => {
    authMocks.login.mockRejectedValue(new Error('用户名或密码错误'))
    const page = useLoginPage()
    page.form.username = 'alice'
    page.form.password = 'wrong-pass'

    await page.onSubmit()

    expect(page.submitError.value).toBe('用户名或密码错误')
    expect(page.loading.value).toBe(false)
  })

  it('应在 probe 解锁时显示提示', () => {
    probeMocks.track.mockReturnValue({
      unlocked: true,
      activated: true,
      count: 4,
    })
    const page = useLoginPage()

    page.handleHeroProbe()

    expect(probeMocks.track).toHaveBeenCalledWith('login-brand', 4)
    expect(page.probeMessage.value).toBe('隐藏入口排查完毕，结果让你失望了。')
  })
})
