import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'

import AWDOperationsPanel from '../contest/AWDOperationsPanel.vue'

const awdMockModule = vi.hoisted(() => ({
  state: null as any,
}))

vi.mock('@/composables/useAdminContestAWD', async () => {
  const { ref } = await vi.importActual<typeof import('vue')>('vue')
  awdMockModule.state = {
    rounds: ref([]),
    selectedRoundId: ref<string | null>(null),
    readiness: ref(null),
    loadingReadiness: ref(false),
    overrideDialogState: ref({
      open: false,
      action: null,
      title: '',
      reason: '',
      readiness: null,
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
      service_id: '',
      challenge_id: '',
      status_group: 'all',
      path_keyword: '',
      page: 1,
      page_size: 20,
    }),
    scoreboardRows: ref([]),
    scoreboardFrozen: ref(false),
    teams: ref([]),
    challengeLinks: ref([]),
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

function getAwdState() {
  return awdMockModule.state
}

function buildReadinessState(overrides: Record<string, unknown> = {}) {
  return {
    contest_id: 'awd-1',
    ready: false,
    total_challenges: 3,
    passed_challenges: 1,
    pending_challenges: 1,
    failed_challenges: 1,
    stale_challenges: 0,
    missing_checker_challenges: 1,
    blocking_count: 2,
    global_blocking_reasons: [],
    blocking_actions: ['create_round', 'run_current_round_check'],
    items: [
      {
        challenge_id: 'challenge-1',
        title: 'Web Checker',
        checker_type: 'http_standard',
        validation_state: 'failed',
        last_preview_at: '2026-04-12T08:00:00.000Z',
        last_access_url: 'http://checker.internal/flag',
        blocking_reason: 'last_preview_failed',
      },
    ],
    ...overrides,
  }
}

function buildDialogState(overrides: Record<string, unknown> = {}) {
  return {
    open: true,
    action: 'create_round',
    title: '创建轮次',
    reason: '',
    readiness: buildReadinessState(),
    ...overrides,
  }
}

describe('AWDOperationsPanel', () => {
  beforeEach(() => {
    const awdState = getAwdState()
    awdState.challengeLinks.value = []
    awdState.challengeCatalog.value = []
    awdState.teams.value = []
    awdState.rounds.value = []
    awdState.selectedRoundId.value = null
    awdState.readiness.value = null
    awdState.loadingReadiness.value = false
    awdState.overrideDialogState.value = {
      open: false,
      action: null,
      title: '',
      reason: '',
      readiness: null,
    }
    awdState.loadChallengeCatalog.mockReset()
    awdState.createChallengeLink.mockReset()
    awdState.updateChallengeLink.mockReset()
    awdState.createRound.mockReset()
    awdState.runSelectedRoundCheck.mockReset()
    awdState.confirmOverrideAction.mockReset()
    awdState.closeOverrideDialog.mockImplementation(() => {
      awdState.overrideDialogState.value = {
        open: false,
        action: null,
        title: '',
        reason: '',
        readiness: null,
      }
    })
  })

  it('应该在没有 AWD 赛事时展示空态', () => {
    const wrapper = mount(AWDOperationsPanel, {
      props: {
        contests: [],
        selectedContestId: null,
      },
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
          AWDRoundInspector: true,
          AWDRoundCreateDialog: true,
          AWDServiceCheckDialog: true,
          AWDAttackLogDialog: true,
          AWDChallengeConfigPanel: true,
          AWDChallengeConfigDialog: true,
        },
      },
    })

    expect(wrapper.get('#awd-contest-selector').text()).toContain('暂无 AWD 赛事')
    expect(wrapper.text()).toContain('当前页没有 AWD 赛事')
  })

  it('应该在赛事列表存在但未选中有效 AWD 赛事时展示显式空态', () => {
    const wrapper = mount(AWDOperationsPanel, {
      props: {
        contests: [
          {
            id: 'awd-1',
            title: '2026 AWD 联赛',
            description: '攻防赛',
            mode: 'awd',
            status: 'running',
            starts_at: '2026-03-18T09:00:00.000Z',
            ends_at: '2026-03-18T18:00:00.000Z',
          },
        ],
        selectedContestId: 'missing-contest',
      },
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
          AWDRoundInspector: true,
          AWDRoundCreateDialog: true,
          AWDServiceCheckDialog: true,
          AWDAttackLogDialog: true,
          AWDChallengeConfigPanel: true,
          AWDChallengeConfigDialog: true,
        },
      },
    })

    expect(wrapper.text()).toContain('选择 AWD 赛事')
    expect(wrapper.text()).toContain('暂无 AWD 赛事')
    expect(wrapper.html()).not.toContain('a-w-d-round-inspector-stub')
  })

  it('应该在选中 AWD 赛事后展示题目配置入口', () => {
    const awdState = getAwdState()
    awdState.challengeLinks.value = [
      {
        id: 'link-1',
        contest_id: 'awd-1',
        challenge_id: 'challenge-1',
        title: 'Web Checker',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 1,
        is_visible: true,
        awd_checker_type: 'http_standard',
        awd_checker_config: {
          get_flag: { method: 'GET', path: '/api/flag' },
        },
        awd_sla_score: 18,
        awd_defense_score: 28,
        awd_checker_validation_state: 'passed',
        awd_checker_last_preview_at: '2026-03-18T09:05:00.000Z',
        awd_checker_last_preview_result: {
          checker_type: 'http_standard',
          service_status: 'up',
          check_result: {
            checker_type: 'http_standard',
            check_source: 'checker_preview',
            status_reason: 'healthy',
          },
          preview_context: {
            access_url: 'http://preview.internal',
            preview_flag: 'flag{preview}',
            round_number: 0,
            team_id: '0',
            challenge_id: 'challenge-1',
          },
        },
        created_at: '2026-03-18T09:00:00.000Z',
      },
    ]

    const wrapper = mount(AWDOperationsPanel, {
      props: {
        contests: [
          {
            id: 'awd-1',
            title: '2026 AWD 联赛',
            description: '攻防赛',
            mode: 'awd',
            status: 'running',
            starts_at: '2026-03-18T09:00:00.000Z',
            ends_at: '2026-03-18T18:00:00.000Z',
          },
        ],
        selectedContestId: 'awd-1',
      },
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
          AWDRoundInspector: true,
          AWDRoundCreateDialog: true,
          AWDServiceCheckDialog: true,
          AWDAttackLogDialog: true,
          AWDChallengeConfigDialog: true,
        },
      },
    })

    expect(wrapper.text()).toContain('题目配置')
    expect(wrapper.text()).toContain('新增题目')
    expect(wrapper.text()).toContain('编辑配置')
    expect(wrapper.text()).toContain('最近通过')
  })

  it('应该渲染 readiness 摘要与系统级阻塞提示', () => {
    const awdState = getAwdState()
    awdState.readiness.value = buildReadinessState({
      global_blocking_reasons: ['no_challenges'],
      items: [],
    })

    const wrapper = mount(AWDOperationsPanel, {
      props: {
        contests: [
          {
            id: 'awd-1',
            title: '2026 AWD 联赛',
            description: '攻防赛',
            mode: 'awd',
            status: 'running',
            starts_at: '2026-03-18T09:00:00.000Z',
            ends_at: '2026-03-18T18:00:00.000Z',
          },
        ],
        selectedContestId: 'awd-1',
      },
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
          AWDRoundInspector: true,
          AWDRoundCreateDialog: true,
          AWDServiceCheckDialog: true,
          AWDAttackLogDialog: true,
          AWDChallengeConfigDialog: true,
        },
      },
    })

    expect(wrapper.text()).toContain('开赛就绪摘要')
    expect(wrapper.text()).toContain('未配 Checker')
    expect(wrapper.text()).toContain('当前赛事还没有关联题目')
  })

  it('未开赛时运行段应显示尚未进入运行阶段', () => {
    const wrapper = mount(AWDOperationsPanel, {
      props: {
        contests: [
          {
            id: 'awd-1',
            title: '2026 AWD 联赛',
            description: '攻防赛',
            mode: 'awd',
            status: 'registering',
            starts_at: '2026-03-18T09:00:00.000Z',
            ends_at: '2026-03-18T18:00:00.000Z',
          },
        ],
        selectedContestId: 'awd-1',
      },
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
          AWDRoundInspector: true,
          AWDRoundCreateDialog: true,
          AWDServiceCheckDialog: true,
          AWDAttackLogDialog: true,
          AWDChallengeConfigDialog: true,
        },
      },
    })

    expect(wrapper.text()).toContain('轮次态势')
    expect(wrapper.text()).toContain('尚未进入运行阶段')
    expect(wrapper.text()).toContain('需先通过赛前检查并开赛')
    expect(wrapper.get('#awd-runtime-shell-create-round').attributes('disabled')).toBeDefined()
    expect(wrapper.get('#awd-runtime-shell-run-check').attributes('disabled')).toBeDefined()
  })

  it('应该在创建轮次被 gate 拦截时打开强制继续弹层', async () => {
    const awdState = getAwdState()
    awdState.readiness.value = buildReadinessState()
    awdState.createRound.mockImplementation(async () => {
      awdState.overrideDialogState.value = buildDialogState()
    })

    const wrapper = mount(AWDOperationsPanel, {
      props: {
        contests: [
          {
            id: 'awd-1',
            title: '2026 AWD 联赛',
            description: '攻防赛',
            mode: 'awd',
            status: 'running',
            starts_at: '2026-03-18T09:00:00.000Z',
            ends_at: '2026-03-18T18:00:00.000Z',
          },
        ],
        selectedContestId: 'awd-1',
      },
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
          AWDRoundInspector: {
            template:
              '<div><button id="stub-open-create-round" type="button" @click="$emit(\'openCreateRoundDialog\')">创建轮次</button></div>',
          },
          AWDRoundCreateDialog: {
            emits: ['save', 'update:open'],
            template:
              '<button id="awd-round-create-submit" type="button" @click="$emit(\'save\', { round_number: 1, status: \'pending\', attack_score: 50, defense_score: 50 })">submit round</button>',
          },
          AWDServiceCheckDialog: true,
          AWDAttackLogDialog: true,
          AWDChallengeConfigPanel: true,
          AWDChallengeConfigDialog: true,
        },
      },
    })

    await wrapper.get('#stub-open-create-round').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-round-create-submit').trigger('click')
    await flushPromises()

    expect(awdState.createRound).toHaveBeenCalled()
    expect(wrapper.text()).toContain('强制继续')
    expect(wrapper.text()).toContain('创建轮次')
  })

  it('应该在当前轮巡检被 gate 拦截时打开强制继续弹层', async () => {
    const awdState = getAwdState()
    awdState.readiness.value = buildReadinessState()
    awdState.runSelectedRoundCheck.mockImplementation(async () => {
      awdState.overrideDialogState.value = buildDialogState({
        action: 'run_current_round_check',
        title: '立即巡检当前轮',
      })
    })

    const wrapper = mount(AWDOperationsPanel, {
      props: {
        contests: [
          {
            id: 'awd-1',
            title: '2026 AWD 联赛',
            description: '攻防赛',
            mode: 'awd',
            status: 'running',
            starts_at: '2026-03-18T09:00:00.000Z',
            ends_at: '2026-03-18T18:00:00.000Z',
          },
        ],
        selectedContestId: 'awd-1',
      },
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
          AWDRoundInspector: {
            template:
              '<div><button id="stub-run-current-check" type="button" @click="$emit(\'runSelectedRoundCheck\')">立即巡检当前轮</button></div>',
          },
          AWDRoundCreateDialog: true,
          AWDServiceCheckDialog: true,
          AWDAttackLogDialog: true,
          AWDChallengeConfigPanel: true,
          AWDChallengeConfigDialog: true,
        },
      },
    })

    await wrapper.get('#stub-run-current-check').trigger('click')
    await flushPromises()

    expect(awdState.runSelectedRoundCheck).toHaveBeenCalled()
    expect(wrapper.text()).toContain('强制继续')
    expect(wrapper.text()).toContain('立即巡检当前轮')
  })

  it('应该在普通失败未进入 gate 状态时保持弹层关闭', async () => {
    const awdState = getAwdState()
    awdState.readiness.value = buildReadinessState()
    awdState.createRound.mockImplementation(async () => {
      await Promise.reject(
        Object.assign(new Error('普通冲突'), { name: 'ApiError', status: 409, code: 14099 })
      ).catch(() => undefined)
    })

    const wrapper = mount(AWDOperationsPanel, {
      props: {
        contests: [
          {
            id: 'awd-1',
            title: '2026 AWD 联赛',
            description: '攻防赛',
            mode: 'awd',
            status: 'running',
            starts_at: '2026-03-18T09:00:00.000Z',
            ends_at: '2026-03-18T18:00:00.000Z',
          },
        ],
        selectedContestId: 'awd-1',
      },
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
          AWDRoundInspector: {
            template:
              '<div><button id="stub-open-create-round" type="button" @click="$emit(\'openCreateRoundDialog\')">创建轮次</button></div>',
          },
          AWDRoundCreateDialog: {
            emits: ['save', 'update:open'],
            template:
              '<button id="awd-round-create-submit" type="button" @click="$emit(\'save\', { round_number: 1, status: \'pending\', attack_score: 50, defense_score: 50 })">submit round</button>',
          },
          AWDServiceCheckDialog: true,
          AWDAttackLogDialog: true,
          AWDChallengeConfigPanel: true,
          AWDChallengeConfigDialog: true,
        },
      },
    })

    await wrapper.get('#stub-open-create-round').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-round-create-submit').trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('填写本次放行原因')
  })

  it('应该把服务检查与攻击日志的 service_id 载荷透传给 composable', async () => {
    const awdState = getAwdState()
    awdState.challengeLinks.value = [
      {
        id: 'link-1',
        contest_id: 'awd-1',
        challenge_id: 'challenge-1',
        awd_service_id: '7009',
        title: 'Web Checker',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 1,
        is_visible: true,
        created_at: '2026-03-18T09:00:00.000Z',
      },
    ]
    awdState.teams.value = [
      {
        id: '12',
        name: 'Red',
        leader_id: '1',
        member_count: 1,
        total_score: 0,
        created_at: '2026-03-18T09:00:00.000Z',
      },
      {
        id: '13',
        name: 'Blue',
        leader_id: '2',
        member_count: 1,
        total_score: 0,
        created_at: '2026-03-18T09:00:00.000Z',
      },
    ]

    const ServiceDialogStub = defineComponent({
      emits: ['save', 'update:open'],
      template:
        '<button id="stub-save-service-check" type="button" @click="$emit(\'save\', { team_id: 12, service_id: 7009, service_status: \'up\', check_result: { latency_ms: 38 } })">save service</button>',
    })
    const AttackDialogStub = defineComponent({
      emits: ['save', 'update:open'],
      template:
        '<button id="stub-save-attack-log" type="button" @click="$emit(\'save\', { attacker_team_id: 12, victim_team_id: 13, service_id: 7009, attack_type: \'flag_capture\', submitted_flag: \'flag{demo}\', is_success: true })">save attack</button>',
    })

    const wrapper = mount(AWDOperationsPanel, {
      props: {
        contests: [
          {
            id: 'awd-1',
            title: '2026 AWD 联赛',
            description: '攻防赛',
            mode: 'awd',
            status: 'running',
            starts_at: '2026-03-18T09:00:00.000Z',
            ends_at: '2026-03-18T18:00:00.000Z',
          },
        ],
        selectedContestId: 'awd-1',
      },
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
          AWDRoundInspector: true,
          AWDRoundCreateDialog: true,
          AWDServiceCheckDialog: ServiceDialogStub,
          AWDAttackLogDialog: AttackDialogStub,
          AWDChallengeConfigPanel: true,
          AWDChallengeConfigDialog: true,
        },
      },
    })

    await wrapper.get('#stub-save-service-check').trigger('click')
    await wrapper.get('#stub-save-attack-log').trigger('click')

    expect(awdState.createServiceCheck).toHaveBeenCalledWith({
      team_id: 12,
      service_id: 7009,
      service_status: 'up',
      check_result: { latency_ms: 38 },
    })
    expect(awdState.createAttackLog).toHaveBeenCalledWith({
      attacker_team_id: 12,
      victim_team_id: 13,
      service_id: 7009,
      attack_type: 'flag_capture',
      submitted_flag: 'flag{demo}',
      is_success: true,
    })
  })
})
