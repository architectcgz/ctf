import { computed, onBeforeUnmount, ref, watch, type ComputedRef, type Ref } from 'vue'

import { getNotifications, markAsRead } from '@/api/notification'
import type { AppRouteTarget } from '@/shared/lib/navigation/routeTarget'
import { useProbeEasterEggs } from '@/shared/model/common/useProbeEasterEggs'
import { useToast } from '@/shared/model/common/useToast'
import { useNotificationStore } from '@/stores/notification'

const notificationsRoute: AppRouteTarget = { name: 'Notifications' }

function isExternalLink(link: string): boolean {
  return /^https?:\/\//.test(link)
}

export function useNotificationDetailPage(notificationId: Ref<string> | ComputedRef<string>) {
  const toast = useToast()
  const notificationStore = useNotificationStore()
  const { track } = useProbeEasterEggs()

  const loading = ref(false)
  const loadFailed = ref(false)
  const isMarkingRead = ref(false)
  const probeMessage = ref('')
  let probeMessageTimer: number | null = null

  const notification = computed(
    () => notificationStore.notifications.find((item) => item.id === notificationId.value) ?? null
  )
  const relatedLink = computed(() => notification.value?.link?.trim() || '')
  const relatedRoute = computed<AppRouteTarget | null>(() => {
    if (!relatedLink.value || isExternalLink(relatedLink.value)) {
      return null
    }
    return relatedLink.value
  })
  const relatedExternalHref = computed<string | null>(() => {
    if (!relatedLink.value || !isExternalLink(relatedLink.value)) {
      return null
    }
    return relatedLink.value
  })
  const hasRelatedLink = computed(() => Boolean(relatedRoute.value || relatedExternalHref.value))

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
    notificationsRoute,
    relatedRoute,
    relatedExternalHref,
    handleIdProbe,
  }
}
