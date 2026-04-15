<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ChevronLeft } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import {
  createAdminContestChallenge,
  getChallenges,
  getContest,
  getContestAWDReadiness,
  listAdminContestChallenges,
  updateAdminContestChallenge,
  updateContest,
} from '@/api/admin'
import type {
  AdminChallengeListItem,
  AdminContestChallengeData,
  AWDReadinessData,
  ContestDetailData,
} from '@/api/contracts'
import type { AdminContestUpdatePayload } from '@/api/admin'
import AdminContestFormPanel from '@/components/admin/contest/AdminContestFormPanel.vue'
import AWDChallengeConfigDialog from '@/components/admin/contest/AWDChallengeConfigDialog.vue'
import AWDChallengeConfigPanel from '@/components/admin/contest/AWDChallengeConfigPanel.vue'
import ContestAwdPreflightPanel from '@/components/admin/contest/ContestAwdPreflightPanel.vue'
import ContestChallengeOrchestrationPanel from '@/components/admin/contest/ContestChallengeOrchestrationPanel.vue'
import ContestWorkbenchStageRail from '@/components/admin/contest/ContestWorkbenchStageRail.vue'
import ContestWorkbenchSummaryStrip from '@/components/admin/contest/ContestWorkbenchSummaryStrip.vue'
import AWDReadinessOverrideDialog from '@/components/admin/contest/AWDReadinessOverrideDialog.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import {
  createFieldLocks,
  createContestStatusOptions,
  createDraftFromContest,
  normalizeEditableStatus,
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
const awdChallengeLinks = ref<AdminContestChallengeData[]>([])
const awdReadiness = ref<AWDReadinessData | null>(null)
const awdChallengeCatalog = ref<AdminChallengeListItem[]>([])
const awdChallengeConfigDialogOpen = ref(false)
const awdChallengeConfigMode = ref<'create' | 'edit'>('create')
const editingAwdChallengeLink = ref<AdminContestChallengeData | null>(null)
const activeAwdChallengeId = ref<string | null>(null)
const awdConfigFocusSource = ref<'pool' | 'preflight' | null>(null)
const awdStartOverrideDialogState = ref<AWDStartOverrideDialogState>(createDefaultAWDStartOverrideDialogState())

const fieldLocks = computed(() => createFieldLocks(editingBaseStatus.value))
const statusOptions = computed(() => createContestStatusOptions(editingBaseStatus.value))
const pageTitle = computed(() => (contest.value ? `编辑《${contest.value.title}》` : '编辑竞赛'))
const workbench = useContestWorkbench(contest)
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

function shouldGateAWDContestStart(mode: ContestDetailData['mode'] | null, targetStatus: AdminContestStatus): boolean {
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
  awdChallengeLinks.value = []
  awdReadiness.value = null
  awdChallengeCatalog.value = []
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
    const [nextChallengeLinks, nextReadiness] = await Promise.all([
      listAdminContestChallenges(nextContestId),
      getContestAWDReadiness(nextContestId),
    ])
    awdChallengeLinks.value = nextChallengeLinks
    awdReadiness.value = nextReadiness

    if (activeAwdChallengeId.value) {
      const hasActiveChallenge = nextChallengeLinks.some(
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
    const result = await getChallenges({ page: 1, page_size: 200 })
    awdChallengeCatalog.value = result.list
  } finally {
    loadingChallengeCatalog.value = false
  }
}

function selectStage(stage: ContestWorkbenchStageKey) {
  selectTab(stage)
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
}

function openAwdChallengeEditDialog(challenge: AdminContestChallengeData) {
  setActiveAwdChallenge(challenge.challenge_id, awdConfigFocusSource.value)
  awdChallengeConfigMode.value = 'edit'
  editingAwdChallengeLink.value = challenge
  awdChallengeConfigDialogOpen.value = true
}

function updateAwdChallengeConfigDialogOpen(value: boolean) {
  awdChallengeConfigDialogOpen.value = value
  if (!value) {
    editingAwdChallengeLink.value = null
  }
}

async function handleSaveAwdChallengeConfig(payload: {
  challenge_id: number
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
      await createAdminContestChallenge(contest.value.id, payload)
      setActiveAwdChallenge(String(payload.challenge_id), null)
    } else if (editingAwdChallengeLink.value) {
      const { challenge_id: _challengeId, ...updatePayload } = payload
      await updateAdminContestChallenge(contest.value.id, editingAwdChallengeLink.value.challenge_id, updatePayload)
      setActiveAwdChallenge(editingAwdChallengeLink.value.challenge_id, awdConfigFocusSource.value)
    }
    awdChallengeConfigDialogOpen.value = false
    editingAwdChallengeLink.value = null
    await refreshAwdWorkbenchData(contest.value.id)
  } finally {
    savingChallengeConfig.value = false
  }
}

function handleOpenAwdConfigFromPool(challenge: AdminContestChallengeData) {
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

  return {
    title: formDraft.value.title.trim(),
    description: formDraft.value.description.trim(),
    mode: formDraft.value.mode,
    starts_at: toISOString(formDraft.value.starts_at),
    ends_at: toISOString(formDraft.value.ends_at),
    status: 'running',
  }
}

async function openPreflightOverrideDialog() {
  const payload = createRunningPayloadFromDraft()
  if (!payload) {
    return
  }

  const readiness = awdReadiness.value ?? (contestId.value ? await getContestAWDReadiness(contestId.value) : null)
  awdStartOverrideDialogState.value = {
    open: true,
    title: '启动赛事',
    readiness,
    confirmLoading: false,
    pendingPayload: payload,
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
    if (detail.mode === 'awd') {
      await refreshAwdWorkbenchData(detail.id)
    } else {
      resetAwdWorkbenchState()
    }
    syncWorkbenchStageSelection()
  } catch (error) {
    loadError.value = humanizeRequestError(error, '竞赛详情加载失败')
  } finally {
    loading.value = false
  }
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
  const readiness = await getContestAWDReadiness(contestId.value)
  awdStartOverrideDialogState.value = {
    open: true,
    title: '启动赛事',
    readiness,
    confirmLoading: false,
    pendingPayload: payload,
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

  saving.value = true
  try {
    const payload: AdminContestUpdatePayload = {
      title: draft.title.trim(),
      description: draft.description.trim(),
      mode: draft.mode,
      starts_at: toISOString(draft.starts_at),
      ends_at: toISOString(draft.ends_at),
      status: draft.status,
    }

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
        <button class="admin-btn admin-btn-ghost" type="button" @click="goBackToContestList">
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
          <button type="button" class="admin-btn admin-btn-ghost" @click="goBackToContestList">
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
              @open:awd-config="handleOpenAwdConfigFromPool"
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
            <ContestAwdPreflightPanel
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
          <section class="workspace-directory-section contest-edit-section">
            <header class="contest-edit-header">
              <div class="workspace-tab-heading__main">
                <div class="workspace-overline">Operations</div>
                <h2 class="workspace-page-title">轮次运行</h2>
                <p class="workspace-page-copy">
                  在这里查看轮次运行信息与现场状态，便于赛事进行中的值守与处理。
                </p>
              </div>
            </header>
          </section>
        </section>
      </template>
    </main>

    <AWDChallengeConfigDialog
      :contest-id="contest?.id || null"
      :open="awdChallengeConfigDialogOpen"
      :mode="awdChallengeConfigMode"
      :challenge-options="awdChallengeCatalog"
      :existing-challenge-ids="existingAwdChallengeIds"
      :draft="editingAwdChallengeLink"
      :loading-challenge-catalog="loadingChallengeCatalog"
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

.contest-edit-header {
  display: grid;
  gap: var(--space-4);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: var(--space-2-5) var(--space-4);
  font-size: var(--font-size-0-875);
  font-weight: 600;
  transition: all 150ms ease;
}

.admin-btn-ghost {
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  color: var(--journal-ink);
}

.admin-btn-ghost:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  color: var(--journal-accent);
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
