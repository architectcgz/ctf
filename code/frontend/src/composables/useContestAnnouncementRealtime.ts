import { useWebSocket } from '@/composables/useWebSocket'

export function useContestAnnouncementRealtime(contestId: string, onUpdated: () => void) {
  const { status, connect, disconnect } = useWebSocket(`contests/${contestId}/announcements`, {
    'contest.announcement.created': () => {
      onUpdated()
    },
    'contest.announcement.deleted': () => {
      onUpdated()
    },
  })

  return {
    status,
    start: connect,
    stop: disconnect,
  }
}
