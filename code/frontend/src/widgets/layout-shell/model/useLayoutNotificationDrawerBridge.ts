import { useNotificationDrawer } from '@/features/notifications'

export function useLayoutNotificationDrawerBridge(...args: Parameters<typeof useNotificationDrawer>) {
  return useNotificationDrawer(...args)
}
