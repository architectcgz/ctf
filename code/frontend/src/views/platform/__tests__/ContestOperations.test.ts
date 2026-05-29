import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestOperations from '../ContestOperations.vue'
import contestOperationsSource from '../ContestOperations.vue?raw'
import contestOperationsPageModelSource from '@/features/platform-contests/model/useContestOperationsPage.ts?raw'
import platformRoutesSource from '@/router/routes/platformRoutes.ts?raw'

const adminApiMocks = vi.hoisted(() => ({
  getContest: vi.fn(),
}))

vi.mock('@/api/admin/contests', async () => {
  const actual =
    await vi.importActual<typeof import('@/api/admin/contests')>('@/api/admin/contests')
  return {
    ...actual,
    getContest: adminApiMocks.getContest,
  }
})

describe('ContestOperations', () => {
  beforeEach(() => {
    adminApiMocks.getContest.mockReset()

    adminApiMocks.getContest.mockResolvedValue({
      id: 'contest-ops-1',
      title: '2026 AWD 运维联赛',
      status: 'running',
    })
  })

  function mountPage(contestId = 'contest-ops-1') {
    return mount(ContestOperations, {
      props: {
        contestId,
      },
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
  }

  it('父页应直接显示轮次态势内容，并只传入运维态能力', async () => {
    const wrapper = mountPage()

    await flushPromises()

    expect(adminApiMocks.getContest).toHaveBeenCalledWith('contest-ops-1')
    expect(wrapper.find('[role="tablist"]').exists()).toBe(false)
    expect(wrapper.find('[role="tabpanel"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="awd-ops-panel"]').text()).toContain(
      'contest-ops-1::inspector::round-inspector::true::true'
    )

    expect(wrapper.find('.ops-topbar').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('返回')
    expect(wrapper.text()).not.toContain('进入竞赛工作室')
  })

  it('路由页应仅负责组合，不直接耦合单场赛事查询流程', () => {
    expect(contestOperationsSource).toContain('useContestOperationsPage')
    expect(contestOperationsSource).not.toContain("from '@/api/admin/contests'")
    expect(contestOperationsPageModelSource).not.toContain("from 'vue-router'")
    expect(platformRoutesSource).toContain("name: 'ContestOperations'")
    expect(platformRoutesSource).toContain("component: () => import('@/views/platform/ContestOperations.vue')")
    expect(platformRoutesSource).toContain('contestId: String(route.params.id || \'\')')
  })

  it('父页不再提供实例编排 tab，而是固定组合运维态面板', async () => {
    const wrapper = mountPage()

    await flushPromises()

    expect(wrapper.find('#contest-ops-tab-instances').exists()).toBe(false)
    expect(wrapper.find('#contest-ops-panel-instances').exists()).toBe(false)
    expect(wrapper.find('#contest-ops-tab-inspector').exists()).toBe(false)
    expect(wrapper.find('#contest-ops-panel-inspector').exists()).toBe(false)
    expect(wrapper.get('[data-testid="awd-ops-panel"]').text()).toContain(
      '::inspector::round-inspector::'
    )
  })

  it('赛事未开赛时才在运维页显示只读就绪摘要', async () => {
    adminApiMocks.getContest.mockResolvedValue({
      id: 'contest-ops-1',
      title: '2026 AWD 运维联赛',
      status: 'registering',
    })

    const wrapper = mountPage()

    await flushPromises()

    expect(wrapper.get('[data-testid="awd-ops-panel"]').text()).toContain(
      '::inspector::readiness::true'
    )
  })

  it('运维页不再渲染 tab 导航', async () => {
    const wrapper = mountPage()

    await flushPromises()

    expect(wrapper.findAll('[role="tab"]')).toHaveLength(0)
    expect(wrapper.find('.top-tabs').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('实例编排')
  })

  it('赛事运维页应通过插槽注入服务告警摘要', async () => {
    const alertWrapper = mount(ContestOperations, {
      props: {
        contestId: 'contest-ops-1',
      },
      global: {
        stubs: {
          AppLoading: {
            template: '<div><slot /></div>',
          },
          AWDOperationsPanel: {
            props: ['operationPanel'],
            methods: {
              getAlertClass() {
                return 'awd-service-alert--danger'
              },
              applyAlertFilter() {},
            },
            template: `
              <div data-testid="awd-ops-panel">
                <slot
                  name="service-alerts"
                  :service-alerts="[{ key: 'service_compromised', label: '服务已失陷', count: 2 }]"
                  selected-alert-key=""
                  :get-service-alert-class="getAlertClass"
                  :apply-service-alert-filter="applyAlertFilter"
                />
              </div>
            `,
          },
        },
      },
    })

    await flushPromises()

    expect(alertWrapper.get('[data-testid="awd-ops-panel"]').text()).toContain('服务已失陷 (2)')
  })

  it('缺少 contestId 时不应发起查询，并会撤掉加载态', async () => {
    const wrapper = mountPage('')

    await flushPromises()

    expect(adminApiMocks.getContest).not.toHaveBeenCalled()
    expect(wrapper.find('.ops-loading-overlay').exists()).toBe(false)
  })
})
