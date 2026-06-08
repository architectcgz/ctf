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

  it('应该请求 AWD readiness 并归一化阻塞摘要', async () => {
    requestMock.mockResolvedValue({
      contest_id: 9,
      ready: false,
      total_challenges: 3,
      passed_challenges: 1,
      pending_challenges: 1,
      failed_challenges: 1,
      stale_challenges: 0,
      missing_checker_challenges: 1,
      blocking_count: 2,
      global_blocking_reasons: ['no_challenges'],
      blocking_actions: ['create_round', 'run_current_round_check'],
      items: [
        {
          awd_challenge_id: 101,
          title: 'web-checker',
          checker_type: 'http_standard',
          validation_state: 'failed',
          last_preview_at: '2026-04-12T08:00:00.000Z',
          last_access_url: 'http://checker.internal/flag',
          blocking_reason: 'last_preview_failed',
        },
      ],
    })

    const result = await getContestAWDReadiness('9')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/9/awd/readiness',
    })
    expect(result).toEqual({
      contest_id: '9',
      ready: false,
      total_challenges: 3,
      passed_challenges: 1,
      pending_challenges: 1,
      failed_challenges: 1,
      stale_challenges: 0,
      missing_checker_challenges: 1,
      blocking_count: 2,
      global_blocking_reasons: ['no_challenges'],
      blocking_actions: ['create_round', 'run_current_round_check'],
      items: [
        {
          awd_challenge_id: '101',
          title: 'web-checker',
          checker_type: 'http_standard',
          validation_state: 'failed',
          last_preview_at: '2026-04-12T08:00:00.000Z',
          last_access_url: 'http://checker.internal/flag',
          blocking_reason: 'last_preview_failed',
        },
      ],
    })
  })

  it('应该在创建轮次时透传 readiness override 字段', async () => {
    requestMock.mockResolvedValue({
      id: 21,
      contest_id: 9,
      round_number: 4,
      status: 'pending',
      attack_score: 60,
      defense_score: 40,
      created_at: '2026-04-12T09:00:00.000Z',
      updated_at: '2026-04-12T09:00:00.000Z',
    })

    await createContestAWDRound('awd-1', {
      round_number: 4,
      status: 'pending',
      attack_score: 60,
      defense_score: 40,
      force_override: true,
      override_reason: 'teacher drill',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/awd-1/awd/rounds',
      data: {
        round_number: 4,
        status: 'pending',
        attack_score: 60,
        defense_score: 40,
        force_override: true,
        override_reason: 'teacher drill',
      },
    })
  })

  it('应该允许当前轮巡检请求携带可选 override body', async () => {
    requestMock.mockResolvedValue({
      round: {
        id: 21,
        contest_id: 9,
        round_number: 4,
        status: 'running',
        attack_score: 60,
        defense_score: 40,
        created_at: '2026-04-12T09:00:00.000Z',
        updated_at: '2026-04-12T09:00:00.000Z',
      },
      services: [],
    })

    await runContestAWDCurrentRoundCheck('awd-1', {
      force_override: true,
      override_reason: 'teacher drill',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/awd-1/awd/current-round/check',
      data: {
        force_override: true,
        override_reason: 'teacher drill',
      },
    })
  })
})
