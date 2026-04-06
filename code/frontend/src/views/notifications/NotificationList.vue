<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Bell, RefreshCw } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import { getNotifications, markAsRead } from '@/api/notification'
import type { NotificationItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AdminNotificationPublishDrawer from '@/components/notifications/AdminNotificationPublishDrawer.vue'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notification'
import { formatDate } from '@/utils/format'

const toast = useToast()
const authStore = useAuthStore()
const notificationStore = useNotificationStore()
const router = useRouter()
const publishDrawerOpen = ref(false)

async function fetchNotifications(params: { page: number; page_size: number }) {
  const data = await getNotifications(params)
  if (params.page === 1) {
    notificationStore.setNotifications(data.list)
  }
  return data
}

const { list, total, page, pageSize, loading, error, changePage, refresh } =
  usePagination<NotificationItem>(fetchNotifications)
const unreadOnPage = computed(() => list.value.filter((item) => item.unread).length)
const readOnPage = computed(() => list.value.length - unreadOnPage.value)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const hasLoadError = computed(() => Boolean(error.value) && list.value.length === 0)
const loadErrorMessage = computed(() => {
  if (error.value instanceof Error && error.value.message.trim().length > 0) {
    return error.value.message
  }
  return '通知加载失败，请稍后重试。'
})

function typeLabel(type: string): string {
  if (type === 'contest') return '竞赛'
  if (type === 'challenge') return '训练'
  if (type === 'team') return '团队'
  return '系统'
}

function openNotificationDetail(item: NotificationItem): void {
  void router.push(`/notifications/${encodeURIComponent(String(item.id))}`)
}

async function markCurrentPageRead(): Promise<void> {
  const unreadItems = list.value.filter((item) => item.unread)
  if (unreadItems.length === 0) return

  const results = await Promise.allSettled(unreadItems.map((item) => markAsRead(String(item.id))))
  const failedCount = results.filter((result) => result.status === 'rejected').length
  unreadItems.forEach((item, index) => {
    if (results[index]?.status === 'fulfilled') {
      const target = list.value.find((entry) => String(entry.id) === String(item.id))
      if (target) {
        target.unread = false
      }
      notificationStore.markAsRead(String(item.id))
    }
  })

  if (failedCount > 0) {
    toast.warning(`部分通知标记失败（${failedCount} 条）`)
  }
}

onMounted(() => {
  void refresh()
})

const summaryStats = computed(() => [
  { key: 'page', label: '本页消息', value: list.value.length, helper: '当前分页已加载的通知数量' },
  { key: 'unread', label: '未读消息', value: unreadOnPage.value, helper: '仍待你确认或处理的消息' },
  { key: 'read', label: '已读消息', value: readOnPage.value, helper: '当前页中已经处理过的消息' },
  { key: 'total', label: '总消息数', value: total.value, helper: '通知中心累计消息总数' },
])

const canPublishNotification = computed(() => authStore.isAdmin)

function openPublishDrawer(): void {
  publishDrawerOpen.value = true
}

function closePublishDrawer(): void {
  publishDrawerOpen.value = false
}

async function handlePublishSuccess(): Promise<void> {
  closePublishDrawer()
  await refresh()
}
</script>

<template>
  <section
    class="journal-shell journal-hero flex min-h-full flex-1 flex-col space-y-6 rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div class="notification-page">
      <header class="notification-topbar">
        <div class="notification-heading">
          <div class="journal-eyebrow">Notifications</div>
          <h1 class="notification-title">通知中心</h1>
          <p class="notification-subtitle">系统、竞赛和训练相关通知会在这里按时间顺序汇总。</p>
        </div>

        <div class="notification-actions">
          <button
            v-if="canPublishNotification"
            type="button"
            class="notification-btn notification-btn-primary"
            @click="openPublishDrawer"
          >
            发布通知
          </button>
          <button type="button" class="notification-btn" @click="markCurrentPageRead">本页已读</button>
          <button type="button" class="notification-btn" @click="refresh">
            <RefreshCw class="h-4 w-4" />
            刷新
          </button>
        </div>
      </header>

      <section class="notification-summary">
        <div class="notification-summary-title">
          <Bell class="h-4 w-4" />
          <span>当前消息概况</span>
        </div>
        <div class="notification-summary-grid">
          <div v-for="stat in summaryStats" :key="stat.key" class="notification-summary-item">
            <div class="notification-summary-label">{{ stat.label }}</div>
            <div class="notification-summary-value">{{ stat.value }}</div>
            <div class="notification-summary-helper">{{ stat.helper }}</div>
          </div>
        </div>
      </section>

      <div class="notification-divider" />

      <div v-if="loading" class="notification-loading">
        <div class="notification-loading-spinner" />
      </div>

      <AppEmpty
        v-else-if="hasLoadError"
        class="notification-empty-state"
        icon="AlertTriangle"
        title="通知加载失败"
        :description="loadErrorMessage"
      >
        <template #action>
          <button type="button" class="notification-btn" @click="refresh">重新加载</button>
        </template>
      </AppEmpty>

      <AppEmpty
        v-else-if="list.length === 0"
        class="notification-empty-state"
        icon="Inbox"
        title="暂无通知"
        description="新的系统、竞赛、团队和训练消息会在这里汇总展示。"
      />

      <template v-else>
        <section class="notification-directory" aria-label="通知目录">
          <div class="notification-directory-top">
            <h2 class="notification-directory-title">消息列表</h2>
            <div class="notification-directory-meta">共 {{ total }} 条</div>
          </div>

          <div class="notification-directory-head">
            <span>类型</span>
            <span>标题与内容</span>
            <span>时间</span>
            <span>状态</span>
          </div>

          <button
            v-for="item in list"
            :key="item.id"
            type="button"
            class="notification-row"
            :class="{ 'notification-row-unread': item.unread }"
            @click="openNotificationDetail(item)"
          >
            <div class="notification-row-type">
              <span class="notification-chip">{{ typeLabel(item.type) }}</span>
            </div>
            <div class="notification-row-main">
              <div class="notification-row-title">{{ item.title }}</div>
              <div class="notification-row-copy">{{ item.content }}</div>
            </div>
            <div class="notification-row-time">{{ formatDate(item.created_at) }}</div>
            <div class="notification-row-state">
              <span class="notification-state-chip" :class="{ 'notification-state-chip-unread': item.unread }">
                {{ item.unread ? '未读' : '已读' }}
              </span>
            </div>
          </button>
        </section>

        <div v-if="total > 0" class="notification-pagination">
          <div>
            <div class="notification-summary-label">Page Control</div>
            <div class="notification-pagination-copy">共 {{ total }} 条，第 {{ page }} / {{ totalPages }} 页</div>
          </div>
          <div class="notification-pagination-actions">
            <button type="button" class="notification-btn" :disabled="page === 1" @click="changePage(page - 1)">
              上一页
            </button>
            <button
              type="button"
              class="notification-btn"
              :disabled="page >= totalPages"
              @click="changePage(page + 1)"
            >
              下一页
            </button>
          </div>
        </div>
      </template>
    </div>

    <AdminNotificationPublishDrawer
      :open="publishDrawerOpen"
      @close="closePublishDrawer"
      @published="handlePublishSuccess"
    />
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-elevated) 74%, var(--color-bg-base));
  --journal-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  font-family:
    'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
    sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 7%, transparent), transparent 22rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base)));
  box-shadow: 0 22px 50px var(--color-shadow-soft);
}

.notification-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.notification-topbar {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  justify-content: space-between;
  gap: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.journal-eyebrow {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.notification-title {
  margin-top: 0.85rem;
  font-size: clamp(32px, 4vw, 46px);
  line-height: 1.02;
  letter-spacing: -0.04em;
  color: var(--journal-ink);
}

.notification-subtitle {
  margin-top: 12px;
  max-width: 720px;
  font-size: 14px;
  line-height: 1.7;
  color: var(--journal-muted);
}

.notification-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.notification-summary {
  display: grid;
  gap: 18px;
  padding: 24px 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.notification-summary-title {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.notification-summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.notification-summary-item {
  min-width: 0;
  padding-left: 16px;
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.notification-summary-label {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.notification-summary-value {
  margin-top: 8px;
  font-size: 22px;
  line-height: 1;
  letter-spacing: -0.03em;
  color: var(--journal-ink);
}

.notification-summary-helper {
  margin-top: 8px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--journal-muted);
}

.notification-divider {
  margin-top: 24px;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.notification-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
}

.notification-loading-spinner {
  width: 32px;
  height: 32px;
  border: 4px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-top-color: var(--journal-accent);
  border-radius: 999px;
  animation: notificationSpin 900ms linear infinite;
}

:deep(.notification-empty-state) {
  margin-top: 24px;
  border-top-style: solid;
  border-bottom-style: solid;
  border-top-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-bottom-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.notification-directory {
  margin-top: 24px;
}

.notification-directory-top {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  justify-content: space-between;
  gap: 8px 16px;
  padding-bottom: 14px;
}

.notification-directory-title {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--journal-ink);
}

.notification-directory-meta,
.notification-pagination-copy {
  font-size: 13px;
  line-height: 1.6;
  color: var(--journal-muted);
}

.notification-directory-head {
  display: grid;
  grid-template-columns: 140px minmax(0, 1fr) 180px 120px;
  gap: 16px;
  padding: 0 0 12px;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.notification-row {
  display: grid;
  grid-template-columns: 140px minmax(0, 1fr) 180px 120px;
  gap: 16px;
  align-items: center;
  width: 100%;
  padding: 18px 0;
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.notification-row-unread {
  box-shadow: inset 2px 0 0 color-mix(in srgb, var(--journal-accent) 56%, transparent);
}

.notification-row:hover,
.notification-row:focus-visible {
  background: color-mix(in srgb, var(--journal-accent) 5%, transparent);
  outline: none;
}

.notification-chip,
.notification-state-chip {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 9px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
}

.notification-chip {
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
}

.notification-row-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--journal-ink);
}

.notification-row-copy {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--journal-muted);
}

.notification-row-time {
  font-size: 13px;
  line-height: 1.6;
  color: var(--journal-muted);
}

.notification-state-chip {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.notification-state-chip-unread {
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
}

.notification-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 24px;
  margin-top: 24px;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.notification-pagination-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.notification-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 40px;
  padding: 0 14px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 12px;
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
  font-size: 14px;
  font-weight: 600;
  color: var(--journal-ink);
  cursor: pointer;
}

.notification-btn-primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: var(--color-bg-base);
}

@keyframes notificationSpin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1180px) {
  .notification-directory-head {
    display: none;
  }

  .notification-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 960px) {
  .notification-summary-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .notification-title {
    font-size: 34px;
  }
}
</style>
