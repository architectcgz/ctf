import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import useLoginPageSource from './useLoginPage.ts?raw'

const authMocks = vi.hoisted(() => ({
  login: vi.fn(),
}))

const routeState = vi.hoisted(() => ({
  value: {
    redirect: undefined as unknown,
  },
}))

const navigationMocks = vi.hoisted(() => ({
  push: vi.fn(),
}))

const probeMocks = vi.hoisted(() => ({
  track: vi.fn(),
}))

vi.mock('./useAuth', () => ({
  useAuth: () => authMocks,
}))

vi.mock('@/composables/routeQueryTransport', () => ({
  useRouteQueryTransport: () => ({
    query: routeState,
  }),
}))

vi.mock('@/composables/routeNavigationTransport', () => ({
  useRouteNavigationTransport: () => navigationMocks,
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
    routeState.value.redirect = undefined
    navigationMocks.push.mockReset()
    probeMocks.track.mockReset()
    probeMocks.track.mockReturnValue({
      unlocked: false,
      activated: false,
      count: 1,
    })
  })

  it('应提交登录并在默认 redirect 时跳到角色首页', async () => {
    authMocks.login.mockResolvedValue({ role: 'teacher' })
    const page = useLoginPage()
    page.form.username = '  alice  '
    page.form.password = 'pass'

    await page.onSubmit()

    expect(authMocks.login).toHaveBeenCalledWith({ username: 'alice', password: 'pass' })
    expect(navigationMocks.push).toHaveBeenCalledWith('/academy/overview')
  })

  it('应支持使用输入框回填值提交', async () => {
    authMocks.login.mockResolvedValue({ role: 'student' })
    const page = useLoginPage()

    await page.onSubmit({
      username: 'alice',
      password: 'browser-saved-password',
    })

    expect(authMocks.login).toHaveBeenCalledWith({
      username: 'alice',
      password: 'browser-saved-password',
    })
    expect(navigationMocks.push).toHaveBeenCalledWith('/student/dashboard')
  })

  it('应回填登录错误并关闭加载态', async () => {
    authMocks.login.mockRejectedValue(new Error('用户名或密码错误'))
    const page = useLoginPage()
    page.form.username = 'alice'
    page.form.password = 'wrong-pass'

    await page.onSubmit()

    expect(page.submitError.value).toBe('用户名或密码错误')
    expect(page.loading.value).toBe(false)
    expect(navigationMocks.push).not.toHaveBeenCalled()
  })

  it('应优先使用登录页 redirect 参数', async () => {
    routeState.value.redirect = '/contests/1'
    authMocks.login.mockResolvedValue({ role: 'teacher' })
    const page = useLoginPage()
    page.form.username = 'alice'
    page.form.password = 'pass'

    await page.onSubmit()

    expect(navigationMocks.push).toHaveBeenCalledWith('/contests/1')
  })

  it('redirect 参数应先做安全清洗', async () => {
    routeState.value.redirect = 'https://evil.example/phish'
    authMocks.login.mockResolvedValue({ role: 'teacher' })
    const page = useLoginPage()
    page.form.username = 'alice'
    page.form.password = 'pass'

    await page.onSubmit()

    expect(navigationMocks.push).toHaveBeenCalledWith('/academy/overview')
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

  it('登录页 route owner 应改为共享 transport 与中性 sanitize util', () => {
    expect(useLoginPageSource).toContain(
      "import { useRouteQueryTransport } from '@/composables/routeQueryTransport'"
    )
    expect(useLoginPageSource).toContain(
      "import { useRouteNavigationTransport } from '@/composables/routeNavigationTransport'"
    )
    expect(useLoginPageSource).toContain("import { sanitizeRedirectPath } from '@/utils/redirectPath'")
    expect(useLoginPageSource).not.toContain("from 'vue-router'")
    expect(useLoginPageSource).not.toContain("from '@/router/guards'")
  })
})
