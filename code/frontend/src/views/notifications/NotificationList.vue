<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">通知中心</h1>
      <button
        type="button"
        class="rounded-lg border border-[var(--color-border-default)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)]"
        @click="refresh"
      >
        刷新
      </button>
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
    </div>

    <div v-else-if="list.length === 0" class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-8 text-center text-[var(--color-text-muted)]">
      暂无通知
    </div>

    <div v-else class="space-y-3">
      <button
        v-for="item in list"
        :key="item.id"
        type="button"
        class="w-full rounded-lg border p-4 text-left transition-colors hover:border-[var(--color-primary)]/40"
        :class="item.unread ? 'border-[var(--color-primary)]/30 bg-[var(--color-primary)]/5' : 'border-[var(--color-border-default)] bg-[var(--color-bg-surface)]'"
        @click="handleMarkAsRead(item)"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <span class="rounded px-2 py-0.5 text-xs font-medium" :class="typeClass(item.type)">
                {{ item.type }}
              </span>
              <span v-if="item.unread" class="rounded bg-[var(--color-primary)] px-2 py-0.5 text-[10px] font-semibold text-white">未读</span>
            </div>
            <div class="text-sm font-medium text-[var(--color-text-primary)]">{{ item.title }}</div>
            <div v-if="item.content" class="text-sm text-[var(--color-text-secondary)]">{{ item.content }}</div>
          </div>
          <div class="shrink-0 text-xs text-[var(--color-text-muted)]">{{ formatDate(item.created_at) }}</div>
        </div>
      </button>
    </div>

    <div v-if="!loading && total > 0" class="flex items-center justify-between">
      <span class="text-sm text-[var(--color-text-secondary)]">共 {{ total }} 条</span>
      <div class="flex items-center gap-2">
        <button
          :disabled="page === 1"
          class="rounded-lg border border-[var(--color-border-default)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-50"
          @click="changePage(page - 1)"
        >
          上一页
        </button>
        <span class="text-sm text-[var(--color-text-secondary)]">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        <button
          :disabled="page >= Math.ceil(total / pageSize)"
          class="rounded-lg border border-[var(--color-border-default)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-50"
          @click="changePage(page + 1)"
        >
          下一页
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'

import { getNotifications, markAsRead } from '@/api/notification'
import type { NotificationItem } from '@/api/contracts'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import { useNotificationStore } from '@/stores/notification'
import { formatDate } from '@/utils/format'

const toast = useToast()
const notificationStore = useNotificationStore()

const { list, total, page, pageSize, loading, changePage, refresh } = usePagination<NotificationItem>(getNotifications)

function typeClass(type: string): string {
  if (type === 'contest') return 'bg-[#f59e0b]/20 text-[#f59e0b]'
  if (type === 'challenge') return 'bg-[#10b981]/20 text-[#10b981]'
  if (type === 'team') return 'bg-[#8b5cf6]/20 text-[#8b5cf6]'
  return 'bg-[#0891b2]/20 text-[#0891b2]'
}

async function handleMarkAsRead(item: NotificationItem): Promise<void> {
  if (!item.unread) return
  try {
    await markAsRead(String(item.id))
    item.unread = false
    notificationStore.markAsRead(String(item.id))
  } catch (error) {
    toast.error('标记已读失败')
  }
}

onMounted(() => {
  refresh()
})
</script>
