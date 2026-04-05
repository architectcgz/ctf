<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Bell, Flag, GraduationCap, Info, RefreshCw, Trophy } from 'lucide-vue-next'
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

function typeIcon(type: string) {
  if (type === 'contest') return Trophy
  if (type === 'challenge') return Flag
  if (type === 'team') return GraduationCap
  return Info
}

type NotificationAccent = 'primary' | 'success' | 'warning' | 'violet'

function typeAccent(type: string): NotificationAccent {
  if (type === 'contest') return 'warning'
  if (type === 'challenge') return 'success'
  if (type === 'team') return 'violet'
  return 'primary'
}

const accentColorMap: Record<NotificationAccent, string> = {
  warning: 'var(--color-warning)',
  success: 'var(--color-success)',
  violet: 'var(--color-cat-reverse)',
  primary: 'var(--color-primary)',
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
  refresh()
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
    <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
      <div>
        <div class="journal-eyebrow">Notification Center</div>
        <h2
          class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
        >
          通知中心
        </h2>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          这里会显示系统、竞赛和训练相关通知。
        </p>

        <div class="mt-6 flex flex-wrap gap-3">
          <button
            v-if="canPublishNotification"
            type="button"
            class="journal-btn journal-btn--primary"
            @click="openPublishDrawer"
          >
            发布通知
          </button>
          <button type="button" class="journal-btn" @click="markCurrentPageRead">本页已读</button>
          <button type="button" class="journal-btn journal-btn--primary" @click="refresh">
            <RefreshCw class="h-4 w-4" />
            刷新
          </button>
        </div>
      </div>

      <article class="journal-brief rounded-[24px] border px-5 py-5">
        <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
          <Bell class="h-5 w-5 text-[var(--journal-accent)]" />
          当前消息概况
        </div>
        <div class="mt-5 grid gap-3 sm:grid-cols-2">
          <div v-for="stat in summaryStats" :key="stat.key" class="journal-note">
            <div class="journal-note-label">{{ stat.label }}</div>
            <div
              class="journal-note-value"
              :style="
                stat.key === 'unread' && unreadOnPage > 0
                  ? { color: 'var(--color-warning)' }
                  : undefined
              "
            >
              {{ stat.value }}
            </div>
            <div class="journal-note-helper">{{ stat.helper }}</div>
          </div>
        </div>
      </article>
    </div>
    <div class="notification-board mt-6 flex-1 px-1 pt-5 md:px-2 md:pt-6">
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-control-border)] border-t-[var(--journal-accent)]"
        />
      </div>

      <AppEmpty
        v-else-if="hasLoadError"
        class="notification-empty-state"
        icon="AlertTriangle"
        title="通知加载失败"
        :description="loadErrorMessage"
      >
        <template #action>
          <button type="button" class="journal-btn" @click="refresh">重新加载</button>
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
        <div class="notification-list mt-5">
          <button
            v-for="item in list"
            :key="item.id"
            type="button"
            class="journal-notification-item w-full text-left"
            :class="{ 'journal-notification-item--unread': item.unread }"
            @click="openNotificationDetail(item)"
          >
            <div class="flex items-start gap-3">
              <div
                class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl border"
                :style="{
                  borderColor: `color-mix(in srgb, ${accentColorMap[typeAccent(item.type)]} 20%, transparent)`,
                  background: `color-mix(in srgb, ${accentColorMap[typeAccent(item.type)]} 10%, transparent)`,
                  color: accentColorMap[typeAccent(item.type)],
                }"
              >
                <component :is="typeIcon(item.type)" class="h-5 w-5" />
              </div>

              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <span
                    class="rounded-full px-2 py-0.5 text-[0.65rem] font-semibold uppercase tracking-wider"
                    :style="{
                      background: `color-mix(in srgb, ${accentColorMap[typeAccent(item.type)]} 12%, transparent)`,
                      color: accentColorMap[typeAccent(item.type)],
                    }"
                    >{{ typeLabel(item.type) }}</span
                  >
                </div>
                <p class="mt-1 text-sm font-medium text-[var(--journal-ink)] line-clamp-1">
                  {{ item.title }}
                </p>
                <p class="mt-0.5 text-xs leading-5 text-[var(--journal-muted)] line-clamp-2">
                  {{ item.content }}
                </p>
              </div>

              <div class="shrink-0 text-right">
                <div class="text-xs text-[var(--journal-muted)]">
                  {{ formatDate(item.created_at) }}
                </div>
                <div v-if="item.unread" class="mt-2 flex justify-end">
                  <span class="h-2 w-2 rounded-full" style="background: var(--journal-accent)" />
                </div>
              </div>
            </div>
          </button>
        </div>

        <div v-if="total > 0" class="notification-pagination">
          <div>
            <div class="journal-note-label">Page Control</div>
            <div class="mt-2 text-sm text-[var(--journal-muted)]">
              共 {{ total }} 条，第 {{ page }} / {{ totalPages }} 页
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button type="button" class="journal-btn" :disabled="page === 1" @click="changePage(page - 1)">
              上一页
            </button>
            <button
              type="button"
              class="journal-btn"
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
  --journal-accent: #4f46e5;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-shell-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --journal-soft-border: color-mix(in srgb, var(--journal-border) 68%, transparent);
  --journal-divider: color-mix(in srgb, var(--journal-border) 56%, transparent);
  --journal-control-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.08), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-brief {
  border-color: var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.journal-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.22);
  background: rgba(99, 102, 241, 0.07);
  padding: 0.2rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border-radius: 18px;
  border: 1px solid var(--journal-shell-border);
  background: var(--journal-surface);
  padding: 0.95rem 1rem;
}

.journal-note-label {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.26em;
  text-transform: uppercase;
  color: #64748b;
}

.journal-note-value {
  margin-top: 0.65rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.45;
  color: var(--journal-muted);
}

.notification-board {
  border-top: 1px dashed var(--journal-divider);
}

.notification-list {
  border: 1px solid var(--journal-shell-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  overflow: hidden;
}

.journal-notification-item {
  border: 0;
  border-bottom: 1px dashed var(--journal-divider);
  background: linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base)));
  padding: 1rem;
  transition:
    border-color 0.2s,
    background 0.2s;
  cursor: pointer;
}

.journal-notification-item:last-child {
  border-bottom: 0;
}

.journal-notification-item:hover {
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
}

.journal-notification-item--unread {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface)),
    color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--color-bg-base))
  );
}

.notification-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-top: 1.5rem;
  padding-top: 1.25rem;
  border-top: 1px dashed var(--journal-divider);
}

.journal-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border-radius: 0.9rem;
  border: 1px solid var(--journal-control-border);
  background: var(--journal-surface);
  padding: 0.55rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-ink);
  transition:
    border-color 0.2s,
    color 0.2s,
    background 0.2s;
  cursor: pointer;
}

.journal-btn:hover:not(:disabled) {
  border-color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
}

.journal-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.journal-btn--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 50%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

.journal-btn--primary:hover:not(:disabled) {
  background: color-mix(in srgb, var(--journal-accent) 14%, transparent);
}

:deep(.notification-empty-state) {
  border-top-style: dashed;
  border-bottom-style: dashed;
  border-top-color: var(--journal-soft-border);
  border-bottom-color: var(--journal-soft-border);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.18), transparent 20rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}
</style>
