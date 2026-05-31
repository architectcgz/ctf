import { useRouter } from 'vue-router'

import type { WebSocketStatus } from '@/shared/model/realtime/useWebSocket'
import { useNotificationDrawer } from '@/features/notifications'

export function useLayoutNotificationDrawerBridge(realtimeStatus: () => WebSocketStatus) {
  const router = useRouter()

  return useNotificationDrawer(realtimeStatus, {
    goToNotifications: () => {
      void router.push('/notifications')
    },
    goToNotificationDetail: (id: string) => {
      void router.push(`/notifications/${encodeURIComponent(id)}`)
    },
  })
}
