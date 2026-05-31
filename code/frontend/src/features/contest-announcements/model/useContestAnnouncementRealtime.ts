import { useWebSocket } from '@/shared/model/realtime/useWebSocket'

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
