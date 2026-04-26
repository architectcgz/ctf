import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, ref } from 'vue'

import { usePlatformContestAwd } from '@/composables/usePlatformContestAwd'
import type { ContestDetailData } from '@/api/contracts'
import { ApiError } from '@/api/request'

const adminApiMocks = vi.hoisted(() => ({
  createAdminContestChallenge: vi.fn(),
  createContestAWDService: vi.fn(),
  createContestAWDAttackLog: vi.fn(),
  createContestAWDRound: vi.fn(),
  createContestAWDServiceCheck: vi.fn(),
  getAdminContestLiveScoreboard: vi.fn(),
  getChallenges: vi.fn(),
  getContestAWDInstanceOrchestration: vi.fn(),
  getContestAWDReadiness: vi.fn(),
  getContestAWDRoundSummary: vi.fn(),
  getContestAWDRoundTrafficSummary: vi.fn(),
  listAdminContestChallenges: vi.fn(),
  listContestAWDServices: vi.fn(),
  listContestAWDRoundAttacks: vi.fn(),
  listContestAWDRoundServices: vi.fn(),
  listContestAWDRounds: vi.fn(),
  listContestAWDRoundTrafficEvents: vi.fn(),
  listContestTeams: vi.fn(),
  runContestAWDCurrentRoundCheck: vi.fn(),
  runContestAWDRoundCheck: vi.fn(),
  startContestAWDTeamServiceInstance: vi.fn(),
  updateContestAWDService: vi.fn(),
  updateAdminContestChallenge: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

vi.mock('@/api/admin', () => adminApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

function buildContest(overrides: Partial<ContestDetailData> = {}): ContestDetailData {
  return {
    id: 'awd-1',
    title: '2026 AWD 联赛',
    description: '攻防赛',
    mode: 'awd',
    status: 'running',
    starts_at: '2026-04-12T09:00:00.000Z',
    ends_at: '2026-04-12T18:00:00.000Z',
    scoreboard_frozen: false,
    ...overrides,
  }
}

function buildReadiness() {
  return {
    contest_id: 'awd-1',
    ready: false,
    total_challenges: 1,
    passed_challenges: 0,
    pending_challenges: 0,
    failed_challenges: 1,
    stale_challenges: 0,
    missing_checker_challenges: 0,
    blocking_count: 1,
    global_blocking_reasons: [],
    blocking_actions: ['create_round', 'run_current_round_check'],
    items: [
      {
        challenge_id: 'challenge-1',
        title: 'Web Checker',
        checker_type: 'http_standard' as const,
        validation_state: 'failed' as const,
        last_preview_at: '2026-04-12T08:00:00.000Z',
        last_access_url: 'http://checker.internal/flag',
        blocking_reason: 'last_preview_failed' as const,
      },
    ],
  }
}

function buildRound(overrides: Record<string, unknown> = {}) {
  return {
    id: 'round-1',
    contest_id: 'awd-1',
    round_number: 1,
    status: 'pending',
    attack_score: 50,
    defense_score: 50,
    created_at: '2026-04-12T09:00:00.000Z',
    updated_at: '2026-04-12T09:00:00.000Z',
    ...overrides,
  }
}

function buildContestAWDService(overrides: Record<string, unknown> = {}) {
  return {
    id: 'service-1',
    contest_id: 'awd-1',
    challenge_id: 'challenge-1',
    template_id: 'template-1',
    display_name: 'Bank Portal',
    order: 1,
    is_visible: true,
    score_config: {
      points: 100,
      awd_sla_score: 20,
      awd_defense_score: 30,
    },
    runtime_config: {
      checker_type: 'http_standard',
      checker_config: {
        get_flag: {
          path: '/flag',
        },
      },
    },
    checker_type: 'http_standard',
    checker_config: {
      get_flag: {
        path: '/flag',
      },
    },
    sla_score: 20,
    defense_score: 30,
    validation_state: 'stale',
    last_preview_at: '2026-04-12T08:00:00.000Z',
    last_preview_result: {
      service_status: 'up',
      check_result: {
        status_code: 200,
      },
      preview_context: {
        access_url: 'http://preview.internal/flag',
        preview_flag: 'FLAG{preview}',
        round_number: 0,
        team_id: 0,
        challenge_id: 0,
      },
    },
    created_at: '2026-04-12T08:00:00.000Z',
    updated_at: '2026-04-12T08:05:00.000Z',
    ...overrides,
  }
}

describe('usePlatformContestAwd', () => {
  beforeEach(() => {
    vi.useRealTimers()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    adminApiMocks.createAdminContestChallenge.mockReset()
    adminApiMocks.createContestAWDService.mockReset()
    adminApiMocks.createContestAWDAttackLog.mockReset()
    adminApiMocks.createContestAWDRound.mockReset()
    adminApiMocks.createContestAWDServiceCheck.mockReset()
    adminApiMocks.getAdminContestLiveScoreboard.mockReset()
    adminApiMocks.getChallenges.mockReset()
    adminApiMocks.getContestAWDInstanceOrchestration.mockReset()
    adminApiMocks.getContestAWDReadiness.mockReset()
    adminApiMocks.getContestAWDRoundSummary.mockReset()
    adminApiMocks.getContestAWDRoundTrafficSummary.mockReset()
    adminApiMocks.listAdminContestChallenges.mockReset()
    adminApiMocks.listContestAWDServices.mockReset()
    adminApiMocks.listContestAWDRoundAttacks.mockReset()
    adminApiMocks.listContestAWDRoundServices.mockReset()
    adminApiMocks.listContestAWDRounds.mockReset()
    adminApiMocks.listContestAWDRoundTrafficEvents.mockReset()
    adminApiMocks.listContestTeams.mockReset()
    adminApiMocks.runContestAWDCurrentRoundCheck.mockReset()
    adminApiMocks.runContestAWDRoundCheck.mockReset()
    adminApiMocks.startContestAWDTeamServiceInstance.mockReset()
    adminApiMocks.updateContestAWDService.mockReset()
    adminApiMocks.updateAdminContestChallenge.mockReset()

    adminApiMocks.listContestAWDRounds.mockResolvedValue([])
    adminApiMocks.getContestAWDInstanceOrchestration.mockResolvedValue({
      contest_id: 'awd-1',
      teams: [],
      services: [],
      instances: [],
    })
    adminApiMocks.listContestTeams.mockResolvedValue([])
    adminApiMocks.listAdminContestChallenges.mockResolvedValue([])
    adminApiMocks.listContestAWDServices.mockResolvedValue([])
    adminApiMocks.getContestAWDReadiness.mockResolvedValue(buildReadiness())
    adminApiMocks.listContestAWDRoundServices.mockResolvedValue([])
    adminApiMocks.listContestAWDRoundAttacks.mockResolvedValue([])
    adminApiMocks.getContestAWDRoundSummary.mockResolvedValue({
      round: buildRound(),
      items: [],
    })
    adminApiMocks.getContestAWDRoundTrafficSummary.mockResolvedValue({
      contest_id: 'awd-1',
      round_id: 'round-1',
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
    adminApiMocks.listContestAWDRoundTrafficEvents.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })
    adminApiMocks.getAdminContestLiveScoreboard.mockResolvedValue({
      contest: {
        id: 'awd-1',
        title: '2026 AWD 联赛',
        status: 'running',
        started_at: '2026-04-12T09:00:00.000Z',
        ends_at: '2026-04-12T18:00:00.000Z',
      },
      scoreboard: {
        list: [],
        total: 0,
        page: 1,
        page_size: 10,
      },
      frozen: false,
      current_team_id: undefined,
    })
    adminApiMocks.getChallenges.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 100,
    })
    adminApiMocks.createContestAWDAttackLog.mockResolvedValue(undefined)
    adminApiMocks.createContestAWDServiceCheck.mockResolvedValue(undefined)
    adminApiMocks.createAdminContestChallenge.mockResolvedValue(undefined)
    adminApiMocks.createContestAWDService.mockResolvedValue(undefined)
    adminApiMocks.updateContestAWDService.mockResolvedValue(undefined)
    adminApiMocks.startContestAWDTeamServiceInstance.mockResolvedValue({
      team_id: 'team-1',
      service_id: 'service-1',
      instance: undefined,
    })
    adminApiMocks.updateAdminContestChallenge.mockResolvedValue(undefined)
  })

  it('初次加载时应从 AWD service 列表合并 checker 与验证信息', async () => {
    adminApiMocks.listContestAWDServices.mockResolvedValue([
      buildContestAWDService({
        challenge_id: 'challenge-1',
        template_id: 'template-9',
        order: 2,
        title: 'Web Checker',
        category: 'web',
        difficulty: 'medium',
      }),
    ])

    let composable!: ReturnType<typeof usePlatformContestAwd>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    const Harness = defineComponent({
      setup() {
        composable = usePlatformContestAwd(selectedContest)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()
    await composable.loadChallengeCatalog()
    await flushPromises()

    expect(adminApiMocks.listContestAWDServices).toHaveBeenCalledWith('awd-1')
    expect(composable.challengeLinks.value).toEqual([
      expect.objectContaining({
        challenge_id: 'challenge-1',
        awd_service_id: 'service-1',
        awd_template_id: 'template-9',
        awd_service_display_name: 'Bank Portal',
        order: 2,
        is_visible: true,
        awd_checker_type: 'http_standard',
        awd_checker_config: {
          get_flag: {
            path: '/flag',
          },
        },
        awd_sla_score: 20,
        awd_defense_score: 30,
        awd_checker_validation_state: 'stale',
        awd_checker_last_preview_at: '2026-04-12T08:00:00.000Z',
        awd_checker_last_preview_result: expect.objectContaining({
          service_status: 'up',
        }),
      }),
    ])

    wrapper.unmount()
  })

  it('应用流量筛选时应提交 service_id 参数', async () => {
    adminApiMocks.listContestAWDRounds.mockResolvedValue([buildRound({ status: 'running' })])
    adminApiMocks.listContestAWDRoundTrafficEvents.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    let composable!: ReturnType<typeof usePlatformContestAwd>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    const Harness = defineComponent({
      setup() {
        composable = usePlatformContestAwd(selectedContest)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    adminApiMocks.listContestAWDRoundTrafficEvents.mockClear()

    await composable.applyTrafficFilters({
      service_id: '7009',
    })
    await flushPromises()

    expect(adminApiMocks.listContestAWDRoundTrafficEvents).toHaveBeenCalledWith('awd-1', 'round-1', {
      page: 1,
      page_size: 20,
      service_id: '7009',
      attacker_team_id: undefined,
      victim_team_id: undefined,
      challenge_id: undefined,
      status_group: undefined,
      path_keyword: undefined,
    })

    wrapper.unmount()
  })

  it('合并管理侧 AWD 题目视图时应只信任 service 顶层配置字段', async () => {
    adminApiMocks.listAdminContestChallenges.mockResolvedValue([
      {
        id: 'link-1',
        contest_id: 'awd-1',
        challenge_id: 'challenge-1',
        title: 'Bank Portal',
        category: 'web',
        difficulty: 'medium',
        points: 100,
        order: 9,
        is_visible: false,
        created_at: '2026-04-12T07:50:00.000Z',
      },
    ])
    adminApiMocks.listContestAWDServices.mockResolvedValue([
      buildContestAWDService({
        challenge_id: 'challenge-1',
        template_id: 'template-9',
        order: 2,
        checker_type: undefined,
        checker_config: undefined,
        sla_score: undefined,
        defense_score: undefined,
        runtime_config: {
          checker_type: 'http_standard',
          checker_config: {
            get_flag: {
              path: '/runtime-only',
            },
          },
        },
        score_config: {
          points: 100,
          awd_sla_score: 99,
          awd_defense_score: 88,
        },
      }),
    ])

    let composable!: ReturnType<typeof usePlatformContestAwd>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    const Harness = defineComponent({
      setup() {
        composable = usePlatformContestAwd(selectedContest)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(composable.challengeLinks.value).toEqual([
      expect.objectContaining({
        challenge_id: 'challenge-1',
        awd_service_id: 'service-1',
        awd_template_id: 'template-9',
        order: 2,
        is_visible: true,
        awd_checker_type: undefined,
        awd_checker_config: {},
        awd_sla_score: 0,
        awd_defense_score: 0,
      }),
    ])

    wrapper.unmount()
  })

  it('在创建轮次被 readiness 门禁拦截后会拉取摘要并允许 override 重试', async () => {
    adminApiMocks.createContestAWDRound
      .mockRejectedValueOnce(
        new ApiError('开赛就绪门禁阻止了创建轮次', { code: 14025, status: 409 })
      )
      .mockResolvedValueOnce(buildRound())

    let composable!: ReturnType<typeof usePlatformContestAwd>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    const Harness = defineComponent({
      setup() {
        composable = usePlatformContestAwd(selectedContest)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    adminApiMocks.getContestAWDReadiness.mockClear()

    await composable.createRound({
      round_number: 1,
      status: 'pending',
      attack_score: 50,
      defense_score: 50,
    })
    await flushPromises()

    expect(adminApiMocks.createContestAWDRound).toHaveBeenCalledWith('awd-1', {
      round_number: 1,
      status: 'pending',
      attack_score: 50,
      defense_score: 50,
    })
    expect(adminApiMocks.getContestAWDReadiness).toHaveBeenCalledWith('awd-1')
    expect(composable.overrideDialogState.value).toMatchObject({
      open: true,
      action: 'create_round',
      title: '创建轮次',
    })

    await composable.confirmOverrideAction(' teacher drill ')
    await flushPromises()

    expect(adminApiMocks.createContestAWDRound).toHaveBeenLastCalledWith('awd-1', {
      round_number: 1,
      status: 'pending',
      attack_score: 50,
      defense_score: 50,
      force_override: true,
      override_reason: 'teacher drill',
    })
    expect(composable.overrideDialogState.value.open).toBe(false)

    wrapper.unmount()
  })

  it('在当前轮巡检被 readiness 门禁拦截后会拉取摘要并允许 override 重试', async () => {
    adminApiMocks.runContestAWDCurrentRoundCheck
      .mockRejectedValueOnce(new ApiError('开赛就绪门禁阻止了巡检', { code: 14025, status: 409 }))
      .mockResolvedValueOnce({
        round: buildRound({ status: 'running' }),
        services: [],
      })

    let composable!: ReturnType<typeof usePlatformContestAwd>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    const Harness = defineComponent({
      setup() {
        composable = usePlatformContestAwd(selectedContest)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    adminApiMocks.getContestAWDReadiness.mockClear()

    await composable.runSelectedRoundCheck()
    await flushPromises()

    expect(adminApiMocks.runContestAWDCurrentRoundCheck).toHaveBeenCalledWith('awd-1')
    expect(adminApiMocks.getContestAWDReadiness).toHaveBeenCalledWith('awd-1')
    expect(composable.overrideDialogState.value).toMatchObject({
      open: true,
      action: 'run_current_round_check',
      title: '立即巡检当前轮',
    })

    await composable.confirmOverrideAction(' emergency drill ')
    await flushPromises()

    expect(adminApiMocks.runContestAWDCurrentRoundCheck).toHaveBeenLastCalledWith('awd-1', {
      force_override: true,
      override_reason: 'emergency drill',
    })
    expect(composable.overrideDialogState.value.open).toBe(false)

    wrapper.unmount()
  })

  it('遇到非 readiness 错误时不会误打开 override 弹层', async () => {
    adminApiMocks.createContestAWDRound.mockRejectedValueOnce(
      new ApiError('普通冲突', { code: 14099, status: 409 })
    )

    let composable!: ReturnType<typeof usePlatformContestAwd>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    const Harness = defineComponent({
      setup() {
        composable = usePlatformContestAwd(selectedContest)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    adminApiMocks.getContestAWDReadiness.mockClear()

    await composable.createRound({
      round_number: 1,
      status: 'pending',
      attack_score: 50,
      defense_score: 50,
    })
    await flushPromises()

    expect(adminApiMocks.getContestAWDReadiness).not.toHaveBeenCalled()
    expect(composable.overrideDialogState.value.open).toBe(false)
    expect(toastMocks.error).toHaveBeenCalledWith('普通冲突')

    wrapper.unmount()
  })

  it('录入服务检查时应提交 service_id 载荷', async () => {
    let composable!: ReturnType<typeof usePlatformContestAwd>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    adminApiMocks.listContestAWDRounds.mockResolvedValueOnce([buildRound()])

    const Harness = defineComponent({
      setup() {
        composable = usePlatformContestAwd(selectedContest)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    await composable.createServiceCheck({
      team_id: 12,
      service_id: 7009,
      service_status: 'up',
      check_result: { latency_ms: 38 },
    })
    await flushPromises()

    expect(adminApiMocks.createContestAWDServiceCheck).toHaveBeenCalledWith('awd-1', 'round-1', {
      team_id: 12,
      service_id: 7009,
      service_status: 'up',
      check_result: { latency_ms: 38 },
    })

    wrapper.unmount()
  })

  it('补录攻击日志时应提交 service_id 载荷', async () => {
    let composable!: ReturnType<typeof usePlatformContestAwd>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    adminApiMocks.listContestAWDRounds.mockResolvedValueOnce([buildRound()])

    const Harness = defineComponent({
      setup() {
        composable = usePlatformContestAwd(selectedContest)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    await composable.createAttackLog({
      attacker_team_id: 12,
      victim_team_id: 13,
      service_id: 7009,
      attack_type: 'flag_capture',
      submitted_flag: 'flag{demo}',
      is_success: true,
    })
    await flushPromises()

    expect(adminApiMocks.createContestAWDAttackLog).toHaveBeenCalledWith('awd-1', 'round-1', {
      attacker_team_id: 12,
      victim_team_id: 13,
      service_id: 7009,
      attack_type: 'flag_capture',
      submitted_flag: 'flag{demo}',
      is_success: true,
    })

    wrapper.unmount()
  })

  it('创建 AWD 配置时应通过显式 service 写入 runtime 字段，关系层只更新分值', async () => {
    let composable!: ReturnType<typeof usePlatformContestAwd>
    const selectedContest = ref<ContestDetailData | null>(buildContest())

    const Harness = defineComponent({
      setup() {
        composable = usePlatformContestAwd(selectedContest)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    await composable.createChallengeLink({
      challenge_id: 101,
      template_id: 5,
      points: 120,
      order: 2,
      is_visible: true,
      awd_checker_type: 'http_standard',
      awd_checker_config: { put_flag: { path: '/flag' } },
      awd_sla_score: 20,
      awd_defense_score: 30,
      awd_checker_preview_token: 'preview-token',
    })
    await flushPromises()

    expect(adminApiMocks.createContestAWDService).toHaveBeenCalledWith('awd-1', {
      template_id: 5,
      points: 120,
      order: 2,
      is_visible: true,
      checker_type: 'http_standard',
      checker_config: { put_flag: { path: '/flag' } },
      awd_sla_score: 20,
      awd_defense_score: 30,
      awd_checker_preview_token: 'preview-token',
    })
    expect(adminApiMocks.updateAdminContestChallenge).not.toHaveBeenCalled()

    wrapper.unmount()
  })

  it('更新 AWD 配置时应优先更新显式 service，关系层只更新分值', async () => {
    let composable!: ReturnType<typeof usePlatformContestAwd>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    adminApiMocks.listContestAWDServices.mockResolvedValueOnce([
      buildContestAWDService({
        challenge_id: '101',
        template_id: '4',
        title: 'Web Checker',
        category: 'web',
        difficulty: 'medium',
      }),
    ])

    const Harness = defineComponent({
      setup() {
        composable = usePlatformContestAwd(selectedContest)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    await composable.updateChallengeLink('101', {
      template_id: 6,
      points: 150,
      order: 3,
      is_visible: false,
      awd_checker_type: 'http_standard',
      awd_checker_config: { get_flag: { path: '/flag' } },
      awd_sla_score: 25,
      awd_defense_score: 35,
      awd_checker_preview_token: 'preview-token-2',
    })
    await flushPromises()

    expect(adminApiMocks.updateContestAWDService).toHaveBeenCalledWith('awd-1', 'service-1', {
      template_id: 6,
      points: 150,
      order: 3,
      is_visible: false,
      checker_type: 'http_standard',
      checker_config: { get_flag: { path: '/flag' } },
      awd_sla_score: 25,
      awd_defense_score: 35,
      awd_checker_preview_token: 'preview-token-2',
    })
    expect(adminApiMocks.updateAdminContestChallenge).not.toHaveBeenCalled()

    wrapper.unmount()
  })
})
