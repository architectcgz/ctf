import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestEdit from '../ContestEdit.vue'
import AWDReadinessOverrideDialog from '@/components/admin/contest/AWDReadinessOverrideDialog.vue'
import { ApiError } from '@/api/request'
import type { ContestDetailData } from '@/api/contracts'
import type { VueWrapper } from '@vue/test-utils'

const pushMock = vi.fn()
const routeState = vi.hoisted(() => ({
  params: { id: 'contest-1' } as Record<string, string>,
}))

const contestApiMocks = vi.hoisted(() => ({
  getContest: vi.fn(),
  updateContest: vi.fn(),
  getContestAWDReadiness: vi.fn(),
  listAdminContestChallenges: vi.fn(),
  getChallenges: vi.fn(),
  createAdminContestChallenge: vi.fn(),
  updateAdminContestChallenge: vi.fn(),
  deleteAdminContestChallenge: vi.fn(),
}))

const destructiveConfirmMock = vi.hoisted(() => vi.fn())
const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
}))
const awdMockModule = vi.hoisted(() => ({
  state: null as any,
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: vi.fn(), back: vi.fn() }),
  }
})

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    getContest: contestApiMocks.getContest,
    updateContest: contestApiMocks.updateContest,
    getContestAWDReadiness: contestApiMocks.getContestAWDReadiness,
    listAdminContestChallenges: contestApiMocks.listAdminContestChallenges,
    getChallenges: contestApiMocks.getChallenges,
    createAdminContestChallenge: contestApiMocks.createAdminContestChallenge,
    updateAdminContestChallenge: contestApiMocks.updateAdminContestChallenge,
    deleteAdminContestChallenge: contestApiMocks.deleteAdminContestChallenge,
  }
})

vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: destructiveConfirmMock,
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

vi.mock('@/composables/useAdminContestAWD', async () => {
  const { ref } = await vi.importActual<typeof import('vue')>('vue')
  awdMockModule.state = {
    rounds: ref([
      {
        id: 'round-1',
        contest_id: 'contest-1',
        round_number: 1,
        status: 'running',
        attack_score: 50,
        defense_score: 50,
        created_at: '2026-03-15T09:00:00.000Z',
        updated_at: '2026-03-15T09:05:00.000Z',
      },
    ]),
    selectedRoundId: ref('round-1'),
    readiness: ref(null),
    loadingReadiness: ref(false),
    overrideDialogState: ref({
      open: false,
      action: null,
      title: '',
      reason: '',
      readiness: null,
      confirmLoading: false,
    }),
    services: ref([]),
    attacks: ref([]),
    summary: ref(null),
    trafficSummary: ref(null),
    trafficEvents: ref([]),
    trafficEventsTotal: ref(0),
    trafficFilters: ref({
      attacker_team_id: '',
      victim_team_id: '',
      challenge_id: '',
      status_group: 'all',
      path_keyword: '',
      page: 1,
      page_size: 20,
    }),
    scoreboardRows: ref([]),
    scoreboardFrozen: ref(false),
    teams: ref([
      {
        id: 'team-1',
        contest_id: 'contest-1',
        name: '蓝队一',
        captain_id: '1001',
        invite_code: 'ABC123',
        max_members: 4,
        member_count: 3,
        created_at: '2026-03-15T08:00:00.000Z',
      },
      {
        id: 'team-2',
        contest_id: 'contest-1',
        name: '红队一',
        captain_id: '1002',
        invite_code: 'DEF456',
        max_members: 4,
        member_count: 3,
        created_at: '2026-03-15T08:01:00.000Z',
      },
    ]),
    challengeLinks: ref([
      {
        id: 'link-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        title: 'Web 入门',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 1,
        is_visible: true,
        awd_checker_type: undefined,
        awd_checker_config: {},
        awd_sla_score: 0,
        awd_defense_score: 0,
        awd_checker_validation_state: 'pending',
        awd_checker_last_preview_at: undefined,
        awd_checker_last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
      },
    ]),
    challengeCatalog: ref([]),
    loadingRounds: ref(false),
    loadingRoundDetail: ref(false),
    loadingTrafficSummary: ref(false),
    loadingTrafficEvents: ref(false),
    loadingChallengeCatalog: ref(false),
    checking: ref(false),
    creatingRound: ref(false),
    savingServiceCheck: ref(false),
    savingAttackLog: ref(false),
    savingChallengeConfig: ref(false),
    shouldAutoRefresh: ref(false),
    refresh: vi.fn(),
    applyTrafficFilters: vi.fn(),
    setTrafficPage: vi.fn(),
    resetTrafficFilters: vi.fn(),
    runSelectedRoundCheck: vi.fn(),
    createRound: vi.fn(),
    confirmOverrideAction: vi.fn(),
    closeOverrideDialog: vi.fn(),
    createServiceCheck: vi.fn(),
    createAttackLog: vi.fn(),
    loadChallengeCatalog: vi.fn(),
    createChallengeLink: vi.fn(),
    updateChallengeLink: vi.fn(),
  }
  return {
    useAdminContestAWD: () => awdMockModule.state,
  }
})

function buildContestDetail(overrides: Partial<ContestDetailData> = {}): ContestDetailData {
  return {
    id: 'contest-1',
    title: '2026 春季校园 CTF',
    description: '校内赛',
    mode: 'jeopardy',
    status: 'registering',
    starts_at: '2026-03-15T09:00:00.000Z',
    ends_at: '2026-03-15T13:00:00.000Z',
    ...overrides,
  }
}

function mountContestEdit() {
  return mount(ContestEdit, {
    global: {
      stubs: {
        ElDialog: {
          props: ['modelValue', 'title'],
          template: '<div><div v-if="title">{{ title }}</div><slot /><slot name="footer" /></div>',
        },
      },
    },
  })
}

function createDeferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

function getWorkbenchStageRail(wrapper: VueWrapper<any>) {
  return wrapper.get('[role="tablist"][aria-label="竞赛工作台阶段切换"]')
}

describe('ContestEdit', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/admin/contests/contest-1/edit')
    pushMock.mockReset()
    contestApiMocks.getContest.mockReset()
    contestApiMocks.updateContest.mockReset()
    contestApiMocks.getContestAWDReadiness.mockReset()
    contestApiMocks.listAdminContestChallenges.mockReset()
    contestApiMocks.getChallenges.mockReset()
    contestApiMocks.createAdminContestChallenge.mockReset()
    contestApiMocks.updateAdminContestChallenge.mockReset()
    contestApiMocks.deleteAdminContestChallenge.mockReset()
    destructiveConfirmMock.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()
    toastMocks.info.mockReset()
    awdMockModule.state.rounds.value = [
      {
        id: 'round-1',
        contest_id: 'contest-1',
        round_number: 1,
        status: 'running',
        attack_score: 50,
        defense_score: 50,
        created_at: '2026-03-15T09:00:00.000Z',
        updated_at: '2026-03-15T09:05:00.000Z',
      },
    ]
    awdMockModule.state.selectedRoundId.value = 'round-1'
    awdMockModule.state.readiness.value = null
    awdMockModule.state.loadingReadiness.value = false
    awdMockModule.state.overrideDialogState.value = {
      open: false,
      action: null,
      title: '',
      reason: '',
      readiness: null,
      confirmLoading: false,
    }
    awdMockModule.state.services.value = []
    awdMockModule.state.attacks.value = []
    awdMockModule.state.summary.value = null
    awdMockModule.state.trafficSummary.value = null
    awdMockModule.state.trafficEvents.value = []
    awdMockModule.state.trafficEventsTotal.value = 0
    awdMockModule.state.trafficFilters.value = {
      attacker_team_id: '',
      victim_team_id: '',
      challenge_id: '',
      status_group: 'all',
      path_keyword: '',
      page: 1,
      page_size: 20,
    }
    awdMockModule.state.scoreboardRows.value = []
    awdMockModule.state.scoreboardFrozen.value = false
    awdMockModule.state.teams.value = [
      {
        id: 'team-1',
        contest_id: 'contest-1',
        name: '蓝队一',
        captain_id: '1001',
        invite_code: 'ABC123',
        max_members: 4,
        member_count: 3,
        created_at: '2026-03-15T08:00:00.000Z',
      },
      {
        id: 'team-2',
        contest_id: 'contest-1',
        name: '红队一',
        captain_id: '1002',
        invite_code: 'DEF456',
        max_members: 4,
        member_count: 3,
        created_at: '2026-03-15T08:01:00.000Z',
      },
    ]
    awdMockModule.state.challengeLinks.value = [
      {
        id: 'link-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        title: 'Web 入门',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 1,
        is_visible: true,
        awd_checker_type: undefined,
        awd_checker_config: {},
        awd_sla_score: 0,
        awd_defense_score: 0,
        awd_checker_validation_state: 'pending',
        awd_checker_last_preview_at: undefined,
        awd_checker_last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
      },
    ]
    awdMockModule.state.challengeCatalog.value = []
    awdMockModule.state.loadingRounds.value = false
    awdMockModule.state.loadingRoundDetail.value = false
    awdMockModule.state.loadingTrafficSummary.value = false
    awdMockModule.state.loadingTrafficEvents.value = false
    awdMockModule.state.loadingChallengeCatalog.value = false
    awdMockModule.state.checking.value = false
    awdMockModule.state.creatingRound.value = false
    awdMockModule.state.savingServiceCheck.value = false
    awdMockModule.state.savingAttackLog.value = false
    awdMockModule.state.savingChallengeConfig.value = false
    awdMockModule.state.shouldAutoRefresh.value = false
    awdMockModule.state.refresh.mockReset()
    awdMockModule.state.applyTrafficFilters.mockReset()
    awdMockModule.state.setTrafficPage.mockReset()
    awdMockModule.state.resetTrafficFilters.mockReset()
    awdMockModule.state.runSelectedRoundCheck.mockReset()
    awdMockModule.state.createRound.mockReset()
    awdMockModule.state.confirmOverrideAction.mockReset()
    awdMockModule.state.closeOverrideDialog.mockReset()
    awdMockModule.state.createServiceCheck.mockReset()
    awdMockModule.state.createAttackLog.mockReset()
    awdMockModule.state.loadChallengeCatalog.mockReset()
    awdMockModule.state.createChallengeLink.mockReset()
    awdMockModule.state.updateChallengeLink.mockReset()
    routeState.params = { id: 'contest-1' }

    contestApiMocks.getContest.mockResolvedValue({
      id: 'contest-1',
      title: '2026 春季校园 CTF',
      description: '校内赛',
      mode: 'jeopardy',
      status: 'registering',
      starts_at: '2026-03-15T09:00:00.000Z',
      ends_at: '2026-03-15T13:00:00.000Z',
    })
    contestApiMocks.updateContest.mockResolvedValue({
      contest: {
        id: 'contest-1',
        title: '2026 春季校园 CTF（更新）',
        description: '校内赛',
        mode: 'jeopardy',
        status: 'registering',
        starts_at: '2026-03-15T09:00:00.000Z',
        ends_at: '2026-03-15T13:00:00.000Z',
      },
    })
    contestApiMocks.getContestAWDReadiness.mockResolvedValue({
      contest_id: 'contest-1',
      ready: false,
      total_challenges: 1,
      passed_challenges: 0,
      pending_challenges: 0,
      failed_challenges: 1,
      stale_challenges: 0,
      missing_checker_challenges: 0,
      blocking_count: 1,
      global_blocking_reasons: [],
      blocking_actions: ['start_contest'],
      items: [
        {
          challenge_id: '101',
          title: 'Challenge 101',
          checker_type: 'http_standard',
          validation_state: 'failed',
          last_preview_at: '2026-04-12T08:00:00.000Z',
          last_access_url: 'http://checker.internal/flag',
          blocking_reason: 'last_preview_failed',
        },
      ],
    })
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      {
        id: 'link-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        title: 'Web 入门',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 1,
        is_visible: true,
        awd_checker_type: undefined,
        awd_checker_config: {},
        awd_sla_score: 0,
        awd_defense_score: 0,
        awd_checker_validation_state: 'pending',
        awd_checker_last_preview_at: undefined,
        awd_checker_last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
      },
    ])
    contestApiMocks.getChallenges.mockResolvedValue({
      list: [
        {
          id: '101',
          title: 'Web 入门',
          description: '现有题目',
          category: 'web',
          difficulty: 'easy',
          points: 120,
          instance_sharing: 'per_user',
          created_by: '9',
          image_id: undefined,
          attachment_url: undefined,
          hints: undefined,
          status: 'published',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
          flag_config: undefined,
        },
        {
          id: '102',
          title: 'Crypto 进阶',
          description: '新增题目',
          category: 'crypto',
          difficulty: 'medium',
          points: 150,
          instance_sharing: 'per_user',
          created_by: '9',
          image_id: undefined,
          attachment_url: undefined,
          hints: undefined,
          status: 'published',
          created_at: '2026-03-02T00:00:00.000Z',
          updated_at: '2026-03-02T00:00:00.000Z',
          flag_config: undefined,
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })
    contestApiMocks.createAdminContestChallenge.mockResolvedValue({
      id: 'link-2',
      contest_id: 'contest-1',
      challenge_id: '102',
      title: 'Crypto 进阶',
      category: 'crypto',
      difficulty: 'medium',
      points: 160,
      order: 3,
      is_visible: false,
      awd_checker_type: undefined,
      awd_checker_config: {},
      awd_sla_score: 0,
      awd_defense_score: 0,
      awd_checker_validation_state: 'pending',
      awd_checker_last_preview_at: undefined,
      awd_checker_last_preview_result: undefined,
      created_at: '2026-03-10T01:00:00.000Z',
    })
    contestApiMocks.updateAdminContestChallenge.mockResolvedValue(undefined)
    contestApiMocks.deleteAdminContestChallenge.mockResolvedValue(undefined)
    destructiveConfirmMock.mockResolvedValue(true)
  })

  it('应该在普通赛下只展示基础信息与题目池阶段', async () => {
    contestApiMocks.getContest.mockResolvedValue(buildContestDetail())

    const wrapper = mount(ContestEdit, {
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
        },
      },
    })

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.text()).toContain('基础信息')
    expect(stageRail.text()).toContain('题目池')
    expect(stageRail.text()).not.toContain('AWD 配置')
    expect(stageRail.text()).not.toContain('赛前检查')
    expect(stageRail.text()).not.toContain('轮次运行')
  })

  it('应该在 AWD 赛事下展示基础信息、题目池、AWD 配置、赛前检查与轮次运行', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )

    const wrapper = mount(ContestEdit, {
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
        },
      },
    })

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.text()).toContain('基础信息')
    expect(stageRail.text()).toContain('题目池')
    expect(stageRail.text()).toContain('AWD 配置')
    expect(stageRail.text()).toContain('赛前检查')
    expect(stageRail.text()).toContain('轮次运行')
  })

  it('应该在赛前检查中列出阻塞项、保留强制开赛入口，并支持返回 AWD 配置后高亮当前题', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-preflight').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('可强制开赛')
    expect(wrapper.text()).toContain('Challenge 101')
    expect(wrapper.text()).toContain('返回 AWD 配置')
    expect(wrapper.text()).toContain('强制开赛')

    await wrapper.get('#awd-readiness-edit-101').trigger('click')
    await flushPromises()

    expect(getWorkbenchStageRail(wrapper).get('[role="tab"][aria-selected="true"]').text()).toContain('AWD 配置')
    expect(wrapper.text()).toContain('当前聚焦题目')
    expect(wrapper.text()).toContain('Challenge 101')
  })

  it('应该支持从题目池跳转到 AWD 配置并保留当前题高亮', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-open-awd-config-link-1').trigger('click')
    await flushPromises()

    expect(getWorkbenchStageRail(wrapper).get('[role="tab"][aria-selected="true"]').text()).toContain('AWD 配置')
    expect(wrapper.text()).toContain('当前聚焦题目')
    expect(wrapper.text()).toContain('Web 入门')
  })

  it('应该在 AWD 赛事已开赛时默认聚焦轮次运行阶段', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'running',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('轮次运行')
    expect(wrapper.text()).toContain('轮次态势')
  })

  it('未开赛时工作台运行段应承接降级壳', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-operations').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('尚未进入运行阶段')
    expect(wrapper.text()).toContain('需先通过赛前检查并开赛')
  })

  it('应该在 AWD 赛事已结束时进入运行段而不是显示赛前降级壳', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'ended',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('轮次运行')
    expect(wrapper.text()).toContain('轮次态势')
    expect(wrapper.text()).not.toContain('尚未进入运行阶段')
  })

  it('AWD 题目列表刷新失败时应保留上次成功数据并避免把摘要误报为 0', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminContestChallenges
      .mockResolvedValueOnce([
        {
          id: 'link-1',
          contest_id: 'contest-1',
          challenge_id: '101',
          title: 'Web 入门',
          category: 'web',
          difficulty: 'easy',
          points: 120,
          order: 1,
          is_visible: true,
          awd_checker_type: undefined,
          awd_checker_config: {},
          awd_sla_score: 0,
          awd_defense_score: 0,
          awd_checker_validation_state: 'pending',
          awd_checker_last_preview_at: undefined,
          awd_checker_last_preview_result: undefined,
          created_at: '2026-03-10T00:00:00.000Z',
        },
      ])
      .mockRejectedValueOnce(new Error('refresh failed'))

    const wrapper = mountContestEdit()

    await flushPromises()

    expect(wrapper.text()).toContain('已关联题目数')
    expect(wrapper.text()).toContain('Web 入门')

    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('刷新列表'))
      ?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('refresh failed')
    expect(wrapper.text()).toContain('Web 入门')
    expect(wrapper.text()).toContain('已关联题目数')
    expect(wrapper.text()).not.toContain('当前竞赛还没有关联题目')
    expect(wrapper.text()).not.toContain('共 0 道题目')
  })

  it('应该在管理页工作台交接时强制落到轮次态势而不是恢复旧子页签', async () => {
    window.sessionStorage.setItem('ctf_admin_awd_ops_panel:contest-1', 'challenges')
    window.history.replaceState({}, '', '/admin/contests/contest-1/edit?panel=operations&opsPanel=inspector')
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'running',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()

    expect(wrapper.get('#awd-ops-tab-inspector').attributes('aria-selected')).toBe('true')
    expect(wrapper.get('#awd-ops-tab-challenges').attributes('aria-selected')).toBe('false')
    expect(wrapper.text()).toContain('轮次态势')
  })

  it('应该在 URL 已指定有效阶段时保留该阶段', async () => {
    window.history.replaceState({}, '', '/admin/contests/contest-1/edit?panel=operations')
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('轮次运行')
    expect(window.location.search).toContain('panel=operations')
  })

  it('应该加载竞赛详情并在保存成功后返回赛事目录', async () => {
    const wrapper = mount(ContestEdit, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('编辑竞赛')

    await wrapper.get('#contest-title').setValue('2026 春季校园 CTF（更新）')
    await wrapper.get('.contest-form-button--primary').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.objectContaining({
        title: '2026 春季校园 CTF（更新）',
      })
    )
    expect(pushMock).toHaveBeenCalledWith({ name: 'ContestManage', query: { panel: 'list' } })
  })

  it('应该在终止进行中的竞赛前弹出二次确认', async () => {
    destructiveConfirmMock.mockResolvedValue(false)
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 春季校园 CTF',
        status: 'running',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-status').setValue('ended')
    await wrapper.get('.contest-form-button--primary').trigger('click')
    await flushPromises()

    expect(destructiveConfirmMock).toHaveBeenCalledWith(
      expect.objectContaining({
        title: '确认结束赛事',
      })
    )
    expect(contestApiMocks.updateContest).not.toHaveBeenCalled()
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('应该在 AWD 启动门禁拦截后展示放行弹层并在确认后回到赛事目录', async () => {
    contestApiMocks.getContest.mockResolvedValue({
      id: 'contest-1',
      title: '2026 AWD 联赛',
      description: '攻防赛',
      mode: 'awd',
      status: 'registering',
      starts_at: '2026-03-15T09:00:00.000Z',
      ends_at: '2026-03-15T13:00:00.000Z',
    })
    contestApiMocks.updateContest
      .mockRejectedValueOnce(new ApiError('AWD 开赛就绪检查未通过', { status: 409, code: 14025 }))
      .mockResolvedValueOnce({
        contest: {
          id: 'contest-1',
          title: '2026 AWD 联赛',
          description: '攻防赛',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      })

    const wrapper = mount(ContestEdit, {
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#contest-status').setValue('running')
    await wrapper.get('.contest-form-button--primary').trigger('click')
    await flushPromises()

    expect(contestApiMocks.getContestAWDReadiness).toHaveBeenCalledWith('contest-1')
    expect(wrapper.text()).toContain('启动赛事')
    expect(wrapper.text()).toContain('强制继续')

    await wrapper.get('#awd-readiness-override-reason').setValue('teacher drill')
    await wrapper.get('#awd-readiness-override-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateContest).toHaveBeenNthCalledWith(
      2,
      'contest-1',
      expect.objectContaining({
        status: 'running',
        force_override: true,
        override_reason: 'teacher drill',
      }),
      { suppressErrorToast: true }
    )
    expect(pushMock).toHaveBeenCalledWith({ name: 'ContestManage', query: { panel: 'list' } })
  })

  it('应该在 14025 后续读取 readiness 失败时给出清晰错误并保持界面可继续操作', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.updateContest.mockRejectedValueOnce(
      new ApiError('AWD 开赛就绪检查未通过', { status: 409, code: 14025 })
    )
    contestApiMocks.getContestAWDReadiness
      .mockResolvedValueOnce({
        contest_id: 'contest-1',
        ready: false,
        total_challenges: 1,
        passed_challenges: 0,
        pending_challenges: 0,
        failed_challenges: 1,
        stale_challenges: 0,
        missing_checker_challenges: 0,
        blocking_count: 1,
        global_blocking_reasons: [],
        blocking_actions: ['start_contest'],
        items: [
          {
            challenge_id: '101',
            title: 'Challenge 101',
            checker_type: 'http_standard',
            validation_state: 'failed',
            last_preview_at: '2026-04-12T08:00:00.000Z',
            last_access_url: 'http://checker.internal/flag',
            blocking_reason: 'last_preview_failed',
          },
        ],
      })
      .mockRejectedValueOnce(new Error('readiness fetch failed'))

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-status').setValue('running')
    await wrapper.get('.contest-form-button--primary').trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('readiness fetch failed')
    expect(contestApiMocks.updateContest).toHaveBeenCalledTimes(1)
    expect(pushMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('编辑竞赛')
  })

  it('应该在赛前检查强制开赛时带上基础表单最新草稿值', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-title').setValue('2026 AWD 联赛（演练版）')
    await wrapper.get('#contest-workbench-stage-tab-preflight').trigger('click')
    await flushPromises()

    await wrapper.get('#contest-awd-preflight-force-start').trigger('click')
    await flushPromises()
    wrapper.findComponent(AWDReadinessOverrideDialog).vm.$emit('confirm', 'teacher drill')
    await flushPromises()

    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.objectContaining({
        title: '2026 AWD 联赛（演练版）',
        status: 'running',
        force_override: true,
        override_reason: 'teacher drill',
      }),
      { suppressErrorToast: true }
    )
  })

  it('应该在 AWD 辅助请求失败时仍保留工作台而不是进入全局加载错误态', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.getContestAWDReadiness.mockRejectedValue(new Error('readiness failed'))

    const wrapper = mountContestEdit()

    await flushPromises()

    expect(wrapper.text()).toContain('竞赛编辑')
    expect(wrapper.text()).toContain('基础信息')
    expect(wrapper.text()).not.toContain('竞赛详情加载失败')
  })

  it('应该为 AWD 配置题库按 published 状态分页拉取完整目录', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.getChallenges
      .mockResolvedValueOnce({
        list: Array.from({ length: 200 }, (_, index) => ({
          id: String(index + 200),
          title: `Challenge ${index + 200}`,
          description: 'page1',
          category: 'web',
          difficulty: 'easy',
          points: 100,
          instance_sharing: 'per_user',
          created_by: '9',
          image_id: undefined,
          attachment_url: undefined,
          hints: undefined,
          status: 'published',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
          flag_config: undefined,
        })),
        total: 201,
        page: 1,
        page_size: 200,
      })
      .mockResolvedValueOnce({
        list: [
          {
            id: '999',
            title: 'Final Challenge',
            description: 'page2',
            category: 'crypto',
            difficulty: 'medium',
            points: 150,
            instance_sharing: 'per_user',
            created_by: '9',
            image_id: undefined,
            attachment_url: undefined,
            hints: undefined,
            status: 'published',
            created_at: '2026-03-02T00:00:00.000Z',
            updated_at: '2026-03-02T00:00:00.000Z',
            flag_config: undefined,
          },
        ],
        total: 201,
        page: 2,
        page_size: 200,
      })

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-create').trigger('click')
    await flushPromises()

    expect(contestApiMocks.getChallenges).toHaveBeenNthCalledWith(1, {
      page: 1,
      page_size: 200,
      status: 'published',
    })
    expect(contestApiMocks.getChallenges).toHaveBeenNthCalledWith(2, {
      page: 2,
      page_size: 200,
      status: 'published',
    })
    expect(wrapper.text()).toContain('Final Challenge')
  })

  it('应该在 AWD 配置题库加载失败时给出错误提示而不是留下未处理异常', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.getChallenges.mockRejectedValueOnce(new Error('catalog failed'))

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-create').trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('catalog failed')
    expect(wrapper.text()).toContain('新增 AWD 题目')
  })

  it('应该在 AWD 辅助数据仍在加载时显示阶段级加载提示而不是空态', async () => {
    const challengeLinksDeferred = createDeferred<any[]>()

    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminContestChallenges.mockReturnValueOnce(challengeLinksDeferred.promise)

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('正在同步 AWD 配置...')
    expect(wrapper.text()).not.toContain('当前赛事还没有关联题目')

    challengeLinksDeferred.resolve([
      {
        id: 'link-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        title: 'Web 入门',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 1,
        is_visible: true,
        awd_checker_type: undefined,
        awd_checker_config: {},
        awd_sla_score: 0,
        awd_defense_score: 0,
        awd_checker_validation_state: 'pending',
        awd_checker_last_preview_at: undefined,
        awd_checker_last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
      },
    ])
    await flushPromises()
  })

  it('应该允许管理员在竞赛编辑页编排题目', async () => {
    const wrapper = mount(ContestEdit, {
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()

    expect(contestApiMocks.listAdminContestChallenges).toHaveBeenCalledWith('contest-1')
    expect(wrapper.text()).toContain('题目池')
    expect(wrapper.text()).toContain('Web 入门')

    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()

    expect(contestApiMocks.getChallenges).toHaveBeenCalledWith({
      page: 1,
      page_size: 100,
      status: 'published',
    })

    await wrapper.get('#contest-challenge-select').setValue('102')
    await wrapper.get('#contest-challenge-points').setValue('160')
    await wrapper.get('#contest-challenge-order').setValue('3')
    await wrapper.get('#contest-challenge-visibility').setValue('false')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.createAdminContestChallenge).toHaveBeenCalledWith('contest-1', {
      challenge_id: 102,
      points: 160,
      order: 3,
      is_visible: false,
    })

    await wrapper.get('#contest-challenge-edit-link-1').trigger('click')
    await flushPromises()

    await wrapper.get('#contest-challenge-points').setValue('140')
    await wrapper.get('#contest-challenge-order').setValue('2')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateAdminContestChallenge).toHaveBeenCalledWith('contest-1', '101', {
      points: 140,
      order: 2,
      is_visible: true,
    })

    await wrapper.get('#contest-challenge-remove-link-1').trigger('click')
    await flushPromises()

    expect(destructiveConfirmMock).toHaveBeenCalled()
    expect(contestApiMocks.deleteAdminContestChallenge).toHaveBeenCalledWith('contest-1', '101')
  })

  it('题目池变更后应同步更新 AWD 配置与赛前检查数据', async () => {
    const challengeLinksState: any[] = []
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminContestChallenges.mockImplementation(async () =>
      challengeLinksState.map((item) => ({ ...item }))
    )
    contestApiMocks.getContestAWDReadiness.mockImplementation(async () => ({
      contest_id: 'contest-1',
      ready: challengeLinksState.length > 0,
      total_challenges: challengeLinksState.length,
      passed_challenges: challengeLinksState.length,
      pending_challenges: 0,
      failed_challenges: 0,
      stale_challenges: 0,
      missing_checker_challenges: 0,
      blocking_count: 0,
      global_blocking_reasons: challengeLinksState.length > 0 ? [] : ['no_challenges'],
      blocking_actions: challengeLinksState.length > 0 ? [] : ['start_contest'],
      items: [],
    }))
    contestApiMocks.getChallenges.mockResolvedValue({
      list: [
        {
          id: '102',
          title: 'Upload Service',
          description: '新增题目',
          category: 'web',
          difficulty: 'medium',
          points: 150,
          instance_sharing: 'per_user',
          created_by: '9',
          image_id: undefined,
          attachment_url: undefined,
          hints: undefined,
          status: 'published',
          created_at: '2026-03-02T00:00:00.000Z',
          updated_at: '2026-03-02T00:00:00.000Z',
          flag_config: undefined,
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    contestApiMocks.createAdminContestChallenge.mockImplementation(async (_contestId, payload) => {
      const created = {
        id: 'link-2',
        contest_id: 'contest-1',
        challenge_id: String(payload.challenge_id),
        title: 'Upload Service',
        category: 'web',
        difficulty: 'medium',
        points: payload.points,
        order: payload.order,
        is_visible: payload.is_visible,
        awd_checker_type: undefined,
        awd_checker_config: {},
        awd_sla_score: 0,
        awd_defense_score: 0,
        awd_checker_validation_state: 'pending',
        awd_checker_last_preview_at: undefined,
        awd_checker_last_preview_result: undefined,
        created_at: '2026-03-10T01:00:00.000Z',
      }
      challengeLinksState.push(created)
      return created
    })

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-select').setValue('102')
    await wrapper.get('#contest-challenge-points').setValue('160')
    await wrapper.get('#contest-challenge-order').setValue('3')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Upload Service')

    await wrapper.get('#contest-workbench-stage-tab-preflight').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('可开赛')
    expect(wrapper.text()).not.toContain('当前赛事还没有关联题目，无法执行开赛关键动作')
  })

  it('AWD 配置变更后题目池应同步更新并阻止重复关联', async () => {
    const challengeLinksState: any[] = [
      {
        id: 'link-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        title: 'Web 入门',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 1,
        is_visible: true,
        awd_checker_type: undefined,
        awd_checker_config: {},
        awd_sla_score: 0,
        awd_defense_score: 0,
        awd_checker_validation_state: 'pending',
        awd_checker_last_preview_at: undefined,
        awd_checker_last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
      },
    ]
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminContestChallenges.mockImplementation(async () =>
      challengeLinksState.map((item) => ({ ...item }))
    )
    contestApiMocks.getChallenges.mockResolvedValue({
      list: [
        {
          id: '101',
          title: 'Web 入门',
          description: '现有题目',
          category: 'web',
          difficulty: 'easy',
          points: 120,
          instance_sharing: 'per_user',
          created_by: '9',
          image_id: undefined,
          attachment_url: undefined,
          hints: undefined,
          status: 'published',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
          flag_config: undefined,
        },
        {
          id: '102',
          title: 'Crypto 进阶',
          description: '新增题目',
          category: 'crypto',
          difficulty: 'medium',
          points: 150,
          instance_sharing: 'per_user',
          created_by: '9',
          image_id: undefined,
          attachment_url: undefined,
          hints: undefined,
          status: 'published',
          created_at: '2026-03-02T00:00:00.000Z',
          updated_at: '2026-03-02T00:00:00.000Z',
          flag_config: undefined,
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })
    contestApiMocks.createAdminContestChallenge.mockImplementation(async (_contestId, payload) => {
      const created = {
        id: 'link-2',
        contest_id: 'contest-1',
        challenge_id: String(payload.challenge_id),
        title: 'Crypto 进阶',
        category: 'crypto',
        difficulty: 'medium',
        points: payload.points,
        order: payload.order,
        is_visible: payload.is_visible,
        awd_checker_type: payload.awd_checker_type,
        awd_checker_config: payload.awd_checker_config ?? {},
        awd_sla_score: payload.awd_sla_score ?? 0,
        awd_defense_score: payload.awd_defense_score ?? 0,
        awd_checker_validation_state: 'pending',
        awd_checker_last_preview_at: undefined,
        awd_checker_last_preview_result: undefined,
        created_at: '2026-03-10T01:00:00.000Z',
      }
      challengeLinksState.push(created)
      return created
    })

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-create').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-challenge').setValue('102')
    await wrapper.get('#awd-challenge-config-points').setValue('160')
    await wrapper.get('#awd-challenge-config-order').setValue('2')
    await wrapper.get('#awd-challenge-config-submit').trigger('click')
    await flushPromises()

    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Crypto 进阶')

    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    expect(wrapper.find('#contest-challenge-select').text()).not.toContain('Crypto 进阶')
  })

  it('AWD 配置保存失败时应提示错误并保持弹层打开', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.createAdminContestChallenge.mockRejectedValueOnce(new Error('save failed'))

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-create').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-challenge').setValue('102')
    await wrapper.get('#awd-challenge-config-points').setValue('160')
    await wrapper.get('#awd-challenge-config-order').setValue('2')
    await wrapper.get('#awd-challenge-config-submit').trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('save failed')
    expect(wrapper.text()).toContain('新增 AWD 题目')
  })

  it('应该在 AWD 赛事的题目池阶段展示摘要列与筛选入口', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      {
        id: 'link-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        title: 'Web 入门',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 1,
        is_visible: true,
        awd_checker_type: 'http_standard',
        awd_checker_config: {},
        awd_sla_score: 18,
        awd_defense_score: 28,
        awd_checker_validation_state: 'stale',
        awd_checker_last_preview_at: '2026-04-12T08:00:00.000Z',
        awd_checker_last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
      },
    ])

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('未配置 AWD')
    expect(wrapper.text()).toContain('预检失败')
    expect(wrapper.text()).toContain('Checker')
    expect(wrapper.text()).toContain('SLA 18 / 防守 28')
    expect(wrapper.text()).toContain('待重新验证')
  })
})
