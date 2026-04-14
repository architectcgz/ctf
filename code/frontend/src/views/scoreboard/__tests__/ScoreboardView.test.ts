import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

import ScoreboardView from '../ScoreboardView.vue'
import scoreboardSource from '../ScoreboardView.vue?raw'

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

    getScoreboardMock.mockImplementation(async (contestId: string) => ({
      contest: {
        id: contestId,
        title: `${contestId}-title`,
        status:
          contestId === 'contest-old'
            ? 'ended'
            : contestId === 'contest-frozen'
              ? 'frozen'
              : 'running',
        started_at: '2026-03-12T00:00:00Z',
        ends_at: '2026-03-12T12:00:00Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: `${contestId}-team`,
            team_name: `${contestId}-team-name`,
            score: 1000,
            solved_count: 5,
            last_submission_at: '2026-03-12T10:15:00Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 20,
      },
      frozen: contestId === 'contest-frozen',
    }))

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/scoreboard', name: 'Scoreboard', component: ScoreboardView }],
    })
    await router.push('/scoreboard')
    await router.isReady()

    const wrapper = mount(ScoreboardView, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    const cards = wrapper.findAll('[data-testid="scoreboard-card"]')

    expect(wrapper.text()).toContain('Scoreboard')
    expect(cards).toHaveLength(3)
    expect(cards[0].text()).toContain('当前竞赛')
    expect(cards[1].text()).toContain('冻结竞赛')
    expect(cards[2].text()).toContain('往期竞赛')
    expect(wrapper.text()).toContain('进行中竞赛支持实时刷新，提交后榜单会自动更新。')
    expect(wrapper.text()).toContain('封榜阶段仅展示冻结前排名，解封后会同步最终成绩。')
    expect(wrapper.text()).toContain('历史竞赛展示最终成绩，可用于复盘队伍解题表现。')
    expect(wrapper.text()).toContain('当前可查看排行的竞赛总数')
    expect(wrapper.text()).toContain('排行榜加载异常的竞赛分区')
    expect(wrapper.text()).not.toContain('报名中竞赛')
    expect(getScoreboardMock).toHaveBeenCalledTimes(3)
  })

  it('同时展示当前排行和历史排行内容', async () => {
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

    getScoreboardMock.mockImplementation(async (contestId: string) => {
      if (contestId === 'contest-history') {
        return {
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
        }
      }

      return {
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
      }
    })

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/scoreboard', name: 'Scoreboard', component: ScoreboardView }],
    })
    await router.push('/scoreboard')
    await router.isReady()

    const wrapper = mount(ScoreboardView, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('Current Champions')
    expect(wrapper.text()).toContain('History Masters')
    expect(wrapper.text()).toContain('按竞赛开始时间倒序展示排行榜')
  })

  it('收到进行中竞赛的 scoreboard.updated 后只刷新对应竞赛', async () => {
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

    getScoreboardMock.mockImplementation(async (contestId: string) => {
      if (contestId === 'contest-history') {
        return {
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
        }
      }

      const refreshCount = getScoreboardMock.mock.calls.filter(
        ([id]) => id === 'contest-running'
      ).length
      return {
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

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/scoreboard', name: 'Scoreboard', component: ScoreboardView }],
    })
    await router.push('/scoreboard')
    await router.isReady()

    const wrapper = mount(ScoreboardView, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    expect(getScoreboardMock).toHaveBeenCalledTimes(2)
    expect(webSocketMocks.getHandlers('contests/contest-history/scoreboard')).toBeUndefined()

    webSocketMocks.getHandlers('contests/contest-running/scoreboard')?.['scoreboard.updated']?.({
      contest_id: 'contest-running',
    })

    await flushPromises()
    await flushPromises()

    expect(getScoreboardMock).toHaveBeenCalledTimes(3)
    expect(getScoreboardMock.mock.calls.at(-1)?.[0]).toBe('contest-running')
    expect(wrapper.text()).toContain('Updated Champions')
  })

  it('排行榜页概况卡片应使用统一 metric-panel 样式类', () => {
    expect(scoreboardSource).toContain('class="scoreboard-summary-grid metric-panel-grid"')
    expect(scoreboardSource).toContain('class="scoreboard-summary-item metric-panel-card"')
    expect(scoreboardSource).toContain('class="scoreboard-summary-label metric-panel-label"')
    expect(scoreboardSource).toContain('class="scoreboard-summary-value metric-panel-value"')
    expect(scoreboardSource).toContain('class="scoreboard-summary-helper metric-panel-helper"')
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

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/scoreboard', name: 'Scoreboard', component: ScoreboardView }],
    })
    await router.push('/scoreboard?tab=points')
    await router.isReady()

    const wrapper = mount(ScoreboardView, {
      global: {
        plugins: [router],
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('竞赛排行榜')
    expect(wrapper.text()).toContain('积分排行榜')
    expect(wrapper.text()).toContain('student_user')
    expect(wrapper.text()).toContain('320')
    expect(wrapper.text()).toContain('4')
    expect(wrapper.text()).toContain('Class A')
    expect(scoreboardSource).toContain('scoreboard-tab-points')
  })
})
