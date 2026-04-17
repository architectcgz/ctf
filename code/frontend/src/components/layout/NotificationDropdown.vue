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
            class="notification-backdrop fixed inset-0 bg-slate-900/20 backdrop-blur-sm"
            @click="close"
          />

          <div
            ref="panel"
            class="notification-drawer fixed top-0 right-0 h-screen w-full sm:w-[420px] bg-white shadow-2xl z-[130] flex flex-col border-l border-slate-200"
            @click.stop
          >
            <div class="px-6 py-6 border-b border-slate-100 flex flex-col gap-4 bg-white z-10 relative">
              <div class="flex justify-between items-start">
                <div>
                  <p
                    class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1.5 flex items-center gap-1.5"
                  >
                    <Bell class="h-3 w-3" /> Notification Hub
                  </p>
                  <h2 class="text-2xl font-black text-slate-900 tracking-tight">通知中心</h2>
                </div>
                <div class="flex items-center gap-3">
                  <div
                    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md border text-[11px] font-bold"
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
                    class="notification-mini-button w-8 h-8 flex items-center justify-center rounded-lg bg-slate-50 text-slate-400 hover:text-slate-900 hover:bg-slate-100 transition-colors border border-slate-200/60"
                    aria-label="关闭通知中心"
                    @click="close"
                  >
                    <X class="h-4 w-4" />
                  </button>
                </div>
              </div>

              <div class="flex items-center justify-between mt-2">
                <span class="text-[12px] font-medium text-slate-500">
                  最近 <span class="font-bold text-slate-900">{{ items.length }}</span> 条消息，<span
                    class="font-bold text-blue-600"
                  >
                    {{ unreadCount }}
                  </span>
                  条未读
                </span>
                <div class="flex items-center gap-3">
                  <button
                    type="button"
                    class="text-[12px] font-bold text-slate-500 hover:text-blue-600 transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                    :disabled="unreadCount === 0"
                    @click="markAllRead"
                  >
                    全部标为已读
                  </button>
                  <div class="w-[1px] h-3.5 bg-slate-200" />
                  <button
                    type="button"
                    class="text-[12px] font-bold text-blue-600 hover:text-blue-700 transition-colors flex items-center gap-0.5"
                    @click="goToNotifications"
                  >
                    查看全部 <ChevronRight class="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
            </div>

            <div class="flex-1 overflow-y-auto relative bg-white">
              <div
                v-if="items.length > 0"
                class="absolute top-0 bottom-0 left-[35px] w-[1px] bg-slate-200/80 z-0"
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

              <div v-else class="py-2">
                <div
                  v-for="item in items"
                  :key="item.id"
                  class="relative pl-14 pr-6 py-6 transition-colors group border-b border-slate-100/60 last:border-b-0"
                  :class="item.unread ? 'bg-blue-50/30' : 'hover:bg-slate-50/50'"
                >
                  <div
                    class="absolute left-[31.5px] top-8 w-[8px] h-[8px] rounded-full ring-[4px] z-10 transition-colors"
                    :class="
                      item.unread
                        ? 'bg-blue-600 ring-blue-100'
                        : 'bg-slate-300 ring-white group-hover:bg-slate-50 group-hover:ring-slate-100'
                    "
                  />

                  <div class="flex justify-between items-center mb-3 gap-3">
                    <div class="flex items-center gap-2 min-w-0">
                      <span
                        class="px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-wider bg-slate-100 border border-slate-200 text-slate-600"
                        :style="typeMeta(item.type).badgeStyle"
                      >
                        {{ typeMeta(item.type).label }}
                      </span>
                      <span
                        v-if="item.unread"
                        class="px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-wider bg-blue-100 border border-blue-200 text-blue-700"
                      >
                        未读
                      </span>
                    </div>
                    <span class="text-[11px] font-mono font-bold text-slate-400 shrink-0">
                      {{ formatDate(item.created_at) }}
                    </span>
                  </div>

                  <h3
                    class="text-[14px] font-black tracking-tight mb-1.5"
                    :class="item.unread ? 'text-slate-900' : 'text-slate-700'"
                  >
                    {{ item.title }}
                  </h3>
                  <p
                    v-if="item.content"
                    class="notification-entry-copy text-[13px] font-medium text-slate-500 leading-relaxed mb-4"
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

                <div class="py-8 text-center flex items-center justify-center gap-3 opacity-60">
                  <div class="w-12 h-[1px] bg-slate-200" />
                  <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
                    End of Notifications
                  </span>
                  <div class="w-12 h-[1px] bg-slate-200" />
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
.notification-trigger {
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.notification-trigger--open {
  border-color: #dbeafe;
  background: #eff6ff;
  color: #2563eb;
}

.notification-trigger-badge {
  background: #ef4444;
  box-shadow: 0 0 0 2px white;
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
