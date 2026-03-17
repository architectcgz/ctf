import { computed, onUnmounted, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

import {
  createTeam,
  getAnnouncements,
  getContestChallenges,
  getContestDetail,
  getMyTeam,
  joinTeam,
  kickTeamMember,
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

interface UseContestDetailPageOptions {
  contestId: MaybeRefOrGetter<string>
  currentUserId: MaybeRefOrGetter<string | undefined>
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
  const showCreateTeam = ref(false)
  const showJoinTeam = ref(false)
  const teamName = ref('')
  const teamIdInput = ref('')
  const creatingTeam = ref(false)
  const joiningTeam = ref(false)

  let countdownTimer: number | null = null
  let requestToken = 0

  const isCaptain = computed(() => {
    const currentUserId = toValue(options.currentUserId)
    return Boolean(team.value && currentUserId && team.value.captain_user_id === currentUserId)
  })

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
      selectedChallenge.value = null
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
  }

  async function submitFlagAction() {
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
      submitResult.value = result

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

  function openCreateTeam() {
    showCreateTeam.value = true
  }

  function closeCreateTeam() {
    showCreateTeam.value = false
    teamName.value = ''
  }

  async function createTeamAction() {
    const name = teamName.value.trim()
    if (!name) {
      toast.warning('请输入队伍名称')
      return
    }
    if (name.length < 2 || name.length > 50) {
      toast.warning('队伍名称长度应在 2-50 字符之间')
      return
    }
    if (!contest.value || creatingTeam.value) {
      return
    }

    creatingTeam.value = true
    try {
      await createTeam(contest.value.id, { name })
      await refreshTeam()
      closeCreateTeam()
      toast.success('创建队伍成功')
    } catch (error) {
      console.error(error)
      toast.error(error instanceof Error ? error.message : '创建队伍失败')
    } finally {
      creatingTeam.value = false
    }
  }

  function openJoinTeam() {
    showJoinTeam.value = true
  }

  function closeJoinTeam() {
    showJoinTeam.value = false
    teamIdInput.value = ''
  }

  async function joinTeamAction() {
    const teamId = teamIdInput.value.trim()
    if (!teamId) {
      toast.warning('请输入队伍 ID')
      return
    }
    if (!contest.value || joiningTeam.value) {
      return
    }

    joiningTeam.value = true
    try {
      await joinTeam(contest.value.id, teamId)
      await refreshTeam()
      closeJoinTeam()
      toast.success('加入队伍成功')
    } catch (error) {
      console.error(error)
      toast.error(error instanceof Error ? error.message : '加入队伍失败')
    } finally {
      joiningTeam.value = false
    }
  }

  async function kickMember(userId: string) {
    if (!contest.value || !team.value || !window.confirm('确定踢出该成员？')) {
      return
    }

    try {
      await kickTeamMember(contest.value.id, team.value.id, userId)
      await refreshTeam()
      toast.success('已踢出成员')
    } catch (error) {
      console.error(error)
      toast.error(error instanceof Error ? error.message : '踢出成员失败')
    }
  }

  watch(
    () => toValue(options.contestId),
    () => {
      void loadPage()
    },
    { immediate: true }
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
  }
}
