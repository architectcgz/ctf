import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'

import AWDOperationsPanel from '../contest/AWDOperationsPanel.vue'

vi.mock('@/composables/useAdminContestAWD', () => ({
  useAdminContestAWD: () => ({
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
    loadingRounds: ref(false),
    loadingRoundDetail: ref(false),
    loadingTrafficSummary: ref(false),
    loadingTrafficEvents: ref(false),
    checking: ref(false),
    creatingRound: ref(false),
    savingServiceCheck: ref(false),
    savingAttackLog: ref(false),
    shouldAutoRefresh: ref(false),
    refresh: vi.fn(),
    applyTrafficFilters: vi.fn(),
    setTrafficPage: vi.fn(),
    resetTrafficFilters: vi.fn(),
    runSelectedRoundCheck: vi.fn(),
    createRound: vi.fn(),
    createServiceCheck: vi.fn(),
    createAttackLog: vi.fn(),
  }),
}))

describe('AWDOperationsPanel', () => {
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
        },
      },
    })

    expect(wrapper.text()).toContain('选择 AWD 赛事')
    expect(wrapper.text()).toContain('暂无 AWD 赛事')
    expect(wrapper.html()).not.toContain('a-w-d-round-inspector-stub')
  })
})
