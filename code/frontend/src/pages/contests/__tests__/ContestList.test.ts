import { describe, it, expect, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'

import ContestList from '@/pages/contests/ContestListRoutePage.vue'
import appRouteLinkSource from '@/shared/ui/navigation/AppRouteLink.vue?raw'
import contestListSource from '@/pages/contests/ContestListRoutePage.vue?raw'
import contestListPageSource from '@/features/contest-detail/model/useContestListPage.ts?raw'
import contestListWorkspaceSource from '@/widgets/contest-list-workspace/ContestListWorkspace.vue?raw'
import contestListRoutesSource from '@/features/contest-detail/model/contestListRoutes.ts?raw'

vi.mock('@/api/contest', () => ({
  getContests: vi.fn().mockResolvedValue({
    list: [
      {
        id: '1',
        title: '2026 春季校园 CTF 挑战赛',
        status: 'running',
        mode: 'jeopardy',
        starts_at: '2024-03-15T09:00:00Z',
        ends_at: '2024-03-15T21:00:00Z',
      },
    ],
    total: 1,
    page: 1,
    page_size: 20,
    summary: {
      draft_count: 0,
      registering_count: 0,
      running_count: 1,
      frozen_count: 0,
      ended_count: 0,
    },
  }),
}))

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      {
        path: '/contests',
        name: 'Contests',
        component: ContestList,
      },
      {
        path: '/contests/:id',
        name: 'ContestDetail',
        component: { template: '<div>contest detail</div>' },
      },
    ],
  })
}

async function mountPage() {
  const router = createTestRouter()
  await router.push('/contests')
  await router.isReady()

  const wrapper = mount(ContestList, {
    global: {
      plugins: [router],
    },
  })

  await flushPromises()
  return { wrapper, router }
}

describe('ContestList', () => {
  it('应该渲染竞赛列表页面', async () => {
    const { wrapper } = await mountPage()

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('Contests')
    expect(wrapper.text()).toContain('竞赛中心')
    expect(wrapper.find('.contest-row-title').attributes('title')).toBe('2026 春季校园 CTF 挑战赛')

    const { getContests } = await import('@/api/contest')
    expect(vi.mocked(getContests)).toHaveBeenCalledWith(
      {
        page: 1,
        page_size: 20,
        statuses: ['registering', 'running', 'frozen', 'ended'],
      },
      { signal: expect.any(AbortSignal) }
    )
  })

  it('路由页应仅负责组合，不直接耦合竞赛列表查询流程', () => {
    expect(contestListSource).toContain('useContestListPage')
    expect(contestListSource).toContain(
      "import { ContestListWorkspace } from '@/widgets/contest-list-workspace'"
    )
    expect(contestListSource).not.toContain("from '@/api/contest'")
    expect(contestListSource).not.toContain('usePagination(getContests)')
    expect(contestListSource).not.toContain('summaryMetricIcon')
    expect(contestListSource).not.toContain('contest-directory-filters')
  })

  it('竞赛详情入口应通过 route target，而不是 page model 直接 push', async () => {
    const { wrapper, router } = await mountPage()

    expect(contestListWorkspaceSource).toContain("from '@/shared/ui/navigation/AppRouteLink.vue'")
    expect(contestListWorkspaceSource).toContain('<AppRouteLink')
    expect(contestListPageSource).not.toContain("from 'vue-router'")
    expect(contestListPageSource).not.toContain('router.push')
    expect(contestListPageSource).toContain('buildContestRoute: (contest: ContestListItem) =>')
    expect(contestListRoutesSource).toContain("name: 'ContestDetail'")
    expect(appRouteLinkSource).toContain('<RouterLink')

    await wrapper.get('a.contest-row').trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('ContestDetail')
    expect(router.currentRoute.value.params).toMatchObject({ id: '1' })
  })

  it('应该为竞赛列表长标题保留省略样式和完整悬浮提示', () => {
    expect(contestListWorkspaceSource).toMatch(
      /class="[^"]*\bcontest-row-title\b[^"]*"[\s\S]*:title="contest\.title"/s
    )
    expect(contestListWorkspaceSource).toMatch(
      /\.contest-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
  })

  it('竞赛页概况卡片应使用统一 metric-panel 样式类', () => {
    expect(contestListWorkspaceSource).toContain('<div class="workspace-overline">Contests</div>')
    expect(contestListWorkspaceSource).toContain(
      '<h1 class="contest-title workspace-page-title">竞赛中心</h1>'
    )
    expect(contestListWorkspaceSource).not.toContain('<div class="journal-eyebrow">Contests</div>')
    expect(contestListWorkspaceSource).not.toContain('journal-eyebrow-text')
    expect(contestListWorkspaceSource).toContain('class="contest-summary-grid metric-panel-grid"')
    expect(contestListWorkspaceSource).toContain(
      'class="contest-summary-item progress-card metric-panel-card"'
    )
    expect(contestListWorkspaceSource).toContain(
      'class="contest-summary-label progress-card-label metric-panel-label"'
    )
    expect(contestListWorkspaceSource).toContain(
      'class="contest-summary-value progress-card-value metric-panel-value"'
    )
    expect(contestListWorkspaceSource).toContain(
      'class="contest-summary-helper progress-card-hint metric-panel-helper"'
    )
    expect(contestListWorkspaceSource).toContain(
      '<component :is="summaryMetricIcon(stat.key)" class="h-4 w-4" />'
    )
  })

  it('竞赛列表错误态操作按钮应接入共享 ui-btn 原语', () => {
    expect(contestListWorkspaceSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(contestListWorkspaceSource).not.toContain('class="contest-btn"')
  })

  it('竞赛列表应拆分开始时间、结束时间和通用操作按钮图标', () => {
    expect(contestListWorkspaceSource).toContain('<span>开始时间</span>')
    expect(contestListWorkspaceSource).toContain('<span>结束时间</span>')
    expect(contestListWorkspaceSource).toContain(
      'class="workspace-directory-compact-text contest-row-start-time"'
    )
    expect(contestListWorkspaceSource).toContain(
      'class="workspace-directory-compact-text contest-row-end-time"'
    )
    expect(contestListWorkspaceSource).not.toContain(
      'formatTime(contest.starts_at) }} - {{ formatTime(contest.ends_at)'
    )
    expect(contestListWorkspaceSource).toContain('class="workspace-directory-row-btn contest-row-cta"')
    expect(contestListWorkspaceSource).toContain('<ArrowRight class="h-4 w-4" />')
    expect(contestListWorkspaceSource).toContain('minmax(10.5rem, 0.85fr) max-content')
    expect(contestListWorkspaceSource).toMatch(/\.contest-row-cta\s*\{[^}]*justify-self:\s*end;/s)
  })

  it('竞赛列表应提供学生通用状态与模式筛选，并透传后端查询参数', async () => {
    const { getContests } = await import('@/api/contest')
    vi.mocked(getContests).mockClear()

    const { wrapper } = await mountPage()

    expect(contestListWorkspaceSource).toContain(
      'class="student-directory-filters contest-directory-filters"'
    )
    expect(wrapper.find('#contest-status-filter').exists()).toBe(true)
    expect(wrapper.find('#contest-mode-filter').exists()).toBe(true)

    await wrapper.get('#contest-status-filter').setValue('running')
    await flushPromises()

    expect(vi.mocked(getContests)).toHaveBeenLastCalledWith(
      {
        page: 1,
        page_size: 20,
        statuses: ['running'],
      },
      { signal: expect.any(AbortSignal) }
    )

    await wrapper.get('#contest-mode-filter').setValue('awd')
    await flushPromises()

    expect(vi.mocked(getContests)).toHaveBeenLastCalledWith(
      {
        page: 1,
        page_size: 20,
        statuses: ['running'],
        mode: 'awd',
      },
      { signal: expect.any(AbortSignal) }
    )
  })

  it('不应该向学生暴露草稿竞赛，也不应把草稿错误渲染为已结束', async () => {
    const { getContests } = await import('@/api/contest')
    vi.mocked(getContests).mockResolvedValueOnce({
      list: [
        {
          id: '1',
          title: '2026 春季校园 CTF 挑战赛',
          status: 'running',
          mode: 'jeopardy',
          starts_at: '2024-03-15T09:00:00Z',
          ends_at: '2024-03-15T21:00:00Z',
        },
        {
          id: '2',
          title: '草稿中的隐藏比赛',
          status: 'draft',
          mode: 'jeopardy',
          starts_at: '2024-03-16T09:00:00Z',
          ends_at: '2024-03-16T21:00:00Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
      summary: {
        draft_count: 1,
        registering_count: 0,
        running_count: 1,
        frozen_count: 0,
        ended_count: 0,
      },
    })

    const { wrapper } = await mountPage()

    expect(wrapper.text()).toContain('2026 春季校园 CTF 挑战赛')
    expect(wrapper.text()).not.toContain('草稿中的隐藏比赛')
    expect(wrapper.text()).not.toContain('草稿')
  })
})
