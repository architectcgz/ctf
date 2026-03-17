import type { Component } from 'vue'
import { Bell, Flag, GraduationCap, Info, Trophy, X } from 'lucide-vue-next'
import { computed, onBeforeUnmount, ref, useTemplateRef, watch } from 'vue'
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

const typeMap: Record<string, NotificationTypeMeta> = {
  system: {
    icon: Info,
    label: '系统',
    accentColor: 'var(--color-primary)',
    iconWrapStyle: {
      backgroundColor: 'var(--color-primary-soft)',
      borderColor: 'color-mix(in srgb, var(--color-primary) 28%, transparent)',
    },
    badgeStyle: {
      color: 'var(--color-primary)',
      borderColor: 'color-mix(in srgb, var(--color-primary) 22%, transparent)',
      backgroundColor: 'color-mix(in srgb, var(--color-primary) 10%, transparent)',
    },
  },
  contest: {
    icon: Trophy,
    label: '竞赛',
    accentColor: 'var(--color-warning)',
    iconWrapStyle: {
      backgroundColor: 'rgba(210, 153, 34, 0.12)',
      borderColor: 'rgba(210, 153, 34, 0.26)',
    },
    badgeStyle: {
      color: 'var(--color-warning)',
      borderColor: 'rgba(210, 153, 34, 0.22)',
      backgroundColor: 'rgba(210, 153, 34, 0.1)',
    },
  },
  challenge: {
    icon: Flag,
    label: '训练',
    accentColor: 'var(--color-success)',
    iconWrapStyle: {
      backgroundColor: 'rgba(63, 185, 80, 0.12)',
      borderColor: 'rgba(63, 185, 80, 0.26)',
    },
    badgeStyle: {
      color: 'var(--color-success)',
      borderColor: 'rgba(63, 185, 80, 0.22)',
      backgroundColor: 'rgba(63, 185, 80, 0.1)',
    },
  },
  team: {
    icon: GraduationCap,
    label: '团队',
    accentColor: '#8b5cf6',
    iconWrapStyle: {
      backgroundColor: 'rgba(139, 92, 246, 0.12)',
      borderColor: 'rgba(139, 92, 246, 0.26)',
    },
    badgeStyle: {
      color: '#a78bfa',
      borderColor: 'rgba(139, 92, 246, 0.22)',
      backgroundColor: 'rgba(139, 92, 246, 0.1)',
    },
  },
}

const fallbackTypeMeta: NotificationTypeMeta = typeMap.system

export function useNotificationDropdown(realtimeStatus: () => WebSocketStatus) {
  const router = useRouter()
  const store = useNotificationStore()
  const toast = useToast()
  const open = ref(false)
  const trigger = useTemplateRef<HTMLButtonElement>('trigger')
  const panel = useTemplateRef<HTMLDivElement>('panel')
  const panelStyle = ref<Record<string, string>>({})
  const repositionPanel = () => updatePanelPosition()

  const unreadCount = computed(() => store.unreadCount)
  const items = computed(() => store.notifications)
  const previewItems = computed(() => items.value.slice(0, 6))
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

  function typeMeta(type: string): NotificationTypeMeta {
    return typeMap[type] || fallbackTypeMeta
  }

  function notificationCardStyle(unread: boolean): Record<string, string> {
    if (unread) {
      return {
        borderColor: 'color-mix(in srgb, var(--color-primary) 22%, var(--color-border-default))',
        background:
          'linear-gradient(180deg, color-mix(in srgb, var(--color-primary) 10%, transparent), rgba(15, 23, 42, 0.38))',
        boxShadow: '0 18px 36px var(--color-shadow-soft)',
      }
    }

    return {
      borderColor: 'var(--color-border-default)',
      backgroundColor: 'color-mix(in srgb, var(--color-bg-base) 62%, transparent)',
    }
  }

  function updatePanelPosition() {
    if (!trigger.value) {
      return
    }

    const rect = trigger.value.getBoundingClientRect()
    const viewportPadding = 12
    const panelWidth = Math.min(420, window.innerWidth - viewportPadding * 2)
    const left = Math.min(
      Math.max(viewportPadding, rect.right - panelWidth),
      window.innerWidth - panelWidth - viewportPadding
    )
    const top = rect.bottom + 10

    panelStyle.value = {
      top: `${top}px`,
      left: `${left}px`,
      width: `${panelWidth}px`,
    }
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

  async function markAsRead(id: string) {
    const target = store.notifications.find((item) => item.id === id)
    if (!target?.unread) {
      return
    }

    try {
      await markAsReadApi(id)
    } catch {
      toast.error('标记已读失败')
      return
    }

    store.markAsRead(id)
  }

  async function markAllRead() {
    const unreadItems = store.notifications.filter((item) => item.unread)
    if (unreadItems.length === 0) {
      return
    }

    const results = await Promise.allSettled(unreadItems.map((item) => markAsReadApi(item.id)))
    const failedCount = results.filter((result) => result.status === 'rejected').length
    if (failedCount > 0) {
      toast.warning(`部分通知标记失败（${failedCount} 条）`)
    }

    store.markAllRead()
  }

  watch(open, (isOpen) => {
    if (!isOpen) {
      return
    }

    updatePanelPosition()
    window.addEventListener('resize', repositionPanel)
    window.addEventListener('scroll', repositionPanel, true)

    const cleanup = () => {
      window.removeEventListener('resize', repositionPanel)
      window.removeEventListener('scroll', repositionPanel, true)
    }

    const stop = watch(open, (next) => {
      if (!next) {
        cleanup()
        stop()
      }
    })
  })

  onBeforeUnmount(() => {
    window.removeEventListener('resize', repositionPanel)
    window.removeEventListener('scroll', repositionPanel, true)
  })

  return {
    Bell,
    X,
    open,
    trigger,
    panel,
    panelStyle,
    unreadCount,
    items,
    previewItems,
    statusMeta,
    statusPillStyle,
    typeMeta,
    notificationCardStyle,
    close,
    toggleOpen,
    goToNotifications,
    markAsRead,
    markAllRead,
  }
}
