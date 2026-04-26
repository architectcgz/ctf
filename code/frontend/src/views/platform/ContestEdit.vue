<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  getContest,
  getContestAWDReadiness,
  updateContest,
} from '@/api/admin'
import type { AWDReadinessData, ContestDetailData } from '@/api/contracts'
import type { AdminContestUpdatePayload } from '@/api/admin'
import AWDChallengeConfigDialog from '@/components/platform/contest/AWDChallengeConfigDialog.vue'
import ContestEditTopbarPanel from '@/components/platform/contest/ContestEditTopbarPanel.vue'
import ContestEditWorkspacePanel from '@/components/platform/contest/ContestEditWorkspacePanel.vue'
import ContestWorkbenchStageTabs from '@/components/platform/contest/ContestWorkbenchStageTabs.vue'
import AWDReadinessOverrideDialog from '@/components/platform/contest/AWDReadinessOverrideDialog.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import {
  buildContestUpdatePayload,
  confirmContestTermination,
  createFieldLocks,
  createContestStatusOptions,
  createDraftFromContest,
  normalizeEditableStatus,
  shouldConfirmContestTermination,
  type PlatformContestStatus,
  type ContestFormDraft,
} from '@/composables/usePlatformContests'
import {
  CONTEST_WORKBENCH_STAGE_ORDER,
  useContestWorkbench,
  type ContestWorkbenchStageKey,
} from '@/composables/useContestWorkbench'
import { useContestEditAwdWorkspace } from '@/composables/useContestEditAwdWorkspace'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
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
const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

const contestId = computed(() => String(route.params.id ?? ''))
const loading = ref(true)
const loadError = ref('')
const saving = ref(false)
const contest = ref<ContestDetailData | null>(null)
const editingBaseStatus = ref<PlatformContestStatus | null>(null)
const formDraft = ref<ContestFormDraft | null>(null)
const awdStartOverrideDialogState = ref<AWDStartOverrideDialogState>(createDefaultAWDStartOverrideDialogState())

const fieldLocks = computed(() => createFieldLocks(editingBaseStatus.value))
const statusOptions = computed(() => createContestStatusOptions(editingBaseStatus.value))
const pageTitle = computed(() => contest.value?.title || '未命名竞赛')
const { activeTab: activeStage, selectTab } = useUrlSyncedTabs<ContestWorkbenchStageKey>({
  orderedTabs: CONTEST_WORKBENCH_STAGE_ORDER,
  defaultTab: 'basics',
})
const {
  activeAwdChallengeId,
  awdChallengeConfigDialogOpen,
  awdChallengeConfigMode,
  awdChallengeLinks,
  awdChallengeLinksLoaded,
  awdConfigFocusSource,
  awdPreflightLoadError,
  awdReadiness,
  awdServiceTemplateCatalog,
  canNavigateNextAwdChallenge,
  canNavigatePreviousAwdChallenge,
  editingAwdChallengeLink,
  existingAwdChallengeIds,
  focusAwdChallengeByOffset,
  handleNavigateAwdChallengeFromOperations,
  handleNavigateAwdChallengeFromPreflight,
  handleOpenAwdConfigFromPool,
  handleSaveAwdChallengeConfig,
  loadingAwdServiceTemplateCatalog,
  loadingAwdStageData,
  openAwdChallengeCreateDialog,
  openAwdChallengeEditDialog,
  refreshAwdWorkbenchData,
  savingChallengeConfig,
} = useContestEditAwdWorkspace({
  contest,
  contestId,
  selectTab,
})
const awdWorkbenchChallengeCount = computed(() =>
  contest.value?.mode === 'awd' && awdChallengeLinksLoaded.value ? awdChallengeLinks.value.length : null
)
const workbench = useContestWorkbench(contest, awdWorkbenchChallengeCount)

function createDefaultAWDStartOverrideDialogState(): AWDStartOverrideDialogState {
  return {
    open: false,
    title: '',
    readiness: null,
    confirmLoading: false,
    pendingPayload: null,
  }
}

function shouldGateAWDContestStart(
  mode: ContestDetailData['mode'] | null,
  targetStatus: PlatformContestStatus
): boolean {
  return mode === 'awd' && targetStatus === 'running'
}

function isAWDReadinessBlockedError(error: unknown): boolean {
  return error instanceof ApiError && error.code === ERR_AWD_READINESS_BLOCKED
}

function humanizeRequestError(error: unknown, fallback: string): string {
  if (error instanceof ApiError && error.message.trim()) return error.message
  if (error instanceof Error && error.message.trim()) return error.message
  return fallback
}

function syncWorkbenchStageSelection(): void {
  const visibleStageKeys = workbench.visibleStages.map((stage) => stage.key)
  const searchParams = new URLSearchParams(window.location.search)
  const requestedStage = searchParams.get('panel') as ContestWorkbenchStageKey | null
  if (requestedStage && visibleStageKeys.includes(requestedStage)) {
    if (activeStage.value !== requestedStage) selectTab(requestedStage)
    return
  }
  if (!requestedStage) {
    if (activeStage.value !== workbench.defaultStage) selectTab(workbench.defaultStage)
    return
  }
  if (!visibleStageKeys.includes(activeStage.value) || activeStage.value !== workbench.defaultStage) {
    selectTab(workbench.defaultStage)
  }
}

function handleDraftChange(nextDraft: ContestFormDraft) {
  formDraft.value = { ...nextDraft }
}

async function openPreflightOverrideDialog() {
  if (!contest.value || !formDraft.value) return
  const payload = buildContestUpdatePayload(
    {
      ...formDraft.value,
      status: 'running',
    },
    fieldLocks.value
  )
  await openAWDStartOverrideDialog(payload)
}

async function openAWDStartOverrideDialog(payload: AdminContestUpdatePayload) {
  if (!contest.value) {
    return
  }

  try {
    const readiness = await getContestAWDReadiness(contest.value.id)
    awdStartOverrideDialogState.value = {
      open: true,
      title: '启动赛事',
      readiness,
      confirmLoading: false,
      pendingPayload: payload,
    }
  } catch (error) {
    toast.error(humanizeRequestError(error, '读取开赛校验失败'))
  }
}

async function loadContestDetail(): Promise<void> {
  if (!contestId.value) {
    setBreadcrumbDetailTitle()
    return
  }
  loading.value = true
  try {
    const detail = await getContest(contestId.value)
    contest.value = detail
    setBreadcrumbDetailTitle(detail.title)
    editingBaseStatus.value = normalizeEditableStatus(detail.status)
    formDraft.value = createDraftFromContest(detail)
    syncWorkbenchStageSelection()
    if (detail.mode === 'awd') await refreshAwdWorkbenchData(detail.id)
  } catch (error) {
    setBreadcrumbDetailTitle()
    loadError.value = humanizeRequestError(error, '竞赛详情加载失败')
  } finally {
    loading.value = false
  }
}

function goBackToContestList() {
  void router.push({ name: 'ContestManage', query: { panel: 'list' } })
}

function goToContestAnnouncements() {
  void router.push({ name: 'ContestAnnouncements', params: { id: contestId.value } })
}

function handleWorkspaceStageNavigation(stage: ContestWorkbenchStageKey) {
  selectTab(stage)
}

async function handleSave(draft: ContestFormDraft): Promise<void> {
  if (!contest.value) return
  saving.value = true
  try {
    const payload = buildContestUpdatePayload(draft, fieldLocks.value)

    if (shouldConfirmContestTermination(editingBaseStatus.value, draft.status)) {
      const confirmed = await confirmContestTermination(draft.title.trim())
      if (!confirmed) {
        return
      }
    }

    if (shouldGateAWDContestStart(draft.mode, draft.status)) {
      try {
        await updateContest(contestId.value, payload, { suppressErrorToast: true })
      } catch (error) {
        if (isAWDReadinessBlockedError(error)) {
          await openAWDStartOverrideDialog(payload)
          return
        }
        throw error
      }
    } else {
      await updateContest(contestId.value, payload)
    }

    toast.success('竞赛已更新')
    goBackToContestList()
  } catch (error) {
    toast.error(humanizeRequestError(error, '更新失败'))
  } finally {
    saving.value = false
  }
}

function handleAwdStartOverrideDialogOpenChange(value: boolean) {
  if (!value) awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
}

async function confirmAWDStartOverride(reason: string) {
  if (!contest.value) {
    return
  }

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
      contest.value.id,
      {
        ...payload,
        force_override: true,
        override_reason: normalizedReason,
      },
      { suppressErrorToast: true }
    )
    toast.success('竞赛已更新')
    awdStartOverrideDialogState.value = createDefaultAWDStartOverrideDialogState()
    goBackToContestList()
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

function getModeLabel(mode: string): string {
  return mode === 'awd' ? 'AWD Mode' : 'Jeopardy'
}

function getStatusLabel(status: string): string {
  switch (status) {
    case 'running': return 'Live'
    case 'registering': return 'Registration'
    case 'ended': return 'Finished'
    case 'frozen': return 'Frozen'
    default: return 'Draft'
  }
}

onMounted(() => {
  void loadContestDetail()
})

onUnmounted(() => {
  setBreadcrumbDetailTitle()
})
</script>

<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero contest-studio-shell">
    <div
      v-if="loading"
      class="studio-loading-overlay"
    >
      <AppLoading>正在同步竞赛工作台...</AppLoading>
    </div>

    <main class="studio-content">
      <ContestEditTopbarPanel
        v-if="contest"
        :page-title="pageTitle"
        :contest-mode="contest.mode"
        :contest-status="contest.status"
        :contest-mode-label="getModeLabel(contest.mode)"
        :contest-status-label="getStatusLabel(contest.status)"
        :active-stage="activeStage"
        :saving="saving"
        @back="goBackToContestList"
        @open-announcements="goToContestAnnouncements"
        @save="formDraft && void handleSave(formDraft)"
      />

      <ContestWorkbenchStageTabs
        v-if="contest"
        :stages="workbench.visibleStages"
        :active-stage="activeStage"
        :select-stage="selectTab"
      />

      <ContestEditWorkspacePanel
        :load-error="loadError"
        :form-draft="formDraft"
        :contest="contest"
        :active-stage="activeStage"
        :saving="saving"
        :status-options="statusOptions"
        :field-locks="fieldLocks"
        :loading-awd-stage-data="loadingAwdStageData"
        :awd-challenge-links="awdChallengeLinks"
        :active-awd-challenge-id="activeAwdChallengeId"
        :awd-config-focus-source="awdConfigFocusSource"
        :can-navigate-previous-awd-challenge="canNavigatePreviousAwdChallenge"
        :can-navigate-next-awd-challenge="canNavigateNextAwdChallenge"
        :awd-preflight-load-error="awdPreflightLoadError"
        :awd-readiness="awdReadiness"
        @go-back="goBackToContestList"
        @update:draft="handleDraftChange"
        @save="handleSave"
        @refresh-awd-workbench="contest && void refreshAwdWorkbenchData(contest.id)"
        @open:awd-config-from-pool="handleOpenAwdConfigFromPool"
        @create:awd-challenge="openAwdChallengeCreateDialog"
        @edit:awd-challenge="openAwdChallengeEditDialog"
        @previous:awd-challenge="focusAwdChallengeByOffset(-1)"
        @next:awd-challenge="focusAwdChallengeByOffset(1)"
        @retry:preflight="contest && void refreshAwdWorkbenchData(contest.id)"
        @navigate:awd-challenge-from-preflight="handleNavigateAwdChallengeFromPreflight"
        @navigate:stage="handleWorkspaceStageNavigation"
        @open:preflight-override="openPreflightOverrideDialog"
        @open:awd-config-from-operations="handleNavigateAwdChallengeFromOperations"
      />
    </main>

    <AWDChallengeConfigDialog
      v-if="contest?.mode === 'awd'"
      :contest-id="contest.id"
      :open="awdChallengeConfigDialogOpen"
      :mode="awdChallengeConfigMode"
      :challenge-options="[]"
      :template-options="awdServiceTemplateCatalog"
      :existing-challenge-ids="existingAwdChallengeIds"
      :draft="editingAwdChallengeLink"
      :loading-challenge-catalog="false"
      :loading-template-catalog="loadingAwdServiceTemplateCatalog"
      :saving="savingChallengeConfig"
      @update:open="awdChallengeConfigDialogOpen = $event"
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
  </div>
</template>

<style scoped>
.contest-studio-shell {
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  display: flex;
  flex-direction: column;
  height: calc(100vh - 64px);
  width: 100%;
  overflow: hidden;
  background: var(--color-bg-base);
}

.studio-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  min-width: 0;
}

.studio-loading-overlay {
  position: absolute;
  inset: 0;
  z-index: 100;
  background: color-mix(in srgb, var(--color-bg-base) 80%, transparent);
  backdrop-filter: blur(12px);
  display: flex;
  align-items: center;
  justify-content: center;
}

</style>
