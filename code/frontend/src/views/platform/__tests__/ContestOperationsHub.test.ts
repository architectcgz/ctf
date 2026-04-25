import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestOperationsHub from '../ContestOperationsHub.vue'
import contestOperationsHubSource from '../ContestOperationsHub.vue?raw'
import contestOperationsHubHeroPanelSource from '@/components/platform/contest/ContestOperationsHubHeroPanel.vue?raw'
import contestOperationsHubWorkspacePanelSource from '@/components/platform/contest/ContestOperationsHubWorkspacePanel.vue?raw'

const pushMock = vi.fn()
const routeState = vi.hoisted(() => ({
  path: '/platform/contest-ops/contests',
  name: 'PlatformContestOpsIndex',
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
        {
          id: 'awd-frozen',
          title: '2026 AWD 冻结赛',
          description: '封榜阶段',
          mode: 'awd',
          status: 'frozen',
          starts_at: '2026-04-16T09:00:00.000Z',
          ends_at: '2026-04-16T18:00:00.000Z',
        },
        {
          id: 'jeopardy-1',
          title: '2026 Jeopardy 校内赛',
          description: '非 AWD',
          mode: 'jeopardy',
          status: 'running',
          starts_at: '2026-04-17T09:00:00.000Z',
          ends_at: '2026-04-17T18:00:00.000Z',
        },
      ],
      total: 3,
      page: 1,
      page_size: 20,
    })
  })

  it('renders contest ops directory copy and lists operable awd contests', async () => {
    const wrapper = mount(ContestOperationsHub)
    await flushPromises()

    expect(wrapper.text()).toContain('赛事运维')
    expect(wrapper.text()).toContain('竞赛列表')
    expect(wrapper.text()).toContain('2026 AWD 联赛')
    expect(wrapper.text()).toContain('2026 AWD 冻结赛')
    expect(wrapper.text()).not.toContain('2026 Jeopardy 校内赛')
    expect(wrapper.text()).toContain('进入运维台')
  })

  it('routes to the per-contest operations workspace from the directory entry', async () => {
    const wrapper = mount(ContestOperationsHub)
    await flushPromises()

    await wrapper.get('#contest-ops-enter-awd-running').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'ContestOperations',
      params: { id: 'awd-running' },
    })
  })

  it('shows an empty state when no awd contest can enter ops', async () => {
    adminApiMocks.getContests.mockResolvedValueOnce({
      list: [
        {
          id: 'jeopardy-1',
          title: '2026 Jeopardy 校内赛',
          description: '非 AWD',
          mode: 'jeopardy',
          status: 'running',
          starts_at: '2026-04-17T09:00:00.000Z',
          ends_at: '2026-04-17T18:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestOperationsHub)
    await flushPromises()

    expect(wrapper.text()).toContain('当前还没有可进入运维台的 AWD 赛事')
    expect(wrapper.text()).toContain('返回竞赛目录')
  })

  it('uses shared directory heading and metric primitives for the ops index shell', () => {
    expect(contestOperationsHubHeroPanelSource).toContain(
      '<header class="list-heading contest-ops-hero workspace-directory-section">'
    )
    expect(contestOperationsHubHeroPanelSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface contest-ops-summary"'
    )
    expect(contestOperationsHubHeroPanelSource).toContain(
      '--metric-panel-columns: repeat(4, minmax(0, 1fr));'
    )
    expect(contestOperationsHubHeroPanelSource).toContain(
      '--metric-panel-columns: repeat(2, minmax(0, 1fr));'
    )
    expect(contestOperationsHubHeroPanelSource).not.toContain('--metric-panel-columns: 4;')
    expect(contestOperationsHubHeroPanelSource).toContain('<Trophy class="h-4 w-4" />')
    expect(contestOperationsHubHeroPanelSource).toContain('<Activity class="h-4 w-4" />')
    expect(contestOperationsHubHeroPanelSource).toContain('<PauseCircle class="h-4 w-4" />')
    expect(contestOperationsHubHeroPanelSource).toContain('<Star class="h-4 w-4" />')
    expect(contestOperationsHubSource).toContain('class="content-pane contest-ops-content"')
    expect(contestOperationsHubSource).toContain(
      'gap: var(--workspace-directory-page-block-gap, var(--space-5));'
    )
    expect(contestOperationsHubSource).toContain('<ContestOperationsHubWorkspacePanel')
    expect(contestOperationsHubWorkspacePanelSource).toContain('contest-ops-directory')
    expect(contestOperationsHubWorkspacePanelSource).toContain(
      'class="workspace-directory-list contest-ops-directory__list"'
    )
    expect(contestOperationsHubWorkspacePanelSource).toContain('class="contest-ops-row"')
    expect(contestOperationsHubWorkspacePanelSource).not.toContain('class="contest-ops-card"')
    expect(contestOperationsHubHeroPanelSource).not.toContain('margin-top: var(--space-5);')
    expect(contestOperationsHubWorkspacePanelSource).toContain('padding: 0;')
    expect(contestOperationsHubWorkspacePanelSource).toContain('gap: var(--space-4);')
    expect(contestOperationsHubWorkspacePanelSource).toContain('border-bottom: 1px solid var(--workspace-directory-row-divider);')
    expect(contestOperationsHubSource).not.toContain('PlatformContestOpsTraffic')
    expect(contestOperationsHubSource).not.toContain('PlatformContestOpsProjector')
    expect(contestOperationsHubSource).not.toContain('PlatformContestOpsScoreboard')
  })
})
