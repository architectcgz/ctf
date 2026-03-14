import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

import ScoreboardView from '../ScoreboardView.vue'

const { getContestsMock, getScoreboardMock } = vi.hoisted(() => ({
  getContestsMock: vi.fn(),
  getScoreboardMock: vi.fn(),
}))

vi.mock('@/api/contest', () => ({
  getContests: getContestsMock,
  getScoreboard: getScoreboardMock,
}))

describe('ScoreboardView', () => {
  beforeEach(() => {
    getContestsMock.mockReset()
    getScoreboardMock.mockReset()
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
        status: contestId === 'contest-old' ? 'ended' : contestId === 'contest-frozen' ? 'frozen' : 'running',
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

    expect(cards).toHaveLength(3)
    expect(cards[0].text()).toContain('当前竞赛')
    expect(cards[1].text()).toContain('冻结竞赛')
    expect(cards[2].text()).toContain('往期竞赛')
    expect(wrapper.text()).not.toContain('报名中竞赛')
    expect(getScoreboardMock).toHaveBeenCalledTimes(3)
  })

  it('同时展示当前排行和历史排行内容', async () => {
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
})
