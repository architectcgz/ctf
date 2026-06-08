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

  it('应该归一化管理员竞赛题目列表中的关系字段', async () => {
    requestMock.mockResolvedValue([
      {
        id: 31,
        contest_id: 7,
        challenge_id: 11,
        title: 'SQL Injection 101',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 2,
        is_visible: true,
        created_at: '2026-03-12T00:00:00.000Z',
      },
    ])

    const result = await listAdminContestChallenges('7')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/challenges',
    })
    expect(result).toEqual([
      {
        id: '31',
        contest_id: '7',
        challenge_id: '11',
        title: 'SQL Injection 101',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 2,
        is_visible: true,
        created_at: '2026-03-12T00:00:00.000Z',
      },
    ])
  })

  it('应该按后端契约创建竞赛题目，并忽略误传的 AWD 兼容字段', async () => {
    requestMock.mockResolvedValue({
      id: 31,
      contest_id: 7,
      challenge_id: 11,
      title: 'SQL Injection 101',
      category: 'web',
      difficulty: 'easy',
      points: 120,
      order: 2,
      is_visible: true,
      created_at: '2026-03-12T00:00:00.000Z',
    })

    const result = await createAdminContestChallenge('7', {
      challenge_id: 11,
      points: 120,
      order: 2,
      is_visible: true,
      ...({
        template_id: 5,
        awd_checker_type: 'http_standard',
        awd_checker_config: {
          put_flag: { method: 'PUT', path: '/api/flag' },
        },
        awd_sla_score: 1,
        awd_defense_score: 2,
        awd_checker_preview_token: 'preview-token-1',
      } as Record<string, unknown>),
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/7/challenges',
      data: {
        challenge_id: 11,
        points: 120,
        order: 2,
        is_visible: true,
      },
    })
    expect(result).toEqual({
      id: '31',
      contest_id: '7',
      challenge_id: '11',
      title: 'SQL Injection 101',
      category: 'web',
      difficulty: 'easy',
      points: 120,
      order: 2,
      is_visible: true,
      created_at: '2026-03-12T00:00:00.000Z',
    })
  })

  it('应该按后端契约更新竞赛题目，并忽略误传的 AWD 兼容字段', async () => {
    requestMock.mockResolvedValue(null)

    await updateAdminContestChallenge('7', '11', {
      points: 150,
      order: 3,
      is_visible: false,
      ...({
        template_id: 7,
        awd_checker_type: 'legacy_probe',
        awd_checker_config: {
          health_path: '/healthz',
        },
        awd_sla_score: 1,
        awd_defense_score: 2,
        awd_checker_preview_token: 'preview-token-2',
      } as Record<string, unknown>),
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'PUT',
      url: '/admin/contests/7/challenges/11',
      data: {
        points: 150,
        order: 3,
        is_visible: false,
      },
    })
  })

  it('应该按后端契约移除竞赛题目', async () => {
    requestMock.mockResolvedValue(null)

    await deleteAdminContestChallenge('7', '11')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'DELETE',
      url: '/admin/contests/7/challenges/11',
    })
  })
})
