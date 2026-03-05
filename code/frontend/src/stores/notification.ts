import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export interface NotificationItem {
  id: string
  type: string
  title: string
  time?: string
  unread: boolean
}

export const useNotificationStore = defineStore('notification', () => {
  const notifications = ref<NotificationItem[]>([])

  const unreadCount = computed(() => notifications.value.filter((n) => n.unread).length)

  function setNotifications(items: NotificationItem[]): void {
    notifications.value = items
  }

  function addNotification(item: NotificationItem): void {
    notifications.value = [item, ...notifications.value].slice(0, 20)
  }

  function markAsRead(id: string): void {
    notifications.value = notifications.value.map((n) => (n.id === id ? { ...n, unread: false } : n))
  }

  function markAllRead(): void {
    notifications.value = notifications.value.map((n) => ({ ...n, unread: false }))
  }

  return { notifications, unreadCount, setNotifications, addNotification, markAsRead, markAllRead }
})
