import type { Component, ComponentPublicInstance } from 'vue'
import { Flag, GraduationCap, Info, Trophy } from 'lucide-vue-next'
import { computed, nextTick, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { markAsRead as markAsReadApi } from '@/api/notification'
import { useToast } from '@/composables/useToast'
import type { WebSocketStatus } from '@/composables/useWebSocket'
import { useNotificationStore } from '@/stores/notification'

export interface NotificationTypeMeta {
  icon: Component
  label: string
  accentColor: string
  iconWrapStyle: Record<string, string>
  badgeStyle: Record<string, string>
}

interface StatusMeta {
  label: string
  accentColor: string
}

function createNotificationTypeMeta(
  icon: Component,
  label: string,
  accentColor: string,
  softBackgroundColor = `color-mix(in srgb, ${accentColor} 12%, transparent)`
): NotificationTypeMeta {
  return {
    icon,
    label,
    accentColor,
    iconWrapStyle: {
      backgroundColor: softBackgroundColor,
      borderColor: `color-mix(in srgb, ${accentColor} 26%, transparent)`,
    },
    badgeStyle: {
      color: accentColor,
      borderColor: `color-mix(in srgb, ${accentColor} 22%, transparent)`,
      backgroundColor: `color-mix(in srgb, ${accentColor} 10%, transparent)`,
    },
  }
}

const typeMap: Record<string, NotificationTypeMeta> = {
  system: createNotificationTypeMeta(Info, '系统', 'var(--color-primary)', 'var(--color-primary-soft)'),
  contest: createNotificationTypeMeta(Trophy, '竞赛', 'var(--color-warning)'),
  challenge: createNotificationTypeMeta(Flag, '训练', 'var(--color-success)'),
  team: createNotificationTypeMeta(GraduationCap, '团队', 'var(--color-brand-swatch-blue)'),
}

const fallbackTypeMeta: NotificationTypeMeta = typeMap.system

export function useNotificationDrawer(realtimeStatus: () => WebSocketStatus) {
  const router = useRouter()
  const store = useNotificationStore()
  const toast = useToast()
  const open = ref(false)
  const triggerRef = ref<HTMLElement | null>(null)
  const isMarkingAllRead = ref(false)

  const unreadCount = computed(() => store.unreadCount)
  const items = computed(() => store.notifications)
  const statusMeta = computed<StatusMeta>(() => {
    if (realtimeStatus() === 'open') {
      return { label: '实时在线', accentColor: 'var(--color-success)' }
    }
    if (realtimeStatus() === 'connecting') {
      return { label: '连接中', accentColor: 'var(--color-warning)' }
    }
    if (realtimeStatus() === 'error') {
      return { label: '连接异常', accentColor: 'var(--color-danger)' }
    }
    return { label: '手动刷新', accentColor: 'var(--color-text-muted)' }
  })
  const statusPillStyle = computed<Record<string, string>>(() => ({
    color: statusMeta.value.accentColor,
    borderColor: `color-mix(in srgb, ${statusMeta.value.accentColor} 22%, var(--color-border-default))`,
    backgroundColor: `color-mix(in srgb, ${statusMeta.value.accentColor} 10%, transparent)`,
  }))

  watch(open, async (isOpen, wasOpen) => {
    if (isOpen || !wasOpen) {
      return
    }

    await nextTick()
    triggerRef.value?.focus()
  })

  function setTriggerRef(element: Element | ComponentPublicInstance | null): void {
    if (element instanceof HTMLElement) {
      triggerRef.value = element
      return
    }

    triggerRef.value = null
  }

  function typeMeta(type: string): NotificationTypeMeta {
    return typeMap[type] || fallbackTypeMeta
  }

  function close() {
    open.value = false
  }

  function toggleOpen() {
    open.value = !open.value
  }

  function goToNotifications() {
    close()
    void router.push('/notifications')
  }

  function goToNotificationDetail(id: string) {
    close()
    void router.push(`/notifications/${encodeURIComponent(id)}`)
  }

  async function markAllRead() {
    if (isMarkingAllRead.value) {
      return
    }

    const unreadItems = store.notifications.filter((item) => item.unread)
    if (unreadItems.length === 0) {
      return
    }

    isMarkingAllRead.value = true

    try {
      const results = await Promise.allSettled(unreadItems.map((item) => markAsReadApi(item.id)))
      const failedCount = results.filter((result) => result.status === 'rejected').length
      unreadItems.forEach((item, index) => {
        if (results[index]?.status === 'fulfilled') {
          store.markAsRead(item.id)
        }
      })

      if (failedCount === 0) {
        store.markAllRead()
      }

      if (failedCount > 0) {
        toast.warning(`部分通知标记失败（${failedCount} 条）`)
      }
    } finally {
      isMarkingAllRead.value = false
    }
  }

  return {
    open,
    setTriggerRef,
    unreadCount,
    isMarkingAllRead,
    items,
    statusMeta,
    statusPillStyle,
    typeMeta,
    close,
    toggleOpen,
    goToNotifications,
    goToNotificationDetail,
    markAllRead,
  }
}
