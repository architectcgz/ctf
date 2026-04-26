import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestOperations from '../ContestOperations.vue'

const pushMock = vi.fn()
const routeState = vi.hoisted(() => ({
  params: { id: 'contest-ops-1' },
  query: {} as Record<string, unknown>,
}))
const adminApiMocks = vi.hoisted(() => ({
  getContest: vi.fn(),
}))

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    getContest: adminApiMocks.getContest,
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

describe('ContestOperations', () => {
  beforeEach(() => {
    pushMock.mockReset()
    adminApiMocks.getContest.mockReset()
    routeState.params.id = 'contest-ops-1'
    routeState.query = {}

    adminApiMocks.getContest.mockResolvedValue({
      id: 'contest-ops-1',
      title: '2026 AWD 运维联赛',
    })
  })

  it('父页应保留主路由动作，并将合法 activeTab 传给运维面板', async () => {
    routeState.query = { activeTab: 'attacks' }

    const wrapper = mount(ContestOperations, {
      global: {
        stubs: {
          AppLoading: {
            template: '<div><slot /></div>',
          },
          AWDOperationsPanel: {
            props: ['initialTab', 'selectedContestId'],
            template:
              '<div data-testid="awd-ops-panel">{{ selectedContestId }}::{{ initialTab }}</div>',
          },
        },
      },
    })

    await flushPromises()

    expect(adminApiMocks.getContest).toHaveBeenCalledWith('contest-ops-1')
    expect(wrapper.get('.workspace-page-title').text()).toBe('2026 AWD 运维联赛')
    expect(wrapper.get('[data-testid="awd-ops-panel"]').text()).toBe('contest-ops-1::attacks')

    expect(wrapper.find('.ops-topbar').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('返回')

    await wrapper.get('.contest-ops-studio-button').trigger('click')
    expect(pushMock).toHaveBeenLastCalledWith({
      name: 'ContestEdit',
      params: { id: 'contest-ops-1' },
    })
  })

  it('父页应在 query 提供非法 activeTab 时回退到 matrix', async () => {
    routeState.query = { activeTab: 'unknown-tab' }

    const wrapper = mount(ContestOperations, {
      global: {
        stubs: {
          AppLoading: {
            template: '<div><slot /></div>',
          },
          AWDOperationsPanel: {
            props: ['initialTab'],
            template: '<div data-testid="awd-ops-panel">{{ initialTab }}</div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.get('[data-testid="awd-ops-panel"]').text()).toBe('matrix')
  })
})
