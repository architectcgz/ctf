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
    })
  })

  it('父页应默认显示轮次态势 tab，并保留主路由动作', async () => {
    const wrapper = mount(ContestOperations, {
      global: {
        stubs: {
          AppLoading: {
            template: '<div><slot /></div>',
          },
          AWDOperationsPanel: {
            props: ['operationPanel', 'runtimeContent', 'selectedContestId'],
            emits: ['open:contest-edit'],
            template:
              '<div data-testid="awd-ops-panel">{{ selectedContestId }}::{{ operationPanel }}::{{ runtimeContent }}<button class="contest-ops-studio-button" @click="$emit(\'open:contest-edit\')">进入竞赛工作室</button></div>',
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
      'contest-ops-1::inspector::readiness'
    )

    expect(wrapper.find('.ops-topbar').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('返回')

    await wrapper.get('.contest-ops-studio-button').trigger('click')
    expect(pushMock).toHaveBeenLastCalledWith({
      name: 'ContestEdit',
      params: { id: 'contest-ops-1' },
    })
  })

  it('父页应根据 panel query 切到实例编排 tab', async () => {
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

    expect(wrapper.get('#contest-ops-tab-instances').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#contest-ops-panel-instances').exists()).toBe(true)
    expect(wrapper.get('#contest-ops-panel-instances').classes()).toContain('active')
    expect(wrapper.find('#contest-ops-panel-inspector').exists()).toBe(false)
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
    expect(wrapper.get('[data-testid="awd-ops-panel"]').text()).toBe('inspector::readiness')
  })

  it('点击实例编排 tab 时应同步更新 panel query', async () => {
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
    await wrapper.get('#contest-ops-tab-instances').trigger('click')

    expect(replaceMock).toHaveBeenCalledWith({
      name: 'ContestOperations',
      params: { id: 'contest-ops-1' },
      query: { panel: 'instances' },
    })
  })
})
