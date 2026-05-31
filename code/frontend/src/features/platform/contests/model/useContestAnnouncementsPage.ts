import { computed, type ComputedRef, type Ref } from 'vue'

import { useContestAnnouncementManagement } from '@/features/contest-announcements'
import { buildContestEditRoute } from './contestManageRoutes'
import { useContestAnnouncementsData } from './useContestAnnouncementsData'

export function useContestAnnouncementsPage(contestId: Ref<string> | ComputedRef<string>) {
  let management!: ReturnType<typeof useContestAnnouncementManagement>
  const contestData = useContestAnnouncementsData(contestId, () => management.loadAnnouncements())
  const contest = contestData.contest
  management = useContestAnnouncementManagement(computed(() => contest.value))

  const loading = contestData.loading
  const loadError = contestData.loadError

  function formatTime(value: string): string {
    return new Date(value).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const loadPage = contestData.loadPage

  async function handleSubmit(): Promise<void> {
    await management.publishAnnouncement()
  }

  async function handleDelete(announcementId: string): Promise<void> {
    await management.deleteAnnouncement(announcementId)
  }

  return {
    contest,
    loading,
    loadError,
    backToStudioRoute: computed(() => buildContestEditRoute(contestId.value)),
    management,
    formatTime,
    loadPage,
    handleSubmit,
    handleDelete,
  }
}
