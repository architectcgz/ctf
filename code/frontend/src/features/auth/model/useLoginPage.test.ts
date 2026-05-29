import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const authMocks = vi.hoisted(() => ({
  login: vi.fn(),
}))

const routeState = vi.hoisted(() => ({
  query: {
    redirect: undefined as unknown,
  },
}))

const routerMocks = vi.hoisted(() => ({
  push: vi.fn(),
}))

const probeMocks = vi.hoisted(() => ({
  track: vi.fn(),
}))

vi.mock('./useAuth', () => ({
  useAuth: () => authMocks,
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
  useRouter: () => routerMocks,
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
    routeState.query.redirect = undefined
    routerMocks.push.mockReset()
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
    expect(routerMocks.push).toHaveBeenCalledWith('/academy/overview')
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
    expect(routerMocks.push).toHaveBeenCalledWith('/student/dashboard')
  })

  it('应回填登录错误并关闭加载态', async () => {
    authMocks.login.mockRejectedValue(new Error('用户名或密码错误'))
    const page = useLoginPage()
    page.form.username = 'alice'
    page.form.password = 'wrong-pass'

    await page.onSubmit()

    expect(page.submitError.value).toBe('用户名或密码错误')
    expect(page.loading.value).toBe(false)
    expect(routerMocks.push).not.toHaveBeenCalled()
  })

  it('应优先使用登录页 redirect 参数', async () => {
    routeState.query.redirect = '/contests/1'
    authMocks.login.mockResolvedValue({ role: 'teacher' })
    const page = useLoginPage()
    page.form.username = 'alice'
    page.form.password = 'pass'

    await page.onSubmit()

    expect(routerMocks.push).toHaveBeenCalledWith('/contests/1')
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
