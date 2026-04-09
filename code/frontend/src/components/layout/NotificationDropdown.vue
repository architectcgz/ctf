<template>
  <div class="relative">
    <button
      ref="trigger"
      type="button"
      class="notification-trigger relative inline-flex h-10 w-10 items-center justify-center"
      aria-label="打开通知中心"
      @click="toggleOpen"
    >
      <Bell class="h-4 w-4" />
      <span
        v-if="unreadCount > 0"
        class="absolute -right-1 -top-1 inline-flex min-w-4 items-center justify-center rounded-full bg-danger px-1 text-[10px] leading-4 text-white"
      >
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </span>
    </button>

    <Teleport to="body">
      <div
        v-if="open"
        class="fixed inset-0 z-[120]"
        @click="close"
      >
        <div
          ref="panel"
          class="fixed z-[130]"
          :style="panelStyle"
          @click.stop
        >
          <AppCard
            variant="panel"
            accent="primary"
            class="notification-panel overflow-hidden border-l-2 border-primary/30"
            :style="notificationPanelStyle"
          >
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0 flex-1">
                <div class="notification-kicker text-text-muted">
                  Notification Hub
                </div>
                <div class="mt-1 text-base font-semibold text-text-primary">
                  通知中心
                </div>
                <div class="mt-1 text-xs leading-5 text-text-secondary">
                  按时间流查看最近 {{ previewItems.length }} 条消息，{{ unreadCount }} 条未读
                </div>
              </div>

              <div class="flex shrink-0 items-center gap-2">
                <span
                  class="inline-flex min-h-8 items-center gap-2 rounded-full border px-3 py-1 text-[11px] font-semibold tracking-[0.14em]"
                  :style="statusPillStyle"
                >
                  <span
                    class="inline-flex h-2 w-2 rounded-full"
                    :style="{ backgroundColor: statusMeta.accentColor }"
                  />
                  {{ statusMeta.label }}
                </span>
                <button
                  type="button"
                  class="notification-mini-button inline-flex h-8 w-8 items-center justify-center"
                  aria-label="关闭通知中心"
                  @click="close"
                >
                  <X class="h-4 w-4" />
                </button>
              </div>
            </div>

            <div class="mt-4 flex flex-wrap items-center gap-2 border-t border-border-subtle pt-4">
              <button
                type="button"
                class="notification-action-button px-3 py-2 text-xs font-medium disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="unreadCount === 0"
                @click="markAllRead"
              >
                全部标记已读
              </button>
              <button
                type="button"
                class="notification-action-button px-3 py-2 text-xs font-medium"
                @click="goToNotifications"
              >
                查看全部通知
              </button>
            </div>

            <div class="mt-4 max-h-[min(520px,calc(100vh-7rem))] overflow-auto pr-1">
              <AppEmpty
                v-if="items.length === 0"
                title="暂无通知"
                description="新的系统、训练或竞赛消息会在这里实时出现。"
                icon="Bell"
              />

              <div
                v-else
                class="notification-timeline"
              >
                <button
                  v-for="item in previewItems"
                  :key="item.id"
                  type="button"
                  class="notification-timeline-item w-full text-left"
                  :class="{ 'notification-timeline-item--unread': item.unread }"
                  @click="goToNotificationDetail(item.id)"
                >
                  <div class="notification-timeline-rail">
                    <span
                      class="notification-timeline-node"
                      :class="{ 'notification-timeline-node--unread': item.unread }"
                      :style="{
                        backgroundColor: item.unread
                          ? typeMeta(item.type).accentColor
                          : 'var(--color-border-default)',
                      }"
                    />
                  </div>

                  <div class="notification-timeline-body">
                    <div class="notification-timeline-header">
                      <div class="flex min-w-0 items-center gap-2">
                        <span
                          class="inline-flex items-center rounded-full border px-2.5 py-1 text-[11px] font-semibold"
                          :style="typeMeta(item.type).badgeStyle"
                        >
                          {{ typeMeta(item.type).label }}
                        </span>
                        <span
                          v-if="item.unread"
                          class="inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.14em] text-primary"
                          :style="{ backgroundColor: 'color-mix(in srgb, var(--color-primary) 12%, transparent)' }"
                        >
                          未读
                        </span>
                      </div>
                      <div class="notification-timeline-time">
                        {{ formatDate(item.created_at) }}
                      </div>
                    </div>

                    <div class="notification-timeline-title">
                      {{ item.title }}
                    </div>
                    <div
                      v-if="item.content"
                      class="notification-timeline-content"
                    >
                      {{ item.content }}
                    </div>

                    <div class="notification-timeline-footer">
                      <component
                        :is="typeMeta(item.type).icon"
                        class="h-3.5 w-3.5"
                        :style="{ color: typeMeta(item.type).accentColor }"
                      />
                      <span>{{ item.unread ? '打开详情并自动已读' : '查看详情' }}</span>
                    </div>
                  </div>
                </button>
              </div>
            </div>
          </AppCard>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { Bell, X } from 'lucide-vue-next'
import { useNotificationDropdown } from '@/composables/useNotificationDropdown'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import type { WebSocketStatus } from '@/composables/useWebSocket'
import { formatDate } from '@/utils/format'

const props = defineProps<{
  realtimeStatus: WebSocketStatus
}>()

const notificationPanelStyle = {
  background: 'linear-gradient(180deg, var(--color-bg-surface), var(--color-bg-base))',
}

const {
  open,
  trigger,
  panel,
  panelStyle,
  unreadCount,
  items,
  previewItems,
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
.notification-trigger,
.notification-mini-button,
.notification-action-button {
  border-radius: 14px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 72%, var(--color-bg-base));
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.04);
}

.notification-trigger:hover,
.notification-mini-button:hover,
.notification-action-button:hover {
  color: var(--color-text-primary);
  border-color: color-mix(in srgb, var(--color-primary) 34%, var(--color-border-default));
  box-shadow: 0 0 18px color-mix(in srgb, var(--color-primary) 14%, transparent);
}

.notification-panel {
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.14);
}

.notification-timeline {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.notification-timeline-item {
  position: relative;
  display: grid;
  grid-template-columns: 1.5rem minmax(0, 1fr);
  gap: 0.85rem;
  border: 1px solid transparent;
  border-radius: 16px;
  background: color-mix(in srgb, var(--color-bg-surface) 70%, var(--color-bg-base));
  padding: 0.9rem 0.95rem;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    transform 0.2s ease;
}

.notification-timeline-item:hover {
  border-color: color-mix(in srgb, var(--color-primary) 26%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-primary) 4%, var(--color-bg-surface));
  transform: translateY(-1px);
}

.notification-timeline-item--unread {
  border-color: color-mix(in srgb, var(--color-primary) 18%, var(--color-border-default));
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface)),
      color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base))
    );
}

.notification-timeline-rail {
  position: relative;
  display: flex;
  justify-content: center;
}

.notification-timeline-rail::before {
  content: '';
  position: absolute;
  top: 0.4rem;
  bottom: -1.15rem;
  left: 50%;
  width: 1px;
  transform: translateX(-50%);
  background: color-mix(in srgb, var(--color-border-subtle) 86%, transparent);
}

.notification-timeline-item:last-child .notification-timeline-rail::before {
  display: none;
}

.notification-timeline-node {
  position: relative;
  z-index: 1;
  margin-top: 0.25rem;
  display: inline-flex;
  height: 0.7rem;
  width: 0.7rem;
  border-radius: 999px;
  box-shadow: 0 0 0 6px color-mix(in srgb, var(--color-bg-surface) 90%, transparent);
}

.notification-timeline-node--unread {
  box-shadow: 0 0 0 6px color-mix(in srgb, var(--color-primary) 12%, var(--color-bg-surface));
}

.notification-timeline-body {
  min-width: 0;
}

.notification-timeline-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.notification-timeline-time {
  flex-shrink: 0;
  font-size: 0.72rem;
  color: var(--color-text-muted);
}

.notification-timeline-title {
  margin-top: 0.5rem;
  font-size: 0.92rem;
  font-weight: 700;
  line-height: 1.5;
  color: var(--color-text-primary);
}

.notification-timeline-content {
  margin-top: 0.35rem;
  display: -webkit-box;
  overflow: hidden;
  color: var(--color-text-secondary);
  font-size: 0.83rem;
  line-height: 1.65;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.notification-timeline-footer {
  margin-top: 0.6rem;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.76rem;
  color: var(--color-text-muted);
}

.notification-kicker {
  font-family: var(--font-family-mono);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}
</style>
