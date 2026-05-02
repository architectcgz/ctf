import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { computed, defineComponent, ref, watch } from 'vue'

import ContestEdit from '../ContestEdit.vue'
import contestEditSource from '../ContestEdit.vue?raw'
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
  listAdminAwdChallenges: vi.fn(),
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

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: vi.fn(), back: vi.fn() }),
  }
})

vi.mock('@/api/admin/contests', async () => {
  const actual =
    await vi.importActual<typeof import('@/api/admin/contests')>('@/api/admin/contests')
  return {
    ...actual,
    getContest: contestApiMocks.getContest,
    updateContest: contestApiMocks.updateContest,
    getContestAWDReadiness: contestApiMocks.getContestAWDReadiness,
    listAdminContestChallenges: contestApiMocks.listAdminContestChallenges,
    listContestAWDServices: contestApiMocks.listContestAWDServices,
    createContestAWDService: contestApiMocks.createContestAWDService,
    deleteContestAWDService: contestApiMocks.deleteContestAWDService,
    updateContestAWDService: contestApiMocks.updateContestAWDService,
    createAdminContestChallenge: contestApiMocks.createAdminContestChallenge,
    updateAdminContestChallenge: contestApiMocks.updateAdminContestChallenge,
    deleteAdminContestChallenge: contestApiMocks.deleteAdminContestChallenge,
  }
})
vi.mock('@/api/admin/awd-authoring', () => ({
  listAdminAwdChallenges: contestApiMocks.listAdminAwdChallenges,
}))
vi.mock('@/api/admin/authoring', () => ({
  getChallenges: contestApiMocks.getChallenges,
}))

vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: destructiveConfirmMock,
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

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

const ContestChallengeEditorDialogStub = defineComponent({
  name: 'ContestChallengeEditorDialog',
  props: {
    open: { type: Boolean, default: false },
    mode: { type: String, default: 'create' },
    contestMode: { type: String, default: 'jeopardy' },
    challengeOptions: { type: Array, default: () => [] },
    awdChallengeOptions: { type: Array, default: () => [] },
    awdChallengeKeyword: { type: String, default: '' },
    awdChallengeServiceType: { type: String, default: '' },
    awdChallengeDeploymentMode: { type: String, default: '' },
    awdChallengeReadiness: { type: String, default: '' },
    existingChallengeIds: { type: Array, default: () => [] },
    draft: { type: Object, default: null },
    loadingChallengeCatalog: { type: Boolean, default: false },
    loadingAwdChallengeCatalog: { type: Boolean, default: false },
    saving: { type: Boolean, default: false },
  },
  emits: [
    'update:open',
    'save',
    'update-awd-challenge-keyword',
    'update-awd-challenge-service-type',
    'update-awd-challenge-deployment-mode',
    'update-awd-challenge-readiness',
    'change-awd-challenge-page',
    'refresh-awd-challenge-catalog',
  ],
  setup(props, { emit }) {
    const challengeId = ref('')
    const awdChallengeId = ref('')
    const awdChallengeIds = ref<string[]>([])
    const points = ref('100')
    const order = ref('0')
    const isVisible = ref('true')

    const isAwdContest = computed(() => props.contestMode === 'awd')
    const isAwdCreateMode = computed(() => isAwdContest.value && props.mode === 'create')
    const selectableChallenges = computed(() =>
      (props.challengeOptions as Array<{ id: string }>).filter(
        (item) =>
          props.mode === 'edit' || !(props.existingChallengeIds as string[]).includes(item.id)
      )
    )

    watch(
      () =>
        [
          props.open,
          props.mode,
          props.draft,
          selectableChallenges.value,
          props.awdChallengeOptions,
        ] as const,
      ([open]) => {
        if (!open) {
          return
        }

        challengeId.value =
          props.mode === 'edit'
            ? String((props.draft as { challenge_id?: string } | null)?.challenge_id ?? '')
            : String(selectableChallenges.value[0]?.id ?? '')
        awdChallengeId.value = isAwdContest.value
          ? String(
              (props.draft as { awd_challenge_id?: string } | null)?.awd_challenge_id ??
                (props.awdChallengeOptions as Array<{ id: string }>)[0]?.id ??
                ''
            )
          : ''
        awdChallengeIds.value =
          isAwdCreateMode.value && awdChallengeId.value ? [awdChallengeId.value] : []
        points.value = String((props.draft as { points?: number } | null)?.points ?? 100)
        order.value = String((props.draft as { order?: number } | null)?.order ?? 0)
        isVisible.value =
          (props.draft as { is_visible?: boolean } | null)?.is_visible === false ? 'false' : 'true'
      },
      { immediate: true, deep: true }
    )

    function submit() {
      emit('save', {
        challenge_id: isAwdContest.value
          ? undefined
          : challengeId.value
            ? Number(challengeId.value)
            : undefined,
        awd_challenge_id: isAwdContest.value ? Number(awdChallengeId.value) : undefined,
        awd_challenge_ids: isAwdCreateMode.value
          ? awdChallengeIds.value.map((id) => Number(id))
          : undefined,
        points: Number(points.value),
        order: Number(order.value),
        is_visible: isVisible.value === 'true',
      })
    }

    function selectAwdChallenge(id: string) {
      if (isAwdCreateMode.value) {
        const selected = new Set(awdChallengeIds.value)
        if (selected.has(id)) {
          if (selected.size > 1) {
            selected.delete(id)
          }
        } else {
          selected.add(id)
        }
        awdChallengeIds.value = (props.awdChallengeOptions as Array<{ id: string }>)
          .map((item) => item.id)
          .filter((itemId) => selected.has(itemId))
        awdChallengeId.value = awdChallengeIds.value[0] ?? ''
        return
      }
      awdChallengeId.value = id
    }

    return {
      challengeId,
      awdChallengeId,
      awdChallengeIds,
      points,
      order,
      isVisible,
      selectableChallenges,
      isAwdContest,
      isAwdCreateMode,
      selectAwdChallenge,
      submit,
    }
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
      <div
        v-if="isAwdCreateMode"
        id="contest-awd-challenge-list"
      >
        <button
          v-for="template in awdChallengeOptions"
          :id="'contest-awd-challenge-option-' + template.id"
          :key="template.id"
          type="button"
          :class="{ 'is-selected': isAwdCreateMode ? awdChallengeIds.includes(template.id) : awdChallengeId === template.id }"
          @click="selectAwdChallenge(template.id)"
        >
          {{ template.name }}
        </button>
      </div>
      <input v-if="!isAwdCreateMode" id="contest-challenge-points" v-model="points" />
      <input v-if="!isAwdCreateMode" id="contest-challenge-order" v-model="order" />
      <select v-if="!isAwdCreateMode" id="contest-challenge-visibility" v-model="isVisible">
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
    awdChallengeOptions: { type: Array, default: () => [] },
    existingChallengeIds: { type: Array, default: () => [] },
    draft: { type: Object, default: null },
    loadingChallengeCatalog: { type: Boolean, default: false },
    loadingAwdChallengeCatalog: { type: Boolean, default: false },
    saving: { type: Boolean, default: false },
  },
  emits: ['update:open', 'save'],
  setup(props, { emit }) {
    const challengeId = ref('')
    const awdChallengeId = ref('')
    const points = ref('100')
    const order = ref('0')
    const isVisible = ref('true')

    const selectableChallenges = computed(() =>
      (props.challengeOptions as Array<{ id: string }>).filter(
        (item) =>
          props.mode === 'edit' || !(props.existingChallengeIds as string[]).includes(item.id)
      )
    )

    watch(
      () =>
        [
          props.open,
          props.mode,
          props.draft,
          selectableChallenges.value,
          props.awdChallengeOptions,
        ] as const,
      ([open]) => {
        if (!open) {
          return
        }

        challengeId.value =
          props.mode === 'edit'
            ? String((props.draft as { challenge_id?: string } | null)?.challenge_id ?? '')
            : String(selectableChallenges.value[0]?.id ?? '')
        awdChallengeId.value = String(
          (props.draft as { awd_challenge_id?: string } | null)?.awd_challenge_id ??
            (props.awdChallengeOptions as Array<{ id: string }>)[0]?.id ??
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
        awd_challenge_id: Number(awdChallengeId.value),
        points: Number(points.value),
        order: Number(order.value),
        is_visible: isVisible.value === 'true',
      })
    }

    return { challengeId, awdChallengeId, points, order, isVisible, selectableChallenges, submit }
  },
  template: `
    <div v-if="open">
      <div>{{ mode === 'create' ? '新增 AWD 题库题目' : '编辑 AWD 题目配置' }}</div>
      <div v-if="mode === 'edit'">{{ draft?.title }}</div>
      <select
        id="awd-challenge-config-template"
        v-model="awdChallengeId"
        :disabled="loadingAwdChallengeCatalog"
      >
        <option
          v-for="template in awdChallengeOptions"
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
        RouterLink: {
          props: ['to'],
          template: '<a><slot /></a>',
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
        RouterLink: {
          props: ['to'],
          template: '<a><slot /></a>',
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

describe('ContestEdit', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/platform/contests/contest-1/edit')
    pushMock.mockReset()
    contestApiMocks.getContest.mockReset()
    contestApiMocks.updateContest.mockReset()
    contestApiMocks.getContestAWDReadiness.mockReset()
    contestApiMocks.listAdminAwdChallenges.mockReset()
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
      blocking_actions: ['start_contest', 'create_round', 'run_current_round_check'],
      items: [
        {
          awd_challenge_id: '1',
          title: 'Challenge 101',
          checker_type: 'http_standard',
          validation_state: 'failed',
          last_preview_at: '2026-04-12T08:00:00.000Z',
          last_access_url: 'http://checker.internal/flag',
          blocking_reason: 'last_preview_failed',
        },
      ],
    })
    contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
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
        awd_challenge_id: '1',
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
      awd_challenge_id: '1',
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
      awd_challenge_id: '1',
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

  it('路由页应仅负责组合，不直接耦合竞赛编辑加载与保存流程', () => {
    expect(contestEditSource).toContain('useContestEditPage')
    expect(contestEditSource).not.toContain("from '@/api/admin/contests'")
  })

  it('顶部应提供公告入口并跳转到单场公告管理页', async () => {
    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-open-announcements').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'ContestAnnouncements',
      params: { id: 'contest-1' },
    })
  })

  it('应该在普通赛下只展示基础信息与题目池阶段', async () => {
    contestApiMocks.getContest.mockResolvedValue(buildContestDetail())

    const wrapper = mountContestEditWithRealChallengeDialog()

    await flushPromises()

    const stageRail = getWorkbenchStageRail(wrapper)

    expect(stageRail.text()).toContain('基础信息')
    expect(stageRail.text()).toContain('题目编排')
    expect(stageRail.text()).not.toContain('AWD 编排')
    expect(stageRail.text()).not.toContain('就绪审计')
    expect(stageRail.text()).not.toContain('轮次运行')
  })

  it('应该在 AWD 赛事下只展示编辑工作台阶段，不混入赛事运维阶段', async () => {
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
    expect(stageRail.text()).toContain('AWD 编排')
    expect(stageRail.text()).toContain('就绪审计')
    expect(stageRail.text()).not.toContain('轮次运行')
    expect(stageRail.text()).not.toContain('实例编排')
  })

  it('应该在赛前检查中列出阻塞项、移除强制开赛入口，并支持返回 AWD 配置', async () => {
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

    expect(wrapper.text()).toContain('开赛已锁定')
    expect(wrapper.text()).toContain('开赛门禁')
    expect(wrapper.text()).toContain('轮次创建')
    expect(wrapper.text()).toContain('即时巡检')
    expect(wrapper.text()).toContain('Challenge 101')
    expect(wrapper.text()).toContain('编辑并试跑')
    expect(wrapper.find('#contest-awd-preflight-force-start').exists()).toBe(false)

    await wrapper.get('#awd-readiness-edit-1').trigger('click')
    await flushPromises()

    expect(pushMock).toHaveBeenCalledWith({
      name: 'ContestAWDConfig',
      params: { id: 'contest-1' },
      query: { service: 'service-1' },
    })
    expect(wrapper.text()).not.toContain('当前焦点题目')
  })

  it('题目池更多菜单不应提供 AWD 编排跳转入口', async () => {
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

    expect(wrapper.find('#contest-challenge-actions-1').exists()).toBe(false)
    expect(document.body.querySelector('#contest-challenge-open-awd-config-1')).toBeNull()
    expect(
      getWorkbenchStageRail(wrapper).get('[role="tab"][aria-selected="true"]').text()
    ).toContain('题目编排')
  })

  it('竞赛编辑页不应渲染轮次运行面板和赛事运维内容', async () => {
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

    expect(wrapper.find('#contest-workbench-stage-tab-operations').exists()).toBe(false)
    expect(wrapper.find('#contest-workbench-stage-tab-instances').exists()).toBe(false)
    expect(wrapper.find('#awd-readiness-edit-101').exists()).toBe(false)
    expect(wrapper.find('.runtime-readiness-strip').exists()).toBe(false)
    expect(wrapper.find('#awd-ops-panel-inspector').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('本轮得分')
    expect(wrapper.text()).not.toContain('攻击流水')
  })

  it('应该在 AWD 赛事已开赛时默认聚焦 AWD 编排阶段', async () => {
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

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('AWD 编排')
    expect(wrapper.text()).not.toContain('轮次态势')
  })

  it('旧 operations URL 在编辑页应回落到默认编辑阶段', async () => {
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

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('题目编排')
    expect(wrapper.find('#contest-workbench-stage-tab-operations').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('尚未进入运行阶段')
  })

  it('AWD 赛事已结束时仍停留在编辑配置阶段，报告导出进入赛事运维页处理', async () => {
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

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('AWD 编排')
    expect(wrapper.text()).not.toContain('轮次态势')
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
          awd_challenge_id: '1',
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

    expect(wrapper.text()).toContain('Web 入门')

    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('同步数据'))
      ?.trigger('click')
    await flushPromises()

    expect(toastMocks.error).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Web 入门')
    expect(wrapper.text()).not.toContain('当前竞赛还没有关联题目')
    expect(wrapper.text()).not.toContain('共 0 道题目')
  })

  it('应该在管理页工作台交接时忽略旧运维子页签并落到编辑阶段', async () => {
    window.sessionStorage.setItem('ctf_admin_awd_ops_panel:contest-1', 'challenges')
    window.history.replaceState(
      {},
      '',
      '/platform/contests/contest-1/edit?panel=operations&opsPanel=inspector'
    )
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
    expect(wrapper.find('#contest-workbench-stage-tab-operations').exists()).toBe(false)
    expect(
      getWorkbenchStageRail(wrapper).get('[role="tab"][aria-selected="true"]').text()
    ).toContain('AWD 编排')
    expect(wrapper.text()).not.toContain('轮次态势')
  })

  it('旧 operations URL 不再作为编辑页有效阶段保留', async () => {
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

    expect(stageRail.get('[role="tab"][aria-selected="true"]').text()).toContain('题目编排')
    expect(window.location.search).not.toContain('panel=operations')
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
  it('应该在 AWD 启动门禁拦截后展示错误并停留在当前页面', async () => {
    contestApiMocks.getContest.mockResolvedValue({
      id: 'contest-1',
      title: '2026 AWD 联赛',
      description: '攻防赛',
      mode: 'awd',
      status: 'registering',
      starts_at: '2026-03-15T09:00:00.000Z',
      ends_at: '2026-03-15T13:00:00.000Z',
    })
    contestApiMocks.updateContest.mockRejectedValueOnce(
      new ApiError('AWD 开赛就绪检查未通过', { status: 409, code: 14025 })
    )

    const wrapper = mountContestEdit()

    await flushPromises()
    contestApiMocks.getContestAWDReadiness.mockClear()
    await wrapper.get('#contest-workbench-stage-tab-basics').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-status').setValue('running')
    await submitContestBasicsForm(wrapper)

    expect(contestApiMocks.getContestAWDReadiness).not.toHaveBeenCalled()
    expect(contestApiMocks.updateContest).toHaveBeenCalledTimes(1)
    expect(contestApiMocks.updateContest).toHaveBeenCalledWith(
      'contest-1',
      expect.objectContaining({ status: 'running' })
    )
    expect(toastMocks.error).toHaveBeenCalledWith('AWD 开赛就绪检查未通过')
    expect(pushMock).not.toHaveBeenCalled()
    expect(wrapper.find('#awd-readiness-override-submit').exists()).toBe(false)
  })

  it('应该在 AWD 启动门禁拦截后不再读取 readiness 放行数据', async () => {
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

    const wrapper = mountContestEdit()

    await flushPromises()
    contestApiMocks.getContestAWDReadiness.mockClear()
    await wrapper.get('#contest-workbench-stage-tab-basics').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-status').setValue('running')
    await submitContestBasicsForm(wrapper)

    expect(contestApiMocks.getContestAWDReadiness).not.toHaveBeenCalled()
    expect(contestApiMocks.updateContest).toHaveBeenCalledTimes(1)
    expect(toastMocks.error).toHaveBeenCalledWith('AWD 开赛就绪检查未通过')
    expect(pushMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('基础信息')
  })

  it('赛前检查页面不应提供强制开赛入口', async () => {
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

    expect(wrapper.find('#contest-awd-preflight-force-start').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('强制启动赛事')
    expect(wrapper.text()).not.toContain('强制放行')
    expect(contestApiMocks.updateContest).not.toHaveBeenCalled()
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

  it('新增 AWD 题目应统一从题目编排打开题库选题弹层', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminAwdChallenges.mockResolvedValueOnce({
      list: [
        {
          id: '999',
          name: 'Final Challenge',
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
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    expect(wrapper.find('#awd-challenge-config-create').exists()).toBe(false)
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()

    expect(wrapper.get('#contest-workbench-stage-tab-pool').attributes('aria-selected')).toBe(
      'true'
    )
    expect(wrapper.find('#contest-challenge-library').exists()).toBe(false)
    expect(wrapper.find('#contest-challenge-select').exists()).toBe(false)
    expect(wrapper.html()).not.toContain(['contest', 'template', 'option', '999'].join('-'))
    expect(wrapper.find('#contest-awd-challenge-option-999').exists()).toBe(true)
    expect(wrapper.find('#awd-challenge-config-template').exists()).toBe(false)
    expect(contestApiMocks.getChallenges).not.toHaveBeenCalled()
    expect(contestApiMocks.listAdminAwdChallenges).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      status: 'published',
    })
    expect(wrapper.text()).toContain('Final Challenge')
  })

  it('应该在题目编排 AWD 题库加载失败时给出错误提示而不是留下未处理异常', async () => {
    contestApiMocks.getContest.mockResolvedValue(
      buildContestDetail({
        title: '2026 AWD 联赛',
        description: '攻防赛',
        mode: 'awd',
        status: 'registering',
      })
    )
    contestApiMocks.listAdminAwdChallenges.mockRejectedValueOnce(new Error('catalog failed'))

    const wrapper = mountContestEdit()

    await flushPromises()
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('catalog failed')
    expect(wrapper.get('#contest-workbench-stage-tab-pool').attributes('aria-selected')).toBe(
      'true'
    )
    expect(wrapper.find('#contest-awd-challenge-list').exists()).toBe(true)
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
        awd_challenge_id: '1',
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

    await wrapper.get('#contest-challenge-edit-101').trigger('click')
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

    await wrapper.get('#contest-challenge-remove-101').trigger('click')
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
    contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
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
        awd_challenge_id: 11,
        points: 100,
        order: 0,
        is_visible: true,
      })
      awdServicesState.push({
        id: 'service-2',
        contest_id: 'contest-1',
        challenge_id: '11',
        awd_challenge_id: '11',
        title: 'Upload Service',
        category: 'web',
        difficulty: 'medium',
        display_name: 'Upload Service',
        order: 0,
        is_visible: true,
        score_config: {
          points: 100,
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
    await wrapper.get('#contest-awd-challenge-option-11').trigger('click')
    expect(wrapper.find('#contest-awd-service-points').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-service-order').exists()).toBe(false)
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    await wrapper.get('#contest-workbench-stage-tab-awd-config').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Upload Service')

    await wrapper.get('#contest-workbench-stage-tab-preflight').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('可以开赛')
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
    await wrapper.get('#contest-challenge-remove-1').trigger('click')
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
        awd_challenge_id: '1',
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
        awd_challenge_id: 1,
        points: 100,
        order: 0,
        is_visible: true,
      })
      const created = {
        id: 'service-2',
        contest_id: 'contest-1',
        challenge_id: '1',
        awd_challenge_id: '1',
        title: 'Bank Portal AWD',
        category: 'web',
        difficulty: 'medium',
        display_name: 'Bank Portal AWD',
        order: 0,
        is_visible: true,
        score_config: {
          points: 100,
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
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-awd-challenge-option-1').trigger('click')
    expect(wrapper.find('#contest-awd-service-points').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-service-order').exists()).toBe(false)
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(contestApiMocks.updateAdminContestChallenge).not.toHaveBeenCalled()

    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Bank Portal AWD')

    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    expect(wrapper.find('#contest-challenge-library').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-challenge-option-1').exists()).toBe(true)
  })

  it('题目编排新增 AWD 题目保存失败时应提示错误并保持弹层打开', async () => {
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
    await wrapper.get('#contest-workbench-stage-tab-pool').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-challenge-add').trigger('click')
    await flushPromises()
    await wrapper.get('#contest-awd-challenge-option-1').trigger('click')
    expect(wrapper.find('#contest-awd-service-points').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-service-order').exists()).toBe(false)
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('部分 AWD 题目关联失败：Bank Portal AWD')
    expect(wrapper.find('#contest-awd-challenge-option-1').exists()).toBe(true)
  })

  it('应该在 AWD 赛事的题目池阶段只展示题目编排信息', async () => {
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
        awd_sla_score: 1,
        awd_defense_score: 2,
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
        awd_challenge_id: '1',
        display_name: 'Web 入门',
        order: 1,
        is_visible: true,
        score_config: {
          points: 120,
          awd_sla_score: 1,
          awd_defense_score: 2,
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

    expect(wrapper.text()).toContain('题目资源')
    expect(wrapper.text()).toContain('可见性')
    expect(wrapper.text()).toContain('分值')
    expect(wrapper.text()).toContain('顺序')
    expect(wrapper.text()).not.toContain('未配置 AWD')
    expect(wrapper.text()).not.toContain('预检失败')
    expect(wrapper.text()).not.toContain('Checker')
    expect(wrapper.text()).not.toContain('SLA 18 / 防守 28')
    expect(wrapper.text()).not.toContain('待重新验证')
  })
})
