import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import type { NotificationItem } from '@/api/contracts'

export interface StoredNotificationItem extends Omit<NotificationItem, 'id'> {
  id: string
}

function normalizeNotification(
  item: NotificationItem | StoredNotificationItem
): StoredNotificationItem {
  return {
    ...item,
    id: String(item.id),
  }
}

function sortByCreatedAtDesc(items: StoredNotificationItem[]): StoredNotificationItem[] {
  return [...items].sort(
    (left, right) => new Date(right.created_at).getTime() - new Date(left.created_at).getTime()
  )
}

export const useNotificationStore = defineStore('notification', () => {
  const notifications = ref<StoredNotificationItem[]>([])

  const unreadCount = computed(() => notifications.value.filter((item) => item.unread).length)

  function setNotifications(items: Array<NotificationItem | StoredNotificationItem>): void {
    const deduped = new Map<string, StoredNotificationItem>()
    items.forEach((item) => {
      const normalized = normalizeNotification(item)
      deduped.set(normalized.id, normalized)
    })
    notifications.value = sortByCreatedAtDesc([...deduped.values()])
  }

  function upsertNotification(item: NotificationItem | StoredNotificationItem): void {
    const normalized = normalizeNotification(item)
    const remaining = notifications.value.filter((current) => current.id !== normalized.id)
    notifications.value = sortByCreatedAtDesc([normalized, ...remaining]).slice(0, 20)
  }

  function markAsRead(id: string): void {
    notifications.value = notifications.value.map((item) =>
      item.id === id ? { ...item, unread: false } : item
    )
  }

  function markAllRead(): void {
    notifications.value = notifications.value.map((item) => ({ ...item, unread: false }))
  }

  return {
    notifications,
    unreadCount,
    setNotifications,
    upsertNotification,
    markAsRead,
    markAllRead,
  }
})
