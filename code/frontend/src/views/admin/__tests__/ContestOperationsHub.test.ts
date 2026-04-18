import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestOperationsHub from '../ContestOperationsHub.vue'
import contestOperationsHubSource from '../ContestOperationsHub.vue?raw'

const pushMock = vi.fn()
const routeState = vi.hoisted(() => ({
  path: '/platform/contest-ops/contests',
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

  it('renders contest management as an awd contest directory and routes from row actions', async () => {
    routeState.path = '/platform/contest-ops/contests'
    routeState.name = 'AdminContestOpsEnvironment'

    const wrapper = mount(ContestOperationsHub)
    await flushPromises()

    expect(wrapper.text()).toContain('竞赛管理')
    expect(wrapper.text()).toContain('全部 AWD 赛事')
    expect(wrapper.text()).toContain('2026 AWD 联赛')
    expect(wrapper.text()).not.toContain('按开始时间查看全部可操作 AWD 赛事')
    expect(wrapper.text()).not.toContain(
      '这里直接承接可运维的 AWD 赛事，用统一目录处理 checker、SLA、防守分和赛前准备，不再通过漂浮入口反复跳转。'
    )

    await wrapper.get('#contest-ops-row-primary-awd-running').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'ContestEdit',
      params: { id: 'awd-running' },
      query: { panel: 'awd-config' },
    })
  })

  it('renders traffic monitoring as an awd contest directory and routes from row actions', async () => {
    routeState.path = '/platform/contest-ops/traffic'
    routeState.name = 'AdminContestOpsTraffic'

    const wrapper = mount(ContestOperationsHub)
    await flushPromises()

    expect(wrapper.text()).toContain('流量监控')
    expect(wrapper.text()).toContain('全部 AWD 赛事')

    await wrapper.get('#contest-ops-row-primary-awd-running').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'ContestEdit',
      params: { id: 'awd-running' },
      query: { panel: 'operations', opsPanel: 'inspector' },
    })
  })

  it('uses a shared workbench header, metric strip, and flat awd directory instead of entry cards', () => {
    expect(contestOperationsHubSource).toContain(
      '<header class="list-heading contest-ops-workbench-head">'
    )
    expect(contestOperationsHubSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface contest-ops-summary"'
    )
    expect(contestOperationsHubSource).toContain(
      'class="workspace-directory-section contest-ops-directory-section"'
    )
    expect(contestOperationsHubSource).toContain(
      'class="contest-ops-directory workspace-directory-list"'
    )
    expect(contestOperationsHubSource).toContain(
      'class="ui-row-actions contest-ops-row__actions"'
    )
    expect(contestOperationsHubSource).not.toContain('class="contest-ops-grid"')
    expect(contestOperationsHubSource).not.toContain('class="contest-ops-card"')
    expect(contestOperationsHubSource).not.toContain('继续处理当前赛事')
  })
})
