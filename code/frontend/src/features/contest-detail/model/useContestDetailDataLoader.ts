import { toValue, type MaybeRefOrGetter, type Ref } from 'vue'

import {
  getAnnouncementSync,
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
import {
  applyContestAnnouncementSyncEvents,
  nextContestAnnouncementSyncCursor,
} from '@/entities/contest-announcement'

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
  let announcementSyncCursor = ''

  async function refreshAnnouncements() {
    if (!contest.value) {
      return
    }

    try {
      const sync = await getAnnouncementSync(contest.value.id)
      announcements.value = await getAnnouncements(contest.value.id)
      announcementSyncCursor = nextContestAnnouncementSyncCursor(sync)
      announcementsError.value = ''
    } catch {
      announcementsError.value = '公告加载失败，请稍后刷新重试'
      announcementSyncCursor = ''
    }
  }

  async function syncAnnouncementsIncrementally() {
    if (!contest.value) {
      announcementSyncCursor = ''
      return
    }

    try {
      if (!announcementSyncCursor) {
        if (announcements.value.length > 0) {
          await refreshAnnouncements()
          return
        }
        const sync = await getAnnouncementSync(contest.value.id)
        announcementSyncCursor = nextContestAnnouncementSyncCursor(sync)
        return
      }

      let afterId = announcementSyncCursor
      while (afterId) {
        const sync = await getAnnouncementSync(contest.value.id, afterId)
        announcements.value = applyContestAnnouncementSyncEvents(announcements.value, sync.events)
        announcementSyncCursor = nextContestAnnouncementSyncCursor(sync)
        if (!sync.has_more) {
          break
        }
        afterId = announcementSyncCursor
      }
      announcementsError.value = ''
    } catch {
      announcementsError.value = '公告同步失败，请稍后刷新重试'
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
      const teamDataPromise = getMyTeam(nextContestId).catch(() => null)
      const challengesDataPromise = getContestChallenges(nextContestId).catch(() => [])
      let announcementsData: ContestAnnouncement[] | null = null
      let announcementSyncAnchor: Awaited<ReturnType<typeof getAnnouncementSync>> | null = null
      try {
        announcementSyncAnchor = await getAnnouncementSync(nextContestId)
        announcementsData = await getAnnouncements(nextContestId)
      } catch {
        announcementsData = null
        announcementSyncAnchor = null
      }
      const [teamData, challengesData] = await Promise.all([teamDataPromise, challengesDataPromise])

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
        if (announcementSyncAnchor) {
          announcementSyncCursor = nextContestAnnouncementSyncCursor(announcementSyncAnchor)
          announcementsError.value = ''
        } else {
          announcementSyncCursor = ''
          announcementsError.value = '公告加载失败，请稍后刷新重试'
        }
      } else {
        announcements.value = []
        announcementsError.value = '公告加载失败，请稍后刷新重试'
        announcementSyncCursor = ''
      }

      startCountdown()
    } catch {
      if (currentToken !== requestToken) {
        return
      }

      resetPageState()
      stopCountdown()
      announcementSyncCursor = ''
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
    syncAnnouncementsIncrementally,
  }
}
