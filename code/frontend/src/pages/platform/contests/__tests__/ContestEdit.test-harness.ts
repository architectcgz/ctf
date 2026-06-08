import { vi } from 'vitest'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { computed, defineComponent, ref, watch } from 'vue'

import ContestEdit from '@/pages/platform/contests/ContestEditRoutePage.vue'
export { default as contestEditSource } from '@/pages/platform/contests/ContestEditRoutePage.vue?raw'
export { default as contestEditPageModelSource } from '@/features/platform/contest-manage/model/useContestEditPage.ts?raw'
export { default as platformContestEditPageSource } from '@/features/platform/contest-manage/ui/PlatformContestEditPage.vue?raw'
export { default as platformRoutesSource } from '@/router/routes/platformRoutes.ts?raw'
export { ApiError } from '@/api/request'
import type { ContestDetailData } from '@/api/contracts'
import type { VueWrapper } from '@vue/test-utils'

export const pushMock = vi.fn()

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
    useRoute: () => ({
      name: 'ContestEdit',
      params: { id: 'contest-1' },
      query: Object.fromEntries(new URLSearchParams(window.location.search)),
    }),
    useRouter: () => ({ push: pushMock, replace: vi.fn(), back: vi.fn() }),
  }
})

vi.mock('@/api/admin/contest-manage', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin/contest-manage')>(
    '@/api/admin/contest-manage'
  )
  return {
    ...actual,
    getContest: contestApiMocks.getContest,
    updateContest: contestApiMocks.updateContest,
    listAdminContestChallenges: contestApiMocks.listAdminContestChallenges,
    createAdminContestChallenge: contestApiMocks.createAdminContestChallenge,
    updateAdminContestChallenge: contestApiMocks.updateAdminContestChallenge,
    deleteAdminContestChallenge: contestApiMocks.deleteAdminContestChallenge,
  }
})
vi.mock('@/api/admin/contest-operations', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin/contest-operations')>(
    '@/api/admin/contest-operations'
  )
  return {
    ...actual,
    getContestAWDReadiness: contestApiMocks.getContestAWDReadiness,
  }
})
vi.mock('@/api/admin/contest-awd-admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin/contest-awd-admin')>(
    '@/api/admin/contest-awd-admin'
  )
  return {
    ...actual,
    listContestAWDServices: contestApiMocks.listContestAWDServices,
    createContestAWDService: contestApiMocks.createContestAWDService,
    deleteContestAWDService: contestApiMocks.deleteContestAWDService,
    updateContestAWDService: contestApiMocks.updateContestAWDService,
  }
})
vi.mock('@/api/admin/awd-authoring', () => ({
  listAdminAwdChallenges: contestApiMocks.listAdminAwdChallenges,
}))
vi.mock('@/api/admin/authoring', () => ({
  getChallenges: contestApiMocks.getChallenges,
}))

vi.mock('@/shared/model/common/useDestructiveConfirm', () => ({
  confirmDestructiveAction: destructiveConfirmMock,
}))

vi.mock('@/shared/model/common/useToast', () => ({
  useToast: () => toastMocks,
}))

export function buildContestDetail(overrides: Partial<ContestDetailData> = {}): ContestDetailData {
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

export function mountContestEdit() {
  return mount(ContestEdit, {
    props: {
      contestId: 'contest-1',
    },
    global: {
      stubs: {
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
        RouterLink: RouterLinkStub,
      },
    },
  })
}

export function mountContestEditWithRealChallengeDialog() {
  return mount(ContestEdit, {
    props: {
      contestId: 'contest-1',
    },
    global: {
      stubs: {
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
        RouterLink: RouterLinkStub,
      },
    },
  })
}

export function createDeferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

export function getWorkbenchStageRail(wrapper: VueWrapper<any>) {
  return wrapper.get('[role="tablist"][aria-label="竞赛工作台阶段切换"]')
}

export function findRouteLink(wrapper: VueWrapper<any>, id: string) {
  return wrapper.findAllComponents(RouterLinkStub).find((link) => link.attributes('id') === id)
}

export async function submitContestBasicsForm(wrapper: VueWrapper<any>) {
  await wrapper.get('.studio-form-canvas form').trigger('submit')
  await flushPromises()
}

export function resetContestEditTestHarness() {
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
}

export { contestApiMocks, destructiveConfirmMock, toastMocks }
