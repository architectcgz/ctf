import { useNotificationRealtime } from '@/features/notifications'

/**
 * 桥接 app shell 与 notification feature 的实时推送。
 * 后续 app shell 迁出 shared 后可移除该桥接，由页面组装层直接调用。
 */
export function useLayoutNotificationRealtimeBridge() {
  return useNotificationRealtime()
}
