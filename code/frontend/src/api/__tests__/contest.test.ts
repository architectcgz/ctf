import { beforeEach, describe, expect, it, vi } from 'vitest'

const requestMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/request', () => ({
  request: requestMock,
}))

import {
  getContestAWDWorkspace,
  startContestChallengeInstance,
  submitContestAWDAttack,
} from '@/api/contest'

describe('contest api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('获取学生 AWD 工作台时应透传 contest id 并标准化字段', async () => {
    requestMock.mockResolvedValue({
      contest_id: 7,
      current_round: {
        id: 41,
        contest_id: 7,
        round_number: 3,
        status: 'running',
        attack_score: 60,
        defense_score: 40,
        created_at: '2026-04-12T08:00:00Z',
        updated_at: '2026-04-12T08:01:00Z',
      },
      my_team: {
        team_id: 13,
        team_name: 'Red',
      },
      services: [
        {
          challenge_id: 9,
          access_url: 'http://red.internal',
          service_status: 'up',
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 40,
          attack_score: 0,
          updated_at: '2026-04-12T08:02:00Z',
        },
      ],
      targets: [
        {
          team_id: 14,
          team_name: 'Blue',
          services: [
            {
              challenge_id: 9,
              access_url: 'http://blue.internal',
            },
          ],
        },
      ],
      recent_events: [
        {
          id: 88,
          direction: 'attack_out',
          challenge_id: 9,
          peer_team_id: 14,
          peer_team_name: 'Blue',
          is_success: true,
          score_gained: 60,
          created_at: '2026-04-12T08:03:00Z',
        },
      ],
    })

    const result = await getContestAWDWorkspace('7')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/contests/7/awd/workspace',
    })
    expect(result.contest_id).toBe('7')
    expect(result.current_round?.id).toBe('41')
    expect(result.my_team?.team_id).toBe('13')
    expect(result.services[0].challenge_id).toBe('9')
    expect(result.targets[0].services[0].challenge_id).toBe('9')
    expect(result.recent_events[0].id).toBe('88')
  })

  it('启动竞赛题目实例时应命中 contest 实例接口并复用实例标准化', async () => {
    requestMock.mockResolvedValue({
      id: 21,
      challenge_id: 9,
      status: 'running',
      share_scope: 'per_team',
      access_url: 'http://red.internal',
      flag_type: 'dynamic',
      expires_at: '2026-04-12T09:00:00Z',
      created_at: '2026-04-12T08:00:00Z',
      max_extends: 2,
      extend_count: 1,
    })

    const result = await startContestChallengeInstance('7', '9')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/contests/7/challenges/9/instances',
      suppressErrorToast: true,
    })
    expect(result).toEqual({
      id: '21',
      challenge_id: '9',
      status: 'running',
      share_scope: 'per_team',
      access_url: 'http://red.internal',
      flag_type: 'dynamic',
      expires_at: '2026-04-12T09:00:00Z',
      created_at: '2026-04-12T08:00:00Z',
      remaining_extends: 1,
    })
  })

  it('提交 stolen flag 时应调用 AWD 提交接口并标准化日志字段', async () => {
    requestMock.mockResolvedValue({
      id: 55,
      round_id: 41,
      attacker_team_id: 13,
      attacker_team: 'Red',
      victim_team_id: 14,
      victim_team: 'Blue',
      challenge_id: 9,
      attack_type: 'flag_capture',
      source: 'submission',
      submitted_flag: 'flag{demo}',
      is_success: true,
      score_gained: 60,
      created_at: '2026-04-12T08:03:00Z',
    })

    const result = await submitContestAWDAttack('7', '9', {
      victim_team_id: 14,
      flag: 'flag{demo}',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/contests/7/awd/challenges/9/submissions',
      data: {
        victim_team_id: 14,
        flag: 'flag{demo}',
      },
    })
    expect(result).toEqual({
      id: '55',
      round_id: '41',
      attacker_team_id: '13',
      attacker_team: 'Red',
      victim_team_id: '14',
      victim_team: 'Blue',
      challenge_id: '9',
      attack_type: 'flag_capture',
      source: 'submission',
      submitted_flag: 'flag{demo}',
      is_success: true,
      score_gained: 60,
      created_at: '2026-04-12T08:03:00Z',
    })
  })
})
