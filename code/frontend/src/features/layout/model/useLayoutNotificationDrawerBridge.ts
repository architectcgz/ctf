import type { WebSocketStatus } from '@/shared/model/realtime/useWebSocket'
import { useNotificationDrawer } from '@/features/notifications'

/**
 * 桥接 app shell 与 notification feature 的通知抽屉。
 * 通过回调注入路由导航，避免 feature 层直接依赖 vue-router。
 */
export function useLayoutNotificationDrawerBridge(
  realtimeStatus: () => WebSocketStatus,
  goToNotifications: () => void,
  goToNotificationDetail: (id: string) => void
) {
  return useNotificationDrawer(realtimeStatus, {
    goToNotifications,
    goToNotificationDetail,
  })
}
