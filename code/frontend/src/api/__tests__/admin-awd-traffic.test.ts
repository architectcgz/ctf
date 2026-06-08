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

  it('应该归一化 AWD 攻击流量摘要数据', async () => {
    requestMock.mockResolvedValue({
      round: {
        id: 41,
        contest_id: 7,
        round_number: 3,
        status: 'running',
        started_at: '2026-03-12T10:00:00.000Z',
        ended_at: null,
        attack_score: 50,
        defense_score: 40,
        created_at: '2026-03-12T09:59:00.000Z',
        updated_at: '2026-03-12T10:01:11.000Z',
      },
      contest_id: 7,
      round_id: 41,
      total_request_count: 18,
      active_attacker_team_count: 3,
      victim_team_count: 2,
      error_request_count: 4,
      unique_path_count: 7,
      latest_event_at: '2026-03-12T10:01:11.000Z',
      trend_buckets: [
        {
          bucket_start_at: '2026-03-12T10:00:00.000Z',
          bucket_end_at: '2026-03-12T10:01:00.000Z',
          request_count: 6,
          error_count: 1,
        },
      ],
      top_victims: [
        {
          team_id: 12,
          team_name: 'Blue',
          request_count: 9,
          error_count: 2,
        },
      ],
      top_attackers: [
        {
          team_id: 11,
          team_name: 'Red',
          request_count: 10,
          error_count: 3,
        },
      ],
      top_challenges: [
        {
          awd_challenge_id: 101,
          awd_challenge_title: 'Web 1',
          request_count: 11,
          error_count: 4,
        },
      ],
      top_paths: [
        {
          path: '/api/flag',
          request_count: 8,
          error_count: 3,
          last_status_code: 500,
        },
      ],
      top_error_paths: [
        {
          path: '/api/flag',
          request_count: 8,
          error_count: 3,
          last_status_code: 500,
        },
      ],
    })

    const result = await getContestAWDRoundTrafficSummary('7', '41')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/awd/rounds/41/traffic/summary',
    })
    expect(result).toEqual({
      round: {
        id: '41',
        contest_id: '7',
        round_number: 3,
        status: 'running',
        started_at: '2026-03-12T10:00:00.000Z',
        ended_at: undefined,
        attack_score: 50,
        defense_score: 40,
        created_at: '2026-03-12T09:59:00.000Z',
        updated_at: '2026-03-12T10:01:11.000Z',
      },
      contest_id: '7',
      round_id: '41',
      total_request_count: 18,
      active_attacker_team_count: 3,
      victim_team_count: 2,
      error_request_count: 4,
      unique_path_count: 7,
      latest_event_at: '2026-03-12T10:01:11.000Z',
      trend_buckets: [
        {
          bucket_start_at: '2026-03-12T10:00:00.000Z',
          bucket_end_at: '2026-03-12T10:01:00.000Z',
          request_count: 6,
          error_count: 1,
        },
      ],
      top_victims: [
        {
          team_id: '12',
          team_name: 'Blue',
          request_count: 9,
          error_count: 2,
        },
      ],
      top_attackers: [
        {
          team_id: '11',
          team_name: 'Red',
          request_count: 10,
          error_count: 3,
        },
      ],
      top_challenges: [
        {
          awd_challenge_id: '101',
          awd_challenge_title: 'Web 1',
          request_count: 11,
          error_count: 4,
        },
      ],
      top_paths: [
        {
          path: '/api/flag',
          request_count: 8,
          error_count: 3,
          last_status_code: 500,
        },
      ],
      top_error_paths: [
        {
          path: '/api/flag',
          request_count: 8,
          error_count: 3,
          last_status_code: 500,
        },
      ],
    })
  })

  it('应该归一化 AWD 攻击流量事件分页数据', async () => {
    requestMock.mockResolvedValue({
      list: [
        {
          id: 301,
          contest_id: 7,
          round_id: 41,
          occurred_at: '2026-03-12T10:00:30.000Z',
          attacker_team_id: 11,
          attacker_team_name: 'Red',
          victim_team_id: 12,
          victim_team_name: 'Blue',
          service_id: 7009,
          awd_challenge_id: 101,
          awd_challenge_title: 'Web 1',
          method: 'GET',
          path: '/api/flag',
          status_code: 500,
          status_group: 'server_error',
          is_error: true,
          source: 'proxy_audit',
          request_id: 'req-1',
        },
      ],
      total: 23,
      page: 2,
      page_size: 20,
    })

    const result = await listContestAWDRoundTrafficEvents('7', '41', {
      page: 2,
      page_size: 20,
      attacker_team_id: '11',
      service_id: '7009',
      status_group: 'server_error',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/awd/rounds/41/traffic/events',
      params: {
        page: 2,
        page_size: 20,
        attacker_team_id: '11',
        service_id: '7009',
        status_group: 'server_error',
      },
    })
    expect(result).toEqual({
      list: [
        {
          id: '301',
          contest_id: '7',
          round_id: '41',
          occurred_at: '2026-03-12T10:00:30.000Z',
          attacker_team_id: '11',
          attacker_team_name: 'Red',
          victim_team_id: '12',
          victim_team_name: 'Blue',
          service_id: '7009',
          awd_challenge_id: '101',
          awd_challenge_title: 'Web 1',
          method: 'GET',
          path: '/api/flag',
          status_code: 500,
          status_group: 'server_error',
          is_error: true,
          source: 'proxy_audit',
          request_id: 'req-1',
        },
      ],
      total: 23,
      page: 2,
      page_size: 20,
    })
  })

  it('应该归一化 AWD 攻击日志来源字段', async () => {
    requestMock.mockResolvedValue([
      {
        id: 71,
        round_id: 41,
        attacker_team_id: 11,
        attacker_team: 'Red',
        victim_team_id: 12,
        victim_team: 'Blue',
        service_id: 7009,
        awd_challenge_id: 101,
        attack_type: 'flag_capture',
        source: 'submission',
        submitted_flag: 'flag{demo}',
        is_success: true,
        score_gained: 80,
        created_at: '2026-03-12T10:07:00.000Z',
      },
    ])

    const result = await listContestAWDRoundAttacks('7', '41')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/awd/rounds/41/attacks',
    })
    expect(result).toEqual([
      {
        id: '71',
        round_id: '41',
        attacker_team_id: '11',
        attacker_team: 'Red',
        victim_team_id: '12',
        victim_team: 'Blue',
        service_id: '7009',
        awd_challenge_id: '101',
        attack_type: 'flag_capture',
        source: 'submission',
        submitted_flag: 'flag{demo}',
        is_success: true,
        score_gained: 80,
        created_at: '2026-03-12T10:07:00.000Z',
      },
    ])
  })
})
