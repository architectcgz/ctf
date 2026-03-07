import type { NotificationItem } from '@/api/contracts'
import { getNotifications } from '@/api/notification'
import { useToast } from '@/composables/useToast'
import { useWebSocket } from '@/composables/useWebSocket'
import { useNotificationStore } from '@/stores/notification'

interface NotificationPayload {
  id?: string | number
  title?: string
  type?: NotificationItem['type']
  content?: string
  unread?: boolean
  created_at?: string
}

function toNotificationItem(payload: unknown): NotificationItem | null {
  if (!payload || typeof payload !== 'object') return null

  const candidate = payload as NotificationPayload
  if (candidate.id == null || !candidate.title || !candidate.type || !candidate.created_at) {
    return null
  }

  return {
    id: String(candidate.id),
    type: candidate.type,
    title: candidate.title,
    content: candidate.content,
    unread: candidate.unread ?? true,
    created_at: candidate.created_at,
  }
}

export function useNotificationRealtime() {
  const store = useNotificationStore()
  const toast = useToast()

  const { status, connect, disconnect } = useWebSocket('notifications', {
    'notification.created': (payload: unknown) => {
      const notification = toNotificationItem(payload)
      if (!notification) return
      store.upsertNotification(notification)
    },
    'notification.read': (payload: unknown) => {
      const notification = toNotificationItem(payload)
      if (!notification) return
      store.markAsRead(String(notification.id))
    },
  })

  async function syncInitialNotifications(): Promise<void> {
    const data = await getNotifications({ page: 1, page_size: 20 })
    store.setNotifications(data.list)
  }

  async function start(): Promise<void> {
    try {
      await syncInitialNotifications()
    } catch (error) {
      toast.error('初始化通知失败')
    }

    try {
      await connect()
    } catch (error) {
      toast.warning('实时通知连接失败，已切换为手动刷新')
    }
  }

  return {
    status,
    start,
    stop: disconnect,
  }
}
