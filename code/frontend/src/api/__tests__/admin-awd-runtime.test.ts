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

  it('应该归一化 AWD 轮次汇总中的运维指标', async () => {
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
      metrics: {
        total_service_count: 6,
        service_up_count: 4,
        service_down_count: 1,
        service_compromised_count: 1,
        attacked_service_count: 2,
        defense_success_count: 1,
        total_attack_count: 5,
        successful_attack_count: 3,
        failed_attack_count: 2,
        scheduler_check_count: 4,
        manual_current_round_check_count: 0,
        manual_selected_round_check_count: 1,
        manual_service_check_count: 1,
        submission_attack_count: 3,
        manual_attack_log_count: 2,
        legacy_attack_log_count: 0,
      },
      items: [
        {
          team_id: 12,
          team_name: 'Blue',
          service_up_count: 1,
          service_down_count: 0,
          service_compromised_count: 0,
          sla_score: 1,
          defense_score: 45,
          attack_score: 0,
          successful_attack_count: 0,
          successful_breach_count: 0,
          unique_attackers_against: 0,
          total_score: 63,
        },
      ],
    })

    const result = await getContestAWDRoundSummary('7', '41')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/awd/rounds/41/summary',
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
      metrics: {
        total_service_count: 6,
        service_up_count: 4,
        service_down_count: 1,
        service_compromised_count: 1,
        attacked_service_count: 2,
        defense_success_count: 1,
        total_attack_count: 5,
        successful_attack_count: 3,
        failed_attack_count: 2,
        scheduler_check_count: 4,
        manual_current_round_check_count: 0,
        manual_selected_round_check_count: 1,
        manual_service_check_count: 1,
        submission_attack_count: 3,
        manual_attack_log_count: 2,
        legacy_attack_log_count: 0,
      },
      items: [
        {
          team_id: '12',
          team_name: 'Blue',
          service_up_count: 1,
          service_down_count: 0,
          service_compromised_count: 0,
          sla_score: 1,
          defense_score: 45,
          attack_score: 0,
          successful_attack_count: 0,
          successful_breach_count: 0,
          unique_attackers_against: 0,
          total_score: 63,
        },
      ],
    })
  })

  it('应该把缺失的 AWD 轮次汇总计数字段回落为 0', async () => {
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
      metrics: {
        total_service_count: 6,
        service_up_count: 4,
        service_down_count: 1,
        service_compromised_count: 1,
        attacked_service_count: 2,
        defense_success_count: 1,
        total_attack_count: 5,
        successful_attack_count: 3,
        failed_attack_count: 2,
        manual_service_check_count: 1,
        submission_attack_count: 3,
      },
      items: [],
    })

    const result = await getContestAWDRoundSummary('7', '41')

    expect(result.metrics).toEqual({
      total_service_count: 6,
      service_up_count: 4,
      service_down_count: 1,
      service_compromised_count: 1,
      attacked_service_count: 2,
      defense_success_count: 1,
      total_attack_count: 5,
      successful_attack_count: 3,
      failed_attack_count: 2,
      scheduler_check_count: 0,
      manual_current_round_check_count: 0,
      manual_selected_round_check_count: 0,
      manual_service_check_count: 1,
      submission_attack_count: 3,
      manual_attack_log_count: 0,
      legacy_attack_log_count: 0,
    })
  })
})
