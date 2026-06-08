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

  it('应该归一化显式 AWD service 列表', async () => {
    requestMock.mockResolvedValue([
      {
        id: 7009,
        contest_id: 7,
        awd_challenge_id: 5,
        display_name: 'Bank Portal',
        order: 2,
        is_visible: true,
        score_config: {
          attack_score: 60,
          awd_sla_score: 1,
          awd_defense_score: 2,
        },
        runtime_config: {
          awd_challenge_id: 5,
          checker_type: 'http_standard',
          checker_config: {
            get_flag: {
              path: '/health',
            },
          },
        },
        validation_state: 'passed',
        last_preview_at: '2026-03-12T00:04:00.000Z',
        last_preview_result: {
          service_status: 'up',
          check_result: {
            status_code: 200,
          },
          preview_context: {
            access_url: 'http://checker.internal/health',
            preview_flag: 'FLAG{preview}',
            round_number: 0,
            team_id: 0,
            awd_challenge_id: 5,
          },
        },
        created_at: '2026-03-12T00:00:00.000Z',
        updated_at: '2026-03-12T00:05:00.000Z',
      },
    ])

    const result = await listContestAWDServices('7')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/awd/services',
    })
    expect(result).toEqual([
      {
        id: '7009',
        contest_id: '7',
        awd_challenge_id: '5',
        title: undefined,
        category: undefined,
        difficulty: undefined,
        display_name: 'Bank Portal',
        order: 2,
        is_visible: true,
        score_config: {
          attack_score: 60,
          awd_sla_score: 1,
          awd_defense_score: 2,
        },
        runtime_config: {
          checker_type: 'http_standard',
          checker_config: {
            get_flag: {
              path: '/health',
            },
          },
        },
        checker_type: 'http_standard',
        checker_config: {
          get_flag: {
            path: '/health',
          },
        },
        sla_score: 1,
        defense_score: 2,
        validation_state: 'passed',
        last_preview_at: '2026-03-12T00:04:00.000Z',
        last_preview_result: {
          checker_type: undefined,
          preview_token: undefined,
          service_status: 'up',
          check_result: {
            status_code: 200,
          },
          preview_context: {
            access_url: 'http://checker.internal/health',
            preview_flag: 'FLAG{preview}',
            round_number: 0,
            team_id: '0',
            awd_challenge_id: '5',
          },
        },
        created_at: '2026-03-12T00:00:00.000Z',
        updated_at: '2026-03-12T00:05:00.000Z',
      },
    ])
  })

  it('应该按后端契约创建显式 AWD service', async () => {
    requestMock.mockResolvedValue({
      id: 7009,
      contest_id: 7,
      awd_challenge_id: 5,
      title: 'Bank Portal',
      category: 'web',
      difficulty: 'medium',
      display_name: 'Bank Portal',
      order: 2,
      is_visible: true,
      score_config: {
        attack_score: 60,
        awd_sla_score: 1,
        awd_defense_score: 2,
      },
      runtime_config: {
        checker_type: 'http_standard',
        checker_config: {
          put_flag: {
            method: 'PUT',
            path: '/api/flag',
          },
        },
      },
      validation_state: 'pending',
      last_preview_at: undefined,
      last_preview_result: undefined,
      created_at: '2026-03-12T00:00:00.000Z',
      updated_at: '2026-03-12T00:05:00.000Z',
    })

    const result = await createContestAWDService('7', {
      awd_challenge_id: 5,
      points: 180,
      display_name: 'Bank Portal',
      order: 2,
      is_visible: true,
      checker_type: 'http_standard',
      checker_config: {
        put_flag: { method: 'PUT', path: '/api/flag' },
      },
      awd_sla_score: 1,
      awd_defense_score: 2,
      awd_checker_preview_token: 'preview-token',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/7/awd/services',
      data: {
        awd_challenge_id: 5,
        points: 180,
        display_name: 'Bank Portal',
        order: 2,
        is_visible: true,
        checker_type: 'http_standard',
        checker_config: {
          put_flag: { method: 'PUT', path: '/api/flag' },
        },
        awd_sla_score: 1,
        awd_defense_score: 2,
        awd_checker_preview_token: 'preview-token',
      },
    })
    expect(result).toEqual({
      id: '7009',
      contest_id: '7',
      awd_challenge_id: '5',
      title: 'Bank Portal',
      category: 'web',
      difficulty: 'medium',
      display_name: 'Bank Portal',
      order: 2,
      is_visible: true,
      score_config: {
        attack_score: 60,
        awd_sla_score: 1,
        awd_defense_score: 2,
      },
      runtime_config: {
        checker_type: 'http_standard',
        checker_config: {
          put_flag: {
            method: 'PUT',
            path: '/api/flag',
          },
        },
      },
      checker_type: 'http_standard',
      checker_config: {
        put_flag: {
          method: 'PUT',
          path: '/api/flag',
        },
      },
      sla_score: 1,
      defense_score: 2,
      validation_state: 'pending',
      last_preview_at: undefined,
      last_preview_result: undefined,
      created_at: '2026-03-12T00:00:00.000Z',
      updated_at: '2026-03-12T00:05:00.000Z',
    })
  })

  it('应该按后端契约更新显式 AWD service', async () => {
    requestMock.mockResolvedValue(null)

    await updateContestAWDService('7', '7009', {
      awd_challenge_id: 6,
      points: 200,
      display_name: 'Bank Portal v2',
      order: 3,
      is_visible: false,
      checker_type: 'legacy_probe',
      checker_config: {
        health: { path: '/healthz' },
      },
      awd_sla_score: 1,
      awd_defense_score: 2,
      awd_checker_preview_token: 'preview-token-2',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'PUT',
      url: '/admin/contests/7/awd/services/7009',
      data: {
        awd_challenge_id: 6,
        points: 200,
        display_name: 'Bank Portal v2',
        order: 3,
        is_visible: false,
        checker_type: 'legacy_probe',
        checker_config: {
          health: { path: '/healthz' },
        },
        awd_sla_score: 1,
        awd_defense_score: 2,
        awd_checker_preview_token: 'preview-token-2',
      },
    })
  })

  it('应该按后端契约删除显式 AWD service', async () => {
    requestMock.mockResolvedValue(null)

    await deleteContestAWDService('7', '7009')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'DELETE',
      url: '/admin/contests/7/awd/services/7009',
    })
  })
})
