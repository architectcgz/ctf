<template>
  <div class="notification-drawer-widget">
    <slot
      name="trigger"
      :open="open"
      :toggle="toggleOpen"
      :close="close"
      :has-unread="hasUnread"
      :unread-count="unreadCount"
      :unread-badge-label="unreadBadgeLabel"
      :set-trigger-ref="setTriggerRef"
    >
      <button
        :ref="setTriggerRef"
        type="button"
        class="notification-drawer-trigger"
        :class="{ 'notification-drawer-trigger--open': open }"
        aria-label="打开通知中心"
        aria-haspopup="dialog"
        :aria-expanded="open ? 'true' : 'false'"
        @click="toggleOpen"
      >
        <Bell class="h-4 w-4" />
        <span v-if="hasUnread" class="notification-drawer-trigger__badge">
          {{ unreadBadgeLabel }}
        </span>
      </button>
    </slot>

    <Teleport to="body">
      <Transition name="notification-drawer-fade">
        <div v-if="open" class="notification-shell" @click.self="close">
          <aside class="notification-panel" role="dialog" aria-label="通知中心" aria-modal="true">
            <div class="panel-inner">
              <NotificationDrawerHeader :has-unread="hasUnread" @close="close" />

              <NotificationDrawerSummary
                :has-unread="hasUnread"
                :unread-count="unreadCount"
                :drawer-summary="drawerSummary"
                :is-marking-all-read="isMarkingAllRead"
                @mark-all-read="markAllRead"
              />

              <NotificationDrawerTabs
                :active-filter="activeFilter"
                :filter-options="filterOptions"
                @select="activeFilter = $event"
              />

              <div class="content-divider" />

              <NotificationDrawerBody
                :items="filteredItems"
                :empty-state="emptyState"
                :type-meta="typeMeta"
                @select="goToNotificationDetail"
              />
            </div>

            <NotificationDrawerFooter @view-all="goToNotifications" />
          </aside>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { Bell } from 'lucide-vue-next'

import type { WebSocketStatus } from '@/shared/model/realtime/useWebSocket'
import { useLayoutNotificationDrawerBridge } from '@/shared/model/layout'

import NotificationDrawerBody from './notification-drawer/NotificationDrawerBody.vue'
import NotificationDrawerFooter from './notification-drawer/NotificationDrawerFooter.vue'
import NotificationDrawerHeader from './notification-drawer/NotificationDrawerHeader.vue'
import NotificationDrawerSummary from './notification-drawer/NotificationDrawerSummary.vue'
import NotificationDrawerTabs from './notification-drawer/NotificationDrawerTabs.vue'
import { useNotificationDrawerViewState } from '@/shared/model/layout/notification-drawer/useNotificationDrawerViewState'
import './notification-drawer/notificationDrawer.css'

defineOptions({
  name: 'NotificationDrawer',
})

const props = defineProps<{
  realtimeStatus: WebSocketStatus
}>()

const {
  open,
  setTriggerRef,
  unreadCount,
  isMarkingAllRead,
  items,
  typeMeta,
  close,
  toggleOpen,
  goToNotifications,
  goToNotificationDetail,
  markAllRead,
} = useLayoutNotificationDrawerBridge(() => props.realtimeStatus)

const {
  activeFilter,
  filterOptions,
  filteredItems,
  emptyState,
  hasUnread,
  unreadBadgeLabel,
  drawerSummary,
} = useNotificationDrawerViewState({
  open,
  close,
  items,
  unreadCount,
})
</script>
