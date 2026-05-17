import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { ApiError } from '@/api/request'
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

interface ChallengeLoadState {
  icon: 'AlertTriangle' | 'Flag' | 'Inbox'
  title: string
  description: string
  retryable: boolean
}

export function useChallengeDetailPage() {
  const route = useRoute()
  const router = useRouter()
  const toast = useToast()
  const { sanitizeHtml } = useSanitize()
  const { track } = useProbeEasterEggs()

  const challengeId = computed(() => String(route.params.id ?? ''))
  const challenge = ref<ChallengeDetailData | null>(null)
  const challengeLoadState = ref<ChallengeLoadState | null>(null)
  const loading = ref(false)
  const solutionsLoading = ref(false)
  const scoreRailProbeMessage = ref('')
  const recommendedSolutions = ref<RecommendedChallengeSolutionData[]>([])
  const communitySolutions = ref<CommunityChallengeSolutionData[]>([])
  const selectedSolutionId = ref<string | null>(null)
  const submissionRecordPage = ref(1)
  const hasLoadedSolutions = ref(false)
  const hasLoadedSubmissionRecords = ref(false)
  const hasLoadedWriteup = ref(false)

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
    solutionsLoading,
    recommendedSolutions,
    communitySolutions,
    onSolutionsLoadFailed: () => {
      toast.error('加载题解失败')
    },
    onChallengeLoadFailed: (error) => {
      challengeLoadState.value = resolveChallengeLoadState(error)
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
    submissionRecordsLoading,
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

  function resolveChallengeLoadState(error: unknown): ChallengeLoadState {
    if (error instanceof ApiError) {
      if (error.code === 13004) {
        return {
          icon: 'Inbox',
          title: '题目不存在',
          description: '该题目可能已被删除，或当前链接已失效。',
          retryable: false,
        }
      }
      if (error.code === 13005) {
        if (error.message.includes('草稿')) {
          return {
            icon: 'Flag',
            title: '草稿题目暂不可访问',
            description: '当前题目还处于草稿状态，尚未开放访问。',
            retryable: false,
          }
        }
        if (error.message.includes('归档')) {
          return {
            icon: 'Flag',
            title: '已归档题目不可访问',
            description: '当前题目已归档，不再提供访问入口。',
            retryable: false,
          }
        }
        return {
          icon: 'Flag',
          title: '题目暂不可访问',
          description: '当前题目尚未发布，暂时不能进入详情页。',
          retryable: false,
        }
      }
      return {
        icon: 'AlertTriangle',
        title: '题目详情加载失败',
        description: error.message.trim() || '当前无法读取题目详情，请稍后重试。',
        retryable: true,
      }
    }

    if (error instanceof Error && error.message.trim()) {
      return {
        icon: 'AlertTriangle',
        title: '题目详情加载失败',
        description: error.message.trim(),
        retryable: true,
      }
    }

    return {
      icon: 'AlertTriangle',
      title: '题目详情加载失败',
      description: '当前无法读取题目详情，请稍后重试。',
      retryable: true,
    }
  }

  function changeSubmissionRecordPage(page: number): void {
    submissionRecordPage.value = page
  }

  function goBackToChallengeList(): void {
    void router.push('/challenges')
  }

  function resetChallengePageState(resetWorkspaceTab = false): void {
    challenge.value = null
    challengeLoadState.value = null
    submissionRecordPage.value = 1
    hasLoadedSolutions.value = false
    hasLoadedSubmissionRecords.value = false
    hasLoadedWriteup.value = false
    resetChallengeInteractions()
    clearSolutions()
    if (resetWorkspaceTab) {
      selectWorkspaceTab('question')
    }
  }

  async function ensureWorkspaceTabData(tab: WorkspaceTab): Promise<void> {
    if (!challenge.value || challengeLoadState.value) {
      return
    }

    if (tab === 'solution') {
      if (!challenge.value.is_solved || hasLoadedSolutions.value || solutionsLoading.value) {
        return
      }
      const loaded = await loadSolutions(challenge.value.id)
      hasLoadedSolutions.value = loaded
      return
    }

    if (tab === 'records') {
      if (hasLoadedSubmissionRecords.value || submissionRecordsLoading.value) {
        return
      }
      const loaded = await loadSubmissionRecords()
      hasLoadedSubmissionRecords.value = loaded
      return
    }

    if (tab === 'writeup') {
      if (hasLoadedWriteup.value || submissionLoading.value) {
        return
      }
      const loaded = await loadMyWriteupSubmission()
      hasLoadedWriteup.value = loaded
    }
  }

  async function initializeChallengePage(): Promise<void> {
    const loaded = await loadChallenge()
    if (!loaded || !challenge.value) {
      return
    }
    await ensureWorkspaceTabData(activeWorkspaceTab.value)
  }

  async function retryChallengeLoad(): Promise<void> {
    if (loading.value) {
      return
    }
    resetChallengePageState()
    await initializeChallengePage()
  }

  watch(
    challengeId,
    () => {
      resetChallengePageState(true)
      void initializeChallengePage()
    },
    { immediate: true }
  )

  watch(activeWorkspaceTab, (tab) => {
    void ensureWorkspaceTabData(tab)
  })

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
    challengeLoadState,
    changeSubmissionRecordPage,
    displayedSolutionCards,
    downloadAttachment,
    extendChallengeInstance,
    flagInput,
    formatSubmissionTime,
    formatWriteupTime,
    goBackToChallengeList,
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
    retryChallengeLoad,
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
    submissionRecordsLoading,
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
    solutionsLoading,
    workspaceTabs,
    writeupContent,
    writeupTitle,
    destroyChallengeInstance,
  }
}
