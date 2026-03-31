<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowLeft, BellRing, CalendarClock, CircleCheckBig, ExternalLink, Inbox } from 'lucide-vue-next'
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
      v-if="loading && !notification"
      class="notification-detail-loading"
    >
      <div class="notification-detail-spinner" />
      <span>正在加载通知详情...</span>
    </section>

    <section
      v-else-if="!notification"
      class="notification-detail-empty"
    >
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

    <article
      v-else
      class="notification-detail-page"
    >
      <header class="notification-detail-header">
        <div class="notification-detail-header-main">
          <button
            type="button"
            class="notification-detail-back"
            @click="goBackToNotifications"
          >
            <ArrowLeft class="h-4 w-4" />
            返回通知列表
          </button>

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
          </div>

          <h1 class="notification-detail-title">
            {{ notification.title }}
          </h1>
          <p class="notification-detail-summary">
            这条通知由 {{ notificationTypeLabel(notification.type) }} 模块发送，进入详情时会自动同步已读状态。
          </p>
        </div>

        <div class="notification-detail-side">
          <div class="notification-detail-side-card">
            <div class="notification-detail-side-label">
              <CalendarClock class="h-3.5 w-3.5" />
              发送时间
            </div>
            <div class="notification-detail-side-value">
              {{ formatDate(notification.created_at) }}
            </div>
          </div>

          <div class="notification-detail-side-card">
            <div class="notification-detail-side-label">
              <BellRing class="h-3.5 w-3.5" />
              通知 ID
            </div>
            <div class="notification-detail-side-value notification-detail-side-value--mono">
              {{ notification.id }}
            </div>
          </div>
        </div>
      </header>

      <section class="notification-detail-content">
        <div class="notification-detail-content-label">
          通知正文
        </div>
        <div class="notification-detail-content-body">
          {{ notification.content || '该通知暂无补充内容。' }}
        </div>
      </section>

      <footer class="notification-detail-footer">
        <button
          type="button"
          class="notification-detail-action notification-detail-action--primary"
          @click="goBackToNotifications"
        >
          返回通知列表
        </button>
        <button
          type="button"
          class="notification-detail-action"
          disabled
        >
          <Inbox class="h-4 w-4" />
          暂无关联对象
        </button>
        <button
          type="button"
          class="notification-detail-action"
          disabled
        >
          <ExternalLink class="h-4 w-4" />
          查看相关对象
        </button>
      </footer>
    </article>
  </div>
</template>

<style scoped>
.notification-detail-shell {
  width: min(72rem, 100%);
  margin: 0 auto;
}

.notification-detail-loading,
.notification-detail-empty {
  margin-top: 1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  border-radius: 24px;
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--color-primary) 8%, transparent), transparent 18rem),
    linear-gradient(180deg, var(--color-bg-surface), var(--color-bg-base));
  box-shadow: 0 20px 40px var(--color-shadow-soft);
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
  margin-top: 1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  border-radius: 24px;
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--color-primary) 8%, transparent), transparent 18rem),
    linear-gradient(180deg, var(--color-bg-surface), var(--color-bg-base));
  padding: 1.5rem;
  box-shadow: 0 20px 40px var(--color-shadow-soft);
}

.notification-detail-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 17rem;
  gap: 1rem;
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
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  cursor: pointer;
}

.notification-detail-back:hover {
  color: var(--color-text-primary);
}

.notification-detail-meta {
  margin-top: 1rem;
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
  border: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  padding: 0.4rem 0.8rem;
  font-size: 0.74rem;
  font-weight: 700;
}

.notification-detail-status {
  color: var(--color-warning);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
}

.notification-detail-status--read {
  color: var(--color-success);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
}

.notification-detail-title {
  margin-top: 1rem;
  font-size: clamp(1.6rem, 2.6vw, 2.2rem);
  font-weight: 700;
  line-height: 1.18;
  letter-spacing: -0.02em;
  color: var(--color-text-primary);
}

.notification-detail-summary {
  margin-top: 0.85rem;
  max-width: 58ch;
  font-size: 0.94rem;
  line-height: 1.8;
  color: var(--color-text-secondary);
}

.notification-detail-side {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.notification-detail-side-card {
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  border-radius: 18px;
  background: color-mix(in srgb, var(--color-bg-surface) 82%, var(--color-bg-base));
  padding: 0.95rem 1rem;
}

.notification-detail-side-label {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.notification-detail-side-value {
  margin-top: 0.65rem;
  font-size: 0.9rem;
  line-height: 1.6;
  color: var(--color-text-primary);
}

.notification-detail-side-value--mono {
  font-family: "JetBrains Mono", "Fira Code", "SFMono-Regular", monospace;
  font-size: 0.76rem;
}

.notification-detail-content {
  margin-top: 1.4rem;
  border-top: 1px solid color-mix(in srgb, var(--color-border-subtle) 86%, transparent);
  padding-top: 1.4rem;
}

.notification-detail-content-label {
  font-size: 0.76rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.notification-detail-content-body {
  margin-top: 0.9rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 68%, transparent);
  border-radius: 18px;
  background: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
  padding: 1rem 1.05rem;
  white-space: pre-wrap;
  font-size: 0.95rem;
  line-height: 1.85;
  color: var(--color-text-primary);
}

.notification-detail-footer {
  margin-top: 1.4rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.notification-detail-action {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  border-radius: 14px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 72%, var(--color-bg-base));
  padding: 0.7rem 1rem;
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.notification-detail-action:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.notification-detail-action--primary {
  border-color: color-mix(in srgb, var(--color-primary) 44%, transparent);
  background: color-mix(in srgb, var(--color-primary) 12%, transparent);
  color: var(--color-primary);
  cursor: pointer;
}

@keyframes notification-detail-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 960px) {
  .notification-detail-page {
    padding: 1.15rem;
  }

  .notification-detail-header {
    grid-template-columns: 1fr;
  }
}
</style>
