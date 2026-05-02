import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getNotifications, markAsRead } from '@/api/notification'
import { useProbeEasterEggs } from '@/composables/useProbeEasterEggs'
import { useToast } from '@/composables/useToast'
import { useNotificationStore } from '@/stores/notification'

export type NotificationAccent = 'primary' | 'success' | 'warning' | 'violet'

export const accentColorMap: Record<NotificationAccent, string> = {
  warning: 'var(--color-warning)',
  success: 'var(--color-success)',
  violet: 'var(--color-cat-reverse)',
  primary: 'var(--color-primary)',
}

export function useNotificationDetailPage() {
  const route = useRoute()
  const router = useRouter()
  const toast = useToast()
  const notificationStore = useNotificationStore()
  const { track } = useProbeEasterEggs()

  const loading = ref(false)
  const loadFailed = ref(false)
  const isMarkingRead = ref(false)
  const probeMessage = ref('')
  let probeMessageTimer: number | null = null

  const notificationId = computed(() => String(route.params.id ?? ''))
  const notification = computed(
    () => notificationStore.notifications.find((item) => item.id === notificationId.value) ?? null
  )
  const hasRelatedLink = computed(() => Boolean(notification.value?.link))

  function notificationAccent(type: string): NotificationAccent {
    if (type === 'contest') return 'warning'
    if (type === 'challenge') return 'success'
    if (type === 'team') return 'violet'
    return 'primary'
  }

  function notificationTypeLabel(type: string): string {
    if (type === 'contest') return '竞赛'
    if (type === 'challenge') return '训练'
    if (type === 'team') return '团队'
    return '系统'
  }

  async function ensureNotificationLoaded(id: string) {
    if (notification.value || !id) {
      return
    }

    loading.value = true
    loadFailed.value = false

    try {
      const data = await getNotifications({ page: 1, page_size: 20 })
      notificationStore.setNotifications(data.list)
    } catch {
      loadFailed.value = true
    } finally {
      loading.value = false
    }
  }

  async function syncReadState(id: string) {
    if (!notification.value?.unread || isMarkingRead.value) {
      return
    }

    isMarkingRead.value = true

    try {
      await markAsRead(id)
      notificationStore.markAsRead(id)
    } catch {
      toast.error('标记已读失败')
    } finally {
      isMarkingRead.value = false
    }
  }

  function goBackToNotifications() {
    void router.push('/notifications')
  }

  function openRelatedLink() {
    const link = notification.value?.link
    if (!link) {
      return
    }

    if (/^https?:\/\//.test(link)) {
      window.open(link, '_blank', 'noopener,noreferrer')
      return
    }

    void router.push(link)
  }

  function showProbeMessage(message: string) {
    probeMessage.value = message
    if (probeMessageTimer) {
      window.clearTimeout(probeMessageTimer)
    }
    probeMessageTimer = window.setTimeout(() => {
      probeMessage.value = ''
      probeMessageTimer = null
    }, 2200)
  }

  function handleIdProbe() {
    const result = track('notification-id', 4)
    if (!result.unlocked) {
      return
    }
    showProbeMessage('值守备注：有人开始认真看编号了。')
  }

  watch(
    notificationId,
    async (id) => {
      if (!id) {
        return
      }

      await ensureNotificationLoaded(id)
      await syncReadState(id)
    },
    { immediate: true }
  )

  onBeforeUnmount(() => {
    if (probeMessageTimer) {
      window.clearTimeout(probeMessageTimer)
    }
  })

  return {
    loading,
    loadFailed,
    probeMessage,
    notification,
    hasRelatedLink,
    notificationAccent,
    notificationTypeLabel,
    goBackToNotifications,
    openRelatedLink,
    handleIdProbe,
  }
}
