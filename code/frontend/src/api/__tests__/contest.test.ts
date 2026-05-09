import { beforeEach, describe, expect, it, vi } from 'vitest'

const requestMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/request', () => ({
  request: requestMock,
}))

import {
  getContestChallenges,
  getContestAWDWorkspace,
  getContests,
  requestContestAWDDefenseSSH,
  requestContestAWDTargetAccess,
  restartContestAWDServiceInstance,
  startContestAWDServiceInstance,
  submitContestAWDAttack,
} from '@/api/contest'

describe('contest api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('获取竞赛列表时应读取 registering_count 汇总字段', async () => {
    requestMock.mockResolvedValue({
      list: [
        {
          id: 7,
          title: '春季赛',
          description: '测试竞赛',
          mode: 'jeopardy',
          status: 'registration',
          start_time: '2026-03-10T09:00:00.000Z',
          end_time: '2026-03-10T12:00:00.000Z',
          freeze_time: null,
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      summary: {
        draft_count: 0,
        registering_count: 3,
        running_count: 1,
        frozen_count: 0,
        ended_count: 2,
      },
    })

    const result = await getContests({ page: 1, page_size: 20, status: 'registering' })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/contests',
      params: {
        page: 1,
        page_size: 20,
        status: 'registration',
        statuses: undefined,
        mode: undefined,
        sort_key: undefined,
        sort_order: undefined,
      },
      signal: undefined,
    })
    expect(result.summary).toEqual({
      draft_count: 0,
      registering_count: 3,
      running_count: 1,
      frozen_count: 0,
      ended_count: 2,
    })
    expect(result.list[0]?.status).toBe('registering')
  })

  it('获取竞赛题目列表时应标准化 awd service 标识', async () => {
    requestMock.mockResolvedValue([
      {
        id: 21,
        awd_challenge_id: 9,
        awd_service_id: 7009,
        title: 'Bank Portal',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 3,
        is_solved: false,
      },
    ])

    const result = await getContestChallenges('7')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/contests/7/challenges',
    })
    expect(result).toEqual([
      {
        id: '21',
        challenge_id: '9',
        awd_challenge_id: '9',
        awd_service_id: '7009',
        title: 'Bank Portal',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        solved_count: 3,
        is_solved: false,
      },
    ])
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
          service_id: 7009,
          awd_challenge_id: 9,
          instance_id: 9001,
          instance_status: 'running',
          defense_connection: {
            entry_mode: 'ssh',
            workspace_status: 'running',
            workspace_revision: 7,
          },
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
              service_id: 7009,
              awd_challenge_id: 9,
              reachable: true,
            },
          ],
        },
      ],
      recent_events: [
        {
          id: 88,
          direction: 'attack_out',
          service_id: 7009,
          awd_challenge_id: 9,
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
    expect(result.services[0].service_id).toBe('7009')
    expect(result.services[0].awd_challenge_id).toBe('9')
    expect(result.services[0].instance_id).toBe('9001')
    expect(result.services[0].instance_status).toBe('running')
    expect(result.services[0].access_url).toBeUndefined()
    expect(result.services[0].defense_connection).toEqual({
      entry_mode: 'ssh',
      workspace_status: 'running',
      workspace_revision: 7,
    })
    expect(result.targets[0].services[0].service_id).toBe('7009')
    expect(result.targets[0].services[0].awd_challenge_id).toBe('9')
    expect(result.targets[0].services[0].reachable).toBe(true)
    expect('access_url' in result.targets[0].services[0]).toBe(false)
    expect(result.recent_events[0].service_id).toBe('7009')
    expect(result.recent_events[0].id).toBe('88')
  })

  it('启动 AWD 服务实例时应命中 service 实例接口并复用实例标准化', async () => {
    requestMock.mockResolvedValue({
      id: 22,
      awd_challenge_id: 9,
      status: 'running',
      share_scope: 'per_team',
      access_url: 'http://red.internal',
      flag_type: 'dynamic',
      expires_at: '2026-04-12T09:00:00Z',
      created_at: '2026-04-12T08:00:00Z',
      max_extends: 2,
      extend_count: 0,
    })

    const result = await startContestAWDServiceInstance('7', '7009')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/contests/7/awd/services/7009/instances',
    })
    expect(result).toEqual({
      id: '22',
      challenge_id: '9',
      awd_challenge_id: '9',
      status: 'running',
      share_scope: 'per_team',
      access_url: 'http://red.internal',
      flag_type: 'dynamic',
      expires_at: '2026-04-12T09:00:00Z',
      created_at: '2026-04-12T08:00:00Z',
      remaining_extends: 2,
    })
  })

  it('重启 AWD 服务实例时应命中 service restart 接口并复用实例标准化', async () => {
    requestMock.mockResolvedValue({
      id: 22,
      awd_challenge_id: 9,
      status: 'pending',
      share_scope: 'per_team',
      flag_type: 'dynamic',
      expires_at: '2026-04-12T09:00:00Z',
      created_at: '2026-04-12T08:00:00Z',
      max_extends: 2,
      extend_count: 0,
    })

    const result = await restartContestAWDServiceInstance('7', '7009')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/contests/7/awd/services/7009/instances/restart',
    })
    expect(result.status).toBe('pending')
    expect(result.id).toBe('22')
  })

  it('提交 stolen flag 时应调用 AWD 提交接口并标准化日志字段', async () => {
    requestMock.mockResolvedValue({
      id: 55,
      round_id: 41,
      attacker_team_id: 13,
      attacker_team: 'Red',
      victim_team_id: 14,
      victim_team: 'Blue',
      service_id: 7009,
      awd_challenge_id: 9,
      attack_type: 'flag_capture',
      source: 'submission',
      submitted_flag: 'flag{demo}',
      is_success: true,
      score_gained: 60,
      created_at: '2026-04-12T08:03:00Z',
    })

    const result = await submitContestAWDAttack('7', '7009', {
      victim_team_id: 14,
      flag: 'flag{demo}',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/contests/7/awd/services/7009/submissions',
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
      service_id: '7009',
      awd_challenge_id: '9',
      attack_type: 'flag_capture',
      source: 'submission',
      submitted_flag: 'flag{demo}',
      is_success: true,
      score_gained: 60,
      created_at: '2026-04-12T08:03:00Z',
    })
  })

  it('请求 AWD 跨队攻击入口时应调用目标代理 access 接口', async () => {
    requestMock.mockResolvedValue({
      access_url: '/api/v1/contests/7/awd/services/7009/targets/14/proxy/?ticket=demo',
    })

    const result = await requestContestAWDTargetAccess('7', '7009', '14')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/contests/7/awd/services/7009/targets/14/access',
    })
    expect(result.access_url).toBe(
      '/api/v1/contests/7/awd/services/7009/targets/14/proxy/?ticket=demo'
    )
  })

  it('请求 AWD 防守 SSH 入口时应调用 defense ssh 接口', async () => {
    requestMock.mockResolvedValue({
      host: '127.0.0.1',
      port: 2222,
      username: 'student+7+7009',
      password: 'ticket-secret',
      command: 'ssh student+7+7009@127.0.0.1 -p 2222',
      expires_at: '2026-04-12T08:15:00Z',
    })

    const result = await requestContestAWDDefenseSSH('7', '7009')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/contests/7/awd/services/7009/defense/ssh',
    })
    expect(result.command).toBe('ssh student+7+7009@127.0.0.1 -p 2222')
    expect(result.username).toBe('student+7+7009')
    expect(result.password).toBe('ticket-secret')
  })
})
