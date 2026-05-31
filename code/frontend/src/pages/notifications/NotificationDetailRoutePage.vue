<template>
  <NotificationDetailWorkspace
    :loading="loading"
    :load-failed="loadFailed"
    :probe-message="probeMessage"
    :notification="notification"
    :has-related-link="hasRelatedLink"
    :notifications-route="notificationsRoute"
    :related-route="relatedRoute"
    :related-external-href="relatedExternalHref"
    :handle-id-probe="handleIdProbe"
  />
</template>

<script setup lang="ts">
import { toRef } from 'vue'

import { useNotificationDetailPage } from '@/features/notifications'
import { NotificationDetailWorkspace } from '@/widgets/notification-detail-workspace'

const props = defineProps<{
  id: string
}>()

const {
  loading,
  loadFailed,
  probeMessage,
  notification,
  hasRelatedLink,
  notificationsRoute,
  relatedRoute,
  relatedExternalHref,
  handleIdProbe,
} = useNotificationDetailPage(toRef(props, 'id'))
</script>
