import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  createAdminContestAnnouncement,
  createAdminContestChallenge,
  createContest,
  createContestAWDService,
  createContestAWDRound,
  deleteAdminContestAnnouncement,
  deleteAdminContestChallenge,
  deleteContestAWDService,
  getAdminContestAnnouncements,
  getAdminContestLiveScoreboard,
  getContestAWDReadiness,
  getContestAWDRoundSummary,
  getContestAWDRoundTrafficSummary,
  getContests,
  listAdminContestChallenges,
  listContestAWDServices,
  listContestAWDRoundAttacks,
  listContestAWDRoundTrafficEvents,
  prewarmContestAWDInstances,
  runContestAWDCheckerPreview,
  runContestAWDCurrentRoundCheck,
  runContestAWDRoundCheck,
  startContestAWDTeamServiceInstance,
  updateAdminContestChallenge,
  updateContest,
  updateContestAWDService,
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

describe('admin contest api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('应该请求指定 AWD 轮次巡检接口并归一化返回值', async () => {
    requestMock.mockResolvedValue({
      round: {
        id: 41,
        contest_id: 7,
        round_number: 3,
        status: 'finished',
        started_at: '2026-03-12T10:00:00.000Z',
        ended_at: '2026-03-12T10:05:00.000Z',
        attack_score: 80,
        defense_score: 45,
        created_at: '2026-03-12T10:00:00.000Z',
        updated_at: '2026-03-12T10:06:00.000Z',
      },
      services: [
        {
          id: 91,
          round_id: 41,
          team_id: 12,
          team_name: 'Blue',
          service_id: 7009,
          service_name: 'Bank Portal',
          awd_challenge_id: 101,
          awd_challenge_title: 'Bank Portal',
          service_status: 'up',
          check_result: { status_reason: 'healthy' },
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 1,
          defense_score: 45,
          attack_score: 0,
          updated_at: '2026-03-12T10:06:00.000Z',
        },
      ],
    })

    const result = await runContestAWDRoundCheck('7', '41')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/7/awd/rounds/41/check',
    })
    expect(result).toEqual({
      round: {
        id: '41',
        contest_id: '7',
        round_number: 3,
        status: 'finished',
        started_at: '2026-03-12T10:00:00.000Z',
        ended_at: '2026-03-12T10:05:00.000Z',
        attack_score: 80,
        defense_score: 45,
        created_at: '2026-03-12T10:00:00.000Z',
        updated_at: '2026-03-12T10:06:00.000Z',
      },
      services: [
        {
          id: '91',
          round_id: '41',
          team_id: '12',
          team_name: 'Blue',
          service_id: '7009',
          service_name: 'Bank Portal',
          awd_challenge_id: '101',
          awd_challenge_title: 'Bank Portal',
          service_status: 'up',
          check_result: { status_reason: 'healthy' },
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 1,
          defense_score: 45,
          attack_score: 0,
          updated_at: '2026-03-12T10:06:00.000Z',
        },
      ],
    })
  })

  it('应该请求 AWD checker 试跑接口并归一化返回值', async () => {
    requestMock.mockResolvedValue({
      checker_type: 'http_standard',
      service_status: 'up',
      check_result: {
        checker_type: 'http_standard',
        check_source: 'checker_preview',
        status_reason: 'healthy',
      },
      preview_context: {
        access_url: 'http://preview.internal',
        preview_flag: 'flag{preview}',
        round_number: 0,
        team_id: 0,
        awd_challenge_id: 101,
      },
      preview_token: 'preview-token-9',
    })

    const result = await runContestAWDCheckerPreview('7', {
      awd_challenge_id: 101,
      checker_type: 'http_standard',
      checker_config: {
        put_flag: { method: 'PUT', path: '/api/flag' },
        get_flag: { method: 'GET', path: '/api/flag' },
      },
      access_url: 'http://preview.internal',
      preview_flag: 'flag{preview}',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/7/awd/checker-preview',
      timeout: 30000,
      data: {
        awd_challenge_id: 101,
        checker_type: 'http_standard',
        checker_config: {
          put_flag: { method: 'PUT', path: '/api/flag' },
          get_flag: { method: 'GET', path: '/api/flag' },
        },
        access_url: 'http://preview.internal',
        preview_flag: 'flag{preview}',
      },
    })
    expect(result).toEqual({
      checker_type: 'http_standard',
      service_status: 'up',
      check_result: {
        checker_type: 'http_standard',
        check_source: 'checker_preview',
        status_reason: 'healthy',
      },
      preview_context: {
        access_url: 'http://preview.internal',
        preview_flag: 'flag{preview}',
        round_number: 0,
        team_id: '0',
        awd_challenge_id: '101',
      },
      preview_token: 'preview-token-9',
    })
  })
})
