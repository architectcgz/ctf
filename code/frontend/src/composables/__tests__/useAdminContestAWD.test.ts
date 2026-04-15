import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, ref } from 'vue'

import { useAdminContestAWD } from '@/composables/useAdminContestAWD'
import type { ContestDetailData } from '@/api/contracts'
import { ApiError } from '@/api/request'

const adminApiMocks = vi.hoisted(() => ({
  createAdminContestChallenge: vi.fn(),
  createContestAWDAttackLog: vi.fn(),
  createContestAWDRound: vi.fn(),
  createContestAWDServiceCheck: vi.fn(),
  getAdminContestLiveScoreboard: vi.fn(),
  getChallenges: vi.fn(),
  getContestAWDReadiness: vi.fn(),
  getContestAWDRoundSummary: vi.fn(),
  getContestAWDRoundTrafficSummary: vi.fn(),
  listAdminContestChallenges: vi.fn(),
  listContestAWDRoundAttacks: vi.fn(),
  listContestAWDRoundServices: vi.fn(),
  listContestAWDRounds: vi.fn(),
  listContestAWDRoundTrafficEvents: vi.fn(),
  listContestTeams: vi.fn(),
  runContestAWDCurrentRoundCheck: vi.fn(),
  runContestAWDRoundCheck: vi.fn(),
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

describe('useAdminContestAWD', () => {
  beforeEach(() => {
    vi.useRealTimers()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    adminApiMocks.createAdminContestChallenge.mockReset()
    adminApiMocks.createContestAWDAttackLog.mockReset()
    adminApiMocks.createContestAWDRound.mockReset()
    adminApiMocks.createContestAWDServiceCheck.mockReset()
    adminApiMocks.getAdminContestLiveScoreboard.mockReset()
    adminApiMocks.getChallenges.mockReset()
    adminApiMocks.getContestAWDReadiness.mockReset()
    adminApiMocks.getContestAWDRoundSummary.mockReset()
    adminApiMocks.getContestAWDRoundTrafficSummary.mockReset()
    adminApiMocks.listAdminContestChallenges.mockReset()
    adminApiMocks.listContestAWDRoundAttacks.mockReset()
    adminApiMocks.listContestAWDRoundServices.mockReset()
    adminApiMocks.listContestAWDRounds.mockReset()
    adminApiMocks.listContestAWDRoundTrafficEvents.mockReset()
    adminApiMocks.listContestTeams.mockReset()
    adminApiMocks.runContestAWDCurrentRoundCheck.mockReset()
    adminApiMocks.runContestAWDRoundCheck.mockReset()
    adminApiMocks.updateAdminContestChallenge.mockReset()

    adminApiMocks.listContestAWDRounds.mockResolvedValue([])
    adminApiMocks.listContestTeams.mockResolvedValue([])
    adminApiMocks.listAdminContestChallenges.mockResolvedValue([])
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
      page_size: 200,
    })
    adminApiMocks.createContestAWDAttackLog.mockResolvedValue(undefined)
    adminApiMocks.createContestAWDServiceCheck.mockResolvedValue(undefined)
    adminApiMocks.createAdminContestChallenge.mockResolvedValue(undefined)
    adminApiMocks.updateAdminContestChallenge.mockResolvedValue(undefined)
  })

  it('在创建轮次被 readiness 门禁拦截后会拉取摘要并允许 override 重试', async () => {
    adminApiMocks.createContestAWDRound
      .mockRejectedValueOnce(
        new ApiError('开赛就绪门禁阻止了创建轮次', { code: 14025, status: 409 })
      )
      .mockResolvedValueOnce(buildRound())

    let composable!: ReturnType<typeof useAdminContestAWD>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    const Harness = defineComponent({
      setup() {
        composable = useAdminContestAWD(selectedContest)
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

    let composable!: ReturnType<typeof useAdminContestAWD>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    const Harness = defineComponent({
      setup() {
        composable = useAdminContestAWD(selectedContest)
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

    let composable!: ReturnType<typeof useAdminContestAWD>
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    const Harness = defineComponent({
      setup() {
        composable = useAdminContestAWD(selectedContest)
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
})
