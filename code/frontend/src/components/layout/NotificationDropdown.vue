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
        class="notification-trigger-badge absolute -right-1 -top-1 inline-flex min-w-4 items-center justify-center rounded-full px-1 text-[10px] leading-4 text-white"
      >
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </span>
    </button>

    <Teleport to="body">
      <Transition appear name="notification-shell">
        <div v-if="open" class="fixed inset-0 z-[120]">
          <div
            class="notification-backdrop fixed inset-0 backdrop-blur-sm"
            @click="close"
          />

          <div
            ref="panel"
            class="notification-drawer fixed top-0 right-0 z-[130] flex h-screen w-full flex-col shadow-2xl sm:w-[420px]"
            @click.stop
          >
            <div class="notification-panel-head relative z-10 flex flex-col gap-4 px-6 py-6">
              <div class="flex justify-between items-start">
                <div>
                  <p class="notification-panel-kicker mb-1.5 flex items-center gap-1.5">
                    <Bell class="h-3 w-3" /> Notification Hub
                  </p>
                  <h2 class="notification-panel-title text-2xl font-black tracking-tight">通知中心</h2>
                </div>
                <div class="flex items-center gap-3">
                  <div
                    class="notification-status-pill inline-flex items-center gap-1.5 rounded-md border px-2.5 py-1 text-[11px] font-bold"
                    :style="statusPillStyle"
                  >
                    <span
                      class="inline-flex h-1.5 w-1.5 rounded-full animate-pulse"
                      :style="{ backgroundColor: statusMeta.accentColor }"
                    />
                    {{ statusMeta.label }}
                  </div>
                  <button
                    type="button"
                    class="notification-mini-button flex h-8 w-8 items-center justify-center rounded-lg border transition-colors"
                    aria-label="关闭通知中心"
                    @click="close"
                  >
                    <X class="h-4 w-4" />
                  </button>
                </div>
              </div>

              <div class="flex items-center justify-between mt-2">
                <span class="notification-summary text-[12px] font-medium">
                  最近 <span class="notification-summary__value font-bold">{{ items.length }}</span> 条消息，<span
                    class="notification-summary__accent font-bold"
                  >
                    {{ unreadCount }}
                  </span>
                  条未读
                </span>
                <div class="flex items-center gap-3">
                  <button
                    type="button"
                    class="notification-summary-action text-[12px] font-bold transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                    :disabled="unreadCount === 0"
                    @click="markAllRead"
                  >
                    全部标为已读
                  </button>
                  <div class="notification-summary-divider h-3.5 w-[1px]" />
                  <button
                    type="button"
                    class="notification-summary-link flex items-center gap-0.5 text-[12px] font-bold transition-colors"
                    @click="goToNotifications"
                  >
                    查看全部 <ChevronRight class="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
            </div>

            <div class="notification-panel-body relative flex-1 overflow-y-auto">
              <div
                v-if="items.length > 0"
                class="notification-rail absolute top-0 bottom-0 left-[35px] z-0 w-[1px]"
              />

              <div
                v-if="items.length === 0"
                class="flex h-full items-center justify-center px-6 py-10"
              >
                <AppEmpty
                  title="暂无通知"
                  description="新的系统、训练或竞赛消息会在这里实时出现。"
                  icon="Bell"
                />
              </div>

              <div v-else class="notification-timeline py-2">
                <div
                  v-for="item in items"
                  :key="item.id"
                  class="notification-item group relative border-b pl-14 pr-6 py-6 transition-colors last:border-b-0"
                  :class="item.unread ? 'notification-item--unread' : 'notification-item--read'"
                >
                  <div
                    class="notification-item-dot absolute left-[31.5px] top-8 z-10 h-[8px] w-[8px] rounded-full ring-[4px] transition-colors"
                    :class="
                      item.unread
                        ? 'notification-item-dot--unread'
                        : 'notification-item-dot--read'
                    "
                  />

                  <div class="flex justify-between items-center mb-3 gap-3">
                    <div class="flex items-center gap-2 min-w-0">
                      <span
                        class="notification-item-type rounded border px-2 py-0.5 text-[10px] font-black uppercase tracking-wider"
                        :style="typeMeta(item.type).badgeStyle"
                      >
                        {{ typeMeta(item.type).label }}
                      </span>
                      <span
                        v-if="item.unread"
                        class="notification-item-unread rounded px-2 py-0.5 text-[10px] font-black uppercase tracking-wider"
                      >
                        未读
                      </span>
                    </div>
                    <span class="notification-item-time shrink-0 text-[11px] font-mono font-bold">
                      {{ formatDate(item.created_at) }}
                    </span>
                  </div>

                  <h3
                    class="notification-item-title mb-1.5 text-[14px] font-black tracking-tight"
                    :class="{ 'notification-item-title--read': !item.unread }"
                  >
                    {{ item.title }}
                  </h3>
                  <p
                    v-if="item.content"
                    class="notification-entry-copy notification-item-copy mb-4 text-[13px] font-medium leading-relaxed"
                  >
                    {{ item.content }}
                  </p>

                  <button
                    type="button"
                    class="inline-flex items-center gap-1.5 text-[11px] font-black transition-colors"
                    :style="{ color: typeMeta(item.type).accentColor }"
                    @click="goToNotificationDetail(item.id)"
                  >
                    <component :is="typeMeta(item.type).icon" class="h-3 w-3" />
                    {{ item.unread ? '打开详情并自动已读' : '查看详情' }}
                  </button>
                </div>

                <div class="notification-endcap flex items-center justify-center gap-3 py-8 text-center opacity-60">
                  <div class="notification-endcap__line h-[1px] w-12" />
                  <span class="notification-endcap__label text-[10px] font-black uppercase tracking-widest">
                    End of Notifications
                  </span>
                  <div class="notification-endcap__line h-[1px] w-12" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { Bell, ChevronRight, X } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import { useNotificationDropdown } from '@/composables/useNotificationDropdown'
import type { WebSocketStatus } from '@/composables/useWebSocket'
import { formatDate } from '@/utils/format'

const props = defineProps<{
  realtimeStatus: WebSocketStatus
}>()

const {
  open,
  trigger,
  panel,
  unreadCount,
  items,
  statusMeta,
  statusPillStyle,
  typeMeta,
  close,
  toggleOpen,
  goToNotifications,
  goToNotificationDetail,
  markAllRead,
} = useNotificationDropdown(() => props.realtimeStatus)
</script>

<style scoped>
.notification-drawer {
  --notification-surface: color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base));
  --notification-surface-subtle: color-mix(in srgb, var(--color-bg-elevated) 82%, var(--color-bg-surface));
  --notification-surface-elevated: color-mix(in srgb, var(--color-bg-elevated) 90%, var(--color-bg-surface));
  --notification-line: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --notification-line-strong: color-mix(in srgb, var(--color-border-default) 92%, transparent);
  --notification-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --notification-muted: color-mix(in srgb, var(--color-text-secondary) 92%, transparent);
  --notification-faint: color-mix(in srgb, var(--color-text-muted) 92%, transparent);
  --notification-accent-soft: color-mix(in srgb, var(--color-primary) 10%, var(--notification-surface));
  border-left: 1px solid var(--notification-line-strong);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--notification-surface) 98%, transparent),
      color-mix(in srgb, var(--notification-surface-subtle) 96%, transparent)
    );
}

.notification-panel-head {
  border-bottom: 1px solid var(--notification-line);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--notification-surface) 98%, transparent),
      color-mix(in srgb, var(--notification-surface-subtle) 92%, transparent)
    );
}

.notification-panel-kicker {
  font-size: 0.625rem;
  font-weight: 900;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--notification-faint);
}

.notification-panel-title {
  color: var(--notification-text);
}

.notification-summary {
  color: var(--notification-muted);
}

.notification-summary__value {
  color: var(--notification-text);
}

.notification-summary__accent,
.notification-summary-link {
  color: color-mix(in srgb, var(--color-primary) 90%, var(--notification-text));
}

.notification-summary-action {
  color: var(--notification-muted);
}

.notification-summary-action:hover {
  color: color-mix(in srgb, var(--color-primary) 90%, var(--notification-text));
}

.notification-summary-divider,
.notification-rail,
.notification-endcap__line {
  background: color-mix(in srgb, var(--notification-line) 86%, transparent);
}

.notification-panel-body {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--notification-surface) 98%, transparent),
    color-mix(in srgb, var(--notification-surface-subtle) 90%, transparent)
  );
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
  background: color-mix(in srgb, var(--color-danger) 88%, var(--color-text-primary));
  box-shadow: 0 0 0 2px var(--notification-surface, var(--color-bg-surface));
}

.notification-backdrop {
  background: color-mix(in srgb, var(--color-bg-base) 44%, transparent);
}

.notification-mini-button {
  border-color: var(--notification-line);
  background: var(--notification-surface-subtle);
  color: var(--notification-faint);
}

.notification-mini-button:hover {
  border-color: var(--notification-line-strong);
  background: var(--notification-surface-elevated);
  color: var(--notification-text);
}

.notification-mini-button:focus-visible,
.notification-trigger:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--color-primary) 44%, white);
  outline-offset: 2px;
}

.notification-entry-copy {
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
}

.notification-item {
  border-bottom-color: color-mix(in srgb, var(--notification-line) 72%, transparent);
}

.notification-item--unread {
  background: color-mix(in srgb, var(--color-primary) 9%, transparent);
}

.notification-item--read:hover {
  background: color-mix(in srgb, var(--notification-line) 18%, transparent);
}

.notification-item-dot--unread {
  background: color-mix(in srgb, var(--color-primary) 90%, var(--notification-text));
  --tw-ring-color: color-mix(in srgb, var(--color-primary) 18%, transparent);
}

.notification-item-dot--read {
  background: color-mix(in srgb, var(--notification-faint) 92%, transparent);
  --tw-ring-color: var(--notification-surface);
}

.group:hover .notification-item-dot--read {
  background: color-mix(in srgb, var(--notification-muted) 92%, transparent);
  --tw-ring-color: color-mix(in srgb, var(--notification-line) 28%, transparent);
}

.notification-item-type {
  border-color: color-mix(in srgb, var(--notification-line-strong) 90%, transparent);
  background: color-mix(in srgb, var(--notification-line) 18%, var(--notification-surface-subtle));
  color: var(--notification-muted);
}

.notification-item-unread {
  border: 1px solid color-mix(in srgb, var(--color-primary) 24%, transparent);
  background: color-mix(in srgb, var(--color-primary) 14%, transparent);
  color: color-mix(in srgb, var(--color-primary) 92%, var(--notification-text));
}

.notification-item-time,
.notification-endcap__label {
  color: var(--notification-faint);
}

.notification-item-title {
  color: var(--notification-text);
}

.notification-item-title--read {
  color: color-mix(in srgb, var(--notification-text) 78%, var(--notification-muted));
}

.notification-item-copy {
  color: var(--notification-muted);
}

:global([data-theme='light']) .notification-drawer {
  --notification-surface: white;
  --notification-surface-subtle: #f8fafc;
  --notification-surface-elevated: white;
  --notification-line: color-mix(in srgb, #e2e8f0 88%, transparent);
  --notification-line-strong: color-mix(in srgb, #d9e1ec 94%, transparent);
  --notification-text: #0f172a;
  --notification-muted: #64748b;
  --notification-faint: #94a3b8;
}

:global([data-theme='dark']) .notification-drawer {
  --notification-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --notification-surface-subtle: color-mix(in srgb, var(--color-bg-elevated) 84%, var(--color-bg-surface));
  --notification-surface-elevated: color-mix(in srgb, var(--color-bg-elevated) 92%, var(--color-bg-surface));
  --notification-line: color-mix(in srgb, var(--color-border-default) 88%, transparent);
  --notification-line-strong: color-mix(in srgb, var(--color-border-default) 94%, transparent);
  --notification-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --notification-muted: color-mix(in srgb, var(--color-text-secondary) 90%, transparent);
  --notification-faint: color-mix(in srgb, var(--color-text-muted) 90%, transparent);
  box-shadow:
    0 24px 56px color-mix(in srgb, var(--color-shadow-strong) 28%, transparent),
    0 0 0 1px color-mix(in srgb, var(--color-border-subtle) 48%, transparent);
}

.notification-shell-enter-active,
.notification-shell-leave-active {
  transition: opacity 0.28s cubic-bezier(0.22, 1, 0.36, 1);
}

.notification-shell-enter-active .notification-backdrop,
.notification-shell-leave-active .notification-backdrop,
.notification-shell-enter-active .notification-drawer,
.notification-shell-leave-active .notification-drawer {
  transition:
    opacity 0.28s cubic-bezier(0.22, 1, 0.36, 1),
    transform 0.36s cubic-bezier(0.16, 1, 0.3, 1);
}

.notification-shell-enter-from,
.notification-shell-leave-to {
  opacity: 0;
}

.notification-shell-enter-from .notification-backdrop,
.notification-shell-leave-to .notification-backdrop {
  opacity: 0;
}

.notification-shell-enter-from .notification-drawer,
.notification-shell-leave-to .notification-drawer {
  opacity: 0;
  transform: translate3d(28px, 0, 0);
}

.notification-shell-enter-to .notification-drawer,
.notification-shell-leave-from .notification-drawer {
  opacity: 1;
  transform: translate3d(0, 0, 0);
}

@media (prefers-reduced-motion: reduce) {
  .notification-trigger,
  .notification-shell-enter-active,
  .notification-shell-leave-active,
  .notification-shell-enter-active .notification-backdrop,
  .notification-shell-leave-active .notification-backdrop,
  .notification-shell-enter-active .notification-drawer,
  .notification-shell-leave-active .notification-drawer {
    transition-duration: 0.01ms !important;
  }

  .notification-shell-enter-from .notification-drawer,
  .notification-shell-leave-to .notification-drawer {
    transform: none;
  }
}
</style>
