import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import AWDOperationsPanel from '../contest/AWDOperationsPanel.vue'

const awdMockModule = vi.hoisted(() => ({
  state: null as any,
}))

vi.mock('@/composables/useAdminContestAWD', async () => {
  const { ref } = await vi.importActual<typeof import('vue')>('vue')
  awdMockModule.state = {
    rounds: ref([]),
    selectedRoundId: ref<string | null>(null),
    services: ref([]),
    attacks: ref([]),
    summary: ref(null),
    trafficSummary: ref(null),
    trafficEvents: ref([]),
    trafficEventsTotal: ref(0),
    trafficFilters: ref({}),
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

describe('AWDOperationsPanel', () => {
  beforeEach(() => {
    const awdState = getAwdState()
    awdState.challengeLinks.value = []
    awdState.challengeCatalog.value = []
    awdState.teams.value = []
    awdState.loadChallengeCatalog.mockReset()
    awdState.createChallengeLink.mockReset()
    awdState.updateChallengeLink.mockReset()
  })

  it('应该在没有 AWD 赛事时展示空态', () => {
    const wrapper = mount(AWDOperationsPanel, {
      props: {
        contests: [],
        selectedContestId: null,
      },
      global: {
        stubs: {
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
})
