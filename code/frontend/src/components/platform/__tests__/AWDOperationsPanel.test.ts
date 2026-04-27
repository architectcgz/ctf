import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'

import AWDOperationsPanel from '../contest/AWDOperationsPanel.vue'

const awdMockModule = vi.hoisted(() => ({
  state: null as any,
}))

vi.mock('@/composables/usePlatformContestAwd', async () => {
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
    instanceOrchestration: ref({
      contest_id: '',
      teams: [],
      services: [],
      instances: [],
    }),
    loadingRounds: ref(false),
    loadingRoundDetail: ref(false),
    loadingTrafficSummary: ref(false),
    loadingTrafficEvents: ref(false),
    loadingChallengeCatalog: ref(false),
    loadingInstanceOrchestration: ref(false),
    checking: ref(false),
    creatingRound: ref(false),
    savingServiceCheck: ref(false),
    savingAttackLog: ref(false),
    savingChallengeConfig: ref(false),
    startingInstanceKey: ref(null),
    shouldAutoRefresh: ref(false),
    refresh: vi.fn(),
    refreshInstanceOrchestration: vi.fn(),
    applyTrafficFilters: vi.fn(),
    setTrafficPage: vi.fn(),
    resetTrafficFilters: vi.fn(),
    runSelectedRoundCheck: vi.fn(),
    startTeamServiceInstance: vi.fn(),
    startTeamAllServices: vi.fn(),
    startAllTeamServices: vi.fn(),
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
    usePlatformContestAwd: () => awdMockModule.state,
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
    awdState.instanceOrchestration.value = {
      contest_id: '',
      teams: [],
      services: [],
      instances: [],
    }
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
    awdState.refreshInstanceOrchestration.mockReset()
    awdState.startTeamServiceInstance.mockReset()
    awdState.startTeamAllServices.mockReset()
    awdState.startAllTeamServices.mockReset()
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

  it('应该在选中 AWD 赛事后只保留轮次态势，并把题目配置收口到外层工作台', async () => {
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
        awd_sla_score: 1,
        awd_defense_score: 2,
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
    awdState.readiness.value = buildReadinessState({
      pending_challenges: 0,
      failed_challenges: 1,
      missing_checker_challenges: 0,
      blocking_count: 1,
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

    expect(wrapper.text()).toContain('轮次态势')
    expect(wrapper.text()).not.toContain('题目配置')
    expect(wrapper.text()).toContain('最近通过')

    await wrapper.get('#awd-readiness-edit-challenge-1').trigger('click')

    expect(wrapper.emitted('open:awd-config')).toEqual([['challenge-1']])
  })

  it('作为独立运维页使用时应隐藏就绪摘要中的编辑动作', () => {
    const awdState = getAwdState()
    awdState.readiness.value = buildReadinessState({
      pending_challenges: 0,
      failed_challenges: 1,
      missing_checker_challenges: 0,
      blocking_count: 1,
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
        hideStudioLink: true,
        hideReadinessActions: true,
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

    expect(wrapper.text()).toContain('Web Checker')
    expect(wrapper.text()).not.toContain('编辑配置')
    expect(wrapper.text()).not.toContain('进入竞赛工作室')
    expect(wrapper.find('#awd-readiness-edit-challenge-1').exists()).toBe(false)
    expect(wrapper.emitted('open:awd-config')).toBeUndefined()
    expect(wrapper.emitted('open:contest-edit')).toBeUndefined()
  })

  it('应该把队伍实例编排放入独立运维 tab', async () => {
    const awdState = getAwdState()
    awdState.instanceOrchestration.value = {
      contest_id: 'awd-1',
      teams: [
        {
          team_id: 'team-1',
          team_name: 'Red Team',
          captain_id: 'captain-1',
        },
      ],
      services: [
        {
          service_id: 'service-1',
          challenge_id: 'challenge-1',
          display_name: 'Web Service',
          is_visible: true,
        },
      ],
      instances: [],
    }

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

    expect(wrapper.get('#awd-ops-tab-inspector').attributes('aria-selected')).toBe('true')
    expect(wrapper.get('#awd-ops-tab-instances').attributes('aria-selected')).toBe('false')
    expect(wrapper.find('#awd-ops-panel-instances').exists()).toBe(false)

    await wrapper.get('#awd-ops-tab-instances').trigger('click')
    await flushPromises()

    expect(wrapper.get('#awd-ops-tab-inspector').attributes('aria-selected')).toBe('false')
    expect(wrapper.get('#awd-ops-tab-instances').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#awd-ops-panel-inspector').exists()).toBe(false)
    expect(wrapper.get('#awd-ops-panel-instances').isVisible()).toBe(true)
    expect(wrapper.text()).toContain('Red Team')
  })

  it('运行内容为 readiness 时只显示运行态摘要', () => {
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
        hideOperationTabs: true,
        operationPanel: 'inspector',
        runtimeContent: 'readiness',
      },
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
          AWDRoundInspector: {
            template: '<section id="awd-ops-panel-inspector">轮次矩阵</section>',
          },
          AWDRoundCreateDialog: true,
          AWDServiceCheckDialog: true,
          AWDAttackLogDialog: true,
          AWDChallengeConfigDialog: true,
        },
      },
    })

    expect(wrapper.find('.runtime-readiness-strip').exists()).toBe(true)
    expect(wrapper.find('#awd-ops-panel-inspector').exists()).toBe(false)
    expect(wrapper.find('#awd-ops-panel-instances').exists()).toBe(false)
  })

  it('运行内容为 round-inspector 时只显示轮次矩阵本体', () => {
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
        hideOperationTabs: true,
        operationPanel: 'inspector',
        runtimeContent: 'round-inspector',
      },
      global: {
        stubs: {
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
          AWDRoundInspector: {
            template: '<section id="awd-ops-panel-inspector">轮次矩阵</section>',
          },
          AWDRoundCreateDialog: true,
          AWDServiceCheckDialog: true,
          AWDAttackLogDialog: true,
          AWDChallengeConfigDialog: true,
        },
      },
    })

    expect(wrapper.find('.runtime-readiness-strip').exists()).toBe(false)
    expect(wrapper.get('#awd-ops-panel-inspector').text()).toBe('轮次矩阵')
    expect(wrapper.find('#awd-ops-panel-instances').exists()).toBe(false)
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

  it('未开赛实例内容应只显示实例编排，不重复显示开赛就绪摘要', () => {
    const awdState = getAwdState()
    awdState.readiness.value = buildReadinessState()
    awdState.instanceOrchestration.value = {
      contest_id: 'awd-1',
      teams: [
        {
          team_id: 'team-1',
          team_name: 'Red Team',
          captain_id: 'captain-1',
        },
      ],
      services: [
        {
          service_id: 'service-1',
          challenge_id: 'challenge-1',
          display_name: 'Web Service',
          is_visible: true,
        },
      ],
      instances: [],
    }

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
        hideOperationTabs: true,
        operationPanel: 'instances',
        runtimeContent: 'instances',
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

    expect(wrapper.get('#awd-ops-panel-instances').isVisible()).toBe(true)
    expect(wrapper.text()).toContain('Red Team')
    expect(wrapper.text()).not.toContain('开赛就绪摘要')
    expect(wrapper.text()).not.toContain('可开赛')
    expect(wrapper.text()).not.toContain('尚未进入运行阶段')
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
          AWDRoundCreateDialog: {
            props: ['open'],
            template:
              '<div v-if="open"><button id="awd-round-create-submit" type="button" @click="$emit(\'save\', { round_number: 1, status: \'pending\', attack_score: 10, defense_score: 20 })">提交</button></div>',
          },
          AWDRoundInspector: {
            template:
              '<div><button id="stub-open-create-round" type="button" @click="$emit(\'openCreateRoundDialog\')">创建轮次</button></div>',
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
          AWDRoundCreateDialog: {
            props: ['open'],
            template:
              '<div v-if="open"><button id="awd-round-create-submit" type="button" @click="$emit(\'save\', { round_number: 1, status: \'pending\', attack_score: 10, defense_score: 20 })">提交</button></div>',
          },
          AWDRoundInspector: {
            template:
              '<div><button id="stub-open-create-round" type="button" @click="$emit(\'openCreateRoundDialog\')">创建轮次</button></div>',
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
