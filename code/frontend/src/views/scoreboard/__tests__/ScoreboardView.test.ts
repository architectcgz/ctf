import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

import ScoreboardDetail from '../ScoreboardDetail.vue'
import ScoreboardView from '../ScoreboardView.vue'
import scoreboardSource from '../ScoreboardView.vue?raw'

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

describe('ScoreboardView', () => {
  beforeEach(() => {
    getContestsMock.mockReset()
    getScoreboardMock.mockReset()
    getPracticeRankingMock.mockReset()
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
    expect(wrapper.text()).toContain('查看完整排行榜')
    expect(wrapper.text()).not.toContain('Current Champions')
    expect(wrapper.text()).not.toContain('History Masters')
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
