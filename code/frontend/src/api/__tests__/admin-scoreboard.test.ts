import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  createContestAWDRound,
  getAdminContestLiveScoreboard,
  getContestAWDReadiness,
  getContestAWDRoundSummary,
  getContestAWDRoundTrafficSummary,
  listContestAWDRoundAttacks,
  listContestAWDRoundTrafficEvents,
  prewarmContestAWDInstances,
  runContestAWDCurrentRoundCheck,
  startContestAWDTeamServiceInstance,
  updateContest,
} from '@/api/admin/contests'

const requestMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/request', () => ({
  request: requestMock,
  ApiError: class ApiError extends Error {
    status?: number

    constructor(message: string, opts?: { status?: number }) {
      super(message)
      this.name = 'ApiError'
      this.status = opts?.status
    }
  },
}))

describe('admin contest AWD runtime api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('应该请求管理员实时排行榜接口并归一化返回值', async () => {
    requestMock.mockResolvedValue({
      contest: {
        id: 7,
        title: '春季赛',
        status: 'frozen',
        started_at: '2026-03-12T09:00:00.000Z',
        ends_at: '2026-03-12T12:00:00.000Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: 11,
            team_name: 'Blue',
            score: 350,
            solved_count: 4,
            last_submission_at: '2026-03-12T11:40:00.000Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })

    const result = await getAdminContestLiveScoreboard('7', { page: 1, page_size: 10 })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/scoreboard/live',
      params: { page: 1, page_size: 10 },
    })
    expect(result).toEqual({
      contest: {
        id: '7',
        title: '春季赛',
        status: 'frozen',
        started_at: '2026-03-12T09:00:00.000Z',
        ends_at: '2026-03-12T12:00:00.000Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: '11',
            team_name: 'Blue',
            score: 350,
            solved_count: 4,
            last_submission_at: '2026-03-12T11:40:00.000Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })
  })
})
