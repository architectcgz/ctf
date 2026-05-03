import { toValue, type MaybeRefOrGetter, type Ref } from 'vue'

import {
  getAnnouncements,
  getContestChallenges,
  getContestDetail,
  getMyTeam,
} from '@/api/contest'
import type {
  ContestAnnouncement,
  ContestChallengeItem,
  ContestDetailData,
  TeamData,
} from '@/api/contracts'

interface UseContestDetailDataLoaderOptions {
  contestId: MaybeRefOrGetter<string>
  contest: Ref<ContestDetailData | null>
  team: Ref<TeamData | null>
  challenges: Ref<ContestChallengeItem[]>
  announcements: Ref<ContestAnnouncement[]>
  announcementsError: Ref<string>
  loading: Ref<boolean>
  resetPageState: () => void
  startCountdown: () => void
  stopCountdown: () => void
  syncSelectedChallengeFromQuery: () => void
  clearSubmissionState: () => void
  onLoadFailed: () => void
}

export function useContestDetailDataLoader({
  contestId,
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
  onLoadFailed,
}: UseContestDetailDataLoaderOptions) {
  let requestToken = 0

  async function refreshAnnouncements() {
    if (!contest.value) {
      return
    }

    try {
      announcements.value = await getAnnouncements(contest.value.id)
      announcementsError.value = ''
    } catch {
      announcementsError.value = '公告加载失败，请稍后刷新重试'
    }
  }

  async function loadPage() {
    const nextContestId = toValue(contestId)
    if (!nextContestId) {
      resetPageState()
      stopCountdown()
      loading.value = false
      return
    }

    const currentToken = ++requestToken
    loading.value = true

    try {
      const contestData = await getContestDetail(nextContestId)
      const [teamData, challengesData, announcementsData] = await Promise.all([
        getMyTeam(nextContestId).catch(() => null),
        getContestChallenges(nextContestId).catch(() => []),
        getAnnouncements(nextContestId).catch(() => null),
      ])

      if (currentToken !== requestToken) {
        return
      }

      contest.value = contestData
      team.value = teamData
      challenges.value = challengesData
      syncSelectedChallengeFromQuery()
      clearSubmissionState()

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
      onLoadFailed()
    } finally {
      if (currentToken === requestToken) {
        loading.value = false
      }
    }
  }

  return {
    loadPage,
    refreshAnnouncements,
  }
}
