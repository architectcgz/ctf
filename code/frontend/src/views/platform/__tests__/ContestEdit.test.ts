import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { computed, defineComponent, ref, watch } from 'vue'

import ContestEdit from '../ContestEdit.vue'
import AWDReadinessOverrideDialog from '@/components/platform/contest/AWDReadinessOverrideDialog.vue'
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
  listAdminAwdServiceTemplates: vi.fn(),
  listAdminContestChallenges: vi.fn(),
  listContestAWDServices: vi.fn(),
  getChallenges: vi.fn(),
  createContestAWDService: vi.fn(),
  deleteContestAWDService: vi.fn(),
  updateContestAWDService: vi.fn(),
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
    listAdminAwdServiceTemplates: contestApiMocks.listAdminAwdServiceTemplates,
    listAdminContestChallenges: contestApiMocks.listAdminContestChallenges,
    listContestAWDServices: contestApiMocks.listContestAWDServices,
    getChallenges: contestApiMocks.getChallenges,
    createContestAWDService: contestApiMocks.createContestAWDService,
    deleteContestAWDService: contestApiMocks.deleteContestAWDService,
    updateContestAWDService: contestApiMocks.updateContestAWDService,
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

vi.mock('@/composables/usePlatformContestAwd', async () => {
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
      service_id: '',
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

const AWDReadinessOverrideDialogStub = defineComponent({
  name: 'AWDReadinessOverrideDialog',
  props: {
    open: { type: Boolean, default: false },
    title: { type: String, default: '' },
    confirmLoading: { type: Boolean, default: false },
  },
  emits: ['update:open', 'confirm'],
  setup(props, { emit }) {
    const reason = ref('')

    watch(
      () => props.open,
      (open) => {
        if (!open) {
          reason.value = ''
        }
      },
      { immediate: true }
    )

    function handleSubmit() {
      emit('confirm', reason.value)
    }

    return { reason, handleSubmit }
  },
  template: `
    <div v-if="open">
      <div>{{ title }}</div>
      <textarea id="awd-readiness-override-reason" v-model="reason" />
      <button
        id="awd-readiness-override-submit"
        type="button"
        @click="handleSubmit"
      >
        {{ confirmLoading ? '强制继续中...' : '强制继续' }}
      </button>
    </div>
  `,
})

const ContestChallengeEditorDialogStub = defineComponent({
  name: 'ContestChallengeEditorDialog',
  props: {
    open: { type: Boolean, default: false },
    mode: { type: String, default: 'create' },
    contestMode: { type: String, default: 'jeopardy' },
    challengeOptions: { type: Array, default: () => [] },
    templateOptions: { type: Array, default: () => [] },
    existingChallengeIds: { type: Array, default: () => [] },
    draft: { type: Object, default: null },
    loadingChallengeCatalog: { type: Boolean, default: false },
    loadingTemplateCatalog: { type: Boolean, default: false },
    saving: { type: Boolean, default: false },
  },
  emits: ['update:open', 'save'],
  setup(props, { emit }) {
    const challengeId = ref('')
    const templateId = ref('')
    const points = ref('100')
    const order = ref('0')
    const isVisible = ref('true')

    const isAwdContest = computed(() => props.contestMode === 'awd')
    const selectableChallenges = computed(() =>
      (props.challengeOptions as Array<{ id: string }>).filter(
        (item) => props.mode === 'edit' || !(props.existingChallengeIds as string[]).includes(item.id)
      )
    )

    watch(
      () =>
        [props.open, props.mode, props.draft, selectableChallenges.value, props.templateOptions] as const,
      ([open]) => {
        if (!open) {
          return
        }

        challengeId.value =
          props.mode === 'edit'
            ? String((props.draft as { challenge_id?: string } | null)?.challenge_id ?? '')
            : String(selectableChallenges.value[0]?.id ?? '')
        templateId.value = isAwdContest.value
          ? String(
              (props.draft as { awd_template_id?: string } | null)?.awd_template_id ??
                (props.templateOptions as Array<{ id: string }>)[0]?.id ??
                ''
            )
          : ''
        points.value = String((props.draft as { points?: number } | null)?.points ?? 100)
        order.value = String((props.draft as { order?: number } | null)?.order ?? 0)
        isVisible.value =
          (props.draft as { is_visible?: boolean } | null)?.is_visible === false ? 'false' : 'true'
      },
      { immediate: true, deep: true }
    )

    function submit() {
      emit('save', {
        challenge_id: isAwdContest.value ? undefined : Number(challengeId.value),
        template_id: isAwdContest.value ? Number(templateId.value) : undefined,
        points: Number(points.value),
        order: Number(order.value),
        is_visible: isVisible.value === 'true',
      })
    }

    return { challengeId, templateId, points, order, isVisible, selectableChallenges, isAwdContest, submit }
  },
  template: `
    <div v-if="open">
      <select
        v-if="mode === 'create' && !isAwdContest"
        id="contest-challenge-select"
        v-model="challengeId"
        :disabled="loadingChallengeCatalog"
      >
        <option
          v-for="challenge in selectableChallenges"
          :key="challenge.id"
          :value="challenge.id"
        >
          {{ challenge.title }}
        </option>
      </select>
      <div v-else>{{ draft?.title }}</div>
      <select
        v-if="isAwdContest"
        id="contest-challenge-template"
        v-model="templateId"
        :disabled="loadingTemplateCatalog"
      >
        <option
          v-for="template in templateOptions"
          :key="template.id"
          :value="template.id"
        >
          {{ template.name }}
        </option>
      </select>
      <input id="contest-challenge-points" v-model="points" />
      <input id="contest-challenge-order" v-model="order" />
      <select id="contest-challenge-visibility" v-model="isVisible">
        <option value="true">可见</option>
        <option value="false">隐藏</option>
      </select>
      <button
        id="contest-challenge-dialog-submit"
        type="button"
        @click="submit"
      >
        {{ saving ? '保存中...' : mode === 'create' ? '关联题目' : '保存变更' }}
      </button>
    </div>
  `,
})

const AWDChallengeConfigDialogStub = defineComponent({
  name: 'AWDChallengeConfigDialog',
  props: {
    open: { type: Boolean, default: false },
    mode: { type: String, default: 'create' },
    challengeOptions: { type: Array, default: () => [] },
    templateOptions: { type: Array, default: () => [] },
    existingChallengeIds: { type: Array, default: () => [] },
    draft: { type: Object, default: null },
    loadingChallengeCatalog: { type: Boolean, default: false },
    loadingTemplateCatalog: { type: Boolean, default: false },
    saving: { type: Boolean, default: false },
  },
  emits: ['update:open', 'save'],
  setup(props, { emit }) {
    const challengeId = ref('')
    const templateId = ref('')
    const points = ref('100')
    const order = ref('0')
    const isVisible = ref('true')

    const selectableChallenges = computed(() =>
      (props.challengeOptions as Array<{ id: string }>).filter(
        (item) => props.mode === 'edit' || !(props.existingChallengeIds as string[]).includes(item.id)
      )
    )

    watch(
      () => [props.open, props.mode, props.draft, selectableChallenges.value, props.templateOptions] as const,
      ([open]) => {
        if (!open) {
          return
        }

        challengeId.value =
          props.mode === 'edit'
            ? String((props.draft as { challenge_id?: string } | null)?.challenge_id ?? '')
            : String(selectableChallenges.value[0]?.id ?? '')
        templateId.value = String(
          (props.draft as { awd_template_id?: string } | null)?.awd_template_id ??
            (props.templateOptions as Array<{ id: string }>)[0]?.id ??
            ''
        )
        points.value = String((props.draft as { points?: number } | null)?.points ?? 100)
        order.value = String((props.draft as { order?: number } | null)?.order ?? 0)
        isVisible.value =
          (props.draft as { is_visible?: boolean } | null)?.is_visible === false ? 'false' : 'true'
      },
      { immediate: true, deep: true }
    )

    function submit() {
      emit('save', {
        challenge_id: props.mode === 'edit' ? Number(challengeId.value) : undefined,
        template_id: Number(templateId.value),
        points: Number(points.value),
        order: Number(order.value),
        is_visible: isVisible.value === 'true',
      })
    }

    return { challengeId, templateId, points, order, isVisible, selectableChallenges, submit }
  },
  template: `
    <div v-if="open">
      <div>{{ mode === 'create' ? '新增 AWD 题库题目' : '编辑 AWD 题目配置' }}</div>
      <div v-if="mode === 'edit'">{{ draft?.title }}</div>
      <select
        id="awd-challenge-config-template"
        v-model="templateId"
        :disabled="loadingTemplateCatalog"
      >
        <option
          v-for="template in templateOptions"
          :key="template.id"
          :value="template.id"
        >
          {{ template.name }}
        </option>
      </select>
      <input id="awd-challenge-config-points" v-model="points" />
      <input id="awd-challenge-config-order" v-model="order" />
      <select id="awd-challenge-config-visible" v-model="isVisible">
        <option value="true">可见</option>
        <option value="false">隐藏</option>
      </select>
      <button
        id="awd-challenge-config-submit"
        type="button"
        @click="submit"
      >
        {{ saving ? '保存中...' : mode === 'create' ? '新增题目' : '保存配置' }}
      </button>
    </div>
  `,
})

function mountContestEdit() {
  return mount(ContestEdit, {
    global: {
      stubs: {
        AWDChallengeConfigDialog: AWDChallengeConfigDialogStub,
        AWDReadinessOverrideDialog: AWDReadinessOverrideDialogStub,
        ContestChallengeEditorDialog: ContestChallengeEditorDialogStub,
        AdminSurfaceModal: {
          props: ['open', 'title'],
          template:
            '<div><div v-if="open"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
        },
        SlideOverDrawer: {
          props: ['open', 'title'],
          template:
            '<div><div v-if="open"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
        },
        AdminSurfaceDrawer: {
          props: ['open', 'title'],
          template:
            '<div><div v-if="open"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
        },
        ElDialog: {
          props: ['modelValue', 'title'],
          template: '<div><div v-if="title">{{ title }}</div><slot /><slot name="footer" /></div>',
        },
      },
    },
  })
}

function mountContestEditWithRealChallengeDialog() {
  return mount(ContestEdit, {
    global: {
      stubs: {
        AWDChallengeConfigDialog: AWDChallengeConfigDialogStub,
        AWDReadinessOverrideDialog: AWDReadinessOverrideDialogStub,
        AdminSurfaceModal: {
          props: ['open', 'title'],
          template:
            '<div><div v-if="open"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
        },
        SlideOverDrawer: {
          props: ['open', 'title'],
          template:
            '<div><div v-if="open"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
        },
        AdminSurfaceDrawer: {
          props: ['open', 'title'],
          template:
            '<div><div v-if="open"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
        },
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

async function submitContestBasicsForm(wrapper: VueWrapper<any>) {
  await wrapper.get('.studio-form-canvas form').trigger('submit')
  await flushPromises()
}

async function openChallengeActionMenu(wrapper: VueWrapper<any>, challengeId = '101') {
  await wrapper.get(`#contest-challenge-actions-${challengeId}`).trigger('click')
  await flushPromises()
}

function getTeleportTarget<T extends Element = Element>(selector: string): T {
  const element = document.body.querySelector(selector)
  expect(element, `expected teleported element ${selector} to exist`).not.toBeNull()
  return element as T
}

describe('ContestEdit', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/platform/contests/contest-1/edit')
    pushMock.mockReset()
    contestApiMocks.getContest.mockReset()
    contestApiMocks.updateContest.mockReset()
    contestApiMocks.getContestAWDReadiness.mockReset()
    contestApiMocks.listAdminAwdServiceTemplates.mockReset()
    contestApiMocks.listAdminContestChallenges.mockReset()
    contestApiMocks.listContestAWDServices.mockReset()
    contestApiMocks.getChallenges.mockReset()
    contestApiMocks.createContestAWDService.mockReset()
    contestApiMocks.deleteContestAWDService.mockReset()
    contestApiMocks.updateContestAWDService.mockReset()
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
      service_id: '',
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
        awd_service_id: 'service-1',
        awd_template_id: '1',
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
    contestApiMocks.listAdminAwdServiceTemplates.mockResolvedValue({
      list: [
        {
          id: '1',
          name: 'Bank Portal AWD',
          slug: 'bank-portal-awd',
          category: 'web',
          difficulty: 'medium',
          description: 'bank target',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: 'v1',
          status: 'published',
          readiness_status: 'passed',
          created_by: '9',
          last_verified_at: '2026-03-01T00:00:00.000Z',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
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
    contestApiMocks.listContestAWDServices.mockResolvedValue([
      {
        id: 'service-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        template_id: '1',
        display_name: 'Web 入门',
        order: 1,
        is_visible: true,
        score_config: {
          points: 120,
          awd_sla_score: 0,
          awd_defense_score: 0,
        },
        runtime_config: {},
        checker_type: undefined,
        checker_config: {},
        sla_score: 0,
        defense_score: 0,
        validation_state: 'pending',
        last_preview_at: undefined,
        last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
        updated_at: '2026-03-10T00:00:00.000Z',
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
      awd_service_id: 'service-2',
      awd_template_id: '1',
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
    contestApiMocks.createContestAWDService.mockResolvedValue({
      id: 'service-2',
      contest_id: 'contest-1',
      challenge_id: '102',
      template_id: '1',
      display_name: 'Crypto 进阶',
      order: 3,
      is_visible: false,
      score_config: {},
      runtime_config: {},
      created_at: '2026-03-10T01:00:00.000Z',
      updated_at: '2026-03-10T01:00:00.000Z',
    })
    contestApiMocks.deleteContestAWDService.mockResolvedValue(undefined)
    contestApiMocks.updateContestAWDService.mockResolvedValue(undefined)
    contestApiMocks.updateAdminContestChallenge.mockResolvedValue(undefined)
    contestApiMocks.deleteAdminContestChallenge.mockResolvedValue(undefined)
    destructiveConfirmMock.mockResolvedValue(true)
  })

  it('应该在普通赛下只展示基础信息与题目池阶段', async () => {
    contestApiMocks.getContest.mockResolvedValue(buildContestDetail())

    const wrapper = mountContestEditWithRealChallengeDialog()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.text()).toContain('基础信息')
    expect(stageRail.text()).toContain('题目编排')
    expect(stageRail.text()).not.toContain('AWD 服务配置')
    expect(stageRail.text()).not.toContain('就绪审计')
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

    const wrapper = mountContestEdit()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.text()).toContain('基础信息')
    expect(stageRail.text()).toContain('题目编排')
    expect(stageRail.text()).toContain('AWD 服务配置')
    expect(stageRail.text()).toContain('就绪审计')
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

    expect(wrapper.text()).toContain('强制启动赛事')
    expect(wrapper.text()).toContain('Challenge 101')
    expect(wrapper.text()).toContain('修正配置')
    expect(wrapper.text()).toContain('强制放行')

    await wrapper.get('#awd-readiness-edit-101').trigger('click')
    await flushPromises()

    expect(getWorkbenchStageRail(wrapper).get('[role="tab"][aria-selected="true"]').text()).toContain('AWD 服务配置')
    expect(wrapper.text()).toContain('正在编辑')
    expect(wrapper.text()).toContain('Web 入门')
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
    await openChallengeActionMenu(wrapper)
    getTeleportTarget<HTMLButtonElement>('#contest-challenge-open-awd-config-101').click()
    await flushPromises()

    expect(getWorkbenchStageRail(wrapper).get('[role="tab"][aria-selected="true"]').text()).toContain('AWD 服务配置')
    expect(wrapper.text()).toContain('正在编辑')
    expect(wrapper.text()).toContain('Web 入门')
  })

  it('应该支持从轮次运行跳转到 AWD 配置并保留当前题高亮', async () => {
    awdMockModule.state.readiness.value = {
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
      blocking_actions: ['create_round'],
      items: [
        {
          challenge_id: '101',
          title: 'Web 入门',
          checker_type: 'http_standard',
          validation_state: 'failed',
          last_preview_at: '2026-04-12T08:00:00.000Z',
          last_access_url: 'http://checker.internal/flag',
          blocking_reason: 'last_preview_failed',
        },
      ],
    }
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
    await wrapper.get('#contest-workbench-stage-tab-operations').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-readiness-edit-101').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('正在编辑')
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

    expect(wrapper.text()).toContain('运维就绪审计')
    expect(wrapper.text()).toContain('开赛前必须修正以下阻塞项')
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
    contestApiMocks.listContestAWDServices
      .mockResolvedValueOnce([
        {
          id: 'service-1',
          contest_id: 'contest-1',
          challenge_id: '101',
          template_id: '1',
          title: 'Web 入门',
          category: 'web',
          difficulty: 'easy',
          display_name: 'Web 入门',
          order: 1,
          is_visible: true,
          score_config: {
            points: 120,
            awd_sla_score: 0,
            awd_defense_score: 0,
          },
          runtime_config: {},
          checker_type: undefined,
          checker_config: {},
          sla_score: 0,
          defense_score: 0,
          validation_state: 'pending',
          last_preview_at: undefined,
          last_preview_result: undefined,
          created_at: '2026-03-10T00:00:00.000Z',
          updated_at: '2026-03-10T00:00:00.000Z',
        },
      ])
      .mockRejectedValueOnce(new Error('refresh failed'))

    const wrapper = mountContestEdit()

    await flushPromises()

    expect(wrapper.text()).toContain('已关联题目')
    expect(wrapper.text()).toContain('Web 入门')

    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('同步数据'))
      ?.trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('refresh failed')
    expect(wrapper.text()).toContain('Web 入门')
    expect(wrapper.text()).toContain('已关联题目')
    expect(wrapper.text()).not.toContain('当前竞赛还没有关联题目')
    expect(wrapper.text()).not.toContain('共 0 道题目')
  })

  it('应该在管理页工作台交接时强制落到轮次态势而不是恢复旧子页签', async () => {
    window.sessionStorage.setItem('ctf_admin_awd_ops_panel:contest-1', 'challenges')
    window.history.replaceState({}, '', '/platform/contests/contest-1/edit?panel=operations&opsPanel=inspector')
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

    expect(wrapper.find('#awd-ops-tab-challenges').exists()).toBe(false)
    expect(getWorkbenchStageRail(wrapper).get('[role="tab"][aria-selected="true"]').text()).toContain('轮次运行')
    expect(wrapper.text()).toContain('轮次态势')
  })

  it('应该在 URL 已指定有效阶段时保留该阶段', async () => {
    window.history.replaceState({}, '', '/platform/contests/contest-1/edit?panel=operations')
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
    const wrapper = mountContestEdit()

    await flushPromises()

    expect(wrapper.text()).toContain('基础信息')

    await wrapper.get('#contest-title').setValue('2026 春季校园 CTF（更新）')
    await submitContestBasicsForm(wrapper)

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
    await submitContestBasicsForm(wrapper)

    expect(destructiveConfirmMock).toHaveBeenCalledWith(
      expect.objectContaining({
        title: '确认结束赛事',
      })
    )
    expect(contestApiMocks.updateContest).not.toHaveBeenCalled()
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('应该在进行中的竞赛切换为已冻结时省略不可修改的时间字段', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 春季校园 CTF',
        status: 'running',
      })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-status').setValue('frozen')
    await submitContestBasicsForm(wrapper)

    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.objectContaining({
        title: '2026 春季校园 CTF',
        status: 'frozen',
      })
    )
    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.not.objectContaining({
        starts_at: expect.anything(),
      })
    )
    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.not.objectContaining({
        ends_at: expect.anything(),
      })
    )
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

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-basics').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-status').setValue('running')
    await submitContestBasicsForm(wrapper)

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
    await wrapper.get('#contest-workbench-stage-tab-basics').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-status').setValue('running')
    await submitContestBasicsForm(wrapper)

    expect(toastMocks.error).toHaveBeenCalledWith('readiness fetch failed')
    expect(contestApiMocks.updateContest).toHaveBeenCalledTimes(1)
    expect(pushMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('基础信息')
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
    await wrapper.get('#contest-workbench-stage-tab-basics').trigger('click')
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

    expect(wrapper.text()).toContain('基础信息')
    expect(wrapper.text()).toContain('基础信息')
    expect(wrapper.text()).not.toContain('竞赛详情加载失败')
  })

  it('应该为 AWD 配置按 published 状态加载服务模板目录', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminAwdServiceTemplates.mockResolvedValueOnce({
      list: [
        {
          id: '999',
          name: 'Final Template',
          slug: 'final-template',
          category: 'crypto',
          difficulty: 'medium',
          description: 'final service',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          created_at: '2026-03-02T00:00:00.000Z',
          updated_at: '2026-03-02T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
    })

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-create').trigger('click')
    await flushPromises()

    expect(contestApiMocks.listAdminAwdServiceTemplates).toHaveBeenCalledWith({
      page: 1,
      page_size: 100,
      status: 'published',
    })
    expect(wrapper.text()).toContain('Final Template')
  })

  it('应该在 AWD 配置服务模板加载失败时给出错误提示而不是留下未处理异常', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminAwdServiceTemplates.mockRejectedValueOnce(new Error('catalog failed'))

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-create').trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('catalog failed')
    expect(wrapper.text()).toContain('新增 AWD 题库题目')
  })

  it('应该在 AWD 辅助数据仍在加载时显示阶段级加载提示而不是空态', async () => {
    const servicesDeferred = createDeferred<any[]>()

    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listContestAWDServices.mockReturnValueOnce(servicesDeferred.promise)

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('正在同步 AWD 配置...')
    expect(wrapper.text()).not.toContain('当前赛事还没有关联题目')

    servicesDeferred.resolve([
      {
        id: 'service-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        template_id: '1',
        title: 'Web 入门',
        category: 'web',
        difficulty: 'easy',
        display_name: 'Web 入门',
        order: 1,
        is_visible: true,
        score_config: {
          points: 120,
          awd_sla_score: 0,
          awd_defense_score: 0,
        },
        runtime_config: {},
        checker_type: undefined,
        checker_config: {},
        sla_score: 0,
        defense_score: 0,
        validation_state: 'pending',
        last_preview_at: undefined,
        last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
        updated_at: '2026-03-10T00:00:00.000Z',
      },
    ])
    await flushPromises()
  })

  it('应该允许管理员在竞赛编辑页编排题目', async () => {
    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()

    expect(contestApiMocks.listAdminContestChallenges).toHaveBeenCalledWith('contest-1')
    expect(wrapper.text()).toContain('题目编排')
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

    await openChallengeActionMenu(wrapper)
    getTeleportTarget<HTMLButtonElement>('#contest-challenge-edit-101').click()
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

    await openChallengeActionMenu(wrapper)
    getTeleportTarget<HTMLButtonElement>('#contest-challenge-remove-101').click()
    await flushPromises()

    expect(destructiveConfirmMock).toHaveBeenCalled()
    expect(contestApiMocks.deleteAdminContestChallenge).toHaveBeenCalledWith('contest-1', '101')
  })

  it('题目池变更后应同步更新 AWD 配置与赛前检查数据', async () => {
    const awdServicesState: any[] = []
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listContestAWDServices.mockImplementation(async () =>
      awdServicesState.map((item) => ({ ...item }))
    )
    contestApiMocks.getContestAWDReadiness.mockImplementation(async () => ({
      contest_id: 'contest-1',
      ready: awdServicesState.length > 0,
      total_challenges: awdServicesState.length,
      passed_challenges: awdServicesState.length,
      pending_challenges: 0,
      failed_challenges: 0,
      stale_challenges: 0,
      missing_checker_challenges: 0,
      blocking_count: 0,
      global_blocking_reasons: awdServicesState.length > 0 ? [] : ['no_challenges'],
      blocking_actions: awdServicesState.length > 0 ? [] : ['start_contest'],
      items: [],
    }))
    contestApiMocks.listAdminAwdServiceTemplates.mockResolvedValue({
      list: [
        {
          id: '11',
          name: 'Upload HTTP 模板',
          slug: 'upload-http',
          category: 'web',
          difficulty: 'medium',
          description: 'http service',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
    })
    contestApiMocks.createContestAWDService.mockImplementation(async (_contestId, payload) => {
      expect(payload).toEqual({
        template_id: 11,
        points: 160,
        order: 3,
        is_visible: true,
      })
      awdServicesState.push({
        id: 'service-2',
        contest_id: 'contest-1',
        challenge_id: '102',
        template_id: '11',
        title: 'Upload Service',
        category: 'web',
        difficulty: 'medium',
        display_name: 'Upload Service',
        order: 3,
        is_visible: true,
        score_config: {
          points: payload.points,
          awd_sla_score: 0,
          awd_defense_score: 0,
        },
        runtime_config: {},
        checker_type: undefined,
        checker_config: {},
        sla_score: 0,
        defense_score: 0,
        validation_state: 'pending',
        last_preview_at: undefined,
        last_preview_result: undefined,
        created_at: '2026-03-10T01:00:00.000Z',
        updated_at: '2026-03-10T01:00:00.000Z',
      })
      return awdServicesState[0]
    })

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-template').setValue('11')
    await wrapper.get('#contest-challenge-points').setValue('160')
    await wrapper.get('#contest-challenge-order').setValue('3')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Upload Service')

    await wrapper.get('#contest-workbench-stage-tab-preflight').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('环境已就绪')
    expect(wrapper.text()).not.toContain('当前赛事还没有关联题目，无法执行开赛关键动作')
  })

  it('AWD 题目从题目池移除时应删除显式 service 而不是只删 challenge 关联', async () => {
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
    await openChallengeActionMenu(wrapper)
    getTeleportTarget<HTMLButtonElement>('#contest-challenge-remove-101').click()
    await flushPromises()

    expect(contestApiMocks.deleteContestAWDService).toHaveBeenCalledWith('contest-1', 'service-1')
    expect(contestApiMocks.deleteAdminContestChallenge).not.toHaveBeenCalled()
  })

  it('AWD 配置变更后题目池应同步更新，并允许继续打开新增对话框', async () => {
    const awdServicesState: any[] = [
      {
        id: 'service-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        template_id: '1',
        title: 'Web 入门',
        category: 'web',
        difficulty: 'easy',
        display_name: 'Web 入门',
        order: 1,
        is_visible: true,
        score_config: {
          points: 120,
          awd_sla_score: 0,
          awd_defense_score: 0,
        },
        runtime_config: {},
        checker_type: undefined,
        checker_config: {},
        sla_score: 0,
        defense_score: 0,
        validation_state: 'pending',
        last_preview_at: undefined,
        last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
        updated_at: '2026-03-10T00:00:00.000Z',
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
    contestApiMocks.listContestAWDServices.mockImplementation(async () =>
      awdServicesState.map((item) => ({ ...item }))
    )
    contestApiMocks.createContestAWDService.mockImplementation(async (_contestId, payload) => {
      expect(payload).toEqual({
        template_id: 1,
        points: 160,
        order: 2,
        is_visible: true,
      })
      const created = {
        id: 'service-2',
        contest_id: 'contest-1',
        challenge_id: '102',
        template_id: '1',
        title: 'Crypto 进阶',
        category: 'crypto',
        difficulty: 'medium',
        display_name: 'Crypto 进阶',
        order: 2,
        is_visible: true,
        score_config: {
          points: 160,
          awd_sla_score: 0,
          awd_defense_score: 0,
        },
        runtime_config: {},
        created_at: '2026-03-10T01:00:00.000Z',
        updated_at: '2026-03-10T01:00:00.000Z',
      }
      awdServicesState.push(created)
      return created
    })

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-create').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-template').setValue('1')
    await wrapper.get('#awd-challenge-config-points').setValue('160')
    await wrapper.get('#awd-challenge-config-order').setValue('2')
    await wrapper.get('#awd-challenge-config-submit').trigger('click')
    await flushPromises()

    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Crypto 进阶')

    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    expect(wrapper.find('#contest-challenge-select').exists()).toBe(false)
    expect(wrapper.find('#contest-challenge-template').exists()).toBe(true)
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
    contestApiMocks.createContestAWDService.mockRejectedValueOnce(new Error('save failed'))

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-create').trigger('click')
    await flushPromises()
    await wrapper.get('#awd-challenge-config-template').setValue('1')
    await wrapper.get('#awd-challenge-config-points').setValue('160')
    await wrapper.get('#awd-challenge-config-order').setValue('2')
    await wrapper.get('#awd-challenge-config-submit').trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('save failed')
    expect(wrapper.text()).toContain('新增 AWD 题库题目')
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
    contestApiMocks.listContestAWDServices.mockResolvedValue([
      {
        id: 'service-1',
        contest_id: 'contest-1',
        challenge_id: '101',
        template_id: '1',
        display_name: 'Web 入门',
        order: 1,
        is_visible: true,
        score_config: {
          points: 120,
          awd_sla_score: 18,
          awd_defense_score: 28,
        },
        runtime_config: {
          checker_type: 'http_standard',
          checker_config: {},
        },
        checker_type: 'http_standard',
        checker_config: {},
        sla_score: 18,
        defense_score: 28,
        validation_state: 'stale',
        last_preview_at: '2026-04-12T08:00:00.000Z',
        last_preview_result: undefined,
        created_at: '2026-03-10T00:00:00.000Z',
        updated_at: '2026-03-10T00:00:00.000Z',
      },
    ])

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('未配置 AWD')
    expect(wrapper.text()).toContain('预检失败')
    expect(wrapper.text()).toContain('Checker')
    expect(wrapper.text()).toContain('S:18 D:28')
    expect(wrapper.text()).toContain('待重新验证')
  })
})
