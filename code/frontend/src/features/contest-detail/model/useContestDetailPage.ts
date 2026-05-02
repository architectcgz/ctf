import { computed, onUnmounted, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

import {
  getAnnouncements,
  getContestChallenges,
  getContestDetail,
  getMyTeam,
  submitContestFlag,
} from '@/api/contest'
import type {
  ContestAnnouncement,
  ContestChallengeItem,
  ContestDetailData,
  SubmitFlagData,
  TeamData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'
import { formatDuration } from '@/utils/format'
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
    countdown: '',
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
  const countdown = ref('')
  const selectedChallenge = ref<ContestChallengeItem | null>(null)
  const flagInput = ref('')
  const submitting = ref(false)
  const submitResult = ref<SubmitFlagData | null>(null)

  let countdownTimer: number | null = null
  let requestToken = 0

  const isCaptain = computed(() => {
    const currentUserId = toValue(options.currentUserId)
    return Boolean(team.value && currentUserId && team.value.captain_user_id === currentUserId)
  })

  function normalizeChallengeId(value: string | null | Array<string | null> | undefined): string {
    if (Array.isArray(value)) {
      return value.find((item): item is string => typeof item === 'string' && item.length > 0) ?? ''
    }
    return typeof value === 'string' ? value : ''
  }

  function requestedChallengeId(): string {
    return normalizeChallengeId(toValue(options.selectedChallengeId))
  }

  function syncSelectedChallengeFromQuery() {
    const challengeId = requestedChallengeId()
    selectedChallenge.value = challengeId
      ? challenges.value.find((challenge) => challenge.id === challengeId) ?? null
      : null
    flagInput.value = ''
    submitResult.value = null
  }

  function stopCountdown() {
    if (countdownTimer) {
      window.clearInterval(countdownTimer)
      countdownTimer = null
    }
  }

  function updateCountdown() {
    if (!contest.value) {
      countdown.value = ''
      stopCountdown()
      return
    }

    const now = Date.now()
    const start = new Date(contest.value.starts_at).getTime()
    const end = new Date(contest.value.ends_at).getTime()

    if (now < start) {
      countdown.value = `距离开始: ${formatDuration(start - now)}`
      return
    }
    if (now < end) {
      countdown.value = `距离结束: ${formatDuration(end - now)}`
      return
    }

    countdown.value = ''
    stopCountdown()
  }

  function startCountdown() {
    stopCountdown()
    if (!contest.value) {
      countdown.value = ''
      return
    }

    updateCountdown()
    if (countdown.value) {
      countdownTimer = window.setInterval(updateCountdown, 1000)
    }
  }

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
    countdown.value = next.countdown
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

  async function refreshAnnouncements() {
    if (!contest.value) {
      return
    }

    try {
      announcements.value = await getAnnouncements(contest.value.id)
      announcementsError.value = ''
    } catch (error) {
      announcementsError.value = '公告加载失败，请稍后刷新重试'
    }
  }

  async function loadPage() {
    const contestId = toValue(options.contestId)
    if (!contestId) {
      resetPageState()
      stopCountdown()
      loading.value = false
      return
    }

    const currentToken = ++requestToken
    loading.value = true

    try {
      const contestData = await getContestDetail(contestId)
      const [teamData, challengesData, announcementsData] = await Promise.all([
        getMyTeam(contestId).catch(() => null),
        getContestChallenges(contestId).catch(() => []),
        getAnnouncements(contestId).catch(() => null),
      ])

      if (currentToken !== requestToken) {
        return
      }

      contest.value = contestData
      team.value = teamData
      challenges.value = challengesData
      syncSelectedChallengeFromQuery()
      flagInput.value = ''
      submitResult.value = null

      if (announcementsData) {
        announcements.value = announcementsData
        announcementsError.value = ''
      } else {
        announcements.value = []
        announcementsError.value = '公告加载失败，请稍后刷新重试'
      }

      startCountdown()
    } catch {
      if (currentToken !== requestToken) {
        return
      }

      resetPageState()
      stopCountdown()
      toast.error('加载竞赛详情失败，请稍后刷新重试')
    } finally {
      if (currentToken === requestToken) {
        loading.value = false
      }
    }
  }

  function selectChallenge(challenge: ContestChallengeItem) {
    selectedChallenge.value = challenge
    flagInput.value = ''
    submitResult.value = null
    options.onSelectedChallengeChange?.(challenge.id)
  }

  function buildContestSubmitMessage(result: SubmitFlagData): string {
    if (result.is_correct) {
      return `正确！+${result.points ?? 0} 分`
    }
    return 'Flag 错误，请重试'
  }

  async function submitFlagAction() {
    if (submitting.value) {
      return
    }

    const flag = flagInput.value.trim()
    if (!flag) {
      toast.warning('请输入 Flag')
      return
    }
    if (flag.length < 5 || flag.length > 200) {
      toast.warning('Flag 长度应在 5-200 字符之间')
      return
    }
    if (!selectedChallenge.value || !contest.value) {
      return
    }

    submitting.value = true
    submitResult.value = null

    try {
      const result = await submitContestFlag(contest.value.id, selectedChallenge.value.id, flag)
      submitResult.value = {
        ...result,
        message: buildContestSubmitMessage(result),
      }

      if (result.is_correct) {
        const solvedChallengeId = selectedChallenge.value.id
        challenges.value = challenges.value.map((challenge) =>
          challenge.id === solvedChallengeId ? { ...challenge, is_solved: true } : challenge
        )
        selectedChallenge.value = {
          ...selectedChallenge.value,
          is_solved: true,
        }
        flagInput.value = ''
      }
    } catch (error) {
      console.error(error)
      toast.error(error instanceof Error ? error.message : '提交失败，请稍后重试')
    } finally {
      submitting.value = false
    }
  }

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
