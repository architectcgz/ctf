<template>
  <div class="relative">
    <button
      ref="trigger"
      type="button"
      class="notification-trigger relative inline-flex h-10 w-10 items-center justify-center"
      :class="{ 'notification-trigger--open': open }"
      aria-label="打开通知中心"
      @click="toggleOpen"
    >
      <Bell class="h-4 w-4" />
      <span
        v-if="unreadCount > 0"
        class="notification-trigger-badge absolute -right-1 -top-1 inline-flex min-w-4 items-center justify-center rounded-full px-1"
      >
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </span>
    </button>

    <SlideOverDrawer
      :open="open"
      class="notification-shell"
      title="通知中心"
      width="26.875rem"
      body-padding="var(--space-0)"
      @update:open="open = $event"
      @close="close"
    >
      <template #icon>
        <Bell class="h-5 w-5" />
      </template>

      <template #header-extra>
        <div class="notification-overview">
          <div class="notification-summary">
            <div class="notification-counts">
              <span class="notification-counts__value">{{ unreadCount }}</span>
              <span class="notification-counts__label">未读</span>
              <span class="notification-counts__split">/</span>
              <span class="notification-counts__total">{{ items.length }} 总计</span>
            </div>

            <button
              v-if="unreadCount > 0"
              type="button"
              class="notification-summary__action"
              @click="markAllRead"
            >
              全部标为已读
            </button>
          </div>

          <div class="notification-filter-tabs" role="tablist" aria-label="通知筛选">
            <button
              v-for="filter in filterOptions"
              :key="filter.value"
              type="button"
              class="notification-filter"
              :class="{ 'notification-filter--active': activeFilter === filter.value }"
              :aria-pressed="activeFilter === filter.value"
              @click="activeFilter = filter.value"
            >
              {{ filter.label }}
            </button>
          </div>
        </div>
      </template>

      <div class="notification-panel-body relative flex-1 overflow-y-auto">
        <div v-if="filteredItems.length === 0" class="notification-empty">
          <div class="notification-empty__icon">
            <Bell class="h-8 w-8" />
          </div>
          <p class="notification-empty__title">{{ emptyState.title }}</p>
          <p class="notification-empty__copy">{{ emptyState.copy }}</p>
        </div>

        <div v-else class="notification-list">
          <button
            v-for="item in filteredItems"
            :key="item.id"
            type="button"
            class="notification-item"
            :class="{ 'notification-item--unread': item.unread }"
            @click="goToNotificationDetail(item.id)"
          >
            <span class="notification-item-icon" :style="typeMeta(item.type).iconWrapStyle">
              <component
                :is="typeMeta(item.type).icon"
                class="h-3.5 w-3.5"
                :style="{ color: typeMeta(item.type).accentColor }"
              />
            </span>

            <span class="notification-item-main">
              <span class="notification-item-meta">
                <span
                  class="notification-item-type"
                  :style="{ color: typeMeta(item.type).accentColor }"
                >
                  {{ typeMeta(item.type).label }}
                </span>
                <span class="notification-item-time">
                  {{ formatDate(item.created_at) }}
                </span>
              </span>

              <span class="notification-item-title-row">
                <span class="notification-item-title">
                  {{ item.title }}
                </span>
                <span v-if="item.unread" class="notification-item-unread-dot" />
              </span>

              <span v-if="item.content" class="notification-item-snippet">
                {{ item.content }}
              </span>
            </span>
          </button>
        </div>
      </div>

      <template #footer>
        <button type="button" class="notification-view-all" @click="goToNotifications">
          <AlignLeft class="notification-view-all__icon" />
          <span>查看全部通知</span>
          <ChevronRight class="notification-view-all__chevron" />
        </button>
      </template>
    </SlideOverDrawer>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { AlignLeft, Bell, ChevronRight } from 'lucide-vue-next'

import SlideOverDrawer from '@/components/common/modal-templates/SlideOverDrawer.vue'
import { useNotificationDropdown } from '@/features/notifications'
import type { WebSocketStatus } from '@/composables/useWebSocket'
import { formatDate } from '@/utils/format'

const props = defineProps<{
  realtimeStatus: WebSocketStatus
}>()

const {
  open,
  trigger,
  unreadCount,
  items,
  typeMeta,
  close,
  toggleOpen,
  goToNotifications,
  goToNotificationDetail,
  markAllRead,
} = useNotificationDropdown(() => props.realtimeStatus)

type NotificationFilter = 'all' | 'unread' | 'read'

const activeFilter = ref<NotificationFilter>('all')

const filterOptions: Array<{ value: NotificationFilter; label: string }> = [
  { value: 'all', label: '全部' },
  { value: 'unread', label: '未读' },
  { value: 'read', label: '已读' },
]

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
      copy: '新的系统、训练或竞赛消息会在这里出现。',
    }
  }
  if (activeFilter.value === 'read') {
    return {
      title: '暂无已读通知',
      copy: '已查看过的通知会在这里保留记录。',
    }
  }
  return {
    title: '暂无新通知',
    copy: '新的系统、训练或竞赛消息会在这里出现。',
  }
})
</script>

<style scoped>
:deep(.notification-shell.modal-template-shell--drawer) {
  --notification-surface: color-mix(in srgb, var(--color-bg-surface) 98%, var(--color-bg-base));
  --notification-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-elevated) 54%,
    var(--color-bg-surface)
  );
  --notification-surface-elevated: color-mix(
    in srgb,
    var(--color-bg-elevated) 68%,
    var(--color-bg-surface)
  );
  --notification-line: color-mix(in srgb, var(--color-border-default) 72%, transparent);
  --notification-line-strong: color-mix(in srgb, var(--color-border-default) 88%, transparent);
  --notification-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --notification-muted: color-mix(in srgb, var(--color-text-secondary) 92%, transparent);
  --notification-faint: color-mix(in srgb, var(--color-text-muted) 92%, transparent);
  --modal-shell-blur: var(--space-0-5);
  --modal-template-shell-overlay: color-mix(in srgb, var(--color-bg-base) 28%, transparent);
  --modal-template-drawer-radius: calc(var(--ui-dialog-radius-wide) + var(--space-1));
  --modal-template-drawer-header-padding-block-start: var(--space-10);
  --modal-template-drawer-header-padding-inline: calc(var(--space-8) - var(--space-0-5));
  --modal-template-drawer-header-padding-block-end: var(--space-0);
  --modal-template-drawer-header-extra-margin-top: var(--space-0);
  --modal-template-drawer-divider-margin-inline: calc(var(--space-8) - var(--space-0-5));
  --modal-template-drawer-footer-padding: var(--space-5) calc(var(--space-8) - var(--space-0-5))
    var(--space-8);
  --modal-template-drawer-icon-size: calc(var(--space-5) + var(--space-4));
  --modal-template-drawer-icon-glyph-size: calc(var(--space-5) + var(--space-0-5));
  --modal-template-drawer-close-size: calc(var(--space-10) - var(--space-0-5));
  --modal-template-drawer-close-glyph-size: var(--font-size-1-00);
  --modal-template-drawer-close-offset: calc(var(--space-8) - var(--space-0-5));
  --modal-template-drawer-title-size: var(--font-size-1-90);
  --modal-template-drawer-title-line-height: 1.18;
  --modal-template-drawer-header-surface: var(--notification-surface);
  --modal-template-drawer-body-surface: var(--notification-surface);
  --modal-template-drawer-footer-surface: var(--notification-surface);
  --modal-template-drawer-panel-border: 1px solid var(--notification-line-strong);
  --modal-template-drawer-panel-shadow:
    calc(var(--space-7) * -1) 0 calc(var(--space-12) + var(--space-3))
      color-mix(in srgb, var(--color-shadow-strong) 16%, transparent),
    calc(var(--space-0-5) * -1) 0 0 color-mix(in srgb, var(--color-border-subtle) 36%, transparent);
  --modal-template-drawer-icon-border: 1px solid
    color-mix(in srgb, var(--color-primary) 16%, var(--notification-line));
  --modal-template-drawer-icon-surface: linear-gradient(
    180deg,
    color-mix(in srgb, var(--color-primary) 8%, var(--notification-surface)),
    color-mix(in srgb, var(--color-primary) 4%, var(--notification-surface-subtle))
  );
  --modal-template-drawer-icon-color: color-mix(
    in srgb,
    var(--color-primary) 88%,
    var(--notification-text)
  );
  --modal-template-drawer-close-border: 1px solid var(--notification-line);
  --modal-template-drawer-close-surface: var(--notification-surface-subtle);
  --modal-template-drawer-close-color: var(--notification-faint);
  --modal-template-drawer-close-hover-surface: var(--notification-surface-elevated);
  --modal-template-drawer-close-hover-color: var(--notification-text);
  --modal-template-drawer-close-hover-transform: none;
}

.notification-overview {
  display: grid;
  gap: var(--space-0);
}

.notification-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}

.notification-counts {
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
}

.notification-counts__value {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-1-80);
  font-weight: 870;
  line-height: 0.9;
  color: var(--color-primary);
  letter-spacing: -0.04em;
}

.notification-counts__label {
  font-size: var(--font-size-1-05);
  font-weight: 700;
  color: var(--notification-text);
}

.notification-counts__split {
  font-size: var(--font-size-1-05);
  color: var(--notification-faint);
}

.notification-counts__total {
  font-size: var(--font-size-1-05);
  font-weight: 700;
  color: var(--notification-faint);
}

.notification-summary__action {
  display: inline-flex;
  align-items: center;
  min-height: var(--ui-control-height-sm);
  color: var(--notification-muted);
  font-size: var(--font-size-12);
  font-weight: 700;
  transition: color var(--ui-motion-fast);
}

.notification-summary__action:hover,
.notification-summary__action:focus-visible {
  color: var(--notification-text);
}

.notification-filter-tabs {
  display: flex;
  align-items: center;
  gap: var(--space-3-5);
  margin-top: var(--space-6);
}

.notification-filter {
  min-width: calc(var(--space-12) + var(--space-3));
  min-height: calc(var(--space-8) + var(--space-1));
  padding: 0 var(--space-4-5);
  border-radius: var(--ui-badge-radius-pill);
  border: 1px solid
    color-mix(in srgb, var(--notification-line-strong) 82%, var(--notification-muted));
  background: color-mix(in srgb, var(--notification-surface-subtle) 52%, var(--notification-surface));
  color: color-mix(in srgb, var(--notification-text) 72%, var(--notification-muted));
  font-size: var(--font-size-14);
  font-weight: 690;
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 68%, transparent),
    0 1px 2px color-mix(in srgb, var(--color-shadow-soft) 16%, transparent);
  transition:
    border-color var(--ui-motion-fast),
    background-color var(--ui-motion-fast),
    box-shadow var(--ui-motion-fast),
    color var(--ui-motion-fast);
}

.notification-filter:hover,
.notification-filter:focus-visible {
  border-color: color-mix(in srgb, var(--color-primary) 18%, var(--notification-line-strong));
  background: color-mix(
    in srgb,
    var(--notification-surface-elevated) 62%,
    var(--notification-surface)
  );
  color: var(--notification-text);
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 82%, transparent),
    0 var(--space-2) var(--space-4) color-mix(in srgb, var(--color-shadow-soft) 18%, transparent);
}

.notification-filter:focus-visible {
  outline: var(--ui-focus-ring-width) solid
    color-mix(in srgb, var(--color-primary) 42%, var(--notification-line-strong));
  outline-offset: var(--space-0-5);
}

.notification-filter--active {
  color: var(--color-bg-surface);
  border-color: color-mix(in srgb, var(--color-primary) 78%, var(--notification-line-strong));
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--color-primary) 88%, var(--color-bg-surface)),
    color-mix(
      in srgb,
      var(--color-primary) 96%,
      color-mix(in srgb, var(--color-primary-hover) 44%, transparent)
    )
  );
  box-shadow:
    0 var(--space-3) var(--space-5) color-mix(in srgb, var(--color-primary) 22%, transparent),
    inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 24%, transparent);
}

.notification-panel-body {
  background: var(--notification-surface);
}

.notification-empty {
  display: flex;
  height: 100%;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-10) var(--space-6);
  text-align: center;
}

.notification-empty__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--notification-line) 14%, transparent);
  color: var(--notification-faint);
}

.notification-empty__title {
  margin-top: var(--space-4);
  font-size: var(--font-size-14);
  font-weight: 700;
  color: var(--notification-text);
}

.notification-empty__copy {
  margin-top: var(--space-1);
  font-size: var(--font-size-12);
  color: var(--notification-muted);
}

.notification-list {
  display: flex;
  flex-direction: column;
  margin-top: calc(var(--space-7) + var(--space-0-5));
  border-top: 1px solid var(--notification-line);
}

.notification-item {
  display: grid;
  grid-template-columns: calc(var(--space-5) + var(--space-4)) minmax(0, 1fr) var(--space-2);
  width: 100%;
  align-items: start;
  gap: var(--space-4);
  min-height: calc(var(--space-12) + var(--space-12) + var(--space-5));
  padding: var(--space-4) var(--space-2-5) var(--space-4) var(--space-1-5);
  border-bottom: 1px solid var(--notification-line);
  text-align: left;
  cursor: pointer;
  background: transparent;
  transition:
    background-color 0.18s ease,
    border-color 0.18s ease;
}

.notification-item:first-child {
  padding-top: var(--space-4);
}

.notification-item:hover,
.notification-item:focus-visible {
  background: color-mix(in srgb, var(--color-primary) 4%, var(--notification-surface-subtle));
  border-bottom-color: color-mix(in srgb, var(--color-primary) 14%, var(--notification-line));
}

.notification-item--unread {
  background: color-mix(in srgb, var(--color-primary) 2%, transparent);
}

.notification-item-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: calc(var(--space-5) + var(--space-4));
  height: calc(var(--space-5) + var(--space-4));
  flex-shrink: 0;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--color-primary) 22%, var(--notification-line));
  background: color-mix(in srgb, var(--color-primary) 8%, var(--notification-surface-subtle));
  margin-top: var(--space-0-5);
}

.notification-item-main {
  display: grid;
  flex: 1;
  min-width: 0;
  gap: var(--space-1);
}

.notification-item-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  line-height: 1.45;
}

.notification-item-type {
  font-size: var(--font-size-1-00);
  font-weight: 800;
}

.notification-item-time {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.02em;
  color: var(--notification-faint);
  white-space: nowrap;
}

.notification-item-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.notification-item-title {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-1-00);
  font-weight: 800;
  line-height: 1.375;
  color: var(--notification-text);
}

.notification-item-snippet {
  overflow: hidden;
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--notification-muted);
  white-space: nowrap;
  text-overflow: ellipsis;
}

.notification-item-unread-dot {
  width: calc(var(--space-1-5) + var(--space-0-5));
  height: calc(var(--space-1-5) + var(--space-0-5));
  flex-shrink: 0;
  border-radius: 999px;
  background: var(--color-primary);
  box-shadow: 0 0 0 var(--space-1) color-mix(in srgb, var(--color-primary) 6%, transparent);
  margin-top: calc(var(--space-8) + var(--space-1-5));
}

.notification-view-all {
  display: grid;
  grid-template-columns: var(--space-6) minmax(0, 1fr) var(--space-4-5);
  align-items: center;
  width: 100%;
  min-height: calc(var(--space-12) + var(--space-3-5));
  gap: var(--space-3);
  padding: 0 var(--space-5);
  border-radius: calc(var(--ui-control-radius-md) + var(--space-0-5));
  border: 1px solid color-mix(in srgb, var(--notification-line-strong) 96%, transparent);
  background: var(--notification-surface);
  color: color-mix(in srgb, var(--notification-text) 92%, transparent);
  font-size: var(--font-size-15);
  font-weight: 710;
  text-align: left;
  box-shadow: 0 1px 2px color-mix(in srgb, var(--color-shadow-soft) 14%, transparent);
  transition:
    border-color var(--ui-motion-fast),
    box-shadow var(--ui-motion-fast),
    transform var(--ui-motion-fast);
}

.notification-view-all:hover,
.notification-view-all:focus-visible {
  border-color: color-mix(in srgb, var(--color-primary) 18%, var(--notification-line-strong));
  box-shadow: 0 var(--space-2) var(--space-5)
    color-mix(in srgb, var(--color-shadow-soft) 12%, transparent);
}

.notification-view-all__icon,
.notification-view-all__chevron {
  width: var(--space-5-5);
  height: var(--space-5-5);
  color: color-mix(in srgb, var(--notification-text) 94%, transparent);
}

.notification-view-all__chevron {
  justify-self: end;
}

.notification-trigger {
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.notification-trigger--open {
  border-color: color-mix(in srgb, var(--color-primary) 28%, transparent);
  background: color-mix(in srgb, var(--color-primary) 12%, transparent);
  color: color-mix(in srgb, var(--color-primary) 92%, var(--color-text-primary));
}

.notification-trigger-badge {
  font-size: var(--font-size-10);
  line-height: 1rem;
  background: color-mix(in srgb, var(--color-danger) 88%, var(--color-text-primary));
  color: var(--color-bg-base);
  box-shadow: 0 0 0 2px var(--notification-surface, var(--color-bg-surface));
}

.notification-trigger:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--color-primary) 44%, var(--notification-line-strong));
  outline-offset: 2px;
}

@media (prefers-reduced-motion: reduce) {
  .notification-trigger {
    transition-duration: 0.01ms !important;
  }
}

@media (max-width: 768px) {
  .notification-summary {
    align-items: flex-start;
    flex-direction: column;
  }

  .notification-filter-tabs {
    flex-wrap: wrap;
  }

  .notification-item {
    grid-template-columns: calc(var(--space-5) + var(--space-4)) minmax(0, 1fr);
  }

  .notification-item-unread-dot {
    display: none;
  }
}
</style>
