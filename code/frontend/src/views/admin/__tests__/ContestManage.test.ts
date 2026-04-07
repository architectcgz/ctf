import { describe, it, expect, beforeEach, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestManage from '../ContestManage.vue'

const exportMocks = vi.hoisted(() => ({
  downloadCSVFile: vi.fn(),
  downloadJSONFile: vi.fn(),
}))

const contestMocks = vi.hoisted(() => ({
  getContests: vi.fn(),
  createContest: vi.fn(),
  updateContest: vi.fn(),
  getAdminContestLiveScoreboard: vi.fn(),
  createContestAWDRound: vi.fn(),
  listContestTeams: vi.fn(),
  listAdminContestChallenges: vi.fn(),
  createContestAWDServiceCheck: vi.fn(),
  createContestAWDAttackLog: vi.fn(),
  listContestAWDRounds: vi.fn(),
  listContestAWDRoundServices: vi.fn(),
  listContestAWDRoundAttacks: vi.fn(),
  getContestAWDRoundSummary: vi.fn(),
  getContestAWDRoundTrafficSummary: vi.fn(),
  listContestAWDRoundTrafficEvents: vi.fn(),
  runContestAWDRoundCheck: vi.fn(),
  runContestAWDCurrentRoundCheck: vi.fn(),
}))

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    getContests: contestMocks.getContests,
    createContest: contestMocks.createContest,
    updateContest: contestMocks.updateContest,
    getAdminContestLiveScoreboard: contestMocks.getAdminContestLiveScoreboard,
    createContestAWDRound: contestMocks.createContestAWDRound,
    listContestTeams: contestMocks.listContestTeams,
    listAdminContestChallenges: contestMocks.listAdminContestChallenges,
    createContestAWDServiceCheck: contestMocks.createContestAWDServiceCheck,
    createContestAWDAttackLog: contestMocks.createContestAWDAttackLog,
    listContestAWDRounds: contestMocks.listContestAWDRounds,
    listContestAWDRoundServices: contestMocks.listContestAWDRoundServices,
    listContestAWDRoundAttacks: contestMocks.listContestAWDRoundAttacks,
    getContestAWDRoundSummary: contestMocks.getContestAWDRoundSummary,
    getContestAWDRoundTrafficSummary: contestMocks.getContestAWDRoundTrafficSummary,
    listContestAWDRoundTrafficEvents: contestMocks.listContestAWDRoundTrafficEvents,
    runContestAWDRoundCheck: contestMocks.runContestAWDRoundCheck,
    runContestAWDCurrentRoundCheck: contestMocks.runContestAWDCurrentRoundCheck,
  }
})

vi.mock('@/utils/csv', () => ({
  downloadCSVFile: exportMocks.downloadCSVFile,
  downloadJSONFile: exportMocks.downloadJSONFile,
}))

describe('ContestManage', () => {
  beforeEach(() => {
    vi.useRealTimers()
    window.sessionStorage.clear()
    exportMocks.downloadCSVFile.mockReset()
    exportMocks.downloadJSONFile.mockReset()
    contestMocks.getContests.mockReset()
    contestMocks.createContest.mockReset()
    contestMocks.updateContest.mockReset()
    contestMocks.getAdminContestLiveScoreboard.mockReset()
    contestMocks.createContestAWDRound.mockReset()
    contestMocks.listContestTeams.mockReset()
    contestMocks.listAdminContestChallenges.mockReset()
    contestMocks.createContestAWDServiceCheck.mockReset()
    contestMocks.createContestAWDAttackLog.mockReset()
    contestMocks.listContestAWDRounds.mockReset()
    contestMocks.listContestAWDRoundServices.mockReset()
    contestMocks.listContestAWDRoundAttacks.mockReset()
    contestMocks.getContestAWDRoundSummary.mockReset()
    contestMocks.getContestAWDRoundTrafficSummary.mockReset()
    contestMocks.listContestAWDRoundTrafficEvents.mockReset()
    contestMocks.runContestAWDRoundCheck.mockReset()
    contestMocks.runContestAWDCurrentRoundCheck.mockReset()

    contestMocks.listContestAWDRounds.mockResolvedValue([])
    contestMocks.listContestTeams.mockResolvedValue([])
    contestMocks.listAdminContestChallenges.mockResolvedValue([])
    contestMocks.createContestAWDRound.mockResolvedValue({
      id: '1',
      contest_id: '1',
      round_number: 1,
      status: 'pending',
      attack_score: 50,
      defense_score: 50,
      created_at: '2026-03-11T00:00:00.000Z',
      updated_at: '2026-03-11T00:00:00.000Z',
    })
    contestMocks.listContestAWDRoundServices.mockResolvedValue([])
    contestMocks.listContestAWDRoundAttacks.mockResolvedValue([])
    contestMocks.createContestAWDServiceCheck.mockResolvedValue({
      id: 'service-x',
      round_id: 'round-x',
      team_id: '1',
      team_name: 'Team 1',
      challenge_id: '101',
      service_status: 'up',
      check_result: {},
      attack_received: 0,
      defense_score: 50,
      attack_score: 0,
      updated_at: '2026-03-11T00:00:00.000Z',
    })
    contestMocks.createContestAWDAttackLog.mockResolvedValue({
      id: 'attack-x',
      round_id: 'round-x',
      attacker_team_id: '1',
      attacker_team: 'Team 1',
      victim_team_id: '2',
      victim_team: 'Team 2',
      challenge_id: '101',
      attack_type: 'flag_capture',
      source: 'manual_attack_log',
      is_success: true,
      score_gained: 50,
      created_at: '2026-03-11T00:00:00.000Z',
    })
    contestMocks.getContestAWDRoundSummary.mockResolvedValue({
      round: {
        id: '0',
        contest_id: '0',
        round_number: 0,
        status: 'pending',
        attack_score: 0,
        defense_score: 0,
        created_at: '2026-03-11T00:00:00.000Z',
        updated_at: '2026-03-11T00:00:00.000Z',
      },
      items: [],
    })
    contestMocks.getContestAWDRoundTrafficSummary.mockResolvedValue({
      contest_id: '0',
      round_id: '0',
      total_request_count: 0,
      active_attacker_team_count: 0,
      victim_team_count: 0,
      error_request_count: 0,
      unique_path_count: 0,
      latest_event_at: undefined,
      top_attackers: [],
      top_victims: [],
      top_challenges: [],
      top_error_paths: [],
      trend_buckets: [],
    })
    contestMocks.listContestAWDRoundTrafficEvents.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })
    contestMocks.runContestAWDRoundCheck.mockResolvedValue({
      round: {
        id: '0',
        contest_id: '0',
        round_number: 0,
        status: 'pending',
        attack_score: 0,
        defense_score: 0,
        created_at: '2026-03-11T00:00:00.000Z',
        updated_at: '2026-03-11T00:00:00.000Z',
      },
      services: [],
    })
    contestMocks.runContestAWDCurrentRoundCheck.mockResolvedValue({
      round: {
        id: '0',
        contest_id: '0',
        round_number: 0,
        status: 'pending',
        attack_score: 0,
        defense_score: 0,
        created_at: '2026-03-11T00:00:00.000Z',
        updated_at: '2026-03-11T00:00:00.000Z',
      },
      services: [],
    })
    contestMocks.getAdminContestLiveScoreboard.mockResolvedValue({
      contest: {
        id: '0',
        title: 'scoreboard',
        status: 'running',
        started_at: '2026-03-11T00:00:00.000Z',
        ends_at: '2026-03-11T02:00:00.000Z',
      },
      scoreboard: {
        list: [],
        total: 0,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })
  })

  it('应该渲染真实竞赛列表', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: '1',
          title: '2026 春季校园 CTF',
          description: '校内赛',
          mode: 'jeopardy',
          status: 'registering',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('赛事编排台')
    expect(wrapper.text()).toContain('2026 春季校园 CTF')
    expect(wrapper.text()).toContain('报名中')
    expect(wrapper.text()).toContain('当前页没有 AWD 赛事')
    expect(contestMocks.getContests).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      status: undefined,
    })
  })

  it('应该将主窗口、赛事列表与 AWD 运维视图拆分为顶部标签页', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: '1',
          title: '2026 春季校园 CTF',
          description: '校内赛',
          mode: 'jeopardy',
          status: 'registering',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
        {
          id: '2',
          title: '2026 AWD 联赛',
          description: '攻防赛',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-03-18T09:00:00.000Z',
          ends_at: '2026-03-18T18:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.get('#contest-tab-overview').attributes('aria-selected')).toBe('true')
    expect(wrapper.get('#contest-tab-list').attributes('aria-selected')).toBe('false')
    expect(wrapper.get('#contest-tab-operations').attributes('aria-selected')).toBe('false')
    expect(wrapper.get('#contest-panel-overview').attributes('aria-hidden')).toBe('false')
    expect(wrapper.get('#contest-panel-list').attributes('aria-hidden')).toBe('true')
    expect(wrapper.get('#contest-panel-operations').attributes('aria-hidden')).toBe('true')
    expect(wrapper.get('#contest-panel-overview').text()).toContain('赛事编排台')
    expect(wrapper.get('#contest-panel-overview').find('select').exists()).toBe(false)
    expect(wrapper.get('#contest-panel-overview').text()).not.toContain('选择 AWD 赛事')

    await wrapper.get('#contest-tab-list').trigger('click')

    expect(wrapper.get('#contest-tab-list').attributes('aria-selected')).toBe('true')
    expect(wrapper.get('#contest-panel-list').attributes('aria-hidden')).toBe('false')
    expect(wrapper.get('#contest-panel-list').text()).toContain('状态筛选')
    expect(wrapper.get('#contest-panel-list').text()).toContain('当前筛选结果')
    expect(wrapper.get('#contest-panel-list').text()).toContain('2026 AWD 联赛')
    expect(wrapper.get('#contest-panel-list').text()).not.toContain('选择 AWD 赛事')

    await wrapper.get('#contest-tab-operations').trigger('click')

    expect(wrapper.get('#contest-tab-operations').attributes('aria-selected')).toBe('true')
    expect(wrapper.get('#contest-panel-operations').attributes('aria-hidden')).toBe('false')
    expect(wrapper.get('#contest-panel-operations').text()).toContain('选择 AWD 赛事')
    expect(wrapper.get('#contest-panel-operations').text()).not.toContain('状态筛选')
    expect(wrapper.get('#contest-panel-operations').text()).not.toContain('赛事编排台')
  })

  it('应该在空列表时展示显式空态', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('暂无竞赛')
    expect(wrapper.text()).not.toContain('mockContests')
  })

  it('应该渲染 AWD 运维面板并触发所选轮次巡检', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'awd-1',
          title: '2026 校赛 AWD',
          description: '攻防对抗赛',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    contestMocks.listContestAWDRounds.mockResolvedValue([
      {
        id: 'round-1',
        contest_id: 'awd-1',
        round_number: 1,
        status: 'running',
        started_at: '2026-03-15T09:00:00.000Z',
        attack_score: 100,
        defense_score: 50,
        created_at: '2026-03-15T09:00:00.000Z',
        updated_at: '2026-03-15T09:05:00.000Z',
      },
    ])
    contestMocks.listContestTeams.mockResolvedValue([
      {
        id: 'team-1',
        contest_id: 'awd-1',
        name: '蓝队一',
        captain_id: '1001',
        invite_code: 'ABC123',
        max_members: 4,
        member_count: 3,
        created_at: '2026-03-15T08:00:00.000Z',
      },
      {
        id: 'team-2',
        contest_id: 'awd-1',
        name: '红队一',
        captain_id: '1002',
        invite_code: 'DEF456',
        max_members: 4,
        member_count: 3,
        created_at: '2026-03-15T08:00:00.000Z',
      },
    ])
    contestMocks.listAdminContestChallenges.mockResolvedValue([
      {
        id: 'link-1',
        contest_id: 'awd-1',
        challenge_id: 'challenge-1',
        points: 100,
        order: 1,
        is_visible: true,
        created_at: '2026-03-15T08:00:00.000Z',
      },
      {
        id: 'link-2',
        contest_id: 'awd-1',
        challenge_id: 'challenge-2',
        points: 150,
        order: 2,
        is_visible: true,
        created_at: '2026-03-15T08:01:00.000Z',
      },
    ])
    contestMocks.listContestAWDRoundServices.mockResolvedValue([
      {
        id: 'service-1',
        round_id: 'round-1',
        team_id: 'team-1',
        team_name: '蓝队一',
        challenge_id: 'challenge-1',
        service_status: 'compromised',
        check_result: {
          check_source: 'manual_selected_round',
          status_reason: 'unexpected_http_status',
          checked_at: '2026-03-15T09:05:00.000Z',
          targets: [
            {
              access_url: 'http://blue-team.internal',
              healthy: false,
              probe: 'http',
              error_code: 'unexpected_http_status',
              error: 'unexpected_http_status:503',
              attempts: [
                {
                  probe: 'http',
                  healthy: false,
                  error_code: 'unexpected_http_status',
                  error: 'unexpected_http_status:503',
                },
              ],
            },
          ],
        },
        attack_received: 2,
        defense_score: 0,
        attack_score: 100,
        updated_at: '2026-03-15T09:05:00.000Z',
      },
      {
        id: 'service-2',
        round_id: 'round-1',
        team_id: 'team-2',
        team_name: '红队一',
        challenge_id: 'challenge-1',
        service_status: 'up',
        check_result: {
          check_source: 'scheduler',
          status_reason: 'healthy',
          checked_at: '2026-03-15T09:05:30.000Z',
        },
        attack_received: 0,
        defense_score: 50,
        attack_score: 0,
        updated_at: '2026-03-15T09:05:30.000Z',
      },
      {
        id: 'service-3',
        round_id: 'round-1',
        team_id: 'team-2',
        team_name: '红队一',
        challenge_id: 'challenge-2',
        service_status: 'down',
        check_result: {
          check_source: 'manual_selected_round',
          status_reason: 'no_running_instances',
          error_code: 'no_running_instances',
          checked_at: '2026-03-15T09:05:45.000Z',
        },
        attack_received: 0,
        defense_score: 0,
        attack_score: 0,
        updated_at: '2026-03-15T09:05:45.000Z',
      },
    ])
    contestMocks.listContestAWDRoundAttacks.mockResolvedValue([
      {
        id: 'attack-1',
        round_id: 'round-1',
        attacker_team_id: 'team-2',
        attacker_team: '红队一',
        victim_team_id: 'team-1',
        victim_team: '蓝队一',
        challenge_id: 'challenge-1',
        attack_type: 'flag_capture',
        source: 'submission',
        is_success: true,
        score_gained: 100,
        created_at: '2026-03-15T09:04:00.000Z',
      },
      {
        id: 'attack-2',
        round_id: 'round-1',
        attacker_team_id: 'team-1',
        attacker_team: '蓝队一',
        victim_team_id: 'team-2',
        victim_team: '红队一',
        challenge_id: 'challenge-1',
        attack_type: 'service_exploit',
        source: 'manual_attack_log',
        is_success: false,
        score_gained: 0,
        created_at: '2026-03-15T09:04:30.000Z',
      },
    ])
    contestMocks.getContestAWDRoundSummary.mockResolvedValue({
      round: {
        id: 'round-1',
        contest_id: 'awd-1',
        round_number: 1,
        status: 'running',
        started_at: '2026-03-15T09:00:00.000Z',
        attack_score: 100,
        defense_score: 50,
        created_at: '2026-03-15T09:00:00.000Z',
        updated_at: '2026-03-15T09:05:00.000Z',
      },
      metrics: {
        total_service_count: 3,
        service_up_count: 1,
        service_down_count: 1,
        service_compromised_count: 1,
        attacked_service_count: 1,
        defense_success_count: 0,
        total_attack_count: 2,
        successful_attack_count: 1,
        failed_attack_count: 1,
        scheduler_check_count: 1,
        manual_current_round_check_count: 0,
        manual_selected_round_check_count: 1,
        manual_service_check_count: 0,
        submission_attack_count: 1,
        manual_attack_log_count: 1,
        legacy_attack_log_count: 0,
      },
      items: [
        {
          team_id: 'team-2',
          team_name: '红队一',
          service_up_count: 1,
          service_down_count: 1,
          service_compromised_count: 0,
          defense_score: 50,
          attack_score: 100,
          successful_attack_count: 1,
          successful_breach_count: 0,
          unique_attackers_against: 0,
          total_score: 150,
        },
        {
          team_id: 'team-1',
          team_name: '蓝队一',
          service_up_count: 0,
          service_down_count: 0,
          service_compromised_count: 1,
          defense_score: 0,
          attack_score: 0,
          successful_attack_count: 0,
          successful_breach_count: 1,
          unique_attackers_against: 1,
          total_score: 0,
        },
      ],
    })
    contestMocks.getContestAWDRoundTrafficSummary.mockResolvedValue({
      contest_id: 'awd-1',
      round_id: 'round-1',
      total_request_count: 18,
      active_attacker_team_count: 2,
      victim_team_count: 2,
      error_request_count: 4,
      unique_path_count: 5,
      latest_event_at: '2026-03-15T09:04:40.000Z',
      trend_buckets: [
        {
          bucket_start_at: '2026-03-15T09:00:00.000Z',
          bucket_end_at: '2026-03-15T09:01:00.000Z',
          request_count: 5,
          error_count: 1,
        },
        {
          bucket_start_at: '2026-03-15T09:01:00.000Z',
          bucket_end_at: '2026-03-15T09:02:00.000Z',
          request_count: 9,
          error_count: 2,
        },
      ],
      top_victims: [
        {
          team_id: 'team-1',
          team_name: '蓝队一',
          request_count: 11,
          error_count: 3,
        },
      ],
      top_attackers: [
        {
          team_id: 'team-2',
          team_name: '红队一',
          request_count: 12,
          error_count: 4,
        },
      ],
      top_challenges: [
        {
          challenge_id: 'challenge-1',
          challenge_title: 'Traffic Alpha',
          request_count: 10,
          error_count: 4,
        },
      ],
      top_error_paths: [
        {
          path: '/api/alpha',
          request_count: 8,
          error_count: 3,
        },
      ],
    })
    contestMocks.listContestAWDRoundTrafficEvents.mockImplementation(async (_contestId, _roundId, params) => {
      const page = params?.page ?? 1
      const pageSize = params?.page_size ?? 20
      const statusGroup = params?.status_group
      if (page === 2) {
        return {
          list: [
            {
              id: 'traffic-3',
              contest_id: 'awd-1',
              round_id: 'round-1',
              occurred_at: '2026-03-15T09:05:30.000Z',
              attacker_team_id: 'team-2',
              attacker_team_name: '红队一',
              victim_team_id: 'team-1',
              victim_team_name: '蓝队一',
              challenge_id: 'challenge-2',
              challenge_title: 'Traffic Gamma',
              method: 'GET',
              path: '/api/gamma',
              status_code: 302,
              status_group: 'redirect',
              is_error: false,
              source: 'proxy_audit',
              request_id: 'req-traffic-3',
            },
          ],
          total: 21,
          page,
          page_size: pageSize,
        }
      }
      if (statusGroup === 'server_error') {
        return {
          list: [
            {
              id: 'traffic-1',
              contest_id: 'awd-1',
              round_id: 'round-1',
              occurred_at: '2026-03-15T09:04:10.000Z',
              attacker_team_id: 'team-2',
              attacker_team_name: '红队一',
              victim_team_id: 'team-1',
              victim_team_name: '蓝队一',
              challenge_id: 'challenge-1',
              challenge_title: 'Traffic Alpha',
              method: 'POST',
              path: '/api/alpha',
              status_code: 500,
              status_group: 'server_error',
              is_error: true,
              source: 'proxy_audit',
              request_id: 'req-traffic-1',
            },
          ],
          total: 1,
          page,
          page_size: pageSize,
        }
      }
      return {
        list: [
          {
            id: 'traffic-1',
            contest_id: 'awd-1',
            round_id: 'round-1',
            occurred_at: '2026-03-15T09:04:10.000Z',
            attacker_team_id: 'team-2',
            attacker_team_name: '红队一',
            victim_team_id: 'team-1',
            victim_team_name: '蓝队一',
            challenge_id: 'challenge-1',
            challenge_title: 'Traffic Alpha',
            method: 'POST',
            path: '/api/alpha',
            status_code: 500,
            status_group: 'server_error',
            is_error: true,
            source: 'proxy_audit',
            request_id: 'req-traffic-1',
          },
          {
            id: 'traffic-2',
            contest_id: 'awd-1',
            round_id: 'round-1',
            occurred_at: '2026-03-15T09:04:40.000Z',
            attacker_team_id: 'team-1',
            attacker_team_name: '蓝队一',
            victim_team_id: 'team-2',
            victim_team_name: '红队一',
            challenge_id: 'challenge-2',
            challenge_title: 'Traffic Beta',
            method: 'GET',
            path: '/api/beta',
            status_code: 200,
            status_group: 'success',
            is_error: false,
            source: 'proxy_audit',
            request_id: 'req-traffic-2',
          },
        ],
        total: 21,
        page,
        page_size: pageSize,
      }
    })
    contestMocks.getAdminContestLiveScoreboard.mockResolvedValue({
      contest: {
        id: 'awd-1',
        title: '2026 校赛 AWD',
        status: 'running',
        started_at: '2026-03-15T09:00:00.000Z',
        ends_at: '2026-03-15T13:00:00.000Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: 'team-2',
            team_name: '红队一',
            score: 150,
            solved_count: 3,
            last_submission_at: '2026-03-15T09:04:00.000Z',
          },
          {
            rank: 2,
            team_id: 'team-1',
            team_name: '蓝队一',
            score: 0,
            solved_count: 0,
            last_submission_at: undefined,
          },
        ],
        total: 2,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })
    contestMocks.runContestAWDRoundCheck.mockResolvedValue({
      round: {
        id: 'round-1',
        contest_id: 'awd-1',
        round_number: 1,
        status: 'running',
        started_at: '2026-03-15T09:00:00.000Z',
        attack_score: 100,
        defense_score: 50,
        created_at: '2026-03-15T09:00:00.000Z',
        updated_at: '2026-03-15T09:06:00.000Z',
      },
      services: [],
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('AWD 运维视图')
    expect(wrapper.text()).toContain('2026 校赛 AWD')
    expect(wrapper.text()).toContain('蓝队一')
    expect(wrapper.text()).toContain('红队一')
    expect(wrapper.text()).toContain('来源: 指定轮次重跑')
    expect(wrapper.text()).toContain('状态: HTTP 状态异常')
    expect(wrapper.text()).toContain('重点告警')
    expect(wrapper.text()).toContain('受影响服务 1 个')
    expect(wrapper.text()).toContain('探测目标 0/1 正常')
    expect(wrapper.text()).toContain('查看探测明细')
    expect(wrapper.text()).toContain('无运行实例')
    expect(wrapper.text()).toContain('学员提交')
    expect(wrapper.text()).toContain('调度巡检')
    expect(wrapper.text()).toContain('人工补录')
    expect(wrapper.text()).toContain('防守成功')
    expect(wrapper.text()).toContain('巡检 调度 1 / 手动 1')
    expect(wrapper.text()).toContain('日志 提交 1 / 人工 1 / 历史 0')
    expect(wrapper.text()).toContain('攻击流量态势')
    expect(wrapper.text()).toContain('趋势摘要')
    expect(wrapper.text()).toContain('热点目标')
    expect(wrapper.text()).toContain('热点攻击队')
    expect(wrapper.text()).toContain('异常路径')
    expect(wrapper.text()).toContain('Traffic Alpha')
    expect(wrapper.text()).toContain('Traffic Beta')
    expect(wrapper.text()).toContain('实时排行榜')
    expect(wrapper.text()).toContain('#1')
    expect(wrapper.text()).toContain('150')
    expect(wrapper.text()).toContain('2026/03/15')
    expect(contestMocks.listContestAWDRounds).toHaveBeenCalledWith('awd-1')
    expect(contestMocks.listContestAWDRoundServices).toHaveBeenCalledWith('awd-1', 'round-1')
    expect(contestMocks.getContestAWDRoundSummary).toHaveBeenCalledWith('awd-1', 'round-1')
    expect(contestMocks.getContestAWDRoundTrafficSummary).toHaveBeenCalledWith('awd-1', 'round-1')
    expect(contestMocks.listContestAWDRoundTrafficEvents).toHaveBeenCalledWith(
      'awd-1',
      'round-1',
      expect.objectContaining({
        page: 1,
        page_size: 20,
      })
    )
    expect(contestMocks.getAdminContestLiveScoreboard).toHaveBeenCalledWith('awd-1', { page: 1, page_size: 10 })

    await wrapper.find('#awd-service-filter-source').setValue('manual_selected_round')
    await flushPromises()
    expect(wrapper.text()).toContain('来源: 指定轮次重跑')
    expect(wrapper.text()).not.toContain('来源: 调度巡检')
    expect(wrapper.text()).toContain('http://blue-team.internal')
    expect(wrapper.text()).toContain('Attempt 1: HTTP · HTTP 状态异常')

    await wrapper.find('#awd-service-filter-alert').setValue('unexpected_http_status')
    await flushPromises()
    expect(wrapper.text()).toContain('蓝队一')

    await wrapper.find('#awd-attack-filter-source').setValue('manual_attack_log')
    await flushPromises()
    expect(wrapper.text()).toContain('人工补录')
    expect(wrapper.text()).not.toContain('学员提交 成功 +100')

    await wrapper.find('#awd-traffic-filter-status-group').setValue('server_error')
    await flushPromises()
    expect(wrapper.text()).toContain('Traffic Alpha')
    expect(wrapper.text()).not.toContain('Traffic Beta')
    expect(contestMocks.listContestAWDRoundTrafficEvents).toHaveBeenLastCalledWith(
      'awd-1',
      'round-1',
      expect.objectContaining({
        page: 1,
        page_size: 20,
        status_group: 'server_error',
      })
    )

    await wrapper.find('#awd-traffic-filter-path').setValue('alpha')
    await wrapper.find('#awd-traffic-filter-search').trigger('click')
    await flushPromises()
    expect(contestMocks.listContestAWDRoundTrafficEvents).toHaveBeenLastCalledWith(
      'awd-1',
      'round-1',
      expect.objectContaining({
        page: 1,
        page_size: 20,
        status_group: 'server_error',
        path_keyword: 'alpha',
      })
    )

    const trafficNextPageButton = wrapper.find('#awd-traffic-page-next')
    expect(trafficNextPageButton.attributes('disabled')).toBeDefined()
    await trafficNextPageButton.trigger('click')
    await flushPromises()
    expect(wrapper.text()).not.toContain('Traffic Gamma')
    expect(contestMocks.listContestAWDRoundTrafficEvents).toHaveBeenLastCalledWith(
      'awd-1',
      'round-1',
      expect.objectContaining({
        page: 1,
        page_size: 20,
        status_group: 'server_error',
        path_keyword: 'alpha',
      })
    )

    await wrapper.find('#awd-export-services').trigger('click')
    expect(exportMocks.downloadCSVFile).toHaveBeenCalledWith(
      '2026-AWD-round-1-services.csv',
      [
        {
          赛事: '2026 校赛 AWD',
          轮次: '第 1 轮',
          筛选队伍: '全部队伍',
          筛选状态: '全部状态',
          筛选来源: '指定轮次重跑',
          筛选告警: 'HTTP 状态异常',
          队伍: '蓝队一',
          靶题: 'Challenge #challenge-1',
          服务状态: '已失陷',
          巡检来源: '指定轮次重跑',
          检查摘要: '来源: 指定轮次重跑 | 状态: HTTP 状态异常 | 时间: 2026/03/15 17:05:00',
          防守得分: 0,
          受攻击次数: 2,
          更新时间: '2026/03/15 17:05:00',
        },
      ]
    )

    await wrapper.find('#awd-export-attacks').trigger('click')
    expect(exportMocks.downloadCSVFile).toHaveBeenLastCalledWith(
      '2026-AWD-round-1-attacks.csv',
      [
        {
          赛事: '2026 校赛 AWD',
          轮次: '第 1 轮',
          筛选队伍: '全部队伍',
          筛选结果: '全部结果',
          筛选来源: '人工补录',
          时间: '2026/03/15 17:04:30',
          攻击方: '蓝队一',
          受害方: '红队一',
          靶题: 'Challenge #challenge-1',
          攻击类型: '服务利用',
          记录来源: '人工补录',
          攻击结果: '失败',
          得分: 0,
          提交Flag: '',
        },
      ]
    )

    await wrapper.find('#awd-export-review-package').trigger('click')
    expect(exportMocks.downloadJSONFile).toHaveBeenCalledWith(
      '2026-AWD-round-1-review-package.json',
      expect.objectContaining({
        contest: expect.objectContaining({
          id: 'awd-1',
          title: '2026 校赛 AWD',
        }),
        round: expect.objectContaining({
          id: 'round-1',
          round_number: 1,
        }),
        summary: expect.objectContaining({
          metrics: expect.objectContaining({
            total_attack_count: 2,
            manual_selected_round_check_count: 1,
            manual_attack_log_count: 1,
          }),
          service_alerts: expect.arrayContaining([
            expect.objectContaining({
              key: 'unexpected_http_status',
              label: 'HTTP 状态异常',
              count: 1,
            }),
            expect.objectContaining({
              key: 'no_running_instances',
              label: '无运行实例',
              count: 1,
            }),
          ]),
        }),
        scoreboard: expect.objectContaining({
          frozen: false,
          rows: expect.arrayContaining([
            expect.objectContaining({
              team_id: 'team-2',
              rank: 1,
              score: 150,
            }),
          ]),
        }),
        filters: expect.objectContaining({
          service: expect.objectContaining({
            check_source: 'manual_selected_round',
            alert_reason: 'unexpected_http_status',
          }),
          attack: expect.objectContaining({
            source: 'manual_attack_log',
          }),
        }),
        services: expect.arrayContaining([
          expect.objectContaining({
            team_id: 'team-1',
            check_source: 'manual_selected_round',
          }),
        ]),
        attacks: expect.arrayContaining([
          expect.objectContaining({
            id: 'attack-2',
            source: 'manual_attack_log',
          }),
        ]),
      })
    )

    const checkButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('立即巡检当前轮'))

    expect(checkButton).toBeDefined()
    await checkButton!.trigger('click')
    await flushPromises()

    expect(contestMocks.runContestAWDRoundCheck).toHaveBeenCalledWith('awd-1', 'round-1')
  })

  it('应该在运行中的 AWD 轮次上自动刷新运维面板', async () => {
    vi.useFakeTimers()

    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'awd-live',
          title: '2026 实时 AWD',
          description: '实时监控',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    contestMocks.listContestAWDRounds.mockResolvedValue([
      {
        id: 'live-round-1',
        contest_id: 'awd-live',
        round_number: 1,
        status: 'running',
        started_at: '2026-03-15T09:00:00.000Z',
        attack_score: 100,
        defense_score: 50,
        created_at: '2026-03-15T09:00:00.000Z',
        updated_at: '2026-03-15T09:05:00.000Z',
      },
    ])
    contestMocks.listContestTeams.mockResolvedValue([
      {
        id: 'team-live-1',
        contest_id: 'awd-live',
        name: 'Live Team',
        captain_id: '1001',
        invite_code: 'LIVE1',
        max_members: 4,
        member_count: 3,
        created_at: '2026-03-15T08:00:00.000Z',
      },
    ])
    contestMocks.listAdminContestChallenges.mockResolvedValue([
      {
        id: 'link-live-1',
        contest_id: 'awd-live',
        challenge_id: 'challenge-live-1',
        title: 'Health Check',
        points: 100,
        order: 1,
        is_visible: true,
        created_at: '2026-03-15T08:00:00.000Z',
      },
    ])
    contestMocks.listContestAWDRoundServices.mockResolvedValue([
      {
        id: 'service-live-1',
        round_id: 'live-round-1',
        team_id: 'team-live-1',
        team_name: 'Live Team',
        challenge_id: 'challenge-live-1',
        service_status: 'up',
        check_result: {
          check_source: 'scheduler',
          status_reason: 'healthy',
          checked_at: '2026-03-15T09:05:00.000Z',
        },
        attack_received: 0,
        defense_score: 50,
        attack_score: 0,
        updated_at: '2026-03-15T09:05:00.000Z',
      },
    ])
    contestMocks.listContestAWDRoundAttacks.mockResolvedValue([])
    contestMocks.getContestAWDRoundSummary.mockResolvedValue({
      round: {
        id: 'live-round-1',
        contest_id: 'awd-live',
        round_number: 1,
        status: 'running',
        started_at: '2026-03-15T09:00:00.000Z',
        attack_score: 100,
        defense_score: 50,
        created_at: '2026-03-15T09:00:00.000Z',
        updated_at: '2026-03-15T09:05:00.000Z',
      },
      metrics: {
        total_service_count: 1,
        service_up_count: 1,
        service_down_count: 0,
        service_compromised_count: 0,
        attacked_service_count: 0,
        defense_success_count: 0,
        total_attack_count: 0,
        successful_attack_count: 0,
        failed_attack_count: 0,
        scheduler_check_count: 1,
        manual_current_round_check_count: 0,
        manual_selected_round_check_count: 0,
        manual_service_check_count: 0,
        submission_attack_count: 0,
        manual_attack_log_count: 0,
        legacy_attack_log_count: 0,
      },
      items: [],
    })
    contestMocks.getAdminContestLiveScoreboard.mockResolvedValue({
      contest: {
        id: 'awd-live',
        title: '2026 实时 AWD',
        status: 'running',
        started_at: '2026-03-15T09:00:00.000Z',
        ends_at: '2026-03-15T13:00:00.000Z',
      },
      scoreboard: {
        list: [],
        total: 0,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('面板会每 15 秒自动刷新一次')
    const initialRoundCalls = contestMocks.listContestAWDRounds.mock.calls.length
    const initialServiceCalls = contestMocks.listContestAWDRoundServices.mock.calls.length
    expect(initialRoundCalls).toBeGreaterThan(0)
    expect(initialServiceCalls).toBeGreaterThan(0)

    await vi.advanceTimersByTimeAsync(15_000)
    await flushPromises()
    await flushPromises()

    expect(contestMocks.listContestAWDRounds).toHaveBeenCalledTimes(initialRoundCalls + 1)
    expect(contestMocks.listContestAWDRoundServices).toHaveBeenCalledTimes(initialServiceCalls + 1)

    wrapper.unmount()
    vi.useRealTimers()
  })

  it('应该允许管理员在 AWD 面板创建新轮次', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'awd-2',
          title: '2026 校赛 AWD 进阶场',
          description: '支持手工建轮',
          mode: 'awd',
          status: 'registering',
          starts_at: '2026-03-20T09:00:00.000Z',
          ends_at: '2026-03-20T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    contestMocks.listContestAWDRounds.mockResolvedValueOnce([])
    contestMocks.listContestTeams.mockResolvedValue([
      {
        id: 'team-9',
        contest_id: 'awd-2',
        name: '蓝队九',
        captain_id: '109',
        invite_code: 'JKL999',
        max_members: 4,
        member_count: 1,
        created_at: '2026-03-20T08:00:00.000Z',
      },
    ])
    contestMocks.listAdminContestChallenges.mockResolvedValue([
      {
        id: 'link-9',
        contest_id: 'awd-2',
        challenge_id: '301',
        points: 100,
        order: 1,
        is_visible: true,
        created_at: '2026-03-20T08:00:00.000Z',
      },
    ])
    contestMocks.createContestAWDRound.mockResolvedValue({
      id: 'round-9',
      contest_id: 'awd-2',
      round_number: 1,
      status: 'pending',
      attack_score: 60,
      defense_score: 40,
      created_at: '2026-03-20T08:00:00.000Z',
      updated_at: '2026-03-20T08:00:00.000Z',
    })
    contestMocks.listContestAWDRounds.mockResolvedValueOnce([
      {
        id: 'round-9',
        contest_id: 'awd-2',
        round_number: 1,
        status: 'pending',
        attack_score: 60,
        defense_score: 40,
        created_at: '2026-03-20T08:00:00.000Z',
        updated_at: '2026-03-20T08:00:00.000Z',
      },
    ])

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue">{{ title }}</div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    await flushPromises()

    const openDialogButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('创建轮次'))

    expect(openDialogButton).toBeDefined()
    await openDialogButton!.trigger('click')
    await flushPromises()

    const numberInput = wrapper.find('#awd-round-number')
    await numberInput.setValue('1')
    await wrapper.find('#awd-attack-score').setValue('60')
    await wrapper.find('#awd-defense-score').setValue('40')

    await wrapper.find('#awd-round-create-submit').trigger('click')
    await flushPromises()
    await flushPromises()

    expect(contestMocks.createContestAWDRound).toHaveBeenCalledWith('awd-2', {
      round_number: 1,
      status: 'pending',
      attack_score: 60,
      defense_score: 40,
    })
    expect(contestMocks.listContestAWDRounds).toHaveBeenLastCalledWith('awd-2')
    expect(wrapper.text()).toContain('第 1 轮')
  })

  it('应该允许管理员补录服务检查结果', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'awd-3',
          title: '2026 校赛 AWD 值守场',
          description: '支持手工补录',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-03-21T09:00:00.000Z',
          ends_at: '2026-03-21T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    contestMocks.listContestAWDRounds.mockResolvedValue([
      {
        id: 'round-3',
        contest_id: 'awd-3',
        round_number: 3,
        status: 'running',
        attack_score: 70,
        defense_score: 40,
        created_at: '2026-03-21T09:00:00.000Z',
        updated_at: '2026-03-21T09:05:00.000Z',
      },
    ])
    contestMocks.listContestAWDRoundServices.mockResolvedValue([])
    contestMocks.listContestTeams.mockResolvedValue([
      {
        id: '11',
        contest_id: 'awd-3',
        name: '蓝队十一',
        captain_id: '1001',
        invite_code: 'XYZ111',
        max_members: 4,
        member_count: 3,
        created_at: '2026-03-21T08:00:00.000Z',
      },
    ])
    contestMocks.listAdminContestChallenges.mockResolvedValue([
      {
        id: 'link-11',
        contest_id: 'awd-3',
        challenge_id: '501',
        points: 100,
        order: 1,
        is_visible: true,
        created_at: '2026-03-21T08:00:00.000Z',
      },
    ])

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue">{{ title }}</div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    await flushPromises()

    const openDialogButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('录入服务检查'))

    expect(openDialogButton).toBeDefined()
    await openDialogButton!.trigger('click')
    await flushPromises()

    await wrapper.find('#awd-service-team').setValue('11')
    await wrapper.find('#awd-service-challenge').setValue('501')
    await wrapper.find('#awd-service-status').setValue('down')
    await wrapper.find('#awd-service-check-result').setValue('{"http_status":503,"reason":"timeout"}')
    contestMocks.listContestAWDRoundServices.mockResolvedValue([
      {
        id: 'service-manual-1',
        round_id: 'round-3',
        team_id: '11',
        team_name: '蓝队十一',
        challenge_id: '501',
        service_status: 'down',
        check_result: {
          check_source: 'manual_service_check',
          checked_at: '2026-03-21T09:06:00.000Z',
          reason: 'timeout',
        },
        attack_received: 0,
        defense_score: 0,
        attack_score: 0,
        updated_at: '2026-03-21T09:06:00.000Z',
      },
    ])
    await wrapper.find('#awd-service-check-submit').trigger('click')
    await flushPromises()
    await flushPromises()

    expect(contestMocks.createContestAWDServiceCheck).toHaveBeenCalledWith('awd-3', 'round-3', {
      team_id: 11,
      challenge_id: 501,
      service_status: 'down',
      check_result: {
        http_status: 503,
        reason: 'timeout',
      },
    })
    expect(wrapper.text()).toContain('来源: 人工补录')
  })

  it('应该允许管理员补录攻击日志', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'awd-4',
          title: '2026 校赛 AWD 攻击补录场',
          description: '支持补录攻击日志',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-03-22T09:00:00.000Z',
          ends_at: '2026-03-22T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    contestMocks.listContestAWDRounds.mockResolvedValue([
      {
        id: 'round-4',
        contest_id: 'awd-4',
        round_number: 2,
        status: 'running',
        attack_score: 80,
        defense_score: 40,
        created_at: '2026-03-22T09:00:00.000Z',
        updated_at: '2026-03-22T09:05:00.000Z',
      },
    ])
    contestMocks.listContestTeams.mockResolvedValue([
      {
        id: '21',
        contest_id: 'awd-4',
        name: '红队二一',
        captain_id: '2001',
        invite_code: 'RT21',
        max_members: 4,
        member_count: 3,
        created_at: '2026-03-22T08:00:00.000Z',
      },
      {
        id: '22',
        contest_id: 'awd-4',
        name: '蓝队二二',
        captain_id: '2002',
        invite_code: 'BT22',
        max_members: 4,
        member_count: 3,
        created_at: '2026-03-22T08:00:00.000Z',
      },
    ])
    contestMocks.listAdminContestChallenges.mockResolvedValue([
      {
        id: 'link-21',
        contest_id: 'awd-4',
        challenge_id: '601',
        points: 120,
        order: 1,
        is_visible: true,
        created_at: '2026-03-22T08:00:00.000Z',
      },
    ])

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue">{{ title }}</div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    await flushPromises()

    const openDialogButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('补录攻击日志'))

    expect(openDialogButton).toBeDefined()
    await openDialogButton!.trigger('click')
    await flushPromises()

    await wrapper.find('#awd-attack-team').setValue('21')
    await wrapper.find('#awd-victim-team').setValue('22')
    await wrapper.find('#awd-attack-challenge').setValue('601')
    await wrapper.find('#awd-attack-type').setValue('service_exploit')
    await wrapper.find('#awd-attack-flag').setValue('awd{manual-log}')
    await wrapper.find('#awd-attack-log-submit').trigger('click')
    await flushPromises()
    await flushPromises()

    expect(contestMocks.createContestAWDAttackLog).toHaveBeenCalledWith('awd-4', 'round-4', {
      attacker_team_id: 21,
      victim_team_id: 22,
      challenge_id: 601,
      attack_type: 'service_exploit',
      submitted_flag: 'awd{manual-log}',
      is_success: true,
    })
  })

  it('应该恢复 AWD 赛事、轮次与筛选状态', async () => {
    window.sessionStorage.setItem('ctf_admin_awd_selected_contest', 'awd-restore-2')
    window.sessionStorage.setItem('ctf_admin_awd_selected_round:awd-restore-2', 'round-restore-2')
    window.sessionStorage.setItem(
      'ctf_admin_awd_filters:awd-restore-2:round-restore-2',
      JSON.stringify({
        service_team_id: 'team-b',
        service_status: 'down',
        service_check_source: 'manual_service_check',
        attack_team_id: 'team-b',
        attack_result: 'failed',
        attack_source: 'manual_attack_log',
      })
    )

    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'awd-restore-1',
          title: 'AWD Restore 1',
          description: 'restore case 1',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-03-25T09:00:00.000Z',
          ends_at: '2026-03-25T13:00:00.000Z',
        },
        {
          id: 'awd-restore-2',
          title: 'AWD Restore 2',
          description: 'restore case 2',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-03-26T09:00:00.000Z',
          ends_at: '2026-03-26T13:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })
    contestMocks.listContestAWDRounds.mockImplementation(async (contestId: string) =>
      contestId === 'awd-restore-2'
        ? [
            {
              id: 'round-restore-1',
              contest_id: 'awd-restore-2',
              round_number: 1,
              status: 'finished',
              attack_score: 60,
              defense_score: 30,
              created_at: '2026-03-26T09:00:00.000Z',
              updated_at: '2026-03-26T09:05:00.000Z',
            },
            {
              id: 'round-restore-2',
              contest_id: 'awd-restore-2',
              round_number: 2,
              status: 'running',
              attack_score: 80,
              defense_score: 40,
              created_at: '2026-03-26T09:10:00.000Z',
              updated_at: '2026-03-26T09:15:00.000Z',
            },
          ]
        : []
    )
    contestMocks.listContestTeams.mockImplementation(async (contestId: string) =>
      contestId === 'awd-restore-2'
        ? [
            {
              id: 'team-a',
              contest_id: 'awd-restore-2',
              name: '红队甲',
              captain_id: '3101',
              invite_code: 'RA',
              max_members: 4,
              member_count: 3,
              created_at: '2026-03-26T08:00:00.000Z',
            },
            {
              id: 'team-b',
              contest_id: 'awd-restore-2',
              name: '蓝队乙',
              captain_id: '3102',
              invite_code: 'RB',
              max_members: 4,
              member_count: 3,
              created_at: '2026-03-26T08:00:00.000Z',
            },
          ]
        : []
    )
    contestMocks.listAdminContestChallenges.mockImplementation(async (contestId: string) =>
      contestId === 'awd-restore-2'
        ? [
            {
              id: 'link-restore-1',
              contest_id: 'awd-restore-2',
              challenge_id: 'restore-challenge-1',
              title: 'Restore Challenge',
              points: 120,
              order: 1,
              is_visible: true,
              created_at: '2026-03-26T08:00:00.000Z',
            },
          ]
        : []
    )
    contestMocks.listContestAWDRoundServices.mockImplementation(async (contestId: string, roundId: string) =>
      contestId === 'awd-restore-2' && roundId === 'round-restore-2'
        ? [
            {
              id: 'restore-service-a',
              round_id: 'round-restore-2',
              team_id: 'team-a',
              team_name: '红队甲',
              challenge_id: 'restore-challenge-1',
              service_status: 'up',
              check_result: {
                check_source: 'scheduler',
                status_reason: 'healthy',
                checked_at: '2026-03-26T09:15:00.000Z',
              },
              attack_received: 0,
              defense_score: 40,
              attack_score: 0,
              updated_at: '2026-03-26T09:15:00.000Z',
            },
            {
              id: 'restore-service-b',
              round_id: 'round-restore-2',
              team_id: 'team-b',
              team_name: '蓝队乙',
              challenge_id: 'restore-challenge-1',
              service_status: 'down',
              check_result: {
                check_source: 'manual_service_check',
                checked_at: '2026-03-26T09:16:00.000Z',
                reason: 'manual timeout',
              },
              attack_received: 1,
              defense_score: 0,
              attack_score: 0,
              updated_at: '2026-03-26T09:16:00.000Z',
            },
          ]
        : []
    )
    contestMocks.listContestAWDRoundAttacks.mockImplementation(async (contestId: string, roundId: string) =>
      contestId === 'awd-restore-2' && roundId === 'round-restore-2'
        ? [
            {
              id: 'restore-attack-a',
              round_id: 'round-restore-2',
              attacker_team_id: 'team-a',
              attacker_team: '红队甲',
              victim_team_id: 'team-b',
              victim_team: '蓝队乙',
              challenge_id: 'restore-challenge-1',
              attack_type: 'flag_capture',
              source: 'submission',
              is_success: true,
              score_gained: 80,
              created_at: '2026-03-26T09:12:00.000Z',
            },
            {
              id: 'restore-attack-b',
              round_id: 'round-restore-2',
              attacker_team_id: 'team-b',
              attacker_team: '蓝队乙',
              victim_team_id: 'team-a',
              victim_team: '红队甲',
              challenge_id: 'restore-challenge-1',
              attack_type: 'service_exploit',
              source: 'manual_attack_log',
              is_success: false,
              score_gained: 0,
              created_at: '2026-03-26T09:13:00.000Z',
            },
          ]
        : []
    )
    contestMocks.getContestAWDRoundSummary.mockResolvedValue({
      round: {
        id: 'round-restore-2',
        contest_id: 'awd-restore-2',
        round_number: 2,
        status: 'running',
        attack_score: 80,
        defense_score: 40,
        created_at: '2026-03-26T09:10:00.000Z',
        updated_at: '2026-03-26T09:15:00.000Z',
      },
      items: [],
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    await flushPromises()

    expect((wrapper.find('#awd-contest-selector').element as HTMLSelectElement).value).toBe(
      'awd-restore-2'
    )
    expect((wrapper.find('#awd-round-selector').element as HTMLSelectElement).value).toBe(
      'round-restore-2'
    )
    expect((wrapper.find('#awd-service-filter-team').element as HTMLSelectElement).value).toBe(
      'team-b'
    )
    expect((wrapper.find('#awd-service-filter-status').element as HTMLSelectElement).value).toBe(
      'down'
    )
    expect((wrapper.find('#awd-service-filter-source').element as HTMLSelectElement).value).toBe(
      'manual_service_check'
    )
    expect((wrapper.find('#awd-attack-filter-team').element as HTMLSelectElement).value).toBe(
      'team-b'
    )
    expect((wrapper.find('#awd-attack-filter-result').element as HTMLSelectElement).value).toBe(
      'failed'
    )
    expect((wrapper.find('#awd-attack-filter-source').element as HTMLSelectElement).value).toBe(
      'manual_attack_log'
    )

    await wrapper.find('#awd-export-services').trigger('click')
    expect(exportMocks.downloadCSVFile).toHaveBeenCalledWith(
      'AWD-Restore-2-round-2-services.csv',
      [
        {
          赛事: 'AWD Restore 2',
          轮次: '第 2 轮',
          筛选队伍: '蓝队乙',
          筛选状态: '下线',
          筛选来源: '人工补录',
          筛选告警: '全部告警',
          队伍: '蓝队乙',
          靶题: 'Restore Challenge',
          服务状态: '下线',
          巡检来源: '人工补录',
          检查摘要: '来源: 人工补录 | 时间: 2026/03/26 17:16:00',
          防守得分: 0,
          受攻击次数: 1,
          更新时间: '2026/03/26 17:16:00',
        },
      ]
    )

    await wrapper.find('#awd-export-attacks').trigger('click')
    expect(exportMocks.downloadCSVFile).toHaveBeenLastCalledWith(
      'AWD-Restore-2-round-2-attacks.csv',
      [
        {
          赛事: 'AWD Restore 2',
          轮次: '第 2 轮',
          筛选队伍: '蓝队乙',
          筛选结果: '仅失败',
          筛选来源: '人工补录',
          时间: '2026/03/26 17:13:00',
          攻击方: '蓝队乙',
          受害方: '红队甲',
          靶题: 'Restore Challenge',
          攻击类型: '服务利用',
          记录来源: '人工补录',
          攻击结果: '失败',
          得分: 0,
          提交Flag: '',
        },
      ]
    )

    await wrapper.find('#awd-export-review-package').trigger('click')
    expect(exportMocks.downloadJSONFile).toHaveBeenLastCalledWith(
      'AWD-Restore-2-round-2-review-package.json',
      expect.objectContaining({
        contest: expect.objectContaining({
          id: 'awd-restore-2',
          title: 'AWD Restore 2',
        }),
        round: expect.objectContaining({
          id: 'round-restore-2',
          round_number: 2,
        }),
        filters: expect.objectContaining({
          service: expect.objectContaining({
            team_id: 'team-b',
            status: 'down',
            check_source: 'manual_service_check',
          }),
          attack: expect.objectContaining({
            team_id: 'team-b',
            result: 'failed',
            source: 'manual_attack_log',
          }),
        }),
        services: [
          expect.objectContaining({
            team_id: 'team-b',
            service_status: 'down',
          }),
        ],
        attacks: [
          expect.objectContaining({
            attacker_team_id: 'team-b',
            source: 'manual_attack_log',
          }),
        ],
      })
    )
  })
})
