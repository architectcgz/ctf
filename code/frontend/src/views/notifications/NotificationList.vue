<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { Bell, Flag, GraduationCap, Info, RefreshCw, Trophy } from 'lucide-vue-next'

import { getNotifications, markAsRead } from '@/api/notification'
import type { NotificationItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import { useNotificationStore } from '@/stores/notification'
import { formatDate } from '@/utils/format'

const toast = useToast()
const notificationStore = useNotificationStore()

async function fetchNotifications(params: { page: number; page_size: number }) {
  const data = await getNotifications(params)
  if (params.page === 1) {
    notificationStore.setNotifications(data.list)
  }
  return data
}

const { list, total, page, pageSize, loading, changePage, refresh } =
  usePagination<NotificationItem>(fetchNotifications)
const unreadOnPage = computed(() => list.value.filter((item) => item.unread).length)

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

async function handleMarkAsRead(item: NotificationItem): Promise<void> {
  if (!item.unread) return
  try {
    await markAsRead(String(item.id))
    const target = list.value.find((entry) => String(entry.id) === String(item.id))
    if (target) {
      target.unread = false
    }
    notificationStore.markAsRead(String(item.id))
  } catch (error) {
    toast.error('标记已读失败')
  }
}

async function markCurrentPageRead(): Promise<void> {
  const unreadItems = list.value.filter((item) => item.unread)
  if (unreadItems.length === 0) return

  const results = await Promise.allSettled(unreadItems.map((item) => markAsRead(String(item.id))))
  const failedCount = results.filter((result) => result.status === 'rejected').length
  list.value.forEach((item) => {
    if (item.unread) item.unread = false
  })
  notificationStore.markAllRead()

  if (failedCount > 0) {
    toast.warning(`部分通知标记失败（${failedCount} 条）`)
  }
}

onMounted(() => {
  refresh()
})
</script>

<template>
  <div class="journal-shell space-y-6">
    <PageHeader
      eyebrow="Notification Center"
      title="通知中心"
      description="统一查看系统、竞赛、团队和训练相关消息，未读状态与顶部铃铛保持同步。"
    >
      <button type="button" class="journal-btn" @click="markCurrentPageRead">本页已读</button>
      <button type="button" class="journal-btn journal-btn--primary" @click="refresh">
        <RefreshCw class="h-4 w-4" />
        刷新
      </button>
    </PageHeader>

    <!-- 统计栏 -->
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 sm:grid-cols-3">
        <div>
          <div class="journal-eyebrow">本页消息</div>
          <div class="mt-2 text-3xl font-semibold tracking-tight text-[var(--journal-ink)]">{{ list.length }}</div>
          <div class="mt-1 text-sm text-[var(--journal-muted)]">当前分页已加载的通知数量</div>
        </div>
        <div>
          <div class="journal-eyebrow">未读消息</div>
          <div
            class="mt-2 text-3xl font-semibold tracking-tight"
            :style="{ color: unreadOnPage > 0 ? 'var(--color-warning)' : 'var(--color-success)' }"
          >{{ unreadOnPage }}</div>
          <div class="mt-1 text-sm text-[var(--journal-muted)]">当前分页仍未处理的通知数</div>
        </div>
        <div>
          <div class="journal-eyebrow">总消息数</div>
          <div class="mt-2 text-3xl font-semibold tracking-tight text-[var(--journal-ink)]">{{ total }}</div>
          <div class="mt-1 text-sm text-[var(--journal-muted)]">通知中心累计消息总数</div>
        </div>
      </div>
    </section>

    <!-- 消息列表 -->
    <section class="journal-panel rounded-[30px] border px-6 py-6 md:px-8">
      <div class="mb-5 flex items-center justify-between">
        <div>
          <div class="journal-eyebrow">Messages</div>
          <h2 class="mt-1 text-lg font-semibold text-[var(--journal-ink)]">消息列表</h2>
        </div>
        <p class="text-xs text-[var(--journal-muted)]">按时间倒序；点击未读消息自动标记已读</p>
      </div>

      <!-- 加载中 -->
      <div v-if="loading" class="flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]" />
      </div>

      <!-- 空状态 -->
      <AppEmpty
        v-else-if="list.length === 0"
        icon="Inbox"
        title="暂无通知"
        description="新的系统、竞赛、团队和训练消息会在这里汇总展示。"
      />

      <!-- 列表 -->
      <div v-else class="space-y-3">
        <button
          v-for="item in list"
          :key="item.id"
          type="button"
          class="journal-notification-item w-full text-left"
          :class="{ 'journal-notification-item--unread': item.unread }"
          @click="handleMarkAsRead(item)"
        >
          <div class="flex items-start gap-3">
            <!-- 图标 -->
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

            <!-- 内容 -->
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span
                  class="rounded-full px-2 py-0.5 text-[0.65rem] font-semibold uppercase tracking-wider"
                  :style="{
                    background: `color-mix(in srgb, ${accentColorMap[typeAccent(item.type)]} 12%, transparent)`,
                    color: accentColorMap[typeAccent(item.type)],
                  }"
                >{{ typeLabel(item.type) }}</span>
              </div>
              <p class="mt-1 text-sm font-medium text-[var(--journal-ink)] line-clamp-1">{{ item.title }}</p>
              <p class="mt-0.5 text-xs leading-5 text-[var(--journal-muted)] line-clamp-2">{{ item.content }}</p>
            </div>

            <!-- 时间 + 未读点 -->
            <div class="shrink-0 text-right">
              <div class="text-xs text-[var(--journal-muted)]">{{ formatDate(item.created_at) }}</div>
              <div v-if="item.unread" class="mt-2 flex justify-end">
                <span class="h-2 w-2 rounded-full" style="background: var(--journal-accent)" />
              </div>
            </div>
          </div>
        </button>
      </div>

      <!-- 分页 -->
      <div
        v-if="!loading && total > 0"
        class="mt-6 flex flex-col gap-3 border-t pt-4 sm:flex-row sm:items-center sm:justify-between"
        style="border-color: var(--journal-border)"
      >
        <span class="text-sm text-[var(--journal-muted)]">
          共 {{ total }} 条，第 {{ page }} / {{ Math.ceil(total / pageSize) }} 页
        </span>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="journal-btn"
            :disabled="page === 1"
            @click="changePage(page - 1)"
          >上一页</button>
          <button
            type="button"
            class="journal-btn"
            :disabled="page >= Math.ceil(total / pageSize)"
            @click="changePage(page + 1)"
          >下一页</button>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #4f46e5;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: rgba(248, 250, 252, 0.9);
  --journal-surface-subtle: rgba(241, 245, 249, 0.7);
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.06), transparent 20rem),
    linear-gradient(180deg, rgba(248, 250, 252, 0.98), rgba(241, 245, 249, 0.95));
}

.journal-panel {
  border-color: var(--journal-border);
  background: var(--journal-surface);
}

.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-notification-item {
  border-radius: 20px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface-subtle);
  padding: 1rem;
  transition: border-color 0.2s, background 0.2s;
  cursor: pointer;
}

.journal-notification-item:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 30%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface-subtle));
}

.journal-notification-item--unread {
  border-color: color-mix(in srgb, var(--journal-accent) 22%, var(--journal-border));
}

.journal-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border-radius: 12px;
  border: 1px solid var(--color-border-default);
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
  background: transparent;
  transition: border-color 0.2s, color 0.2s;
  cursor: pointer;
}

.journal-btn:hover:not(:disabled) {
  border-color: var(--journal-accent);
  color: var(--journal-accent);
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

:global([data-theme='dark']) .journal-shell {
  --journal-ink: #f1f5f9;
  --journal-muted: #94a3b8;
  --journal-border: rgba(51, 65, 85, 0.72);
  --journal-surface: rgba(15, 23, 42, 0.85);
  --journal-surface-subtle: rgba(30, 41, 59, 0.6);
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.18), transparent 20rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}

:global([data-theme='dark']) .journal-panel {
  background: rgba(15, 23, 42, 0.6);
}
</style>
