import { beforeEach, describe, expect, it, vi } from 'vitest'

const authMocks = vi.hoisted(() => ({
  logout: vi.fn(),
}))

const routerMocks = vi.hoisted(() => ({
  push: vi.fn(),
}))

vi.mock('@/features/auth', () => ({
  useAuth: () => authMocks,
}))

vi.mock('vue-router', () => ({
  useRouter: () => routerMocks,
}))

import { useLayoutSessionActionsBridge } from './useLayoutSessionActionsBridge'

describe('useLayoutSessionActionsBridge', () => {
  beforeEach(() => {
    authMocks.logout.mockReset()
    routerMocks.push.mockReset()
  })

  it('应在退出登录后跳回登录页', async () => {
    authMocks.logout.mockResolvedValue(undefined)

    const bridge = useLayoutSessionActionsBridge()

    await bridge.logout()

    expect(authMocks.logout).toHaveBeenCalledTimes(1)
    expect(routerMocks.push).toHaveBeenCalledWith('/login')
  })
})
