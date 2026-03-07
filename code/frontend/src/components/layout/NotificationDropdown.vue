<template>
  <div class="relative">
    <button
      type="button"
      class="relative inline-flex h-10 w-10 items-center justify-center rounded-lg border transition-all duration-200"
      :class="
        theme === 'light'
          ? 'border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:bg-slate-50 hover:text-slate-900'
          : 'border-slate-700 bg-slate-800 text-slate-300 hover:border-slate-600 hover:bg-slate-700 hover:text-slate-100'
      "
      @click="open = !open"
    >
      <Bell class="h-5 w-5" />
      <span
        v-if="unreadCount > 0"
        class="absolute -right-1 -top-1 inline-flex min-w-4 items-center justify-center rounded-full bg-danger px-1 text-[10px] leading-4 text-white"
      >
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </span>
    </button>

    <div
      v-if="open"
      class="absolute right-0 mt-2 w-[360px] max-w-[calc(100vw-2rem)] overflow-hidden rounded-lg border border-border bg-surface shadow-xl"
    >
      <div class="flex items-center justify-between px-4 py-3">
        <div class="text-sm font-semibold">通知</div>
        <div class="flex items-center gap-2">
          <button class="text-xs text-text-muted hover:text-text-primary" @click="markAllRead">
            全部已读
          </button>
          <button class="text-xs text-text-muted hover:text-text-primary" @click="open = false">
            关闭
          </button>
        </div>
      </div>

      <div class="max-h-[420px] overflow-auto border-t border-border">
        <div v-if="items.length === 0" class="px-4 py-6 text-sm text-text-muted">暂无通知</div>
        <button
          v-for="item in items"
          :key="item.id"
          type="button"
          class="w-full border-b border-border-subtle px-4 py-3 text-left hover:bg-elevated"
          @click="markAsRead(item.id)"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="truncate text-xs text-text-muted">{{ item.type }}</div>
              <div
                class="mt-0.5 break-words text-sm"
                :class="item.unread ? 'text-text-primary' : 'text-text-secondary'"
              >
                {{ item.title }}
              </div>
              <div
                v-if="item.content"
                class="mt-1 line-clamp-2 break-words text-xs text-text-secondary"
              >
                {{ item.content }}
              </div>
              <div class="mt-1 text-xs text-text-muted">{{ formatDate(item.created_at) }}</div>
            </div>
            <span
              v-if="item.unread"
              class="mt-1 inline-block h-2 w-2 shrink-0 rounded-full bg-primary"
            />
          </div>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Bell } from 'lucide-vue-next'
import { computed, ref } from 'vue'

import { markAsRead as markAsReadApi } from '@/api/notification'
import { useTheme } from '@/composables/useTheme'
import { useToast } from '@/composables/useToast'
import { useNotificationStore } from '@/stores/notification'
import { formatDate } from '@/utils/format'

const store = useNotificationStore()
const { theme } = useTheme()
const toast = useToast()
const open = ref(false)

const unreadCount = computed(() => store.unreadCount)
const items = computed(() => store.notifications)

async function markAsRead(id: string) {
  const target = store.notifications.find((item) => item.id === id)
  if (!target?.unread) return

  try {
    await markAsReadApi(id)
  } catch (error) {
    toast.error('标记已读失败')
    return
  }

  store.markAsRead(id)
}

async function markAllRead() {
  const unreadItems = store.notifications.filter((item) => item.unread)
  if (unreadItems.length === 0) return

  const results = await Promise.allSettled(unreadItems.map((item) => markAsReadApi(item.id)))
  const failedCount = results.filter((result) => result.status === 'rejected').length
  if (failedCount > 0) {
    toast.warning(`部分通知标记失败（${failedCount} 条）`)
  }

  store.markAllRead()
}
</script>
