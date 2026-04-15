import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestOperationsHub from '../ContestOperationsHub.vue'

const pushMock = vi.fn()
const routeState = vi.hoisted(() => ({
  path: '/admin/contest-ops/environment',
  name: 'AdminContestOpsEnvironment',
}))
const adminApiMocks = vi.hoisted(() => ({
  getContests: vi.fn(),
}))

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    getContests: adminApiMocks.getContests,
  }
})

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock }),
  }
})

describe('ContestOperationsHub', () => {
  beforeEach(() => {
    pushMock.mockReset()
    adminApiMocks.getContests.mockReset()
    adminApiMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'awd-running',
          title: '2026 AWD 联赛',
          description: '运行中赛事',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-04-15T09:00:00.000Z',
          ends_at: '2026-04-15T18:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
  })

  it('renders environment management copy and routes to awd config for the preferred contest', async () => {
    routeState.path = '/admin/contest-ops/environment'
    routeState.name = 'AdminContestOpsEnvironment'

    const wrapper = mount(ContestOperationsHub)
    await flushPromises()

    expect(wrapper.text()).toContain('环境管理')
    expect(wrapper.text()).toContain('2026 AWD 联赛')

    await wrapper.get('#contest-ops-primary-action').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'ContestEdit',
      params: { id: 'awd-running' },
      query: { panel: 'awd-config' },
    })
  })

  it('renders traffic monitoring copy and routes to the operations inspector', async () => {
    routeState.path = '/admin/contest-ops/traffic'
    routeState.name = 'AdminContestOpsTraffic'

    const wrapper = mount(ContestOperationsHub)
    await flushPromises()

    expect(wrapper.text()).toContain('流量监控')

    await wrapper.get('#contest-ops-primary-action').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'ContestEdit',
      params: { id: 'awd-running' },
      query: { panel: 'operations', opsPanel: 'inspector' },
    })
  })
})
