import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type {
  ChallengeDetailData,
  CommunityChallengeSolutionData,
  RecommendedChallengeSolutionData,
} from '@/api/contracts'
import { useProbeEasterEggs } from '@/composables/useProbeEasterEggs'
import { useSanitize } from '@/composables/useSanitize'
import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'
import { useToast } from '@/composables/useToast'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'

import {
  useChallengeDetailInteractions,
  useChallengeDetailPresentation,
  type ChallengeSolutionTab,
} from '.'
import { useChallengeDetailDataLoader } from './useChallengeDetailDataLoader'
import { useChallengeInstance } from './useChallengeInstance'

type WorkspaceTab = 'question' | 'solution' | 'records' | 'writeup'

export function useChallengeDetailPage() {
  const route = useRoute()
  const router = useRouter()
  const toast = useToast()
  const { sanitizeHtml } = useSanitize()
  const { track } = useProbeEasterEggs()

  const challengeId = computed(() => String(route.params.id ?? ''))
  const challenge = ref<ChallengeDetailData | null>(null)
  const loading = ref(false)
  const scoreRailProbeMessage = ref('')
  const recommendedSolutions = ref<RecommendedChallengeSolutionData[]>([])
  const communitySolutions = ref<CommunityChallengeSolutionData[]>([])
  const selectedSolutionId = ref<string | null>(null)
  const submissionRecordPage = ref(1)

  let scoreRailProbeTimer: number | null = null

  const {
    instance,
    loading: instanceLoading,
    creating: instanceCreating,
    opening: instanceOpening,
    extending: instanceExtending,
    destroying: instanceDestroying,
    start: startInstance,
    open: openInstance,
    extend: extendChallengeInstance,
    destroy: destroyChallengeInstance,
  } = useChallengeInstance(challengeId)

  const workspaceTabs: Array<{ id: WorkspaceTab; label: string }> = [
    { id: 'question', label: '题目' },
    { id: 'solution', label: '题解' },
    { id: 'records', label: '提交记录' },
    { id: 'writeup', label: '编写题解' },
  ]
  const workspaceTabOrder = workspaceTabs.map((tab) => tab.id) as WorkspaceTab[]
  const {
    activeTab: activeWorkspaceTab,
    setTabButtonRef,
    selectTab: selectWorkspaceTab,
    handleTabKeydown: handleWorkspaceTabKeydown,
  } = useUrlSyncedTabs<WorkspaceTab>({
    orderedTabs: workspaceTabOrder,
    defaultTab: 'question',
  })

  const solutionTabOrder: ChallengeSolutionTab[] = ['recommended', 'community']
  const submissionRecordPageSize = 10
  const { clearSolutions, loadSolutions, loadChallenge } = useChallengeDetailDataLoader({
    challengeId,
    challenge,
    loading,
    recommendedSolutions,
    communitySolutions,
    onSolutionsLoadFailed: () => {
      toast.error('加载题解失败')
    },
    onChallengeLoadFailed: () => {
      toast.error('加载题目详情失败')
      void router.push('/challenges')
    },
  })

  const {
    myWriteup,
    submitting,
    submissionLoading,
    submissionSaving,
    writeupTitle,
    writeupContent,
    flagInput,
    submitResult,
    submissionRecords,
    resetChallengeInteractions,
    loadMyWriteupSubmission,
    loadSubmissionRecords,
    isHintExpanded,
    toggleHint,
    submitFlagHandler,
    downloadAttachment,
    saveWriteup,
  } = useChallengeDetailInteractions({
    challengeId,
    challenge,
    loadSolutions,
  })

  const submissionRecordTotal = computed(() => submissionRecords.value.length)
  const submissionRecordTotalPages = computed(() =>
    Math.max(1, Math.ceil(submissionRecordTotal.value / submissionRecordPageSize))
  )
  const paginatedSubmissionRecords = computed(() => {
    const start = (submissionRecordPage.value - 1) * submissionRecordPageSize
    return submissionRecords.value.slice(start, start + submissionRecordPageSize)
  })

  const needTarget = computed(() => challenge.value?.need_target ?? true)
  const {
    activeSolutionTab,
    sanitizedDescription,
    displayedSolutionCards,
    activeSolution,
    sanitizedActiveSolutionContent,
    submitPlaceholder,
    submitPanelTitle,
    submitPanelCopy,
    submitFieldLabel,
    submitInputClass,
    submissionStatusLabel,
    submissionStatusText,
    submissionRecordMessage,
    formatWriteupTime,
    formatSubmissionTime,
  } = useChallengeDetailPresentation({
    challenge,
    recommendedSolutions,
    communitySolutions,
    myWriteup,
    selectedSolutionId,
    submitResult,
    sanitizeHtml,
  })

  const {
    setTabButtonRef: setSolutionTabButtonRef,
    handleTabKeydown: handleSolutionTabKeydown,
  } = useTabKeyboardNavigation<ChallengeSolutionTab>({
    orderedTabs: solutionTabOrder,
    selectTab: selectSolutionTab,
  })

  function showScoreRailProbe(message: string) {
    scoreRailProbeMessage.value = message
    if (scoreRailProbeTimer) {
      window.clearTimeout(scoreRailProbeTimer)
    }
    scoreRailProbeTimer = window.setTimeout(() => {
      scoreRailProbeMessage.value = ''
      scoreRailProbeTimer = null
    }, 2800)
  }

  function handleScoreRailProbe() {
    const result = track('challenge-side-rail', 4)
    if (!result.unlocked) {
      return
    }
    showScoreRailProbe('这块区域的情报价值，低于你现在的期待。')
  }

  function selectSolutionTab(tab: ChallengeSolutionTab): void {
    activeSolutionTab.value = tab
  }

  function changeSubmissionRecordPage(page: number): void {
    submissionRecordPage.value = page
  }

  watch(
    challengeId,
    () => {
      challenge.value = null
      submissionRecordPage.value = 1
      resetChallengeInteractions()
      clearSolutions()
      selectWorkspaceTab('question')
      void Promise.all([loadChallenge(), loadMyWriteupSubmission(), loadSubmissionRecords()])
    },
    { immediate: true }
  )

  watch(
    submissionRecords,
    () => {
      submissionRecordPage.value = 1
    },
    { deep: true }
  )

  onBeforeUnmount(() => {
    if (scoreRailProbeTimer) {
      window.clearTimeout(scoreRailProbeTimer)
    }
  })

  return {
    activeSolution,
    activeSolutionTab,
    activeWorkspaceTab,
    challenge,
    changeSubmissionRecordPage,
    displayedSolutionCards,
    downloadAttachment,
    extendChallengeInstance,
    flagInput,
    formatSubmissionTime,
    formatWriteupTime,
    handleScoreRailProbe,
    handleSolutionTabKeydown,
    handleWorkspaceTabKeydown,
    instance,
    instanceCreating,
    instanceDestroying,
    instanceExtending,
    instanceLoading,
    instanceOpening,
    isHintExpanded,
    loadChallenge,
    loadMyWriteupSubmission,
    loadSubmissionRecords,
    loading,
    myWriteup,
    needTarget,
    openInstance,
    paginatedSubmissionRecords,
    communitySolutions,
    recommendedSolutions,
    saveWriteup,
    sanitizedActiveSolutionContent,
    sanitizedDescription,
    scoreRailProbeMessage,
    selectSolutionTab,
    selectWorkspaceTab,
    selectedSolutionId,
    setSolutionTabButtonRef,
    setTabButtonRef,
    startInstance,
    submissionLoading,
    submissionRecordMessage,
    submissionRecordPage,
    submissionRecordTotal,
    submissionRecordTotalPages,
    submissionRecords,
    submissionSaving,
    submissionStatusLabel,
    submissionStatusText,
    submitFieldLabel,
    submitFlagHandler,
    submitInputClass,
    submitPanelCopy,
    submitPanelTitle,
    submitPlaceholder,
    submitResult,
    submitting,
    toggleHint,
    workspaceTabs,
    writeupContent,
    writeupTitle,
    destroyChallengeInstance,
  }
}
