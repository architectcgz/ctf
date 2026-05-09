import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import type { NotificationItem, NotificationType } from '@/api/contracts'
import { getNotifications, markAsRead } from '@/api/notification'
import { usePagination } from '@/composables/usePagination'
import { useProbeEasterEggs } from '@/composables/useProbeEasterEggs'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notification'

export function useNotificationListPage() {
  const toast = useToast()
  const authStore = useAuthStore()
  const notificationStore = useNotificationStore()
  const router = useRouter()
  const { track } = useProbeEasterEggs()
  const publishDrawerOpen = ref(false)
  const probeMessage = ref('')
  const selectedCategory = ref<NotificationType | 'all'>('all')
  let probeMessageTimer: number | null = null

  async function fetchNotifications(params: { page: number; page_size: number }) {
    const data = await getNotifications({
      ...params,
      ...(selectedCategory.value === 'all' ? {} : { type: selectedCategory.value }),
    })
    if (params.page === 1) {
      notificationStore.setNotifications(data.list)
    }
    return data
  }

  const { list, total, page, pageSize, loading, error, changePage, refresh } =
    usePagination<NotificationItem>(fetchNotifications)
  const unreadOnPage = computed(() => list.value.filter((item) => item.unread).length)
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
  const hasLoadError = computed(() => Boolean(error.value) && list.value.length === 0)
  const loadErrorMessage = computed(() => {
    if (error.value instanceof Error && error.value.message.trim().length > 0) {
      return error.value.message
    }
    return '通知加载失败，请稍后重试。'
  })

  function typeLabel(type: string): string {
    if (type === 'contest') return '竞赛'
    if (type === 'challenge') return '训练'
    if (type === 'team') return '团队'
    return '系统'
  }

  const categoryOptions: Array<{ key: NotificationType | 'all'; label: string }> = [
    { key: 'all', label: '全部' },
    { key: 'system', label: '系统' },
    { key: 'contest', label: '竞赛' },
    { key: 'challenge', label: '训练' },
    { key: 'team', label: '团队' },
  ]

  const selectedCategoryLabel = computed(
    () =>
      categoryOptions.find((option) => option.key === selectedCategory.value)?.label ??
      categoryOptions[0].label
  )

  async function selectCategory(next: NotificationType | 'all'): Promise<void> {
    if (selectedCategory.value === next) return
    selectedCategory.value = next
    await changePage(1)
  }

  function openNotificationDetail(item: NotificationItem): void {
    void router.push(`/notifications/${encodeURIComponent(String(item.id))}`)
  }

  function showProbeMessage(message: string) {
    probeMessage.value = message
    if (probeMessageTimer) {
      window.clearTimeout(probeMessageTimer)
    }
    probeMessageTimer = window.setTimeout(() => {
      probeMessage.value = ''
      probeMessageTimer = null
    }, 3000)
  }

  async function markCurrentPageRead(): Promise<void> {
    const unreadItems = list.value.filter((item) => item.unread)
    if (unreadItems.length === 0) return

    const results = await Promise.allSettled(unreadItems.map((item) => markAsRead(String(item.id))))
    const failedCount = results.filter((result) => result.status === 'rejected').length
    unreadItems.forEach((item, index) => {
      if (results[index]?.status === 'fulfilled') {
        const target = list.value.find((entry) => String(entry.id) === String(item.id))
        if (target) {
          target.unread = false
        }
        notificationStore.markAsRead(String(item.id))
      }
    })

    if (failedCount > 0) {
      toast.warning(`部分通知标记失败（${failedCount} 条）`)
    }
  }

  const headStats = computed(() => [
    { key: 'total', label: '消息数', value: total.value },
    { key: 'unread', label: '未读数', value: unreadOnPage.value },
  ])

  const canPublishNotification = computed(() => authStore.isAdmin)

  function openPublishDrawer(): void {
    publishDrawerOpen.value = true
  }

  function closePublishDrawer(): void {
    publishDrawerOpen.value = false
  }

  async function handlePublishSuccess(): Promise<void> {
    closePublishDrawer()
    await refresh()
  }

  async function handleRefresh() {
    const result = track('notification-refresh', 3)
    if (result.unlocked) {
      showProbeMessage('新消息不会因为执念刷新得更快。')
    }
    await refresh()
  }

  onMounted(() => {
    void refresh()
  })

  onBeforeUnmount(() => {
    if (probeMessageTimer) {
      window.clearTimeout(probeMessageTimer)
    }
  })

  return {
    publishDrawerOpen,
    probeMessage,
    list,
    total,
    page,
    pageSize,
    loading,
    changePage,
    refresh,
    totalPages,
    hasLoadError,
    loadErrorMessage,
    headStats,
    categoryOptions,
    selectedCategory,
    selectedCategoryLabel,
    canPublishNotification,
    typeLabel,
    selectCategory,
    openNotificationDetail,
    markCurrentPageRead,
    openPublishDrawer,
    closePublishDrawer,
    handlePublishSuccess,
    handleRefresh,
  }
}
