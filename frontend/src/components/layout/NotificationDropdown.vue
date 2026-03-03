<template>
  <div class="relative">
    <button
      type="button"
      class="relative inline-flex h-9 w-9 items-center justify-center rounded-lg border border-border bg-elevated text-text-secondary hover:text-text-primary"
      @click="open = !open"
    >
      <Bell class="h-4 w-4" />
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
          <button class="text-xs text-text-muted hover:text-text-primary" @click="markAllRead">全部已读</button>
          <button class="text-xs text-text-muted hover:text-text-primary" @click="open = false">关闭</button>
        </div>
      </div>

      <div class="max-h-[420px] overflow-auto border-t border-border">
        <div v-if="items.length === 0" class="px-4 py-6 text-sm text-text-muted">暂无通知</div>
        <button
          v-for="n in items"
          :key="n.id"
          type="button"
          class="w-full border-b border-border-subtle px-4 py-3 text-left hover:bg-elevated"
          @click="markAsRead(String(n.id))"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="truncate text-xs text-text-muted">{{ n.type }}</div>
              <div class="mt-0.5 break-words text-sm" :class="n.unread ? 'text-text-primary' : 'text-text-secondary'">
                {{ n.title }}
              </div>
              <div v-if="n.time" class="mt-1 text-xs text-text-muted">{{ n.time }}</div>
            </div>
            <span v-if="n.unread" class="mt-1 inline-block h-2 w-2 shrink-0 rounded-full bg-primary" />
          </div>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Bell } from 'lucide-vue-next'
import { computed, ref } from 'vue'

import { useNotificationStore } from '@/stores/notification'

const store = useNotificationStore()
const open = ref(false)

const unreadCount = computed(() => store.unreadCount)
const items = computed(() => store.notifications)

function markAsRead(id: string) {
  store.markAsRead(id)
}

function markAllRead() {
  store.markAllRead()
}
</script>

