<template>
  <span
    class="notification-read-state-pill"
    :class="[
      variant === 'detail'
        ? 'notification-read-state-pill--detail'
        : 'notification-read-state-pill--directory',
      unread
        ? 'notification-read-state-pill--unread'
        : 'notification-read-state-pill--read',
    ]"
  >
    <CircleCheckBig v-if="showIcon" class="h-3.5 w-3.5" />
    {{ getNotificationReadStateLabel(unread) }}
  </span>
</template>

<style scoped>
.notification-read-state-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border-radius: 999px;
  font-weight: 700;
}

.notification-read-state-pill--directory.notification-read-state-pill--unread {
  color: var(--color-primary);
}

.notification-read-state-pill--directory.notification-read-state-pill--read {
  color: var(--color-text-secondary);
}

.notification-read-state-pill--detail {
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 84%, transparent);
  padding: 0.4rem 0.8rem;
  font-size: var(--font-size-0-74);
}

.notification-read-state-pill--detail.notification-read-state-pill--unread {
  color: var(--color-warning);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
}

.notification-read-state-pill--detail.notification-read-state-pill--read {
  color: var(--color-success);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
}
</style>

<script setup lang="ts">
import { CircleCheckBig } from 'lucide-vue-next'

import { getNotificationReadStateLabel } from '@/entities/notification/model'

interface Props {
  unread: boolean
  variant?: 'directory' | 'detail'
  showIcon?: boolean
}

withDefaults(defineProps<Props>(), {
  variant: 'directory',
  showIcon: false,
})
</script>
