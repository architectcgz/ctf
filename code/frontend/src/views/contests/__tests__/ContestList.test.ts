import { describe, it, expect, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ContestList from '../ContestList.vue'
import contestListSource from '../ContestList.vue?raw'

const pushMock = vi.fn()

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

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

describe('ContestList', () => {
  it('应该渲染竞赛列表页面', async () => {
    pushMock.mockReset()

    const wrapper = mount(ContestList)

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

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
    expect(contestListSource).not.toContain("from '@/api/contest'")
    expect(contestListSource).not.toContain('usePagination(getContests)')
  })

  it('应该为竞赛列表长标题保留省略样式和完整悬浮提示', () => {
    expect(contestListSource).toMatch(
      /class="[^"]*\bcontest-row-title\b[^"]*"[\s\S]*:title="contest\.title"/s
    )
    expect(contestListSource).toMatch(
      /\.contest-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
  })

  it('竞赛页概况卡片应使用统一 metric-panel 样式类', () => {
    expect(contestListSource).toContain('<div class="workspace-overline">Contests</div>')
    expect(contestListSource).toContain('<h1 class="contest-title workspace-page-title">竞赛中心</h1>')
    expect(contestListSource).not.toContain('<div class="journal-eyebrow">Contests</div>')
    expect(contestListSource).not.toContain('journal-eyebrow-text')
    expect(contestListSource).toContain('class="contest-summary-grid metric-panel-grid"')
    expect(contestListSource).toContain('class="contest-summary-item progress-card metric-panel-card"')
    expect(contestListSource).toContain(
      'class="contest-summary-label progress-card-label metric-panel-label"'
    )
    expect(contestListSource).toContain(
      'class="contest-summary-value progress-card-value metric-panel-value"'
    )
    expect(contestListSource).toContain(
      'class="contest-summary-helper progress-card-hint metric-panel-helper"'
    )
    expect(contestListSource).toContain(
      '<component :is="summaryMetricIcon(stat.key)" class="h-4 w-4" />'
    )
  })

  it('竞赛列表错误态操作按钮应接入共享 ui-btn 原语', () => {
    expect(contestListSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(contestListSource).not.toContain('class="contest-btn"')
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

    const wrapper = mount(ContestList)
    await flushPromises()

    expect(wrapper.text()).toContain('2026 春季校园 CTF 挑战赛')
    expect(wrapper.text()).not.toContain('草稿中的隐藏比赛')
    expect(wrapper.text()).not.toContain('草稿')
  })
})
