import { computed, onUnmounted, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

import { getMyTeam } from '@/api/contest'
import type {
  ContestAnnouncement,
  ContestChallengeItem,
  ContestDetailData,
  SubmitFlagData,
  TeamData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'
import { useContestDetailCountdown } from './useContestDetailCountdown'
import { useContestDetailDataLoader } from './useContestDetailDataLoader'
import { useContestDetailSelectionSync } from './useContestDetailSelectionSync'
import { useContestFlagSubmission } from './useContestFlagSubmission'
import { useContestTeamActions } from './useContestTeamActions'

interface UseContestDetailPageOptions {
  contestId: MaybeRefOrGetter<string>
  currentUserId: MaybeRefOrGetter<string | undefined>
  selectedChallengeId?: MaybeRefOrGetter<string | null | Array<string | null> | undefined>
  onSelectedChallengeChange?: (challengeId: string | null) => void
}

function createEmptyState() {
  return {
    contest: null as ContestDetailData | null,
    team: null as TeamData | null,
    challenges: [] as ContestChallengeItem[],
    announcements: [] as ContestAnnouncement[],
    announcementsError: '',
    selectedChallenge: null as ContestChallengeItem | null,
    flagInput: '',
    submitResult: null as SubmitFlagData | null,
    showCreateTeam: false,
    showJoinTeam: false,
    teamName: '',
    teamIdInput: '',
  }
}

export function useContestDetailPage(options: UseContestDetailPageOptions) {
  const toast = useToast()

  const contest = ref<ContestDetailData | null>(null)
  const team = ref<TeamData | null>(null)
  const challenges = ref<ContestChallengeItem[]>([])
  const announcements = ref<ContestAnnouncement[]>([])
  const announcementsError = ref('')
  const loading = ref(false)
  const { countdown, startCountdown, stopCountdown } = useContestDetailCountdown({
    contest,
  })
  const {
    selectedChallenge,
    flagInput,
    submitting,
    submitResult,
    clearSubmissionState,
    syncSelectedChallengeById,
    selectChallenge,
    submitFlagAction,
  } = useContestFlagSubmission({
    contest,
    challenges,
    onSelectedChallengeChange: options.onSelectedChallengeChange,
  })
  const { syncSelectedChallengeFromQuery } = useContestDetailSelectionSync({
    selectedChallengeId: options.selectedChallengeId,
    syncSelectedChallengeById,
  })

  const isCaptain = computed(() => {
    const currentUserId = toValue(options.currentUserId)
    return Boolean(team.value && currentUserId && team.value.captain_user_id === currentUserId)
  })

  function resetPageState() {
    const next = createEmptyState()
    contest.value = next.contest
    team.value = next.team
    challenges.value = next.challenges
    announcements.value = next.announcements
    announcementsError.value = next.announcementsError
    selectedChallenge.value = next.selectedChallenge
    flagInput.value = next.flagInput
    submitResult.value = next.submitResult
    showCreateTeam.value = next.showCreateTeam
    showJoinTeam.value = next.showJoinTeam
    teamName.value = next.teamName
    teamIdInput.value = next.teamIdInput
  }

  async function refreshTeam() {
    if (!contest.value) {
      return
    }
    team.value = await getMyTeam(contest.value.id)
  }

  const {
    showCreateTeam,
    showJoinTeam,
    teamName,
    teamIdInput,
    creatingTeam,
    joiningTeam,
    openCreateTeam,
    closeCreateTeam,
    createTeamAction,
    openJoinTeam,
    closeJoinTeam,
    joinTeamAction,
    kickMember,
  } = useContestTeamActions({
    contest,
    team,
    refreshTeam,
  })
  const { loadPage, refreshAnnouncements } = useContestDetailDataLoader({
    contestId: options.contestId,
    contest,
    team,
    challenges,
    announcements,
    announcementsError,
    loading,
    resetPageState,
    startCountdown,
    stopCountdown,
    syncSelectedChallengeFromQuery,
    clearSubmissionState,
    onLoadFailed: () => {
      toast.error('加载竞赛详情失败，请稍后刷新重试')
    },
  })

  watch(
    () => toValue(options.contestId),
    () => {
      void loadPage()
    },
    { immediate: true }
  )

  watch(
    () => toValue(options.selectedChallengeId),
    () => {
      syncSelectedChallengeFromQuery()
    }
  )

  onUnmounted(() => {
    stopCountdown()
  })

  return {
    contest,
    team,
    challenges,
    announcements,
    announcementsError,
    loading,
    countdown,
    selectedChallenge,
    flagInput,
    submitting,
    submitResult,
    showCreateTeam,
    showJoinTeam,
    teamName,
    teamIdInput,
    creatingTeam,
    joiningTeam,
    isCaptain,
    selectChallenge,
    submitFlagAction,
    openCreateTeam,
    closeCreateTeam,
    createTeamAction,
    openJoinTeam,
    closeJoinTeam,
    joinTeamAction,
    kickMember,
    refreshAnnouncements,
  }
}
