<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowLeft, BellRing, CalendarClock, CircleCheckBig, Inbox } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getNotifications, markAsRead } from '@/api/notification'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useToast } from '@/composables/useToast'
import { useNotificationStore } from '@/stores/notification'
import { formatDate } from '@/utils/format'

type NotificationAccent = 'primary' | 'success' | 'warning' | 'violet'

const accentColorMap: Record<NotificationAccent, string> = {
  warning: 'var(--color-warning)',
  success: 'var(--color-success)',
  violet: 'var(--color-cat-reverse)',
  primary: 'var(--color-primary)',
}

const route = useRoute()
const router = useRouter()
const toast = useToast()
const notificationStore = useNotificationStore()

const loading = ref(false)
const loadFailed = ref(false)
const isMarkingRead = ref(false)

const notificationId = computed(() => String(route.params.id ?? ''))
const notification = computed(
  () => notificationStore.notifications.find((item) => item.id === notificationId.value) ?? null
)

function notificationAccent(type: string): NotificationAccent {
  if (type === 'contest') return 'warning'
  if (type === 'challenge') return 'success'
  if (type === 'team') return 'violet'
  return 'primary'
}

function notificationTypeLabel(type: string): string {
  if (type === 'contest') return '竞赛'
  if (type === 'challenge') return '训练'
  if (type === 'team') return '团队'
  return '系统'
}

async function ensureNotificationLoaded(id: string) {
  if (notification.value || !id) {
    return
  }

  loading.value = true
  loadFailed.value = false

  try {
    const data = await getNotifications({ page: 1, page_size: 20 })
    notificationStore.setNotifications(data.list)
  } catch {
    loadFailed.value = true
  } finally {
    loading.value = false
  }
}

async function syncReadState(id: string) {
  if (!notification.value?.unread || isMarkingRead.value) {
    return
  }

  isMarkingRead.value = true

  try {
    await markAsRead(id)
    notificationStore.markAsRead(id)
  } catch {
    toast.error('标记已读失败')
  } finally {
    isMarkingRead.value = false
  }
}

function goBackToNotifications() {
  void router.push('/notifications')
}

watch(
  notificationId,
  async (id) => {
    if (!id) {
      return
    }

    await ensureNotificationLoaded(id)
    await syncReadState(id)
  },
  { immediate: true }
)
</script>

<template>
  <div class="notification-detail-shell">
    <section
      class="journal-shell journal-shell-user journal-hero notification-workspace min-h-full rounded-[30px] border"
    >
      <section v-if="loading && !notification" class="notification-detail-loading">
        <div class="notification-detail-spinner" />
        <span>正在加载通知详情...</span>
      </section>

      <section v-else-if="!notification" class="notification-detail-empty">
        <AppEmpty
          :icon="loadFailed ? 'AlertTriangle' : 'Inbox'"
          :title="loadFailed ? '通知加载失败' : '通知不存在'"
          :description="
            loadFailed
              ? '当前无法读取通知详情，请稍后重试。'
              : '这条通知可能已被移除，或不在当前可读取的通知范围内。'
          "
        >
          <template #action>
            <button
              type="button"
              class="notification-detail-action notification-detail-action--primary"
              @click="goBackToNotifications"
            >
              返回通知列表
            </button>
          </template>
        </AppEmpty>
      </section>

      <article v-else class="notification-detail-page">
        <header class="notification-detail-header">
          <div class="notification-detail-header-main">
            <button type="button" class="notification-detail-back" @click="goBackToNotifications">
              <ArrowLeft class="h-4 w-4" />
              返回通知列表
            </button>

            <div class="notification-overline">Notification</div>
            <h1 class="notification-detail-title workspace-page-title">
              {{ notification.title }}
            </h1>

            <div class="notification-detail-meta">
              <span
                class="notification-detail-badge"
                :style="{
                  color: accentColorMap[notificationAccent(notification.type)],
                  borderColor: `color-mix(in srgb, ${accentColorMap[notificationAccent(notification.type)]} 22%, transparent)`,
                  backgroundColor: `color-mix(in srgb, ${accentColorMap[notificationAccent(notification.type)]} 12%, transparent)`,
                }"
              >
                {{ notificationTypeLabel(notification.type) }}
              </span>
              <span
                class="notification-detail-status"
                :class="{ 'notification-detail-status--read': !notification.unread }"
              >
                <CircleCheckBig class="h-3.5 w-3.5" />
                {{ notification.unread ? '未读' : '已读' }}
              </span>
              <span class="notification-detail-meta-text">
                {{ formatDate(notification.created_at) }}
              </span>
            </div>
          </div>

          <aside class="notification-detail-side">
            <div class="notification-detail-side-card">
              <div class="notification-overline">Meta</div>
              <div class="notification-detail-side-item">
                <CalendarClock class="h-3.5 w-3.5" />
                <span>{{ formatDate(notification.created_at) }}</span>
              </div>
              <div class="notification-detail-side-item">
                <BellRing class="h-3.5 w-3.5" />
                <span>{{ notificationTypeLabel(notification.type) }}</span>
              </div>
            </div>

            <div class="notification-detail-side-card">
              <div class="notification-overline">ID</div>
              <div class="notification-detail-side-value notification-detail-side-value--mono">
                {{ notification.id }}
              </div>
            </div>
          </aside>
        </header>

        <div class="notification-divider" />

        <section class="notification-detail-content">
          <div class="notification-detail-content-head">
            <div>
              <div class="notification-overline">Message</div>
              <h2 class="notification-section-title">通知正文</h2>
            </div>
          </div>
          <div class="notification-detail-content-body">
            {{ notification.content || '该通知暂无补充内容。' }}
          </div>
        </section>

        <div class="notification-divider" />

        <footer class="notification-detail-footer">
          <button
            type="button"
            class="notification-detail-action notification-detail-action--primary"
            @click="goBackToNotifications"
          >
            返回通知列表
          </button>
          <button type="button" class="notification-detail-action" disabled>
            <Inbox class="h-4 w-4" />
            暂无关联对象
          </button>
        </footer>
      </article>
    </section>
  </div>
</template>

<style scoped>
.notification-detail-shell {
  --journal-shell-accent: var(--color-primary);
  --journal-shell-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-shell-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --journal-shell-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-surface) 76%,
    var(--color-bg-base)
  );
  --journal-shell-hero-radial-strength: 8%;
  --journal-shell-hero-radial-size: 18rem;
  --journal-shell-hero-top-strength: 97%;
  --journal-shell-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle) 95%,
    var(--color-bg-base)
  );
  --journal-shell-hero-shadow: none;
  width: 100%;
  min-width: 0;
  flex: 1 1 auto;
}

.notification-workspace {
  padding: 1.5rem;
}

.notification-detail-loading,
.notification-detail-empty {
  min-height: 14rem;
}

.notification-detail-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.85rem;
  min-height: 14rem;
  color: var(--color-text-secondary);
}

.notification-detail-spinner {
  height: 1.6rem;
  width: 1.6rem;
  border-radius: 999px;
  border: 3px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  border-top-color: var(--color-primary);
  animation: notification-detail-spin 0.8s linear infinite;
}

.notification-detail-page {
  min-height: 100%;
  background: transparent;
}

.notification-detail-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 16rem;
  gap: 1.5rem;
}

.notification-detail-header-main {
  min-width: 0;
}

.notification-detail-back {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  border: 0;
  background: transparent;
  padding: 0;
  font-size: var(--font-size-0-86);
  font-weight: 600;
  color: var(--journal-muted);
  cursor: pointer;
}

.notification-detail-back:hover {
  color: var(--journal-ink);
}

.notification-overline {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.notification-detail-title {
  margin-top: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.notification-detail-meta {
  margin-top: 1.1rem;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.notification-detail-badge,
.notification-detail-status {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  padding: 0.4rem 0.8rem;
  font-size: var(--font-size-0-74);
  font-weight: 700;
}

.notification-detail-meta-text {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.notification-detail-status {
  color: var(--color-warning);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
}

.notification-detail-status--read {
  color: var(--color-success);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
}

.notification-detail-side {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-self: start;
  min-width: 0;
  border-inline-start: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  padding-inline-start: 1.25rem;
}

.notification-detail-side-card {
  display: grid;
  gap: 0.7rem;
}

.notification-detail-side-value {
  font-size: var(--font-size-0-90);
  line-height: 1.6;
  color: var(--journal-ink);
}

.notification-detail-side-item {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-size: var(--font-size-0-85);
  line-height: 1.6;
  color: var(--journal-muted);
}

.notification-detail-side-value--mono {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-76);
}

.notification-divider {
  margin: 1.4rem 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
}

.notification-detail-content {
  min-width: 0;
}

.notification-detail-content-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
}

.notification-section-title {
  margin-top: 0.35rem;
  font-size: var(--font-size-1-10);
  font-weight: 700;
  color: var(--journal-ink);
}

.notification-detail-content-body {
  margin-top: 0.9rem;
  padding: 0;
  white-space: pre-wrap;
  font-size: var(--font-size-0-95);
  line-height: 1.85;
  color: var(--journal-ink);
}

.notification-detail-footer {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.notification-detail-action {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 80%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 95%, var(--color-bg-base));
  padding: 0.72rem 1rem;
  font-size: var(--font-size-0-88);
  font-weight: 600;
  color: var(--journal-ink);
}

.notification-detail-action:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.notification-detail-action--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 32%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink));
  cursor: pointer;
}

@keyframes notification-detail-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 960px) {
  .notification-workspace {
    padding: 1rem;
  }

  .notification-detail-header {
    grid-template-columns: 1fr;
  }

  .notification-detail-side {
    border-inline-start: 0;
    border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
    padding-inline-start: 0;
    padding-top: 1rem;
  }
}
</style>
