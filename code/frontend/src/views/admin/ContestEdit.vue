<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ChevronLeft } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import {
  createContestAWDService,
  getChallenges,
  getContest,
  getContestAWDReadiness,
  listAdminAwdServiceTemplates,
  listAdminContestChallenges,
  listContestAWDServices,
  updateContestAWDService,
  updateAdminContestChallenge,
  updateContest,
} from '@/api/admin'
import type {
  AdminAwdServiceTemplateData,
  AdminChallengeListItem,
  AdminContestChallengeViewData,
  AWDReadinessData,
  ContestDetailData,
} from '@/api/contracts'
import type { AdminContestUpdatePayload } from '@/api/admin'
import AdminContestFormPanel from '@/components/admin/contest/AdminContestFormPanel.vue'
import AWDChallengeConfigDialog from '@/components/admin/contest/AWDChallengeConfigDialog.vue'
import AWDChallengeConfigPanel from '@/components/admin/contest/AWDChallengeConfigPanel.vue'
import AWDOperationsPanel from '@/components/admin/contest/AWDOperationsPanel.vue'
import ContestAwdPreflightPanel from '@/components/admin/contest/ContestAwdPreflightPanel.vue'
import ContestChallengeOrchestrationPanel from '@/components/admin/contest/ContestChallengeOrchestrationPanel.vue'
import ContestWorkbenchStageRail from '@/components/admin/contest/ContestWorkbenchStageRail.vue'
import ContestWorkbenchSummaryStrip from '@/components/admin/contest/ContestWorkbenchSummaryStrip.vue'
import AWDReadinessOverrideDialog from '@/components/admin/contest/AWDReadinessOverrideDialog.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import {
  buildContestUpdatePayload,
  confirmContestTermination,
  createFieldLocks,
  createContestStatusOptions,
  createDraftFromContest,
  normalizeEditableStatus,
  shouldConfirmContestTermination,
  type AdminContestStatus,
  type ContestFormDraft,
} from '@/composables/useAdminContests'
import {
  CONTEST_WORKBENCH_STAGE_ORDER,
  useContestWorkbench,
  type ContestWorkbenchStageKey,
} from '@/composables/useContestWorkbench'
import { ApiError } from '@/api/request'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'
import { useToast } from '@/composables/useToast'
import { mergeAdminContestChallengesWithAWDServices } from '@/utils/adminContestAwdChallengeLinks'

interface AWDStartOverrideDialogState {
  open: boolean
  title: string
  readiness: AWDReadinessData | null
  confirmLoading: boolean
  pendingPayload: AdminContestUpdatePayload | null
}

const ERR_AWD_READINESS_BLOCKED = 14025

const route = useRoute()
const router = useRouter()
const toast = useToast()

const contestId = computed(() => String(route.params.id ?? ''))
const loading = ref(true)
const loadError = ref('')
const saving = ref(false)
const loadingAwdStageData = ref(false)
const loadingChallengeCatalog = ref(false)
const savingChallengeConfig = ref(false)
const contest = ref<ContestDetailData | null>(null)
const editingBaseStatus = ref<AdminContestStatus | null>(null)
const formDraft = ref<ContestFormDraft | null>(null)
const awdConfigLoadError = ref('')
const awdPreflightLoadError = ref('')
const awdChallengeLinks = ref<AdminContestChallengeViewData[]>([])
const awdChallengeLinksLoaded = ref(false)
const awdReadiness = ref<AWDReadinessData | null>(null)
const awdChallengeCatalog = ref<AdminChallengeListItem[]>([])
const awdServiceTemplateCatalog = ref<AdminAwdServiceTemplateData[]>([])
const awdChallengeConfigDialogOpen = ref(false)
const awdChallengeConfigMode = ref<'create' | 'edit'>('create')
const editingAwdChallengeLink = ref<AdminContestChallengeViewData | null>(null)
const activeAwdChallengeId = ref<string | null>(null)
const awdConfigFocusSource = ref<'pool' | 'preflight' | null>(null)
const loadingAwdServiceTemplateCatalog = ref(false)
const awdStartOverrideDialogState = ref<AWDStartOverrideDialogState>(createDefaultAWDStartOverrideDialogState())

const fieldLocks = computed(() => createFieldLocks(editingBaseStatus.value))
const statusOptions = computed(() => createContestStatusOptions(editingBaseStatus.value))
const pageTitle = computed(() => (contest.value ? `编辑《${contest.value.title}》` : '编辑竞赛'))
const awdWorkbenchChallengeCount = computed(() =>
  contest.value?.mode === 'awd' && awdChallengeLinksLoaded.value ? awdChallengeLinks.value.length : null
)
const workbench = useContestWorkbench(contest, awdWorkbenchChallengeCount)
const { activeTab: activeStage, selectTab } = useUrlSyncedTabs<ContestWorkbenchStageKey>({
  orderedTabs: CONTEST_WORKBENCH_STAGE_ORDER,
  defaultTab: 'basics',
})
const sortedAwdChallengeLinks = computed(() =>
  [...awdChallengeLinks.value].sort(
    (left, right) => left.order - right.order || left.challenge_id.localeCompare(right.challenge_id)
  )
)
const activeAwdChallengeIndex = computed(() =>
  sortedAwdChallengeLinks.value.findIndex((item) => item.challenge_id === activeAwdChallengeId.value)
)
const canNavigatePreviousAwdChallenge = computed(() => activeAwdChallengeIndex.value > 0)
const canNavigateNextAwdChallenge = computed(
  () => activeAwdChallengeIndex.value >= 0 && activeAwdChallengeIndex.value < sortedAwdChallengeLinks.value.length - 1
)
const existingAwdChallengeIds = computed(() => awdChallengeLinks.value.map((item) => item.challenge_id))

function createDefaultAWDStartOverrideDialogState(): AWDStartOverrideDialogState {
  return {
    open: false,
    title: '',
    readiness: null,
    confirmLoading: false,
    pendingPayload: null,
  }
}

function toISOString(value: string): string {
  return new Date(value).toISOString()
}

function shouldGateAWDContestStart(
  mode: ContestDetailData['mode'] | null,
  targetStatus: AdminContestStatus
): boolean {
  return mode === 'awd' && targetStatus === 'running'
}

function isAWDReadinessBlockedError(error: unknown): boolean {
  return error instanceof ApiError && error.code === ERR_AWD_READINESS_BLOCKED
}

function humanizeRequestError(error: unknown, fallback: string): string {
  if (error instanceof ApiError && error.message.trim()) {
    return error.message
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

function readRequestedWorkbenchStage(): ContestWorkbenchStageKey | null {
  if (typeof window === 'undefined') {
    return null
  }

  const requestedStage = new URLSearchParams(window.location.search).get('panel')
  if (!requestedStage) {
    return null
  }

  if (CONTEST_WORKBENCH_STAGE_ORDER.includes(requestedStage as ContestWorkbenchStageKey)) {
    return requestedStage as ContestWorkbenchStageKey
  }

  return null
}

function syncWorkbenchStageSelection(): void {
  const visibleStageKeys = workbench.visibleStages.map((stage) => stage.key)
  const requestedStage = readRequestedWorkbenchStage()

  if (requestedStage && visibleStageKeys.includes(requestedStage)) {
    if (activeStage.value !== requestedStage) {
      selectTab(requestedStage)
    }
    return
  }

  if (requestedStage || !visibleStageKeys.includes(activeStage.value) || activeStage.value !== workbench.defaultStage) {
    selectTab(workbench.defaultStage)
  }
}

function resetAwdWorkbenchState() {
  awdConfigLoadError.value = ''
  awdPreflightLoadError.value = ''
  awdChallengeLinks.value = []
  awdChallengeLinksLoaded.value = false
  awdReadiness.value = null
  awdChallengeCatalog.value = []
  awdServiceTemplateCatalog.value = []
  awdChallengeConfigDialogOpen.value = false
  awdChallengeConfigMode.value = 'create'
  editingAwdChallengeLink.value = null
  activeAwdChallengeId.value = null
  awdConfigFocusSource.value = null
}

async function refreshAwdWorkbenchData(nextContestId = contestId.value): Promise<void> {
  if (!contest.value || contest.value.mode !== 'awd' || !nextContestId) {
    resetAwdWorkbenchState()
    return
  }

  loadingAwdStageData.value = true
  try {
    awdConfigLoadError.value = ''
    awdPreflightLoadError.value = ''

    const [challengeLinksResult, awdServicesResult, readinessResult] = await Promise.allSettled([
      listAdminContestChallenges(nextContestId),
      listContestAWDServices(nextContestId),
      getContestAWDReadiness(nextContestId),
    ])

    if (challengeLinksResult.status === 'fulfilled' && awdServicesResult.status === 'fulfilled') {
      awdChallengeLinks.value = mergeAdminContestChallengesWithAWDServices(
        challengeLinksResult.value,
        awdServicesResult.value
      )
      awdChallengeLinksLoaded.value = true
    } else {
      awdConfigLoadError.value = humanizeRequestError(
        challengeLinksResult.status === 'rejected'
          ? challengeLinksResult.reason
          : awdServicesResult.status === 'rejected'
            ? awdServicesResult.reason
            : null,
        'AWD 配置数据加载失败'
      )
    }

    if (readinessResult.status === 'fulfilled') {
      awdReadiness.value = readinessResult.value
    } else {
      awdReadiness.value = null
      awdPreflightLoadError.value = humanizeRequestError(readinessResult.reason, '赛前检查数据加载失败')
    }

    if (challengeLinksResult.status === 'fulfilled' && activeAwdChallengeId.value) {
      const hasActiveChallenge = awdChallengeLinks.value.some(
        (item) => item.challenge_id === activeAwdChallengeId.value
      )
      if (!hasActiveChallenge) {
        activeAwdChallengeId.value = null
        awdConfigFocusSource.value = null
      }
    }
  } finally {
    loadingAwdStageData.value = false
  }
}

async function loadAwdChallengeCatalog(): Promise<void> {
  if (loadingChallengeCatalog.value) {
    return
  }
  if (awdChallengeCatalog.value.length > 0) {
    return
  }

  loadingChallengeCatalog.value = true
  try {
    const pageSize = 200
    const catalog: AdminChallengeListItem[] = []
    let page = 1
    let total = 0

    do {
      const result = await getChallenges({
        page,
        page_size: pageSize,
        status: 'published',
      })
      catalog.push(...result.list)
      total = result.total
      page += 1
    } while (catalog.length < total)

    awdChallengeCatalog.value = catalog
  } catch (error) {
    awdChallengeCatalog.value = []
    toast.error(humanizeRequestError(error, 'AWD 题目目录加载失败'))
  } finally {
    loadingChallengeCatalog.value = false
  }
}

async function loadAwdServiceTemplateCatalog(): Promise<void> {
  if (loadingAwdServiceTemplateCatalog.value) {
    return
  }
  if (awdServiceTemplateCatalog.value.length > 0) {
    return
  }

  loadingAwdServiceTemplateCatalog.value = true
  try {
    const pageSize = 100
    const templates: AdminAwdServiceTemplateData[] = []
    let page = 1
    let total = 0

    do {
      const result = await listAdminAwdServiceTemplates({
        page,
        page_size: pageSize,
        status: 'published',
      })
      templates.push(...result.list)
      total = result.total
      page += 1
    } while (templates.length < total)

    awdServiceTemplateCatalog.value = templates
  } catch (error) {
    awdServiceTemplateCatalog.value = []
    toast.error(humanizeRequestError(error, 'AWD 服务模板加载失败'))
  } finally {
    loadingAwdServiceTemplateCatalog.value = false
  }
}

function selectStage(stage: ContestWorkbenchStageKey) {
  selectTab(stage)
}

function handleDraftChange(nextDraft: ContestFormDraft) {
  formDraft.value = {
    ...nextDraft,
  }
}

function setActiveAwdChallenge(challengeId: string | null, source: 'pool' | 'preflight' | null) {
  activeAwdChallengeId.value = challengeId
  awdConfigFocusSource.value = challengeId ? source : null
}

function focusAwdChallengeByOffset(offset: -1 | 1) {
  if (activeAwdChallengeIndex.value < 0) {
    return
  }

  const nextChallenge = sortedAwdChallengeLinks.value[activeAwdChallengeIndex.value + offset]
  if (!nextChallenge) {
    return
  }

  setActiveAwdChallenge(nextChallenge.challenge_id, awdConfigFocusSource.value)
}

function openAwdChallengeCreateDialog() {
  awdChallengeConfigMode.value = 'create'
  editingAwdChallengeLink.value = null
  awdChallengeConfigDialogOpen.value = true
  void loadAwdChallengeCatalog()
  void loadAwdServiceTemplateCatalog()
}

function openAwdChallengeEditDialog(challenge: AdminContestChallengeViewData) {
  setActiveAwdChallenge(challenge.challenge_id, awdConfigFocusSource.value)
  awdChallengeConfigMode.value = 'edit'
  editingAwdChallengeLink.value = challenge
  awdChallengeConfigDialogOpen.value = true
  void loadAwdServiceTemplateCatalog()
}

function updateAwdChallengeConfigDialogOpen(value: boolean) {
  awdChallengeConfigDialogOpen.value = value
  if (!value) {
    editingAwdChallengeLink.value = null
  }
}

async function handleSaveAwdChallengeConfig(payload: {
  challenge_id: number
  template_id: number
  points: number
  order: number
  is_visible: boolean
  awd_checker_type: 'legacy_probe' | 'http_standard'
  awd_checker_config: Record<string, unknown>
  awd_sla_score: number
  awd_defense_score: number
  awd_checker_preview_token?: string
}) {
  if (!contest.value) {
    return
  }

  savingChallengeConfig.value = true
  try {
    if (awdChallengeConfigMode.value === 'create') {
      await createContestAWDService(contest.value.id, {
        challenge_id: payload.challenge_id,
        template_id: payload.template_id,
        order: payload.order,
        is_visible: payload.is_visible,
        checker_type: payload.awd_checker_type,
        checker_config: payload.awd_checker_config,
        awd_sla_score: payload.awd_sla_score,
        awd_defense_score: payload.awd_defense_score,
        awd_checker_preview_token: payload.awd_checker_preview_token,
      })
      await updateAdminContestChallenge(contest.value.id, String(payload.challenge_id), {
        points: payload.points,
      })
      setActiveAwdChallenge(String(payload.challenge_id), null)
    } else if (editingAwdChallengeLink.value) {
      if (editingAwdChallengeLink.value.awd_service_id) {
        await updateContestAWDService(contest.value.id, editingAwdChallengeLink.value.awd_service_id, {
          template_id: payload.template_id,
          order: payload.order,
          is_visible: payload.is_visible,
          checker_type: payload.awd_checker_type,
          checker_config: payload.awd_checker_config,
          awd_sla_score: payload.awd_sla_score,
          awd_defense_score: payload.awd_defense_score,
          awd_checker_preview_token: payload.awd_checker_preview_token,
        })
      } else {
        await createContestAWDService(contest.value.id, {
          challenge_id: Number(editingAwdChallengeLink.value.challenge_id),
          template_id: payload.template_id,
          order: payload.order,
          is_visible: payload.is_visible,
          checker_type: payload.awd_checker_type,
          checker_config: payload.awd_checker_config,
          awd_sla_score: payload.awd_sla_score,
          awd_defense_score: payload.awd_defense_score,
          awd_checker_preview_token: payload.awd_checker_preview_token,
        })
      }
      await updateAdminContestChallenge(contest.value.id, editingAwdChallengeLink.value.challenge_id, {
        points: payload.points,
      })
      setActiveAwdChallenge(editingAwdChallengeLink.value.challenge_id, awdConfigFocusSource.value)
    }
    awdChallengeConfigDialogOpen.value = false
    editingAwdChallengeLink.value = null
    await refreshAwdWorkbenchData(contest.value.id)
  } catch (error) {
    toast.error(
      humanizeRequestError(error, awdChallengeConfigMode.value === 'create' ? '新增 AWD 题目失败' : '更新 AWD 题目失败')
    )
  } finally {
    savingChallengeConfig.value = false
  }
}

function handleOpenAwdConfigFromPool(challenge: AdminContestChallengeViewData) {
  setActiveAwdChallenge(challenge.challenge_id, 'pool')
  selectStage('awd-config')
}

function handleNavigateAwdChallengeFromPreflight(challengeId: string) {
  setActiveAwdChallenge(challengeId, 'preflight')
}

function createRunningPayloadFromDraft(): AdminContestUpdatePayload | null {
  if (!formDraft.value) {
    return null
  }

  return buildContestUpdatePayload(
    {
      ...formDraft.value,
      status: 'running',
    },
    fieldLocks.value
  )
}

async function openPreflightOverrideDialog() {
  const payload = createRunningPayloadFromDraft()
  if (!payload) {
    return
  }

  try {
    const readiness = awdReadiness.value ?? (contestId.value ? await getContestAWDReadiness(contestId.value) : null)
    awdStartOverrideDialogState.value = {
      open: true,
      title: '启动赛事',
      readiness,
      confirmLoading: false,
      pendingPayload: payload,
    }
  } catch (error) {
    awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
    toast.error(humanizeRequestError(error, '赛前检查数据加载失败'))
  }
}

async function loadContestDetail(): Promise<void> {
  if (!contestId.value) {
    loadError.value = '缺少赛事 ID，无法进入编辑页。'
    loading.value = false
    return
  }

  loading.value = true
  loadError.value = ''
  try {
    const detail = await getContest(contestId.value)
    contest.value = detail
    editingBaseStatus.value = normalizeEditableStatus(detail.status)
    formDraft.value = createDraftFromContest(detail)
    syncWorkbenchStageSelection()
  } catch (error) {
    loadError.value = humanizeRequestError(error, '竞赛详情加载失败')
  } finally {
    loading.value = false
  }

  if (contest.value?.mode === 'awd') {
    await refreshAwdWorkbenchData(contest.value.id)
    return
  }

  resetAwdWorkbenchState()
}

function goBackToContestList() {
  void router.push({ name: 'ContestManage', query: { panel: 'list' } })
}

async function finalizeContestUpdateSuccess() {
  toast.success('竞赛已更新')
  awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
  goBackToContestList()
}

async function openAWDStartOverrideDialog(payload: AdminContestUpdatePayload) {
  try {
    const readiness = await getContestAWDReadiness(contestId.value)
    awdStartOverrideDialogState.value = {
      open: true,
      title: '启动赛事',
      readiness,
      confirmLoading: false,
      pendingPayload: payload,
    }
  } catch (error) {
    awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
    toast.error(humanizeRequestError(error, '赛前检查数据加载失败'))
  }
}

function handleAwdStartOverrideDialogOpenChange(value: boolean) {
  if (!value) {
    awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
  }
}

async function confirmAWDStartOverride(reason: string) {
  const payload = awdStartOverrideDialogState.value.pendingPayload
  const normalizedReason = reason.trim()
  if (!payload || !normalizedReason) {
    return
  }

  awdStartOverrideDialogState.value = {
    ...awdStartOverrideDialogState.value,
    confirmLoading: true,
  }

  try {
    await updateContest(
      contestId.value,
      {
        ...payload,
        force_override: true,
        override_reason: normalizedReason,
      },
      { suppressErrorToast: true }
    )
    await finalizeContestUpdateSuccess()
  } catch (error) {
    if (isAWDReadinessBlockedError(error)) {
      await openAWDStartOverrideDialog(payload)
      return
    }
    toast.error(humanizeRequestError(error, '竞赛更新失败'))
  } finally {
    if (awdStartOverrideDialogState.value.open) {
      awdStartOverrideDialogState.value = {
        ...awdStartOverrideDialogState.value,
        confirmLoading: false,
      }
    }
  }
}

async function handleSave(draft: ContestFormDraft): Promise<void> {
  if (!contest.value) {
    return
  }

  if (shouldConfirmContestTermination(editingBaseStatus.value, draft.status)) {
    const confirmed = await confirmContestTermination(draft.title)
    if (!confirmed) {
      return
    }
  }

  saving.value = true
  try {
    const payload = buildContestUpdatePayload(draft, fieldLocks.value)

    if (shouldGateAWDContestStart(contest.value.mode, draft.status)) {
      try {
        await updateContest(contestId.value, payload, { suppressErrorToast: true })
        await finalizeContestUpdateSuccess()
      } catch (error) {
        if (isAWDReadinessBlockedError(error)) {
          await openAWDStartOverrideDialog(payload)
          return
        }
        toast.error(humanizeRequestError(error, '竞赛更新失败'))
      }
      return
    }

    await updateContest(contestId.value, payload)
    await finalizeContestUpdateSuccess()
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadContestDetail()
})
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Contest Workspace</span>
        <span class="class-chip">竞赛编辑</span>
      </div>
      <div class="topbar-actions">
        <button class="ui-btn ui-btn--ghost" type="button" @click="goBackToContestList">
          <ChevronLeft class="h-4 w-4" />
          返回竞赛目录
        </button>
      </div>
    </header>

    <main class="content-pane">
      <div v-if="loading" class="flex justify-center py-12">
        <AppLoading>正在同步竞赛详情...</AppLoading>
      </div>

      <AppEmpty
        v-else-if="loadError"
        title="竞赛详情加载失败"
        :description="loadError"
        icon="AlertTriangle"
      >
        <template #action>
          <button type="button" class="ui-btn ui-btn--ghost" @click="goBackToContestList">
            返回竞赛目录
          </button>
        </template>
      </AppEmpty>

      <template v-else-if="formDraft && contest">
        <ContestWorkbenchStageRail
          :stages="workbench.visibleStages"
          :active-stage="activeStage"
          :select-stage="selectTab"
        />

        <ContestWorkbenchSummaryStrip :items="workbench.summaryItems" />

        <section
          id="contest-workbench-panel-basics"
          class="tab-panel contest-edit-panel"
          :class="{ active: activeStage === 'basics' }"
          role="tabpanel"
          aria-labelledby="contest-workbench-stage-tab-basics"
          :aria-hidden="activeStage === 'basics' ? 'false' : 'true'"
        >
          <section class="workspace-directory-section contest-edit-section">
            <header class="contest-edit-header">
              <div class="workspace-tab-heading__main">
                <div class="workspace-overline">Contest Editor</div>
                <h1 class="workspace-page-title">编辑竞赛</h1>
                <p class="workspace-page-copy">
                  {{ pageTitle }}。保存成功后会回到赛事目录，便于继续查看列表或进入后续编排。
                </p>
              </div>
            </header>

            <AdminContestFormPanel
              :mode="'edit'"
              :draft="formDraft"
              :saving="saving"
              :status-options="statusOptions"
              :field-locks="fieldLocks"
              :show-cancel="false"
              :note="'保存后将返回赛事目录列表；AWD 赛事切换到进行中时仍会执行就绪检查。'"
              @update:draft="handleDraftChange"
              @save="handleSave"
            />
          </section>
        </section>

        <section
          id="contest-workbench-panel-pool"
          class="tab-panel contest-edit-panel"
          :class="{ active: activeStage === 'pool' }"
          role="tabpanel"
          aria-labelledby="contest-workbench-stage-tab-pool"
          :aria-hidden="activeStage === 'pool' ? 'false' : 'true'"
        >
          <section class="contest-edit-section contest-edit-section--flat">
            <ContestChallengeOrchestrationPanel
              :contest-id="contest.id"
              :contest-mode="contest.mode"
              :challenge-links="contest.mode === 'awd' ? awdChallengeLinks : undefined"
              :loading-external="contest.mode === 'awd' ? loadingAwdStageData : undefined"
              :load-error-external="contest.mode === 'awd' ? awdConfigLoadError : undefined"
              @open:awd-config="handleOpenAwdConfigFromPool"
              @updated="refreshAwdWorkbenchData(contest.id)"
            />
          </section>
        </section>

        <section
          v-if="contest.mode === 'awd'"
          id="contest-workbench-panel-awd-config"
          class="tab-panel contest-edit-panel"
          :class="{ active: activeStage === 'awd-config' }"
          role="tabpanel"
          aria-labelledby="contest-workbench-stage-tab-awd-config"
          :aria-hidden="activeStage === 'awd-config' ? 'false' : 'true'"
        >
          <section class="contest-edit-section contest-edit-section--flat">
            <div
              v-if="loadingAwdStageData && !awdConfigLoadError && awdChallengeLinks.length === 0"
              class="workspace-directory-section contest-edit-section"
            >
              <AppLoading>正在同步 AWD 配置...</AppLoading>
            </div>
            <AppEmpty
              v-else-if="awdConfigLoadError && !awdChallengeLinksLoaded"
              title="AWD 配置数据暂时不可用"
              :description="awdConfigLoadError"
              icon="AlertTriangle"
            >
              <template #action>
                <button type="button" class="ui-btn ui-btn--ghost" @click="refreshAwdWorkbenchData(contest.id)">
                  重试加载
                </button>
              </template>
            </AppEmpty>
            <template v-else>
              <p v-if="awdConfigLoadError && awdChallengeLinksLoaded" class="contest-edit-inline-warning" role="status">
                AWD 题目刷新失败，当前显示上次成功同步的数据。{{ awdConfigLoadError }}
              </p>
              <AWDChallengeConfigPanel
                :challenge-links="awdChallengeLinks"
                :active-challenge-id="activeAwdChallengeId"
                :focus-source="awdConfigFocusSource"
                :can-navigate-previous="canNavigatePreviousAwdChallenge"
                :can-navigate-next="canNavigateNextAwdChallenge"
                @create="openAwdChallengeCreateDialog"
                @edit="openAwdChallengeEditDialog"
                @previous="focusAwdChallengeByOffset(-1)"
                @next="focusAwdChallengeByOffset(1)"
              />
            </template>
          </section>
        </section>

        <section
          v-if="contest.mode === 'awd'"
          id="contest-workbench-panel-preflight"
          class="tab-panel contest-edit-panel"
          :class="{ active: activeStage === 'preflight' }"
          role="tabpanel"
          aria-labelledby="contest-workbench-stage-tab-preflight"
          :aria-hidden="activeStage === 'preflight' ? 'false' : 'true'"
        >
          <section class="contest-edit-section contest-edit-section--flat">
            <AppEmpty
              v-if="awdPreflightLoadError"
              title="赛前检查暂时不可用"
              :description="awdPreflightLoadError"
              icon="AlertTriangle"
            >
              <template #action>
                <button type="button" class="ui-btn ui-btn--ghost" @click="refreshAwdWorkbenchData(contest.id)">
                  重试加载
                </button>
              </template>
            </AppEmpty>
            <ContestAwdPreflightPanel
              v-else
              :readiness="awdReadiness"
              :loading="loadingAwdStageData"
              @navigate:challenge="handleNavigateAwdChallengeFromPreflight"
              @navigate:stage="selectStage"
              @open:override="openPreflightOverrideDialog"
            />
          </section>
        </section>

        <section
          v-if="contest.mode === 'awd'"
          id="contest-workbench-panel-operations"
          class="tab-panel contest-edit-panel"
          :class="{ active: activeStage === 'operations' }"
          role="tabpanel"
          aria-labelledby="contest-workbench-stage-tab-operations"
          :aria-hidden="activeStage === 'operations' ? 'false' : 'true'"
        >
          <section v-if="activeStage === 'operations'" class="contest-edit-section contest-edit-section--flat">
            <AWDOperationsPanel
              :contests="[contest]"
              :selected-contest-id="contest.id"
              :hide-contest-selector="true"
            />
          </section>
        </section>
      </template>
    </main>

    <AWDChallengeConfigDialog
      :contest-id="contest?.id || null"
      :open="awdChallengeConfigDialogOpen"
      :mode="awdChallengeConfigMode"
      :challenge-options="awdChallengeCatalog"
      :template-options="awdServiceTemplateCatalog"
      :existing-challenge-ids="existingAwdChallengeIds"
      :draft="editingAwdChallengeLink"
      :loading-challenge-catalog="loadingChallengeCatalog"
      :loading-template-catalog="loadingAwdServiceTemplateCatalog"
      :saving="savingChallengeConfig"
      @update:open="updateAwdChallengeConfigDialogOpen"
      @save="handleSaveAwdChallengeConfig"
    />

    <AWDReadinessOverrideDialog
      :open="awdStartOverrideDialogState.open"
      :title="awdStartOverrideDialogState.title"
      :readiness="awdStartOverrideDialogState.readiness"
      :confirm-loading="awdStartOverrideDialogState.confirmLoading"
      @update:open="handleAwdStartOverrideDialogOpenChange"
      @confirm="confirmAWDStartOverride"
    />
  </section>
</template>

<style scoped>
.content-pane {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: var(--space-6);
}

.contest-edit-panel {
  display: none;
}

.contest-edit-panel.active {
  display: grid;
}

.contest-edit-section {
  display: grid;
  gap: var(--space-5);
  padding: var(--space-5) var(--space-5-5);
}

.contest-edit-section--flat {
  padding: 0;
}

.contest-edit-inline-warning {
  margin: 0 0 var(--space-4);
  border: 1px solid color-mix(in srgb, var(--journal-danger, #d9594c) 32%, transparent);
  border-radius: 1rem;
  padding: var(--space-3) var(--space-4);
  background: color-mix(in srgb, var(--journal-danger, #d9594c) 12%, transparent);
  color: var(--journal-ink);
  font-size: var(--font-size-0-875);
}

.contest-edit-header {
  display: grid;
  gap: var(--space-4);
}

@media (max-width: 767px) {
  .content-pane {
    gap: var(--space-5);
    padding: var(--space-5) var(--space-4) var(--space-6);
  }

  .contest-edit-section {
    padding: var(--space-4-5) var(--space-4);
  }
}
</style>
