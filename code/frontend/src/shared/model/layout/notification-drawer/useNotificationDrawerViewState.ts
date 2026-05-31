import { computed, onBeforeUnmount, ref, watch, type Ref } from 'vue'

import type { NotificationFilter, NotificationFilterOption } from './types'

interface NotificationDrawerListItem {
  unread: boolean
}

interface UseNotificationDrawerViewStateOptions<Item extends NotificationDrawerListItem> {
  open: Ref<boolean>
  close: () => void
  items: Ref<Item[]>
  unreadCount: Ref<number>
}

const filterOptions: NotificationFilterOption[] = [
  { value: 'all', label: '全部' },
  { value: 'unread', label: '未读' },
  { value: 'read', label: '已读' },
]

export function useNotificationDrawerViewState<Item extends NotificationDrawerListItem>({
  open,
  close,
  items,
  unreadCount,
}: UseNotificationDrawerViewStateOptions<Item>) {
  const activeFilter = ref<NotificationFilter>('all')
  const hasUnread = computed(() => unreadCount.value > 0)
  const unreadBadgeLabel = computed(() =>
    unreadCount.value > 99 ? '99+' : String(unreadCount.value)
  )
  const drawerSummary = computed(() => {
    if (unreadCount.value > 0) {
      return '条未读通知待处理'
    }
    if (items.value.length > 0) {
      return '全部通知已读'
    }
    return '当前没有新通知'
  })

  const filteredItems = computed(() => {
    if (activeFilter.value === 'unread') {
      return items.value.filter((item) => item.unread)
    }
    if (activeFilter.value === 'read') {
      return items.value.filter((item) => !item.unread)
    }
    return items.value
  })

  const emptyState = computed(() => {
    if (activeFilter.value === 'unread') {
      return {
        title: '暂无未读通知',
        copy: '新的系统、竞赛与训练动态会优先显示在这里。',
      }
    }
    if (activeFilter.value === 'read') {
      return {
        title: '暂无已读通知',
        copy: '已处理的通知会保留在这里，方便回看。',
      }
    }
    return {
      title: '暂无新通知',
      copy: '新的系统、竞赛与训练动态会显示在这里。',
    }
  })

  function handleWindowKeydown(event: KeyboardEvent): void {
    if (event.key !== 'Escape') return
    close()
  }

  watch(
    open,
    (isOpen, _wasOpen, onCleanup) => {
      if (typeof window === 'undefined') return

      if (!isOpen) {
        document.body.style.overflow = ''
        return
      }

      document.body.style.overflow = 'hidden'
      window.addEventListener('keydown', handleWindowKeydown)
      onCleanup(() => {
        window.removeEventListener('keydown', handleWindowKeydown)
      })
    },
    { immediate: true }
  )

  onBeforeUnmount(() => {
    if (typeof window === 'undefined') return
    window.removeEventListener('keydown', handleWindowKeydown)
    document.body.style.overflow = ''
  })

  return {
    activeFilter,
    filterOptions,
    filteredItems,
    emptyState,
    hasUnread,
    unreadBadgeLabel,
    drawerSummary,
  }
}
