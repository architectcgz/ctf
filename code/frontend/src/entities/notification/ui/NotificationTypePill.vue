<template>
  <span
    class="notification-type-pill"
    :class="
      variant === 'detail'
        ? 'notification-type-pill--detail'
        : ['workspace-directory-status-pill', 'workspace-directory-status-pill--primary', 'notification-type-pill--directory']
    "
    :style="notificationTypePillStyle(type)"
  >
    {{ getNotificationTypeLabel(type) }}
  </span>
</template>

<style scoped>
.notification-type-pill {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  font-weight: 700;
}

.notification-type-pill--directory {
  border: 1px solid var(--notification-type-pill-border);
  background: var(--notification-type-pill-bg);
  color: var(--notification-type-pill-color);
  letter-spacing: 0.02em;
}

.notification-type-pill--detail {
  border: 1px solid var(--notification-type-pill-border);
  background: var(--notification-type-pill-bg);
  color: var(--notification-type-pill-color);
  padding: 0.4rem 0.8rem;
  font-size: var(--font-size-0-74);
}
</style>

<script setup lang="ts">
import type { NotificationType } from '@/entities/notification/model'
import { getNotificationTypeAccentColor, getNotificationTypeLabel } from '@/entities/notification/model'

interface Props {
  type: NotificationType | string
  variant?: 'directory' | 'detail'
}

withDefaults(defineProps<Props>(), {
  variant: 'directory',
})

function notificationTypePillStyle(type: NotificationType | string): Record<string, string> {
  const color = getNotificationTypeAccentColor(type)
  return {
    '--notification-type-pill-color': color,
    '--notification-type-pill-bg': `color-mix(in srgb, ${color} 12%, transparent)`,
    '--notification-type-pill-border': `color-mix(in srgb, ${color} 22%, transparent)`,
  }
}
</script>
