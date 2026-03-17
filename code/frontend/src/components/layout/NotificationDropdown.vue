<template>
  <div class="relative">
    <button
      ref="trigger"
      type="button"
      class="relative inline-flex h-10 w-10 items-center justify-center rounded-xl border border-border bg-surface text-text-secondary transition hover:border-primary/45 hover:bg-elevated hover:text-text-primary"
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
            class="overflow-hidden rounded-[28px] shadow-[0_32px_80px_var(--color-shadow-strong)] backdrop-blur-xl"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="text-[11px] font-semibold uppercase tracking-[0.22em] text-text-muted">
                  Notification Hub
                </div>
                <div class="mt-1 text-base font-semibold text-text-primary">
                  通知中心
                </div>
                <div class="mt-1 text-xs leading-5 text-text-secondary">
                  最近 {{ items.length }} 条消息，{{ unreadCount }} 条未读
                </div>
              </div>

              <div class="flex items-center gap-2">
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
                  class="inline-flex h-8 w-8 items-center justify-center rounded-xl border border-border bg-base/60 text-text-muted transition hover:border-primary/40 hover:text-text-primary"
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
                class="rounded-xl border border-border bg-base/70 px-3 py-2 text-xs font-medium text-text-secondary transition hover:border-primary/45 hover:text-text-primary disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="unreadCount === 0"
                @click="markAllRead"
              >
                全部标记已读
              </button>
              <button
                type="button"
                class="rounded-xl border border-border bg-base/70 px-3 py-2 text-xs font-medium text-text-secondary transition hover:border-primary/45 hover:text-text-primary"
                @click="goToNotifications"
              >
                查看全部通知
              </button>
            </div>

            <div class="mt-4 max-h-[min(520px,calc(100vh-7rem))] overflow-auto">
              <AppEmpty
                v-if="items.length === 0"
                title="暂无通知"
                description="新的系统、训练或竞赛消息会在这里实时出现。"
                icon="Bell"
              />

              <div
                v-else
                class="space-y-3"
              >
                <AppCard
                  v-for="item in previewItems"
                  :key="item.id"
                  as="button"
                  variant="action"
                  :accent="item.unread ? 'primary' : 'neutral'"
                  interactive
                  class="w-full text-left"
                  :style="notificationCardStyle(item.unread)"
                  @click="markAsRead(item.id)"
                >
                  <div class="flex items-start gap-3">
                    <div
                      class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl border"
                      :style="typeMeta(item.type).iconWrapStyle"
                    >
                      <component
                        :is="typeMeta(item.type).icon"
                        class="h-4 w-4"
                        :style="{ color: typeMeta(item.type).accentColor }"
                      />
                    </div>

                    <div class="min-w-0 flex-1">
                      <div class="flex flex-wrap items-center gap-2">
                        <span
                          class="inline-flex items-center rounded-full border px-2.5 py-1 text-[11px] font-semibold"
                          :style="typeMeta(item.type).badgeStyle"
                        >
                          {{ typeMeta(item.type).label }}
                        </span>
                        <span
                          v-if="item.unread"
                          class="inline-flex items-center rounded-full bg-primary/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.14em] text-primary"
                        >
                          未读
                        </span>
                      </div>

                      <div class="mt-2 flex items-start justify-between gap-3">
                        <div class="min-w-0">
                          <div class="break-words text-sm font-semibold text-text-primary">
                            {{ item.title }}
                          </div>
                          <div
                            v-if="item.content"
                            class="mt-1 line-clamp-2 break-words text-sm leading-6 text-text-secondary"
                          >
                            {{ item.content }}
                          </div>
                        </div>
                        <span
                          v-if="item.unread"
                          class="mt-1 inline-flex h-2.5 w-2.5 shrink-0 rounded-full"
                          :style="{ backgroundColor: typeMeta(item.type).accentColor }"
                        />
                      </div>

                      <div class="mt-3 flex items-center justify-between gap-3">
                        <div class="text-xs text-text-muted">
                          {{ formatDate(item.created_at) }}
                        </div>
                        <div class="text-xs font-medium text-text-secondary">
                          {{ item.unread ? '点击标记已读' : '已读消息' }}
                        </div>
                      </div>
                    </div>
                  </div>
                </AppCard>
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
  notificationCardStyle,
  close,
  toggleOpen,
  goToNotifications,
  markAsRead,
  markAllRead,
} = useNotificationDropdown(() => props.realtimeStatus)
</script>
