import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

import ScoreboardDetail from '../ScoreboardDetail.vue'
import ScoreboardView from '../ScoreboardView.vue'
import scoreboardSource from '../ScoreboardView.vue?raw'
import scoreboardDetailSource from '../ScoreboardDetail.vue?raw'

function createScoreboardRouter(initialPath = '/scoreboard') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/scoreboard', name: 'Scoreboard', component: ScoreboardView },
      { path: '/scoreboard/:contestId', name: 'ScoreboardDetail', component: ScoreboardDetail },
    ],
  })
  return router.push(initialPath).then(() => router.isReady()).then(() => router)
}

const { getContestsMock, getScoreboardMock, getPracticeRankingMock } = vi.hoisted(() => ({
  getContestsMock: vi.fn(),
  getScoreboardMock: vi.fn(),
  getPracticeRankingMock: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
}))

const webSocketMocks = vi.hoisted(() => {
  const connect = vi.fn().mockResolvedValue(undefined)
  const disconnect = vi.fn()
  const send = vi.fn()
  const handlersByEndpoint = new Map<string, Record<string, (payload: unknown) => void>>()

  return {
    connect,
    disconnect,
    send,
    getHandlers: (endpoint: string) => handlersByEndpoint.get(endpoint),
    reset: () => handlersByEndpoint.clear(),
    useWebSocket: vi.fn(
      (endpoint: string, handlers: Record<string, (payload: unknown) => void>) => {
        handlersByEndpoint.set(endpoint, handlers)
        return {
          status: { value: 'idle' as const },
          connect,
          disconnect,
          send,
        }
      }
    ),
  }
})

vi.mock('@/api/contest', () => ({
  getContests: getContestsMock,
  getScoreboard: getScoreboardMock,
}))

vi.mock('@/api/scoreboard', () => ({
  getPracticeRanking: getPracticeRankingMock,
}))

vi.mock('@/composables/useWebSocket', () => ({
  useWebSocket: webSocketMocks.useWebSocket,
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

describe('ScoreboardView', () => {
  beforeEach(() => {
    getContestsMock.mockReset()
    getScoreboardMock.mockReset()
    getPracticeRankingMock.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()
    toastMocks.info.mockReset()
    webSocketMocks.connect.mockClear()
    webSocketMocks.disconnect.mockClear()
    webSocketMocks.send.mockClear()
    webSocketMocks.useWebSocket.mockClear()
    webSocketMocks.reset()
  })

  it('按最新竞赛在前的顺序展示排行榜列表', async () => {
    getContestsMock.mockResolvedValue({
      list: [
        {
          id: 'contest-old',
          title: '往期竞赛',
          mode: 'jeopardy',
          status: 'ended',
          starts_at: '2026-03-01T00:00:00Z',
          ends_at: '2026-03-01T12:00:00Z',
        },
        {
          id: 'contest-running',
          title: '当前竞赛',
          mode: 'jeopardy',
          status: 'running',
          starts_at: '2026-03-12T00:00:00Z',
          ends_at: '2026-03-12T12:00:00Z',
        },
        {
          id: 'contest-registering',
          title: '报名中竞赛',
          mode: 'jeopardy',
          status: 'registering',
          starts_at: '2026-03-14T00:00:00Z',
          ends_at: '2026-03-14T12:00:00Z',
        },
        {
          id: 'contest-frozen',
          title: '冻结竞赛',
          mode: 'jeopardy',
          status: 'frozen',
          starts_at: '2026-03-10T00:00:00Z',
          ends_at: '2026-03-10T12:00:00Z',
        },
      ],
      total: 4,
      page: 1,
      page_size: 100,
    })

    const router = await createScoreboardRouter()

    const wrapper = mount(ScoreboardView, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    const cards = wrapper.findAll('[data-testid="scoreboard-card"]')

    expect(wrapper.text()).toContain('竞赛排行榜')
    expect(wrapper.text()).toContain('积分排行榜')
    expect(wrapper.find('#scoreboard-panel-contest').classes()).toContain('active')
    expect(wrapper.find('#scoreboard-panel-points').classes()).not.toContain('active')
    expect(cards).toHaveLength(3)
    expect(cards[0].text()).toContain('当前竞赛')
    expect(cards[1].text()).toContain('冻结竞赛')
    expect(cards[2].text()).toContain('往期竞赛')
    expect(wrapper.text()).toContain('进行中竞赛进入详情后支持实时刷新，提交后榜单会自动更新。')
    expect(wrapper.text()).toContain('封榜阶段先展示竞赛入口，进入后查看冻结前排名。')
    expect(wrapper.text()).toContain('历史竞赛进入详情后展示最终成绩，可用于复盘队伍解题表现。')
    expect(wrapper.text()).toContain('当前可查看排行的竞赛总数')
    expect(wrapper.text()).toContain('支持进入后实时刷新的竞赛数量')
    expect(wrapper.text()).not.toContain('报名中竞赛')
    expect(wrapper.find('a[href="/scoreboard/contest-running"]').exists()).toBe(true)
    expect(getScoreboardMock).not.toHaveBeenCalled()
  })

  it('排行详情路由页应仅负责组合，不直接耦合排行榜详情加载流程', () => {
    expect(scoreboardDetailSource).toContain('useScoreboardDetailPage')
    expect(scoreboardDetailSource).not.toContain("from '@/api/contest'")
    expect(scoreboardDetailSource).not.toContain('watch(')
  })

  it('排行榜路由页应仅做组合，不直接持有路由查询tab编排逻辑', () => {
    expect(scoreboardSource).toContain('useScoreboardRoutePage')
    expect(scoreboardSource).not.toContain('useRouteQueryTabs')
    expect(scoreboardSource).not.toContain('useRoute')
    expect(scoreboardSource).not.toContain('useRouter')
  })

  it('竞赛排行列表不直接展开当前排行和历史排行内容', async () => {
    getPracticeRankingMock.mockResolvedValue([])
    getContestsMock.mockResolvedValue({
      list: [
        {
          id: 'contest-history',
          title: '往期竞赛',
          mode: 'jeopardy',
          status: 'ended',
          starts_at: '2026-03-01T00:00:00Z',
          ends_at: '2026-03-01T12:00:00Z',
        },
        {
          id: 'contest-running',
          title: '当前竞赛',
          mode: 'jeopardy',
          status: 'running',
          starts_at: '2026-03-12T00:00:00Z',
          ends_at: '2026-03-12T12:00:00Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 100,
    })

    const router = await createScoreboardRouter()

    const wrapper = mount(ScoreboardView, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('当前竞赛')
    expect(wrapper.text()).toContain('往期竞赛')
    expect(wrapper.text()).toContain('点击进入排行详情')
    expect(wrapper.text()).not.toContain('Current Champions')
    expect(wrapper.text()).not.toContain('History Masters')
    expect(wrapper.text()).not.toContain('查看完整排行榜')
    expect(getScoreboardMock).not.toHaveBeenCalled()
  })

  it('排行详情页收到 scoreboard.updated 后刷新当前竞赛', async () => {
    getScoreboardMock.mockImplementation(async (contestId: string) => {
      const refreshCount = getScoreboardMock.mock.calls.filter(([id]) => id === contestId).length
      return {
        contest: {
          id: contestId,
          title: '当前竞赛',
          status: 'running',
          started_at: '2026-03-12T00:00:00Z',
          ends_at: '2026-03-12T12:00:00Z',
        },
        scoreboard: {
          list: [
            {
              rank: 1,
              team_id: 'team-running',
              team_name: refreshCount >= 2 ? 'Updated Champions' : 'Current Champions',
              score: refreshCount >= 2 ? 2550 : 2450,
              solved_count: 8,
              last_submission_at: '2026-03-12T10:15:00Z',
            },
          ],
          total: 1,
          page: 1,
          page_size: 20,
        },
        frozen: false,
      }
    })

    const router = await createScoreboardRouter('/scoreboard/contest-running')
    const wrapper = mount(ScoreboardDetail, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    expect(getScoreboardMock).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('Current Champions')

    webSocketMocks.getHandlers('contests/contest-running/scoreboard')?.['scoreboard.updated']?.({
      contest_id: 'contest-running',
    })

    await flushPromises()
    await flushPromises()

    expect(getScoreboardMock).toHaveBeenCalledTimes(2)
    expect(getScoreboardMock.mock.calls.at(-1)?.[0]).toBe('contest-running')
    expect(wrapper.text()).toContain('Updated Champions')
  })

  it('排行详情页只展示指定竞赛的排行内容', async () => {
    getScoreboardMock.mockResolvedValue({
      contest: {
        id: 'contest-history',
        title: '往期竞赛',
        status: 'ended',
        started_at: '2026-03-01T00:00:00Z',
        ends_at: '2026-03-01T12:00:00Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: 'team-history',
            team_name: 'History Masters',
            score: 1980,
            solved_count: 6,
            last_submission_at: '2026-03-01T10:15:00Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 20,
      },
      frozen: false,
    })

    const router = await createScoreboardRouter('/scoreboard/contest-history')
    const wrapper = mount(ScoreboardDetail, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    expect(getScoreboardMock).toHaveBeenCalledWith('contest-history', { page: 1, page_size: 100 })
    expect(wrapper.text()).toContain('往期竞赛')
    expect(wrapper.text()).toContain('History Masters')
    expect(wrapper.findAll('[data-testid="scoreboard-detail-row"]')).toHaveLength(1)
  })

  it('排行详情静默刷新失败时应保留旧数据，只弹一次刷新失败提示', async () => {
    getScoreboardMock
      .mockResolvedValueOnce({
        contest: {
          id: 'contest-running',
          title: '当前竞赛',
          status: 'running',
          started_at: '2026-03-12T00:00:00Z',
          ends_at: '2026-03-12T12:00:00Z',
        },
        scoreboard: {
          list: [
            {
              rank: 1,
              team_id: 'team-running',
              team_name: 'Current Champions',
              score: 2450,
              solved_count: 8,
              last_submission_at: '2026-03-12T10:15:00Z',
            },
          ],
          total: 1,
          page: 1,
          page_size: 20,
        },
        frozen: false,
      })
      .mockRejectedValueOnce(new Error('refresh failed'))

    const router = await createScoreboardRouter('/scoreboard/contest-running')
    const wrapper = mount(ScoreboardDetail, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('Current Champions')

    await wrapper.get('button.ui-btn--secondary').trigger('click')
    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('Current Champions')
    expect(wrapper.text()).not.toContain('排行榜加载失败')
    expect(toastMocks.error).toHaveBeenCalledWith('排行榜刷新失败')
  })

  it('列表页不会为竞赛列表建立实时排行榜连接', async () => {
    getPracticeRankingMock.mockResolvedValue([])
    getContestsMock.mockResolvedValue({
      list: [
        {
          id: 'contest-history',
          title: '往期竞赛',
          mode: 'jeopardy',
          status: 'ended',
          starts_at: '2026-03-01T00:00:00Z',
          ends_at: '2026-03-01T12:00:00Z',
        },
        {
          id: 'contest-running',
          title: '当前竞赛',
          mode: 'jeopardy',
          status: 'running',
          starts_at: '2026-03-12T00:00:00Z',
          ends_at: '2026-03-12T12:00:00Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 100,
    })

    const router = await createScoreboardRouter()
    mount(ScoreboardView, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    expect(getScoreboardMock).not.toHaveBeenCalled()
    expect(webSocketMocks.getHandlers('contests/contest-history/scoreboard')).toBeUndefined()
    expect(webSocketMocks.getHandlers('contests/contest-running/scoreboard')).toBeUndefined()
  })

  it('排行榜页概况卡片应使用统一 metric-panel 样式类', () => {
    expect(scoreboardSource).toContain('class="scoreboard-summary-grid metric-panel-grid"')
    expect(scoreboardSource).toContain('class="scoreboard-summary-item metric-panel-card"')
    expect(scoreboardSource).toContain('class="scoreboard-summary-label metric-panel-label"')
    expect(scoreboardSource).toContain('class="scoreboard-summary-value metric-panel-value"')
    expect(scoreboardSource).toContain('class="scoreboard-summary-helper metric-panel-helper"')
  })

  it('tabs 应直接位于页面顶部，points 页签不应重复渲染局部页头', () => {
    expect(scoreboardSource).not.toContain('<header class="scoreboard-topbar">')
    expect(scoreboardSource).not.toContain('<h2 class="scoreboard-directory-title">积分排行榜</h2>')
    expect(scoreboardSource).toContain('Contest Scoreboard')
    expect(scoreboardSource).toContain('Points Scoreboard')
  })

  it('排行榜顶部页签应接入统一 workspace tab 样式', () => {
    expect(scoreboardSource).toContain('class="workspace-tabbar top-tabs"')
    expect(scoreboardSource).toContain('class="workspace-tab top-tab"')
    expect(scoreboardSource).not.toContain('--page-top-tabs-gap: var(--space-7);')
    expect(scoreboardSource).not.toContain('--page-top-tabs-padding: 0 var(--space-7);')
    expect(scoreboardSource).not.toContain('--page-top-tab-min-height: 52px;')
  })

  it('竞赛排行列表分页展示并支持切换下一页', async () => {
    getPracticeRankingMock.mockResolvedValue([])
    getContestsMock.mockResolvedValue({
      list: Array.from({ length: 7 }, (_, index) => {
        const day = String(12 - index).padStart(2, '0')
        return {
          id: `contest-${index + 1}`,
          title: `竞赛 ${index + 1}`,
          mode: 'jeopardy',
          status: 'running',
          starts_at: `2026-03-${day}T00:00:00Z`,
          ends_at: `2026-03-${day}T12:00:00Z`,
        }
      }),
      total: 7,
      page: 1,
      page_size: 100,
    })

    const router = await createScoreboardRouter()
    const wrapper = mount(ScoreboardView, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    const firstPageCards = wrapper.findAll('[data-testid="scoreboard-card"]')
    expect(firstPageCards).toHaveLength(6)
    expect(wrapper.find('.scoreboard-pagination').text()).toContain('共 7 个竞赛')
    expect(wrapper.find('.scoreboard-pagination').text()).toContain('1 / 2')
    expect(firstPageCards.some((card) => card.text().includes('竞赛 1'))).toBe(true)
    expect(firstPageCards.some((card) => card.text().includes('竞赛 7'))).toBe(false)

    const nextButton = wrapper
      .findAll('.page-pagination-controls__button')
      .find((button) => button.text().trim() === '下一页')
    expect(nextButton).toBeTruthy()
    await nextButton?.trigger('click')

    const secondPageCards = wrapper.findAll('[data-testid="scoreboard-card"]')
    expect(secondPageCards).toHaveLength(1)
    expect(wrapper.find('.scoreboard-pagination').text()).toContain('2 / 2')
    expect(secondPageCards[0].text()).toContain('竞赛 7')
    expect(secondPageCards[0].text()).not.toContain('竞赛 1')
  })

  it('竞赛排行列表只有一页时也应显示分页控件', async () => {
    getPracticeRankingMock.mockResolvedValue([])
    getContestsMock.mockResolvedValue({
      list: [
        {
          id: 'contest-running',
          title: '当前竞赛',
          mode: 'jeopardy',
          status: 'running',
          starts_at: '2026-03-12T00:00:00Z',
          ends_at: '2026-03-12T12:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
    })

    const router = await createScoreboardRouter()
    const wrapper = mount(ScoreboardView, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    const pagination = wrapper.find('.scoreboard-pagination')
    expect(pagination.exists()).toBe(true)
    expect(pagination.text()).toContain('共 1 个竞赛')
    expect(pagination.text()).toContain('上一页')
    expect(pagination.text()).toContain('1 / 1')
    expect(pagination.text()).toContain('下一页')

    const paginationButtons = wrapper.findAll('.page-pagination-controls__button')
    expect(paginationButtons[0].attributes('disabled')).toBeDefined()
    expect(paginationButtons[1].attributes('disabled')).toBeDefined()
  })

  it('排行榜页级 shell 不应继续携带 journal-eyebrow-text 修饰类', () => {
    expect(scoreboardSource).toContain(
      'class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"'
    )
    expect(scoreboardSource).not.toContain('journal-eyebrow-text')
  })

  it('排行榜空态操作按钮应接入共享 ui-btn 原语', () => {
    expect(scoreboardSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(scoreboardSource).not.toContain('class="scoreboard-btn"')
  })

  it('提供竞赛排行榜与积分排行榜两个页签，并展示积分榜字段', async () => {
    getPracticeRankingMock.mockResolvedValue([
      {
        rank: 1,
        user_id: 'student-1',
        username: 'student_user',
        total_score: 320,
        solved_count: 4,
        class_name: 'Class A',
      },
    ])
    getContestsMock.mockResolvedValue({
      list: [
        {
          id: 'contest-running',
          title: '当前竞赛',
          mode: 'jeopardy',
          status: 'running',
          starts_at: '2026-03-12T00:00:00Z',
          ends_at: '2026-03-12T12:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
    })
    getScoreboardMock.mockResolvedValue({
      contest: {
        id: 'contest-running',
        title: '当前竞赛',
        status: 'running',
        started_at: '2026-03-12T00:00:00Z',
        ends_at: '2026-03-12T12:00:00Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: 'team-running',
            team_name: 'Current Champions',
            score: 2450,
            solved_count: 8,
            last_submission_at: '2026-03-12T10:15:00Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 20,
      },
      frozen: false,
    })

    const router = await createScoreboardRouter('/scoreboard?panel=points')

    const wrapper = mount(ScoreboardView, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('竞赛排行榜')
    expect(wrapper.text()).toContain('积分排行榜')
    expect(wrapper.find('#scoreboard-panel-points').classes()).toContain('active')
    expect(wrapper.find('#scoreboard-panel-contest').classes()).not.toContain('active')
    expect(wrapper.text()).toContain('student_user')
    expect(wrapper.text()).toContain('320')
    expect(wrapper.text()).toContain('4')
    expect(wrapper.text()).toContain('Class A')
    expect(scoreboardSource).toContain('scoreboard-tab-points')
  })
})
