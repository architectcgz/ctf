import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestEdit from '../ContestEdit.vue'
import type { ContestDetailData } from '@/api/contracts'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeState = vi.hoisted(() => ({
  params: { id: 'contest-1' } as Record<string, string>,
}))

const contestApiMocks = vi.hoisted(() => ({
  getContest: vi.fn(),
  updateContest: vi.fn(),
  getContestAWDReadiness: vi.fn(),
  listAdminAwdServiceTemplates: vi.fn(),
  listAdminContestChallenges: vi.fn(),
  listContestAWDServices: vi.fn(),
  getChallenges: vi.fn(),
  createContestAWDService: vi.fn(),
  updateContestAWDService: vi.fn(),
  updateAdminContestChallenge: vi.fn(),
}))

const awdMockModule = vi.hoisted(() => ({
  state: null as any,
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: replaceMock, back: vi.fn() }),
  }
})

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    getContest: contestApiMocks.getContest,
    updateContest: contestApiMocks.updateContest,
    getContestAWDReadiness: contestApiMocks.getContestAWDReadiness,
    listAdminAwdServiceTemplates: contestApiMocks.listAdminAwdServiceTemplates,
    listAdminContestChallenges: contestApiMocks.listAdminContestChallenges,
    listContestAWDServices: contestApiMocks.listContestAWDServices,
    getChallenges: contestApiMocks.getChallenges,
    createContestAWDService: contestApiMocks.createContestAWDService,
    updateContestAWDService: contestApiMocks.updateContestAWDService,
    updateAdminContestChallenge: contestApiMocks.updateAdminContestChallenge,
  }
})

vi.mock('@/composables/useToast', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
  }),
}))

vi.mock('@/composables/usePlatformContestAwd', async () => {
  const { ref } = await vi.importActual<typeof import('vue')>('vue')

  if (!awdMockModule.state) {
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
  }

  return {
    usePlatformContestAwd: () => awdMockModule.state,
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
        PlatformContestFormPanel: {
          props: ['draft'],
          template: '<div data-testid="platform-contest-form">{{ draft?.title }}</div>',
        },
        AWDChallengeConfigDialog: true,
        AWDChallengeConfigPanel: true,
        AWDOperationsPanel: true,
        ContestAwdPreflightPanel: true,
        ContestChallengeOrchestrationPanel: true,
        ContestWorkbenchStageRail: {
          props: ['stages'],
          template: '<div data-testid="stage-rail">{{ stages.map((item) => item.label).join(" / ") }}</div>',
        },
        AWDReadinessOverrideDialog: true,
      },
    },
  })
}

describe('ContestEdit', () => {
  beforeEach(() => {
    pushMock.mockReset()
    replaceMock.mockReset()
    Object.values(contestApiMocks).forEach((mock) => mock.mockReset())
    contestApiMocks.getContest.mockResolvedValue(buildContestDetail())
  })

  it('普通竞赛仍应保留编辑页入口', async () => {
    const wrapper = mountContestEdit()

    await flushPromises()

    expect(contestApiMocks.getContest).toHaveBeenCalledWith('contest-1')
    expect(wrapper.text()).toContain('2026 春季校园 CTF')
    expect(replaceMock).not.toHaveBeenCalled()
  })

  it('AWD 赛事应直接切到新的管理员工作台入口', async () => {
    contestApiMocks.getContest.mockResolvedValueOnce(
      buildContestDetail({
        title: '2026 AWD 联赛',
        mode: 'awd' as ContestDetailData['mode'],
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()

    expect(wrapper.text()).toContain('正在进入 AWD 工作台')
    expect(replaceMock).toHaveBeenCalledWith({
      name: 'AdminAwdOverview',
      params: { id: 'contest-1' },
    })
  })

  it('AWD 赛事进入编辑页时不应再拉取旧 AWD 工作台附属数据', async () => {
    contestApiMocks.getContest.mockResolvedValueOnce(
      buildContestDetail({
        title: '2026 AWD 联赛',
        mode: 'awd' as ContestDetailData['mode'],
      })
    )

    mountContestEdit()

    await flushPromises()

    expect(contestApiMocks.getContestAWDReadiness).not.toHaveBeenCalled()
    expect(contestApiMocks.listAdminContestChallenges).not.toHaveBeenCalled()
    expect(contestApiMocks.listContestAWDServices).not.toHaveBeenCalled()
  })
})
