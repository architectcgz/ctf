import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestOperations from '../ContestOperations.vue'

const pushMock = vi.fn()
const replaceMock = vi.fn()
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
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
  }
})

describe('ContestOperations', () => {
  beforeEach(() => {
    pushMock.mockReset()
    replaceMock.mockReset()
    adminApiMocks.getContest.mockReset()
    routeState.params.id = 'contest-ops-1'
    routeState.query = {}

    adminApiMocks.getContest.mockResolvedValue({
      id: 'contest-ops-1',
      title: '2026 AWD 运维联赛',
      status: 'running',
    })
  })

  it('父页应默认显示轮次态势 tab，并只传入运维态能力', async () => {
    const wrapper = mount(ContestOperations, {
      global: {
        stubs: {
          AppLoading: {
            template: '<div><slot /></div>',
          },
          AWDOperationsPanel: {
            props: [
              'operationPanel',
              'runtimeContent',
              'selectedContestId',
              'hideStudioLink',
              'hideReadinessActions',
            ],
            template:
              '<div data-testid="awd-ops-panel">{{ selectedContestId }}::{{ operationPanel }}::{{ runtimeContent }}::{{ hideStudioLink }}::{{ hideReadinessActions }}</div>',
          },
        },
      },
    })

    await flushPromises()

    expect(adminApiMocks.getContest).toHaveBeenCalledWith('contest-ops-1')
    expect(wrapper.get('#contest-ops-tab-inspector').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#contest-ops-panel-inspector').exists()).toBe(true)
    expect(wrapper.get('#contest-ops-panel-inspector').classes()).toContain('active')
    expect(wrapper.get('[data-testid="awd-ops-panel"]').text()).toContain(
      'contest-ops-1::inspector::round-inspector::true::true'
    )

    expect(wrapper.find('.ops-topbar').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('返回')
    expect(wrapper.text()).not.toContain('进入竞赛工作室')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('父页不再提供实例编排 tab，panel=instances 应回退到轮次态势', async () => {
    routeState.query = { panel: 'instances' }

    const wrapper = mount(ContestOperations, {
      global: {
        stubs: {
          AppLoading: {
            template: '<div><slot /></div>',
          },
          AWDOperationsPanel: {
            props: ['operationPanel', 'runtimeContent'],
            template: '<div data-testid="awd-ops-panel">{{ operationPanel }}::{{ runtimeContent }}</div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('#contest-ops-tab-instances').exists()).toBe(false)
    expect(wrapper.find('#contest-ops-panel-instances').exists()).toBe(false)
    expect(wrapper.get('#contest-ops-tab-inspector').attributes('aria-selected')).toBe('true')
    expect(wrapper.get('#contest-ops-panel-inspector').classes()).toContain('active')
    expect(wrapper.get('[data-testid="awd-ops-panel"]').text()).toBe('inspector::round-inspector')
  })

  it('父页应在 query 提供非法 panel 时回退到轮次态势', async () => {
    routeState.query = { panel: 'unknown-panel' }

    const wrapper = mount(ContestOperations, {
      global: {
        stubs: {
          AppLoading: {
            template: '<div><slot /></div>',
          },
          AWDOperationsPanel: {
            props: ['operationPanel', 'runtimeContent'],
            template: '<div data-testid="awd-ops-panel">{{ operationPanel }}::{{ runtimeContent }}</div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.get('#contest-ops-tab-inspector').attributes('aria-selected')).toBe('true')
    expect(wrapper.get('[data-testid="awd-ops-panel"]').text()).toBe('inspector::round-inspector')
  })

  it('赛事未开赛时才在运维页显示只读就绪摘要', async () => {
    adminApiMocks.getContest.mockResolvedValue({
      id: 'contest-ops-1',
      title: '2026 AWD 运维联赛',
      status: 'registering',
    })

    const wrapper = mount(ContestOperations, {
      global: {
        stubs: {
          AppLoading: {
            template: '<div><slot /></div>',
          },
          AWDOperationsPanel: {
            props: ['operationPanel', 'runtimeContent', 'hideReadinessActions'],
            template:
              '<div data-testid="awd-ops-panel">{{ operationPanel }}::{{ runtimeContent }}::{{ hideReadinessActions }}</div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.get('[data-testid="awd-ops-panel"]').text()).toBe('inspector::readiness::true')
  })

  it('运维页只保留轮次态势 tab', async () => {
    const wrapper = mount(ContestOperations, {
      global: {
        stubs: {
          AppLoading: {
            template: '<div><slot /></div>',
          },
          AWDOperationsPanel: {
            props: ['operationPanel'],
            template: '<div data-testid="awd-ops-panel">{{ operationPanel }}</div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.findAll('[role="tab"]')).toHaveLength(1)
    expect(wrapper.get('#contest-ops-tab-inspector').text()).toContain('轮次态势')
    expect(wrapper.text()).not.toContain('实例编排')
    expect(replaceMock).not.toHaveBeenCalled()
  })
})
