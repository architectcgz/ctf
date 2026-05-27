<template>
  <button
    :ref="setTriggerRef"
    type="button"
    class="topnav-icon-button topnav-notification-button"
    :class="{ 'topnav-notification-button--active': open }"
    aria-label="打开通知中心"
    aria-haspopup="dialog"
    :aria-expanded="open ? 'true' : 'false'"
    @click="$emit('toggle')"
  >
    <Bell class="h-4 w-4" />
    <span
      v-if="hasUnread"
      class="topnav-notification-badge"
    >
      {{ unreadBadgeLabel }}
    </span>
  </button>
</template>

<script setup lang="ts">
import type { VNodeRef } from 'vue'
import { Bell } from 'lucide-vue-next'

defineProps<{
  open: boolean
  hasUnread: boolean
  unreadBadgeLabel: string
  setTriggerRef: VNodeRef
}>()

defineEmits<{
  toggle: []
}>()
</script>
