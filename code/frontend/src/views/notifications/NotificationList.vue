<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Notification Center"
      title="通知中心"
      description="统一查看系统、竞赛、团队和训练相关消息，未读状态与顶部铃铛保持同步。"
    >
      <ElButton plain @click="markCurrentPageRead">本页已读</ElButton>
      <ElButton type="primary" @click="refresh">刷新</ElButton>
    </PageHeader>

    <section class="grid gap-4 md:grid-cols-3">
      <MetricCard label="本页消息" :value="list.length" hint="当前分页已加载的通知数量" accent="primary" />
      <MetricCard label="未读消息" :value="unreadOnPage" hint="当前分页仍未处理的通知数" :accent="unreadOnPage > 0 ? 'warning' : 'success'" />
      <MetricCard label="总消息数" :value="total" hint="通知中心累计消息总数" accent="primary" />
    </section>

    <SectionCard title="消息列表" subtitle="按时间倒序展示；点击未读消息会自动标记已读。">
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"
        />
      </div>

      <div
        v-else-if="list.length === 0"
        class="rounded-2xl border border-dashed border-[var(--color-border-default)] bg-[var(--color-bg-base)] p-8 text-center text-[var(--color-text-muted)]"
      >
        暂无通知
      </div>

      <div v-else class="space-y-3">
        <button
          v-for="item in list"
          :key="item.id"
          type="button"
          class="w-full rounded-[24px] border p-4 text-left transition hover:-translate-y-0.5 hover:border-[var(--color-primary)]/40"
          :class="
            item.unread
              ? 'border-[var(--color-primary)]/30 bg-[var(--color-primary)]/6 shadow-[0_18px_36px_var(--color-shadow-soft)]'
              : 'border-[var(--color-border-default)] bg-[var(--color-bg-base)]'
          "
          @click="handleMarkAsRead(item)"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0 space-y-2">
              <div class="flex flex-wrap items-center gap-2">
                <span class="rounded-full border px-2.5 py-1 text-xs font-medium" :class="typeClass(item.type)">
                  {{ typeLabel(item.type) }}
                </span>
                <span
                  v-if="item.unread"
                  class="rounded-full bg-[var(--color-primary)] px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.14em] text-white"
                >
                  未读
                </span>
              </div>
              <div class="text-sm font-semibold text-[var(--color-text-primary)]">{{ item.title }}</div>
              <div v-if="item.content" class="text-sm leading-6 text-[var(--color-text-secondary)]">
                {{ item.content }}
              </div>
            </div>
            <div class="shrink-0 text-right">
              <div class="text-xs text-[var(--color-text-muted)]">
                {{ formatDate(item.created_at) }}
              </div>
              <div v-if="item.unread" class="mt-3 flex justify-end">
                <span class="inline-block h-2.5 w-2.5 rounded-full bg-[var(--color-primary)]" />
              </div>
            </div>
          </div>
        </button>
      </div>

      <div v-if="!loading && total > 0" class="mt-6 flex flex-col gap-3 border-t border-border-subtle pt-4 sm:flex-row sm:items-center sm:justify-between">
        <span class="text-sm text-[var(--color-text-secondary)]">共 {{ total }} 条，第 {{ page }} / {{ Math.ceil(total / pageSize) }} 页</span>
        <div class="flex items-center gap-2">
          <button
            :disabled="page === 1"
            class="rounded-xl border border-[var(--color-border-default)] px-3 py-2 text-sm text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-50"
            @click="changePage(page - 1)"
          >
            上一页
          </button>
          <button
            :disabled="page >= Math.ceil(total / pageSize)"
            class="rounded-xl border border-[var(--color-border-default)] px-3 py-2 text-sm text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-50"
            @click="changePage(page + 1)"
          >
            下一页
          </button>
        </div>
      </div>
    </SectionCard>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'

import { getNotifications, markAsRead } from '@/api/notification'
import type { NotificationItem } from '@/api/contracts'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
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

function typeClass(type: string): string {
  if (type === 'contest') return 'border-[#f59e0b]/30 bg-[#f59e0b]/12 text-[#f59e0b]'
  if (type === 'challenge') return 'border-[#10b981]/30 bg-[#10b981]/12 text-[#10b981]'
  if (type === 'team') return 'border-[#8b5cf6]/30 bg-[#8b5cf6]/12 text-[#8b5cf6]'
  return 'border-[var(--color-primary)]/30 bg-[var(--color-primary)]/12 text-[var(--color-primary)]'
}

function typeLabel(type: string): string {
  if (type === 'contest') return '竞赛'
  if (type === 'challenge') return '训练'
  if (type === 'team') return '团队'
  return '系统'
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
