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

  it('应在启动队伍服务实例时把错误展示权交给调用方', async () => {
    requestMock.mockResolvedValue({
      team_id: 12,
      service_id: 99,
      instance: null,
    })

    await startContestAWDTeamServiceInstance('7', {
      team_id: '12',
      service_id: '99',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/7/awd/instances',
      data: {
        team_id: 12,
        service_id: 99,
      },
    })
  })

  it('应在赛前预热 AWD 实例时提交可选 team_id 并保留结果矩阵', async () => {
    requestMock.mockResolvedValue({
      contest_id: 7,
      results: [
        {
          team_id: 12,
          service_id: 99,
          outcome: 'reused',
          instance: {
            id: 88,
            contest_mode: 'awd',
            challenge_id: 3,
            status: 'running',
            share_scope: 'per_team',
            access_url: null,
            flag_type: 'dynamic',
            expires_at: '2026-04-12T18:00:00.000Z',
            remaining_extends: 0,
            created_at: '2026-04-12T09:00:00.000Z',
          },
          error_message: '',
        },
      ],
      summary: {
        total: 1,
        started: 0,
        reused: 1,
        failed: 0,
      },
    })

    await prewarmContestAWDInstances('7', {
      team_id: '12',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/7/awd/instances/prewarm',
      data: {
        team_id: 12,
      },
    })
  })
})
