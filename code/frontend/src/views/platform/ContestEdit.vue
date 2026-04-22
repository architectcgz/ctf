<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Bell, ChevronLeft, Trophy, Save, ShieldCheck } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import {
  createContestAWDService,
  getContest,
  getContestAWDReadiness,
  listAdminAwdServiceTemplates,
  listContestAWDServices,
  updateContestAWDService,
  updateContest,
} from '@/api/admin'
import type {
  AdminAwdServiceTemplateData,
  AdminContestAWDServiceSavePayload,
  AdminContestChallengeViewData,
  AWDReadinessData,
  ContestDetailData,
} from '@/api/contracts'
import type { AdminContestUpdatePayload } from '@/api/admin'
import PlatformContestFormPanel from '@/components/platform/contest/PlatformContestFormPanel.vue'
import AWDChallengeConfigDialog from '@/components/platform/contest/AWDChallengeConfigDialog.vue'
import AWDChallengeConfigPanel from '@/components/platform/contest/AWDChallengeConfigPanel.vue'
import AWDOperationsPanel from '@/components/platform/contest/AWDOperationsPanel.vue'
import ContestAwdPreflightPanel from '@/components/platform/contest/ContestAwdPreflightPanel.vue'
import ContestChallengeOrchestrationPanel from '@/components/platform/contest/ContestChallengeOrchestrationPanel.vue'
import ContestWorkbenchStageTabs from '@/components/platform/contest/ContestWorkbenchStageTabs.vue'
import AWDReadinessOverrideDialog from '@/components/platform/contest/AWDReadinessOverrideDialog.vue'
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
  type PlatformContestStatus,
  type ContestFormDraft,
} from '@/composables/usePlatformContests'
import {
  CONTEST_WORKBENCH_STAGE_ORDER,
  useContestWorkbench,
  type ContestWorkbenchStageKey,
} from '@/composables/useContestWorkbench'
import { ApiError } from '@/api/request'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'
import { useToast } from '@/composables/useToast'
import { mapPlatformContestAwdServicesToChallengeLinks } from '@/utils/platformContestAwdChallengeLinks'

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
const savingChallengeConfig = ref(false)
const contest = ref<ContestDetailData | null>(null)
const editingBaseStatus = ref<PlatformContestStatus | null>(null)
const formDraft = ref<ContestFormDraft | null>(null)
const awdConfigLoadError = ref('')
const awdPreflightLoadError = ref('')
const awdChallengeLinks = ref<AdminContestChallengeViewData[]>([])
const awdChallengeLinksLoaded = ref(false)
const awdReadiness = ref<AWDReadinessData | null>(null)
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
const pageTitle = computed(() => contest.value?.title || '未命名竞赛')
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

function resetAwdWorkbenchState() {
  awdConfigLoadError.value = ''
  awdPreflightLoadError.value = ''
  awdChallengeLinks.value = []
  awdChallengeLinksLoaded.value = false
  awdReadiness.value = null
  awdServiceTemplateCatalog.value = []
  awdChallengeConfigDialogOpen.value = false
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
    const [awdServicesResult, readinessResult] = await Promise.allSettled([
      listContestAWDServices(nextContestId),
      getContestAWDReadiness(nextContestId),
    ])
    if (awdServicesResult.status === 'fulfilled') {
      awdChallengeLinks.value = mapPlatformContestAwdServicesToChallengeLinks(
        awdServicesResult.value
      )
      awdChallengeLinksLoaded.value = true
    } else {
      awdConfigLoadError.value = humanizeRequestError(awdServicesResult.reason, 'AWD 配置同步失败')
      toast.error(awdConfigLoadError.value)
    }
    if (readinessResult.status === 'fulfilled') {
      awdReadiness.value = readinessResult.value
    } else {
      awdPreflightLoadError.value = humanizeRequestError(
        readinessResult.reason,
        '赛前检查同步失败'
      )
      toast.error(awdPreflightLoadError.value)
    }
  } finally {
    loadingAwdStageData.value = false
  }
}

async function loadAwdServiceTemplateCatalog(): Promise<void> {
  if (loadingAwdServiceTemplateCatalog.value || awdServiceTemplateCatalog.value.length > 0) return
  loadingAwdServiceTemplateCatalog.value = true
  try {
    const result = await listAdminAwdServiceTemplates({ page: 1, page_size: 100, status: 'published' })
    awdServiceTemplateCatalog.value = result.list
  } catch (error) {
    toast.error(humanizeRequestError(error, '服务模板加载失败'))
  } finally {
    loadingAwdServiceTemplateCatalog.value = false
  }
}

function handleDraftChange(nextDraft: ContestFormDraft) {
  formDraft.value = { ...nextDraft }
}

function setActiveAwdChallenge(challengeId: string | null, source: 'pool' | 'preflight' | null) {
  activeAwdChallengeId.value = challengeId
  awdConfigFocusSource.value = challengeId ? source : null
}

function focusAwdChallengeByOffset(offset: -1 | 1) {
  if (activeAwdChallengeIndex.value < 0) return
  const nextChallenge = sortedAwdChallengeLinks.value[activeAwdChallengeIndex.value + offset]
  if (!nextChallenge) return
  setActiveAwdChallenge(nextChallenge.challenge_id, awdConfigFocusSource.value)
}

function openAwdChallengeCreateDialog() {
  awdChallengeConfigMode.value = 'create'
  editingAwdChallengeLink.value = null
  awdChallengeConfigDialogOpen.value = true
  void loadAwdServiceTemplateCatalog()
}

function openAwdChallengeEditDialog(challenge: AdminContestChallengeViewData) {
  setActiveAwdChallenge(challenge.challenge_id, awdConfigFocusSource.value)
  awdChallengeConfigMode.value = 'edit'
  editingAwdChallengeLink.value = challenge
  awdChallengeConfigDialogOpen.value = true
  void loadAwdServiceTemplateCatalog()
}

async function handleSaveAwdChallengeConfig(payload: AdminContestAWDServiceSavePayload) {
  if (!contest.value) return
  savingChallengeConfig.value = true
  try {
    if (awdChallengeConfigMode.value === 'create') {
      await createContestAWDService(contest.value.id, payload)
    } else if (editingAwdChallengeLink.value) {
      await updateContestAWDService(contest.value.id, editingAwdChallengeLink.value.awd_service_id!, payload)
    }
    awdChallengeConfigDialogOpen.value = false
    await refreshAwdWorkbenchData(contest.value.id)
  } catch (error) {
    toast.error(humanizeRequestError(error, '保存 AWD 配置失败'))
  } finally {
    savingChallengeConfig.value = false
  }
}

function handleOpenAwdConfigFromPool(challenge: AdminContestChallengeViewData) {
  activeAwdChallengeId.value = challenge.challenge_id
  awdConfigFocusSource.value = 'pool'
  selectTab('awd-config')
}

function handleNavigateAwdChallengeFromPreflight(challengeId: string) {
  setActiveAwdChallenge(challengeId, 'preflight')
  selectTab('awd-config')
}

function handleNavigateAwdChallengeFromOperations(challengeId: string) {
  setActiveAwdChallenge(challengeId, null)
  selectTab('awd-config')
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
  if (!contestId.value) return
  loading.value = true
  try {
    const detail = await getContest(contestId.value)
    contest.value = detail
    editingBaseStatus.value = normalizeEditableStatus(detail.status)
    formDraft.value = createDraftFromContest(detail)
    syncWorkbenchStageSelection()
    if (detail.mode === 'awd') await refreshAwdWorkbenchData(detail.id)
  } catch (error) {
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
</script>

<template>
  <div class="contest-studio-shell">
    <div
      v-if="loading"
      class="studio-loading-overlay"
    >
      <AppLoading>正在同步竞赛工作台...</AppLoading>
    </div>

    <main class="studio-content">
      <header
        v-if="contest"
        class="studio-topbar"
      >
        <div class="studio-topbar-left">
          <button
            class="studio-back-btn"
            title="退出工作室"
            @click="goBackToContestList"
          >
            <ChevronLeft class="h-5 w-5" />
          </button>
          
          <div class="studio-title-group">
            <h1
              class="studio-contest-title"
              :title="pageTitle"
            >
              {{ pageTitle }}
            </h1>
            <div class="studio-contest-meta">
              <span
                class="meta-tag"
                :class="`meta-tag--${contest.mode}`"
              >
                <Trophy class="h-3 w-3" /> {{ getModeLabel(contest.mode) }}
              </span>
              <span class="meta-tag meta-tag--status">
                <ShieldCheck class="h-3 w-3" /> {{ getStatusLabel(contest.status) }}
              </span>
            </div>
          </div>
        </div>

        <div class="studio-topbar-right">
          <button
            id="contest-open-announcements"
            type="button"
            class="studio-toolbar-btn"
            @click="goToContestAnnouncements"
          >
            <Bell class="h-4 w-4" />
            <span>公告</span>
          </button>
          <button
            v-if="activeStage === 'basics'"
            type="button"
            class="studio-save-btn"
            :disabled="saving"
            @click="() => (formDraft && handleSave(formDraft))"
          >
            <Save class="h-4 w-4" />
            <span>{{ saving ? '正在保存...' : '保存变更' }}</span>
          </button>
        </div>
      </header>

      <!-- Horizontal Stage Navigation -->
      <ContestWorkbenchStageTabs
        v-if="contest"
        :stages="workbench.visibleStages"
        :active-stage="activeStage"
        :select-stage="selectTab"
      />

      <div class="studio-canvas">
        <div class="studio-scroll-area">
          <AppEmpty
            v-if="loadError"
            title="竞赛详情加载失败"
            :description="loadError"
            icon="AlertTriangle"
          >
            <template #action>
              <button
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="goBackToContestList"
              >
                返回竞赛目录
              </button>
            </template>
          </AppEmpty>

          <template v-else-if="formDraft && contest">
            <!-- 基础配置 -->
            <div
              v-if="activeStage === 'basics'"
              class="studio-pane studio-pane--full fade-in"
            >
              <div class="studio-form-canvas">
                <PlatformContestFormPanel
                  :mode="'edit'"
                  :draft="formDraft"
                  :saving="saving"
                  :status-options="statusOptions"
                  :field-locks="fieldLocks"
                  :show-cancel="false"
                  @update:draft="handleDraftChange"
                  @save="handleSave"
                />
              </div>
            </div>

            <!-- 题目编排 -->
            <div
              v-if="activeStage === 'pool'"
              class="studio-pane fade-in"
            >
              <ContestChallengeOrchestrationPanel
                :contest-id="contest.id"
                :contest-mode="contest.mode"
                :challenge-links="contest.mode === 'awd' ? awdChallengeLinks : undefined"
                :loading-external="loadingAwdStageData"
                @open:awd-config="handleOpenAwdConfigFromPool"
                @updated="refreshAwdWorkbenchData(contest.id)"
              />
            </div>

            <!-- AWD 服务配置 -->
            <div
              v-if="contest.mode === 'awd' && activeStage === 'awd-config'"
              class="studio-pane fade-in"
            >
              <template v-if="loadingAwdStageData && awdChallengeLinks.length === 0">
                <AppLoading>正在同步 AWD 配置...</AppLoading>
              </template>
              <AWDChallengeConfigPanel
                v-else
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
            </div>

            <!-- 赛前就绪检查 -->
            <div
              v-if="contest.mode === 'awd' && activeStage === 'preflight'"
              class="studio-pane fade-in"
            >
              <AppEmpty
                v-if="awdPreflightLoadError"
                title="赛前检查暂时不可用"
                :description="awdPreflightLoadError"
                icon="AlertTriangle"
              >
                <template #action>
                  <button
                    type="button"
                    class="ui-btn ui-btn--ghost"
                    @click="refreshAwdWorkbenchData(contest.id)"
                  >
                    重试加载
                  </button>
                </template>
              </AppEmpty>
              <ContestAwdPreflightPanel
                v-else
                :readiness="awdReadiness"
                :loading="loadingAwdStageData"
                @navigate:challenge="handleNavigateAwdChallengeFromPreflight"
                @navigate:stage="selectTab"
                @open:override="openPreflightOverrideDialog"
              />
            </div>

            <!-- 赛场运维 -->
            <div
              v-if="contest.mode === 'awd' && activeStage === 'operations'"
              class="studio-pane fade-in"
            >
              <AWDOperationsPanel
                :contests="[contest]"
                :selected-contest-id="contest.id"
                :hide-contest-selector="true"
                @open:awd-config="handleNavigateAwdChallengeFromOperations"
              />
            </div>
          </template>
        </div>
      </div>
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

.studio-topbar {
  height: 4.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2rem;
  background: var(--color-bg-surface);
  border-bottom: 1px solid var(--workspace-line-soft);
  z-index: 10;
}

.studio-topbar-left {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.studio-topbar-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.studio-back-btn {
  width: 2.5rem;
  height: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.75rem;
  color: var(--journal-muted);
  border: 1px solid var(--workspace-line-soft);
  transition: all 0.2s ease;
  cursor: pointer;
}
.studio-back-btn:hover {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
  border-color: var(--color-border-default);
}

.studio-toolbar-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  height: 2.4rem;
  padding: 0 1rem;
  border-radius: 0.85rem;
  border: 1px solid var(--workspace-line-soft);
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  font-size: 12px;
  font-weight: 800;
  transition: all 0.2s ease;
}

.studio-toolbar-btn:hover {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
  border-color: var(--color-border-default);
}

.studio-save-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  height: 2.4rem;
  padding: 0 1.25rem;
  background: var(--color-primary);
  color: white;
  border-radius: 0.85rem;
  font-size: 12px;
  font-weight: 800;
  box-shadow: 0 8px 20px color-mix(in srgb, var(--color-primary) 24%, transparent);
  transition: all 0.2s ease;
}

.studio-save-btn:hover {
  background: var(--color-primary-hover);
  transform: translateY(-1px);
  box-shadow: 0 10px 24px color-mix(in srgb, var(--color-primary) 30%, transparent);
}

.studio-title-group {
  display: flex;
  align-items: baseline;
  gap: 1.25rem;
}

.studio-contest-title {
  font-size: 1rem;
  font-weight: 900;
  letter-spacing: -0.01em;
  color: var(--color-text-primary);
  margin: 0;
  max-width: 24rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.studio-contest-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.meta-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.1rem 0.55rem;
  border-radius: 4px;
  font-size: 9px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border: 1px solid transparent;
}

.meta-tag--awd {
  background: color-mix(in srgb, var(--color-primary) 8%, transparent);
  color: var(--color-primary);
  border-color: color-mix(in srgb, var(--color-primary) 20%, transparent);
}

.meta-tag--status {
  background: color-mix(in srgb, var(--journal-muted) 8%, transparent);
  color: var(--journal-muted);
  border-color: color-mix(in srgb, var(--journal-muted) 20%, transparent);
}

.studio-canvas {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0; /* 关键：防止 Flex 子项溢出导致滚动失效 */
  position: relative;
  background: var(--color-bg-surface, #ffffff);
}

.studio-scroll-area {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0;
  display: flex;
  flex-direction: column;
}

.studio-scroll-area::-webkit-scrollbar { width: 6px; }
.studio-scroll-area::-webkit-scrollbar-thumb { background: color-mix(in srgb, var(--workspace-line-soft) 50%, transparent); border-radius: 10px; }

.studio-pane {
  width: 100%;
  flex: 1 0 auto; /* 确保子项能撑开内容 */
}

/* 基础设置表单：去卡片化，自然平铺 */
.studio-form-canvas {
  background: transparent;
  border: none;
  padding: 2rem;
  box-shadow: none;
  width: 100%;
  max-width: 64rem;
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

.fade-in { animation: studioFadeIn 0.5s cubic-bezier(0.4, 0, 0.2, 1); }
@keyframes studioFadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
</style>
