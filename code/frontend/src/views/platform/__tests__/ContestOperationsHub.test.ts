import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestOperationsHub from '../ContestOperationsHub.vue'
import contestOperationsHubSource from '../ContestOperationsHub.vue?raw'

const pushMock = vi.fn()
const routeState = vi.hoisted(() => ({
  path: '/platform/contest-ops/environment',
  name: 'PlatformContestOpsEnvironment',
}))
const adminApiMocks = vi.hoisted(() => ({
  getContests: vi.fn(),
  listContestAWDRounds: vi.fn(),
  getContestAWDRoundSummary: vi.fn(),
  getContestAWDRoundTrafficSummary: vi.fn(),
  listContestAWDRoundServices: vi.fn(),
  listContestAWDRoundAttacks: vi.fn(),
  getAdminContestLiveScoreboard: vi.fn(),
}))

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    getContests: adminApiMocks.getContests,
    listContestAWDRounds: adminApiMocks.listContestAWDRounds,
    getContestAWDRoundSummary: adminApiMocks.getContestAWDRoundSummary,
    getContestAWDRoundTrafficSummary: adminApiMocks.getContestAWDRoundTrafficSummary,
    listContestAWDRoundServices: adminApiMocks.listContestAWDRoundServices,
    listContestAWDRoundAttacks: adminApiMocks.listContestAWDRoundAttacks,
    getAdminContestLiveScoreboard: adminApiMocks.getAdminContestLiveScoreboard,
  }
})

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock }),
  }
})

describe('ContestOperationsHub', () => {
  beforeEach(() => {
    pushMock.mockReset()
    adminApiMocks.getContests.mockReset()
    adminApiMocks.listContestAWDRounds.mockReset()
    adminApiMocks.getContestAWDRoundSummary.mockReset()
    adminApiMocks.getContestAWDRoundTrafficSummary.mockReset()
    adminApiMocks.listContestAWDRoundServices.mockReset()
    adminApiMocks.listContestAWDRoundAttacks.mockReset()
    adminApiMocks.getAdminContestLiveScoreboard.mockReset()
    adminApiMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'awd-running',
          title: '2026 AWD 联赛',
          description: '运行中赛事',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-04-15T09:00:00.000Z',
          ends_at: '2026-04-15T18:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    adminApiMocks.listContestAWDRounds.mockResolvedValue([])
    adminApiMocks.getContestAWDRoundSummary.mockResolvedValue({
      round: {
        id: '41',
        contest_id: 'awd-running',
        round_number: 4,
        status: 'running',
        attack_score: 60,
        defense_score: 40,
        created_at: '2026-04-15T10:00:00.000Z',
        updated_at: '2026-04-15T10:15:00.000Z',
      },
      metrics: {
        total_service_count: 12,
        service_up_count: 9,
        service_down_count: 2,
        service_compromised_count: 1,
        attacked_service_count: 4,
        defense_success_count: 8,
        total_attack_count: 18,
        successful_attack_count: 6,
        failed_attack_count: 12,
        scheduler_check_count: 12,
        manual_current_round_check_count: 0,
        manual_selected_round_check_count: 0,
        manual_service_check_count: 0,
        submission_attack_count: 18,
        manual_attack_log_count: 0,
        legacy_attack_log_count: 0,
      },
      items: [
        {
          team_id: '13',
          team_name: 'Red',
          service_up_count: 3,
          service_down_count: 0,
          service_compromised_count: 0,
          sla_score: 54,
          defense_score: 40,
          attack_score: 120,
          successful_attack_count: 2,
          successful_breach_count: 0,
          unique_attackers_against: 1,
          total_score: 214,
        },
      ],
    })
    adminApiMocks.getContestAWDRoundTrafficSummary.mockResolvedValue({
      contest_id: 'awd-running',
      round_id: '41',
      total_request_count: 324,
      active_attacker_team_count: 4,
      victim_team_count: 3,
      unique_path_count: 17,
      error_request_count: 28,
      latest_event_at: '2026-04-15T10:14:00.000Z',
      top_attackers: [{ team_id: '13', team_name: 'Red', request_count: 112, error_count: 8 }],
      top_victims: [{ team_id: '14', team_name: 'Blue', request_count: 96, error_count: 11 }],
      top_challenges: [{ challenge_id: '101', challenge_title: 'Bank Portal', request_count: 160, error_count: 14 }],
      top_paths: [{ path: '/login', request_count: 66, error_count: 3, last_status_code: 200 }],
      top_error_paths: [{ path: '/api/transfer', request_count: 18, error_count: 9, last_status_code: 500 }],
      trend_buckets: [],
    })
    adminApiMocks.listContestAWDRoundServices.mockResolvedValue([
      {
        id: '1',
        round_id: '41',
        team_id: '13',
        team_name: 'Red',
        service_id: '7009',
        challenge_id: '101',
        service_status: 'up',
        checker_type: 'http_standard',
        check_result: {},
        attack_received: 1,
        sla_score: 18,
        defense_score: 40,
        attack_score: 60,
        updated_at: '2026-04-15T10:12:00.000Z',
      },
      {
        id: '2',
        round_id: '41',
        team_id: '14',
        team_name: 'Blue',
        service_id: '7009',
        challenge_id: '101',
        service_status: 'compromised',
        checker_type: 'http_standard',
        check_result: {},
        attack_received: 2,
        sla_score: 0,
        defense_score: 0,
        attack_score: 0,
        updated_at: '2026-04-15T10:13:00.000Z',
      },
    ])
    adminApiMocks.listContestAWDRoundAttacks.mockResolvedValue([
      {
        id: '88',
        round_id: '41',
        attacker_team_id: '13',
        attacker_team: 'Red',
        victim_team_id: '14',
        victim_team: 'Blue',
        service_id: '7009',
        challenge_id: '101',
        attack_type: 'flag_capture',
        source: 'submission',
        submitted_flag: 'flag{demo}',
        is_success: true,
        score_gained: 60,
        created_at: '2026-04-15T10:14:00.000Z',
      },
    ])
    adminApiMocks.getAdminContestLiveScoreboard.mockResolvedValue({
      contest: {
        id: 'awd-running',
        title: '2026 AWD 联赛',
        status: 'running',
        started_at: '2026-04-15T09:00:00.000Z',
        ends_at: '2026-04-15T18:00:00.000Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: '13',
            team_name: 'Red',
            score: 214,
            solved_count: 0,
            last_submission_at: '2026-04-15T10:14:00.000Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })
  })

  it('renders environment management copy and routes to awd config for the preferred contest', async () => {
    routeState.path = '/platform/contest-ops/environment'
    routeState.name = 'PlatformContestOpsEnvironment'

    const wrapper = mount(ContestOperationsHub)
    await flushPromises()

    expect(wrapper.text()).toContain('环境管理')
    expect(wrapper.text()).toContain('2026 AWD 联赛')

    await wrapper.get('#contest-ops-primary-action').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'ContestEdit',
      params: { id: 'awd-running' },
      query: { panel: 'awd-config' },
    })
  })

  it('renders traffic monitoring copy and routes to the operations inspector', async () => {
    routeState.path = '/platform/contest-ops/traffic'
    routeState.name = 'PlatformContestOpsTraffic'

    const wrapper = mount(ContestOperationsHub)
    await flushPromises()

    expect(wrapper.text()).toContain('流量监控')

    await wrapper.get('#contest-ops-primary-action').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'ContestEdit',
      params: { id: 'awd-running' },
      query: { panel: 'operations', opsPanel: 'inspector' },
    })
  })

  it('renders projector telemetry for the preferred contest instead of only jump links', async () => {
    routeState.path = '/platform/contest-ops/projector'
    routeState.name = 'PlatformContestOpsProjector'
    adminApiMocks.listContestAWDRounds.mockResolvedValueOnce([
      {
        id: '41',
        contest_id: 'awd-running',
        round_number: 4,
        status: 'running',
        attack_score: 60,
        defense_score: 40,
        created_at: '2026-04-15T10:00:00.000Z',
        updated_at: '2026-04-15T10:15:00.000Z',
      },
    ])

    const wrapper = mount(ContestOperationsHub)
    await flushPromises()

    expect(adminApiMocks.listContestAWDRounds).toHaveBeenCalledWith('awd-running')
    expect(adminApiMocks.getContestAWDRoundSummary).toHaveBeenCalledWith('awd-running', '41')
    expect(adminApiMocks.getContestAWDRoundTrafficSummary).toHaveBeenCalledWith('awd-running', '41')
    expect(adminApiMocks.listContestAWDRoundServices).toHaveBeenCalledWith('awd-running', '41')
    expect(adminApiMocks.listContestAWDRoundAttacks).toHaveBeenCalledWith('awd-running', '41')
    expect(adminApiMocks.getAdminContestLiveScoreboard).toHaveBeenCalledWith('awd-running', {
      page: 1,
      page_size: 10,
    })
    expect(wrapper.text()).toContain('当前轮次')
    expect(wrapper.text()).toContain('实时榜单')
    expect(wrapper.text()).toContain('最新攻击')
    expect(wrapper.text()).toContain('Round 4')
    expect(wrapper.text()).toContain('Red')
    expect(wrapper.text()).toContain('Blue')
    expect(wrapper.text()).toContain('Bank Portal')
  })

  it('uses shared list heading and metric panel primitives for the operations workspace header', () => {
    expect(contestOperationsHubSource).toContain(
      '<header class="list-heading contest-ops-hero workspace-directory-section">'
    )
    expect(contestOperationsHubSource).not.toContain(
      '<header class="contest-ops-hero workspace-directory-section">'
    )
    expect(contestOperationsHubSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface contest-ops-summary"'
    )
    expect(contestOperationsHubSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(contestOperationsHubSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(contestOperationsHubSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(contestOperationsHubSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
  })
})
