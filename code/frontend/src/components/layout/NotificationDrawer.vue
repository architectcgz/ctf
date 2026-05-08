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
              <header class="panel-header">
                <div class="title-wrap">
                  <div class="bell-wrap">
                    <BellRing class="bell-icon" />
                    <span v-if="hasUnread" class="bell-dot" />
                  </div>

                  <div>
                    <h1>NOTIFICATIONS</h1>
                    <p>通知中心</p>
                  </div>
                </div>

                <button type="button" class="close-btn" aria-label="关闭抽屉" @click="close">
                  <X class="close-btn__icon" />
                </button>
              </header>

              <section class="summary-row">
                <div class="summary-main">
                  <span v-if="hasUnread" class="summary-number">
                    {{ unreadCount }}
                  </span>
                  <span class="summary-text">{{ drawerSummary }}</span>
                </div>

                <nav class="summary-actions" aria-label="通知操作">
                  <button
                    type="button"
                    class="text-action"
                    :disabled="!hasUnread || isMarkingAllRead"
                    :aria-busy="isMarkingAllRead ? 'true' : 'false'"
                    @click="markAllRead"
                  >
                    全部设为已读
                  </button>
                </nav>
              </section>

              <section class="tabs" aria-label="通知筛选">
                <button
                  v-for="filter in filterOptions"
                  :key="filter.value"
                  type="button"
                  class="tab-btn"
                  :class="{ 'is-active': activeFilter === filter.value }"
                  :aria-pressed="activeFilter === filter.value"
                  @click="activeFilter = filter.value"
                >
                  {{ filter.label }}
                </button>
              </section>

              <div class="content-divider" />

              <Transition name="notification-list-fade" mode="out-in">
                <section v-if="filteredItems.length === 0" key="empty" class="notification-empty">
                  <div class="notification-empty__icon">
                    <Bell class="h-8 w-8" />
                  </div>
                  <p class="notification-empty__title">
                    {{ emptyState.title }}
                  </p>
                  <p class="notification-empty__copy">
                    {{ emptyState.copy }}
                  </p>
                </section>

                <section v-else key="list" class="notification-list" aria-label="通知列表">
                  <button
                    v-for="item in filteredItems"
                    :key="item.id"
                    type="button"
                    class="notice-card"
                    :class="{ 'is-unread': item.unread, 'is-read': !item.unread }"
                    @click="goToNotificationDetail(item.id)"
                  >
                    <span class="notice-icon" :style="{ color: typeMeta(item.type).accentColor }">
                      <component :is="typeMeta(item.type).icon" class="notice-icon__glyph" />
                    </span>

                    <span class="notice-body">
                      <span
                        class="notice-category"
                        :style="{ color: typeMeta(item.type).accentColor }"
                      >
                        {{ typeMeta(item.type).label }}
                      </span>

                      <span class="notice-title-row">
                        <span class="notice-title">{{ item.title }}</span>
                        <time>{{ formatDate(item.created_at) }}</time>
                      </span>

                      <span v-if="item.content" class="notice-copy">
                        {{ item.content }}
                      </span>
                    </span>

                    <span v-if="item.unread" class="unread-dot" aria-label="未读" />
                  </button>
                </section>
              </Transition>
            </div>

            <footer class="panel-footer">
              <button type="button" class="view-all-btn" @click="goToNotifications">
                <span class="footer-icon">
                  <AlignLeft class="footer-icon__glyph" />
                </span>
                <span class="view-all-btn__label"> 查看全部通知 </span>
                <ChevronRight class="arrow-icon" />
              </button>
            </footer>
          </aside>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { AlignLeft, Bell, BellRing, ChevronRight, X } from 'lucide-vue-next'

import { useNotificationDrawer } from '@/features/notifications'
import type { WebSocketStatus } from '@/composables/useWebSocket'
import { formatDate } from '@/utils/format'

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
} = useNotificationDrawer(() => props.realtimeStatus)

type NotificationFilter = 'all' | 'unread' | 'read'

const activeFilter = ref<NotificationFilter>('all')
const hasUnread = computed(() => unreadCount.value > 0)
const unreadBadgeLabel = computed(() =>
  unreadCount.value > 99 ? '99+' : String(unreadCount.value)
)
const drawerSummary = computed(() => {
  if (unreadCount.value > 0) {
    return '条未读通知待处理'
  }
  if (items.value.length > 0) {
    return '全部通知已读'
  }
  return '当前没有新通知'
})

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
      copy: '新的系统、竞赛与训练动态会优先显示在这里。',
    }
  }
  if (activeFilter.value === 'read') {
    return {
      title: '暂无已读通知',
      copy: '已处理的通知会保留在这里，方便回看。',
    }
  }
  return {
    title: '暂无新通知',
    copy: '新的系统、竞赛与训练动态会显示在这里。',
  }
})

function handleWindowKeydown(event: KeyboardEvent): void {
  if (!open.value || event.key !== 'Escape') return
  close()
}

watch(
  open,
  (isOpen) => {
    if (typeof window === 'undefined') return

    window.removeEventListener('keydown', handleWindowKeydown)
    if (isOpen) {
      window.addEventListener('keydown', handleWindowKeydown)
      document.body.style.overflow = 'hidden'
      return
    }

    document.body.style.overflow = ''
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  if (typeof window === 'undefined') return
  window.removeEventListener('keydown', handleWindowKeydown)
  document.body.style.overflow = ''
})
</script>

<style scoped>
.notification-shell {
  position: fixed;
  inset: 0;
  z-index: var(--ui-dialog-z-index);
  display: flex;
  align-items: stretch;
  justify-content: flex-end;
  padding: 0;
  background:
    linear-gradient(
      90deg,
      color-mix(in srgb, var(--color-bg-base) 18%, transparent),
      color-mix(in srgb, var(--color-bg-base) 42%, transparent) 52%,
      color-mix(in srgb, var(--color-bg-base) 66%, transparent)
    ),
    radial-gradient(
      circle at 26% 34%,
      color-mix(in srgb, var(--color-primary) 16%, transparent),
      transparent 34%
    ),
    radial-gradient(
      circle at 23% 76%,
      color-mix(in srgb, var(--color-success) 12%, transparent),
      transparent 26%
    );
  backdrop-filter: blur(7px);
  -webkit-backdrop-filter: blur(7px);
}

:global([data-theme='dark']) .notification-shell {
  background:
    linear-gradient(
      90deg,
      color-mix(in srgb, var(--color-bg-base) 30%, transparent),
      color-mix(in srgb, var(--color-bg-base) 62%, transparent) 52%,
      color-mix(in srgb, var(--color-bg-base) 86%, transparent)
    ),
    radial-gradient(
      circle at 26% 34%,
      color-mix(in srgb, var(--color-primary) 16%, transparent),
      transparent 34%
    ),
    radial-gradient(
      circle at 23% 76%,
      color-mix(in srgb, var(--color-success) 12%, transparent),
      transparent 26%
    );
}

.notification-drawer-fade-enter-active,
.notification-drawer-fade-leave-active {
  transition: opacity var(--ui-motion-normal);
}

.notification-drawer-fade-enter-from,
.notification-drawer-fade-leave-to {
  opacity: 0;
}

.notification-drawer-widget {
  display: inline-flex;
}

.notification-list-fade-enter-active,
.notification-list-fade-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.notification-list-fade-enter-from {
  opacity: 0;
  transform: translateY(var(--space-2));
}

.notification-list-fade-leave-to {
  opacity: 0;
  transform: translateY(calc(var(--space-2) * -1));
}

.notification-panel {
  --notification-panel-width: min(40.05vw, 25.3125rem);
  --notification-panel-text: var(--color-text-primary);
  --notification-panel-muted: var(--color-text-secondary);
  --notification-panel-subtle: var(--color-text-muted);
  --notification-panel-surface: rgb(255 255 255);
  --notification-panel-surface-end: rgb(244 247 251);
  --notification-panel-shell-bg: linear-gradient(180deg, rgb(255 255 255), rgb(244 247 251));
  --notification-panel-edge: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --notification-panel-edge-soft: color-mix(in srgb, var(--color-border-default) 62%, transparent);
  --notification-panel-shadow: -1.25rem 0 3.25rem rgb(15 23 42 / 0.12);
  --notification-panel-sheen:
    linear-gradient(90deg, rgb(255 255 255 / 0.32), transparent 28%),
    linear-gradient(180deg, rgb(255 255 255 / 0.38), transparent 24%);
  --notification-icon-color: color-mix(
    in srgb,
    var(--color-primary) 84%,
    var(--notification-panel-text)
  );
  --notification-signal: color-mix(
    in srgb,
    var(--color-primary) 88%,
    var(--color-brand-swatch-blue)
  );
  --notification-title: var(--notification-panel-text);
  --notification-summary: color-mix(in srgb, var(--notification-panel-text) 92%, transparent);
  --notification-action: color-mix(
    in srgb,
    var(--color-primary) 84%,
    var(--color-brand-swatch-blue)
  );
  --notification-tab-shell-bg: rgb(240 242 245);
  --notification-tab-shell-border: transparent;
  --notification-tab-text: var(--color-text-secondary);
  --notification-tab-bg: transparent;
  --notification-tab-border: transparent;
  --notification-tab-hover-text: var(--color-primary);
  --notification-tab-hover-bg: rgb(255 255 255 / 0.54);
  --notification-tab-active-text: var(--color-primary);
  --notification-tab-active-bg: rgb(255 255 255);
  --notification-tab-active-border: transparent;
  --notification-tab-active-shadow:
    0 0.125rem 0.25rem rgb(15 23 42 / 0.05), 0 1px 2px rgb(15 23 42 / 0.1);
  --notification-divider: linear-gradient(
    90deg,
    transparent,
    color-mix(in srgb, var(--color-border-default) 72%, transparent) 8%,
    color-mix(in srgb, var(--color-border-default) 62%, transparent) 92%,
    transparent
  );
  --notification-card-bg: linear-gradient(180deg, rgb(255 255 255), rgb(248 250 252));
  --notification-card-border: color-mix(in srgb, var(--color-border-default) 70%, transparent);
  --notification-card-border-hover: color-mix(
    in srgb,
    var(--color-primary) 28%,
    var(--color-border-default)
  );
  --notification-card-shadow: 0 0.875rem 2rem rgb(15 23 42 / 0.08);
  --notification-card-shadow-hover: 0 1rem 2.25rem rgb(15 23 42 / 0.12);
  --notification-card-title: var(--notification-panel-text);
  --notification-card-time: color-mix(in srgb, var(--notification-panel-muted) 86%, transparent);
  --notification-card-copy: color-mix(in srgb, var(--notification-panel-muted) 92%, transparent);
  --notification-empty-icon-bg: color-mix(in srgb, var(--color-border-default) 42%, transparent);
  --notification-footer-bg: linear-gradient(180deg, rgb(255 255 255), rgb(244 247 251));
  --notification-footer-text: var(--notification-panel-text);
  --notification-footer-icon-bg: rgb(248 250 252);
  --notification-footer-icon-border: color-mix(
    in srgb,
    var(--color-border-default) 72%,
    transparent
  );
  position: relative;
  align-self: stretch;
  width: var(--notification-panel-width);
  min-width: 23.4375rem;
  max-width: 25.3125rem;
  height: 100vh;
  height: 100dvh;
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--notification-panel-text);
  background-color: rgb(255 255 255);
  background-image: var(--notification-panel-shell-bg);
  border-left: 1px solid var(--notification-panel-edge);
  box-shadow: var(--notification-panel-shadow);
}

:global([data-theme='dark']) .notification-panel {
  --notification-panel-text: rgb(244 247 251);
  --notification-panel-muted: rgb(200 210 225 / 0.72);
  --notification-panel-subtle: rgb(188 199 213 / 0.78);
  --notification-panel-surface: rgb(14 23 34);
  --notification-panel-surface-end: rgb(9 18 29);
  --notification-panel-shell-bg: linear-gradient(180deg, rgb(14 23 34), rgb(9 18 29));
  --notification-panel-edge: rgb(151 173 202 / 0.24);
  --notification-panel-edge-soft: rgb(141 158 179 / 0.14);
  --notification-panel-shadow:
    -1.75rem 0 5.125rem rgb(0 0 0 / 0.34), inset 1px 0 0 rgb(255 255 255 / 0.03);
  --notification-panel-sheen:
    linear-gradient(90deg, rgb(255 255 255 / 0.028), transparent 28%),
    linear-gradient(180deg, rgb(255 255 255 / 0.025), transparent 24%);
  --notification-icon-color: rgb(238 244 252 / 0.93);
  --notification-signal: rgb(66 165 255);
  --notification-title: rgb(244 247 251);
  --notification-summary: rgb(245 248 252 / 0.94);
  --notification-action: rgb(112 173 247 / 0.82);
  --notification-tab-shell-bg: rgb(9 14 20);
  --notification-tab-shell-border: rgb(255 255 255 / 0.05);
  --notification-tab-text: rgb(255 255 255 / 0.5);
  --notification-tab-bg: transparent;
  --notification-tab-border: transparent;
  --notification-tab-hover-text: rgb(255 255 255);
  --notification-tab-hover-bg: rgb(255 255 255 / 0.03);
  --notification-tab-active-text: rgb(255 255 255);
  --notification-tab-active-bg: rgb(30 41 59);
  --notification-tab-active-border: rgb(255 255 255 / 0.1);
  --notification-tab-active-shadow: 0 0.25rem 0.75rem rgb(0 0 0 / 0.3);
  --notification-divider: linear-gradient(
    90deg,
    transparent,
    rgb(150 166 188 / 0.26) 8%,
    rgb(150 166 188 / 0.22) 92%,
    transparent
  );
  --notification-card-bg: linear-gradient(180deg, rgb(18 31 45), rgb(13 23 35));
  --notification-card-border: rgb(139 158 181 / 0.17);
  --notification-card-border-hover: rgb(143 167 199 / 0.28);
  --notification-card-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.025), 0 1rem 2.625rem rgb(0 0 0 / 0.16);
  --notification-card-shadow-hover:
    inset 0 1px 0 rgb(255 255 255 / 0.035), 0 1.125rem 2.875rem rgb(0 0 0 / 0.2);
  --notification-card-title: rgb(249 251 255);
  --notification-card-time: rgb(188 199 213 / 0.78);
  --notification-card-copy: rgb(217 225 237 / 0.84);
  --notification-empty-icon-bg: rgb(152 166 184 / 0.14);
  --notification-footer-bg: linear-gradient(180deg, rgb(10 19 30), rgb(8 17 27));
  --notification-footer-text: rgb(241 246 253);
  --notification-footer-icon-bg: rgb(15 25 39);
  --notification-footer-icon-border: rgb(138 158 184 / 0.16);
  background-color: rgb(14 23 34);
}

.notification-panel::before {
  content: '';
  position: absolute;
  inset: 0;
  background: var(--notification-panel-sheen);
  pointer-events: none;
}

.panel-inner {
  position: relative;
  z-index: 1;
  flex: 1 1 0;
  min-height: 0;
  padding: 3rem 2rem var(--space-4) 1.5rem;
  overflow-y: auto;
  background-color: rgb(255 255 255);
  background-image: linear-gradient(180deg, rgb(255 255 255), rgb(244 247 251));
}

:global([data-theme='dark']) .panel-inner {
  background-color: rgb(14 23 34);
  background-image: linear-gradient(180deg, rgb(14 23 34), rgb(9 18 29));
}

.panel-header {
  position: relative;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  min-height: 3rem;
}

.title-wrap {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
}

.bell-wrap {
  position: relative;
  width: 2rem;
  height: 2rem;
  color: var(--notification-icon-color);
  display: flex;
  align-items: center;
  justify-content: center;
}

.bell-icon {
  width: var(--space-5);
  height: var(--space-5);
  filter: drop-shadow(0 0 0.4375rem rgb(90 163 255 / 0.12));
}

.bell-dot {
  position: absolute;
  top: var(--space-0-5);
  right: var(--space-1);
  width: var(--space-1-5);
  height: var(--space-1-5);
  border-radius: 50%;
  background: var(--notification-signal);
  box-shadow: 0 0 0.875rem color-mix(in srgb, var(--notification-signal) 70%, transparent);
}

.panel-header h1 {
  margin: 0;
  color: var(--notification-title);
  font-size: var(--font-size-16);
  line-height: 1.18;
  font-weight: 500;
  letter-spacing: 0.14rem;
}

.panel-header p {
  margin: var(--space-2) 0 0;
  color: var(--notification-panel-muted);
  font-size: var(--font-size-12);
  line-height: 1.2;
  letter-spacing: 0.0125rem;
}

.close-btn {
  position: relative;
  top: 0.1875rem;
  width: 1.875rem;
  height: 1.875rem;
  padding: 0;
  color: var(--notification-panel-muted);
  border: none;
  border-radius: 0.75rem;
  background: transparent;
  cursor: pointer;
}

.close-btn__icon {
  width: var(--space-4-5);
  height: var(--space-4-5);
}

.close-btn:hover,
.close-btn:focus-visible {
  background: color-mix(in srgb, var(--notification-panel-text) 7%, transparent);
}

.close-btn:focus-visible,
.text-action:focus-visible,
.tab-btn:focus-visible,
.notice-card:focus-visible,
.view-all-btn:focus-visible {
  outline: var(--ui-focus-ring-width) solid rgb(114 184 255 / 0.46);
  outline-offset: var(--space-1);
}

.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.375rem;
  margin-top: var(--space-8);
  min-height: 2rem;
}

.summary-main {
  display: flex;
  align-items: baseline;
  gap: var(--space-3);
  min-width: 0;
  white-space: nowrap;
}

.summary-number {
  color: var(--notification-signal);
  font-size: var(--font-size-26);
  line-height: 1;
  font-weight: 300;
  letter-spacing: -0.0625rem;
  text-shadow: 0 0 1.125rem color-mix(in srgb, var(--notification-signal) 32%, transparent);
}

.summary-text {
  min-width: 0;
  color: var(--notification-summary);
  font-size: var(--font-size-13);
  line-height: 1;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
}

.summary-actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  white-space: nowrap;
}

.text-action {
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--notification-action);
  font-size: var(--font-size-13);
  font-weight: 500;
  cursor: pointer;
}

.text-action:hover {
  color: color-mix(in srgb, var(--notification-action) 86%, var(--notification-panel-text));
}

.text-action:disabled {
  cursor: default;
  opacity: 0.54;
}

.tabs {
  margin-top: var(--space-6);
  min-height: 2.25rem;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  align-items: center;
  column-gap: var(--space-1-5);
  padding: var(--space-0-5);
  border: 1px solid var(--notification-tab-shell-border);
  border-radius: var(--ui-control-radius-lg);
  background: var(--notification-tab-shell-bg);
}

.tab-btn {
  height: 1.75rem;
  border: 1px solid var(--notification-tab-border);
  border-radius: var(--ui-control-radius-md);
  background: var(--notification-tab-bg);
  color: var(--notification-tab-text);
  font-size: var(--font-size-13);
  font-weight: 500;
  cursor: pointer;
  transition:
    border-color var(--ui-motion-fast),
    background-color var(--ui-motion-fast),
    color var(--ui-motion-fast),
    box-shadow var(--ui-motion-fast);
}

.tab-btn:hover:not(.is-active) {
  color: var(--notification-tab-hover-text);
  background: var(--notification-tab-hover-bg);
}

.tab-btn.is-active {
  color: var(--notification-tab-active-text);
  font-weight: 600;
  background: var(--notification-tab-active-bg);
  border-color: var(--notification-tab-active-border);
  box-shadow: var(--notification-tab-active-shadow);
}

.tab-btn:focus-visible {
  outline-color: var(--color-primary);
  outline-offset: calc(var(--space-0-5) * -1);
}

.content-divider {
  height: 1px;
  margin-top: var(--space-4);
  background: var(--notification-divider);
}

.notification-empty {
  display: flex;
  min-height: 18rem;
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
  background: var(--notification-empty-icon-bg);
  color: var(--notification-panel-muted);
}

.notification-empty__title {
  margin-top: var(--space-4);
  font-size: var(--font-size-14);
  font-weight: 700;
  color: var(--notification-summary);
}

.notification-empty__copy {
  margin-top: var(--space-1);
  font-size: var(--font-size-12);
  color: var(--notification-panel-muted);
}

.notification-list {
  display: grid;
  gap: var(--space-2-5);
  margin-top: var(--space-4);
}

.notice-card {
  position: relative;
  min-height: 5.875rem;
  display: grid;
  grid-template-columns: 2.75rem minmax(0, 1fr);
  gap: var(--space-2-5);
  width: 100%;
  padding: var(--space-4) 2.25rem var(--space-4) var(--space-4);
  border: 1px solid var(--notification-card-border);
  border-radius: var(--ui-control-radius-lg);
  background: var(--notification-card-bg);
  box-shadow: var(--notification-card-shadow);
  text-align: left;
  cursor: pointer;
  transition:
    border-color var(--ui-motion-fast),
    transform var(--ui-motion-fast),
    box-shadow var(--ui-motion-fast);
}

.notice-card:hover,
.notice-card:focus-visible {
  border-color: var(--notification-card-border-hover);
  transform: translateY(-0.0625rem);
  box-shadow: var(--notification-card-shadow-hover);
}

.notice-card.is-unread {
  border-color: var(--notification-card-border-hover);
}

.notice-card.is-read {
  opacity: 0.95;
}

.notice-icon {
  width: 2.125rem;
  height: 2.125rem;
  margin-top: var(--space-0-5);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(
    circle at 48% 43%,
    rgb(81 240 130 / 0.2),
    rgb(58 188 101 / 0.14) 52%,
    rgb(58 188 101 / 0.09)
  );
  box-shadow:
    0 0 1.75rem rgb(67 225 121 / 0.09),
    inset 0 0 0 1px rgb(76 230 126 / 0.04);
}

.notice-icon__glyph {
  width: var(--space-4);
  height: var(--space-4);
}

.notice-body {
  min-width: 0;
  display: block;
}

.notice-category {
  display: block;
  margin-bottom: var(--space-1);
  font-size: var(--font-size-12);
  line-height: 1;
  font-weight: 600;
}

.notice-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
}

.notice-title {
  min-width: 0;
  flex: 0 1 auto;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--notification-card-title);
  font-size: var(--font-size-15);
  line-height: 1.18;
  font-weight: 700;
  letter-spacing: 0.0125rem;
}

.notice-title-row time {
  min-width: 0;
  color: var(--notification-card-time);
  font-size: var(--font-size-12);
  line-height: 1.1;
  font-weight: 400;
  white-space: nowrap;
}

.notice-copy {
  display: -webkit-box;
  width: 100%;
  max-width: 20.625rem;
  margin: var(--space-2) 0 0;
  color: var(--notification-card-copy);
  font-size: var(--font-size-13);
  line-height: 1.45;
  font-weight: 400;
  letter-spacing: 0.00625rem;
  overflow: hidden;
  text-overflow: ellipsis;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.unread-dot {
  position: absolute;
  top: var(--space-5);
  right: var(--space-4);
  width: var(--space-2);
  height: var(--space-2);
  border-radius: 50%;
  background: var(--notification-signal);
  box-shadow:
    0 0 1.125rem color-mix(in srgb, var(--notification-signal) 72%, transparent),
    0 0 0.125rem rgb(255 255 255 / 0.4) inset;
}

.panel-footer {
  position: relative;
  flex: 0 0 auto;
  margin-top: auto;
  width: 100%;
  padding: var(--space-3) 0 calc(var(--space-3) + env(safe-area-inset-bottom, 0px));
  z-index: 1;
  border-top: 1px solid var(--notification-panel-edge-soft);
  background-color: rgb(255 255 255);
  background-image: var(--notification-footer-bg);
  box-shadow: none;
}

:global([data-theme='dark']) .panel-footer {
  background-color: rgb(10 19 30);
}

.view-all-btn {
  width: 100%;
  min-height: var(--ui-control-height-lg);
  display: grid;
  grid-template-columns: 2.25rem minmax(0, 1fr) 1.75rem;
  align-items: center;
  column-gap: var(--space-3);
  padding: 0 1.75rem 0 1.5rem;
  border: 0;
  background-color: rgb(255 255 255);
  background-image: var(--notification-footer-bg);
  color: var(--notification-footer-text);
  cursor: pointer;
  text-align: left;
}

:global([data-theme='dark']) .view-all-btn {
  background-color: rgb(10 19 30);
}

.footer-icon {
  width: 2.125rem;
  height: 2.125rem;
  border-radius: 50%;
  color: var(--notification-footer-text);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--notification-footer-icon-bg);
  border: 1px solid var(--notification-footer-icon-border);
}

.footer-icon__glyph {
  width: var(--space-4);
  height: var(--space-4);
}

.view-all-btn__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-14);
  line-height: 1;
  font-weight: 500;
  letter-spacing: 0.00625rem;
}

.arrow-icon {
  justify-self: end;
  width: var(--space-5);
  height: var(--space-5);
  color: var(--notification-footer-text);
}

.notification-drawer-trigger {
  position: relative;
  display: inline-flex;
  width: var(--space-10);
  height: var(--space-10);
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 88%, transparent);
  background: color-mix(in srgb, var(--color-bg-elevated) 58%, var(--color-bg-surface));
  color: var(--color-text-secondary);
  box-shadow: 0 1px 2px color-mix(in srgb, var(--color-shadow-soft) 12%, transparent);
  cursor: pointer;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.notification-drawer-trigger--open {
  border-color: color-mix(in srgb, var(--color-primary) 28%, transparent);
  background: color-mix(in srgb, var(--color-primary) 12%, transparent);
  color: color-mix(in srgb, var(--color-primary) 92%, var(--color-text-primary));
}

.notification-drawer-trigger:hover {
  border-color: color-mix(in srgb, var(--color-border-default) 92%, transparent);
  background: color-mix(in srgb, var(--color-bg-elevated) 72%, var(--color-bg-surface));
  color: var(--color-text-primary);
}

.notification-drawer-trigger__badge {
  position: absolute;
  top: calc(var(--space-1) * -1);
  right: calc(var(--space-1) * -1);
  display: inline-flex;
  min-width: var(--space-4);
  align-items: center;
  justify-content: center;
  padding: 0 var(--space-1);
  border-radius: 999px;
  font-size: var(--font-size-10);
  line-height: 1rem;
  background: color-mix(in srgb, var(--color-danger) 88%, var(--color-text-primary));
  color: var(--color-bg-base);
  box-shadow: 0 0 0 2px var(--color-bg-surface);
}

.notification-drawer-trigger:focus-visible {
  outline: var(--ui-focus-ring-width) solid
    color-mix(in srgb, var(--color-primary) 44%, var(--color-border-default));
  outline-offset: var(--space-0-5);
}

@media (prefers-reduced-motion: reduce) {
  .notification-drawer-trigger {
    transition-duration: 0.01ms;
  }
}

@media (max-width: 768px) {
  .notification-panel {
    width: 100vw;
    min-width: 0;
    max-width: none;
  }

  .panel-inner {
    padding: var(--space-8) var(--space-4) var(--space-4);
  }

  .summary-row {
    align-items: flex-start;
    flex-direction: column;
    margin-top: var(--space-7);
  }

  .tabs {
    margin-top: var(--space-5);
  }

  .notice-card {
    grid-template-columns: 2.5rem minmax(0, 1fr);
    gap: var(--space-2-5);
    min-height: 5.75rem;
    padding: var(--space-3-5) 2rem var(--space-3-5) var(--space-3-5);
  }

  .notice-icon {
    width: var(--space-8);
    height: var(--space-8);
  }

  .notice-title {
    font-size: var(--font-size-14);
  }

  .notice-title-row time {
    font-size: var(--font-size-13);
  }

  .notice-copy {
    font-size: var(--font-size-13);
  }
}
</style>
