import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getContest, updateContest } from '@/api/admin/contests'
import type { AdminContestChallengeViewData, ContestDetailData } from '@/api/contracts'
import {
  buildContestUpdatePayload,
  confirmContestTermination,
  createContestStatusOptions,
  createDraftFromContest,
  createFieldLocks,
  normalizeEditableStatus,
  shouldConfirmContestTermination,
  type ContestFormDraft,
  type PlatformContestStatus,
} from '@/features/platform-contests'
import {
  CONTEST_WORKBENCH_STAGE_ORDER,
  type ContestWorkbenchStageKey,
  useContestEditAwdWorkspace,
  useContestWorkbench,
} from '@/features/contest-workbench'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useToast } from '@/composables/useToast'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'

export function useContestEditPage() {
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

  const fieldLocks = computed(() => createFieldLocks(editingBaseStatus.value))
  const statusOptions = computed(() => createContestStatusOptions(editingBaseStatus.value))
  const pageTitle = computed(() => contest.value?.title || '未命名竞赛')
  const { activeTab: activeStage, selectTab } = useUrlSyncedTabs<ContestWorkbenchStageKey>({
    orderedTabs: CONTEST_WORKBENCH_STAGE_ORDER,
    defaultTab: 'basics',
  })
  const {
    awdChallengeLinks,
    awdChallengeLinksLoaded,
    awdChallengePoolCreateRequestKey,
    awdPreflightLoadError,
    awdReadiness,
    loadingAwdStageData,
    refreshAwdWorkbenchData,
  } = useContestEditAwdWorkspace({
    contest,
    contestId,
    selectTab,
  })
  const awdWorkbenchChallengeCount = computed(() =>
    contest.value?.mode === 'awd' && awdChallengeLinksLoaded.value ? awdChallengeLinks.value.length : null
  )
  const workbench = useContestWorkbench(contest, awdWorkbenchChallengeCount)

  function humanizeRequestError(error: unknown, fallback: string): string {
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

  function openAwdConfigPage(challenge: AdminContestChallengeViewData) {
    if (!contest.value) return
    void router.push({
      name: 'ContestAWDConfig',
      params: { id: contest.value.id },
      query: challenge.awd_service_id ? { service: challenge.awd_service_id } : undefined,
    })
  }

  function handleNavigateAwdChallengeFromPreflight(challengeId: string) {
    const challenge = awdChallengeLinks.value.find(
      (item) => item.awd_challenge_id === challengeId || item.challenge_id === challengeId
    )
    if (challenge) {
      openAwdConfigPage(challenge)
      return
    }
    if (contest.value) {
      void router.push({ name: 'ContestAWDConfig', params: { id: contest.value.id } })
    }
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
      await updateContest(contestId.value, payload)
      toast.success('竞赛已更新')
      goBackToContestList()
    } catch (error) {
      toast.error(humanizeRequestError(error, '更新失败'))
    } finally {
      saving.value = false
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

  return {
    loading,
    loadError,
    saving,
    contest,
    formDraft,
    fieldLocks,
    statusOptions,
    pageTitle,
    activeStage,
    selectTab,
    workbench,
    awdChallengeLinks,
    awdChallengePoolCreateRequestKey,
    awdPreflightLoadError,
    awdReadiness,
    loadingAwdStageData,
    refreshAwdWorkbenchData,
    handleDraftChange,
    goBackToContestList,
    goToContestAnnouncements,
    handleWorkspaceStageNavigation,
    openAwdConfigPage,
    handleNavigateAwdChallengeFromPreflight,
    handleSave,
    getModeLabel,
    getStatusLabel,
  }
}
