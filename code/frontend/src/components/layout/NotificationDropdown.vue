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
      width="24.5rem"
      @update:open="open = $event"
      @close="close"
    >
      <template #icon>
        <Bell class="h-5 w-5" />
      </template>

      <template #header-extra>
        <div class="notification-overview">
          <div class="notification-overview-row">
            <div class="notification-counts">
              <span class="notification-counts__value">{{ unreadCount }}</span>
              <span class="notification-counts__label">未读</span>
              <span class="notification-counts__split">/</span>
              <span class="notification-counts__total">{{ items.length }} 总计</span>
            </div>

            <div
              class="notification-connection"
              :title="statusMeta.label"
            >
              <span
                class="notification-connection__dot"
                :style="{ backgroundColor: statusMeta.accentColor }"
              />
              <span class="notification-connection__label">{{ statusMeta.label }}</span>
            </div>
          </div>

          <div class="notification-toolbar">
            <button
              type="button"
              class="notification-toolbar__link"
              :disabled="unreadCount === 0"
              @click="markAllRead"
            >
              全部标为已读
            </button>
            <span
              class="notification-toolbar__divider"
              aria-hidden="true"
            />
            <button
              type="button"
              class="notification-toolbar__link"
              @click="goToNotifications"
            >
              查看全部
            </button>
          </div>
        </div>
      </template>

      <div class="notification-panel-body relative flex-1 overflow-y-auto">
        <div
          v-if="items.length === 0"
          class="notification-empty"
        >
          <div class="notification-empty__icon">
            <Bell class="h-8 w-8" />
          </div>
          <p class="notification-empty__title">暂无新通知</p>
          <p class="notification-empty__copy">新的系统、训练或竞赛消息会在这里出现。</p>
        </div>

        <div
          v-else
          class="notification-list"
        >
          <button
            v-for="item in items"
            :key="item.id"
            type="button"
            class="notification-item"
            :class="{ 'notification-item--unread': item.unread }"
            @click="goToNotificationDetail(item.id)"
          >
            <span
              class="notification-item-icon"
              :style="typeMeta(item.type).iconWrapStyle"
            >
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
                <span
                  v-if="item.unread"
                  class="notification-item-unread-dot"
                />
              </span>

              <span
                v-if="item.content"
                class="notification-item-snippet"
              >
                {{ item.content }}
              </span>
            </span>
          </button>
        </div>
      </div>
    </SlideOverDrawer>
  </div>
</template>

<script setup lang="ts">
import { Bell } from 'lucide-vue-next'

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
  statusMeta,
  typeMeta,
  close,
  toggleOpen,
  goToNotifications,
  goToNotificationDetail,
  markAllRead,
} = useNotificationDropdown(() => props.realtimeStatus)
</script>

<style scoped>
:deep(.notification-shell.modal-template-shell--drawer) {
  --modal-template-shell-overlay: color-mix(in srgb, var(--color-bg-base) 45%, transparent);
}

:deep(.notification-shell .modal-template-panel--drawer) {
  --notification-surface: color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base));
  --notification-surface-subtle: color-mix(in srgb, var(--color-bg-elevated) 82%, var(--color-bg-surface));
  --notification-surface-elevated: color-mix(in srgb, var(--color-bg-elevated) 90%, var(--color-bg-surface));
  --notification-line: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --notification-line-strong: color-mix(in srgb, var(--color-border-default) 92%, transparent);
  --notification-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --notification-muted: color-mix(in srgb, var(--color-text-secondary) 92%, transparent);
  --notification-faint: color-mix(in srgb, var(--color-text-muted) 92%, transparent);
  border-left: 1px solid var(--notification-line-strong);
  box-shadow:
    0 24px 56px color-mix(in srgb, var(--color-shadow-strong) 24%, transparent),
    0 0 0 1px color-mix(in srgb, var(--color-border-subtle) 40%, transparent);
}

:deep(.notification-shell .modal-template-drawer) {
  background: var(--notification-surface);
}

:deep(.notification-shell .modal-template-drawer__header) {
  padding: var(--space-5) var(--space-7) var(--space-3);
  border-bottom: 1px solid var(--notification-line);
  background: var(--notification-surface);
}

:deep(.notification-shell .modal-template-drawer__icon) {
  border-color: color-mix(in srgb, var(--color-primary) 18%, var(--notification-line));
  background: color-mix(in srgb, var(--color-primary) 10%, var(--notification-surface));
  color: color-mix(in srgb, var(--color-primary) 92%, var(--notification-text));
}

:deep(.notification-shell .modal-template-drawer__title) {
  margin-top: var(--space-3);
  font-size: 1.375rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--notification-text);
}

:deep(.notification-shell .modal-template-drawer__header-extra) {
  margin-top: var(--space-3);
}

:deep(.notification-shell .modal-template-drawer__body) {
  padding: 0;
  background: var(--notification-surface);
}

:deep(.notification-shell .modal-template-drawer__close) {
  border-color: var(--notification-line);
  background: var(--notification-surface-subtle);
  color: var(--notification-faint);
}

:deep(.notification-shell .modal-template-drawer__close:hover) {
  border-color: var(--notification-line-strong);
  background: var(--notification-surface-elevated);
  color: var(--notification-text);
}

.notification-overview {
  display: grid;
  gap: var(--space-2);
}

.notification-overview-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.notification-counts {
  display: flex;
  align-items: baseline;
  gap: var(--space-1-5);
}

.notification-counts__value {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-18);
  font-weight: 800;
  line-height: 1;
  color: var(--color-primary);
}

.notification-counts__label {
  font-size: var(--font-size-11);
  font-weight: 700;
  color: var(--notification-muted);
}

.notification-counts__split {
  font-size: var(--font-size-11);
  color: var(--notification-faint);
}

.notification-counts__total {
  font-size: var(--font-size-11);
  font-weight: 600;
  color: var(--notification-faint);
}

.notification-connection {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  padding: var(--space-0-5) var(--space-1-5);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--notification-line) 12%, transparent);
}

.notification-connection__dot {
  width: 0.375rem;
  height: 0.375rem;
  border-radius: 999px;
}

.notification-connection__label {
  font-size: var(--font-size-10);
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--notification-faint);
}

.notification-toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.notification-toolbar__link {
  font-size: var(--font-size-11);
  font-weight: 700;
  color: var(--notification-muted);
  transition: color 0.18s ease;
}

.notification-toolbar__link:hover:not(:disabled),
.notification-toolbar__link:focus-visible:not(:disabled) {
  color: color-mix(in srgb, var(--color-primary) 92%, var(--notification-text));
}

.notification-toolbar__link:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.notification-toolbar__divider {
  width: 1px;
  height: 0.75rem;
  background: color-mix(in srgb, var(--notification-line-strong) 88%, transparent);
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
}

.notification-item {
  display: flex;
  width: 100%;
  align-items: flex-start;
  gap: var(--space-2-5);
  padding: var(--space-3) var(--space-4) var(--space-3) var(--space-3);
  border-bottom: 1px solid var(--notification-line);
  text-align: left;
  cursor: pointer;
  background: transparent;
  transition:
    background-color 0.18s ease,
    border-color 0.18s ease;
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
  width: 1.875rem;
  height: 1.875rem;
  flex-shrink: 0;
  border-radius: 999px;
  border: 1px solid transparent;
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
  gap: var(--space-2);
}

.notification-item-type {
  font-size: var(--font-size-10);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.notification-item-time {
  font-family: var(--font-family-mono);
  font-size: 0.5625rem;
  font-weight: 500;
  letter-spacing: 0.04em;
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
  font-size: var(--font-size-13);
  font-weight: 700;
  line-height: 1.4;
  color: var(--notification-text);
}

.notification-item-snippet {
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  font-size: var(--font-size-12);
  line-height: 1.5;
  color: var(--notification-muted);
}

.notification-item-unread-dot {
  width: 6px;
  height: 6px;
  flex-shrink: 0;
  border-radius: 999px;
  background: var(--color-primary);
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
  outline: 2px solid color-mix(in srgb, var(--color-primary) 44%, white);
  outline-offset: 2px;
}

@media (prefers-reduced-motion: reduce) {
  .notification-trigger {
    transition-duration: 0.01ms !important;
  }
}
</style>
