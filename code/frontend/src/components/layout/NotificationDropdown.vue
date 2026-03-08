<template>
  <div class="relative">
    <button
      ref="trigger"
      type="button"
      class="relative inline-flex h-10 w-10 items-center justify-center rounded-xl border border-border bg-surface text-text-secondary transition hover:border-primary/45 hover:bg-elevated hover:text-text-primary"
      @click="toggleOpen"
    >
      <Bell class="h-4 w-4" />
      <span
        v-if="unreadCount > 0"
        class="absolute -right-1 -top-1 inline-flex min-w-4 items-center justify-center rounded-full bg-danger px-1 text-[10px] leading-4 text-white"
      >
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </span>
    </button>

    <Teleport to="body">
      <div v-if="open" class="fixed inset-0 z-[120]" @click="close">
        <div
          ref="panel"
          class="fixed z-[130] overflow-hidden rounded-2xl border border-border bg-[var(--color-bg-surface)] shadow-[0_32px_80px_var(--color-shadow-strong)] ring-1 ring-black/25"
          :style="panelStyle"
          @click.stop
        >
          <div class="h-px bg-[linear-gradient(90deg,transparent,rgba(8,145,178,0.55),transparent)]" />
          <div class="flex items-center justify-between border-b border-border px-4 py-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">Realtime</div>
              <div class="text-sm font-semibold text-text-primary">通知中心</div>
            </div>
            <div class="flex items-center gap-2">
              <button class="text-xs text-text-muted hover:text-text-primary" @click="markAllRead">
                全部已读
              </button>
              <button class="text-xs text-text-muted hover:text-text-primary" @click="close">
                关闭
              </button>
            </div>
          </div>

          <div class="max-h-[min(420px,calc(100vh-7rem))] overflow-auto bg-[var(--color-bg-surface)]">
            <div v-if="items.length === 0" class="px-4 py-6 text-sm text-text-muted">暂无通知</div>
            <button
              v-for="item in items"
              :key="item.id"
              type="button"
              class="w-full border-b border-border-subtle px-4 py-3 text-left transition hover:bg-elevated/85"
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
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { Bell } from 'lucide-vue-next'
import { computed, onBeforeUnmount, ref, useTemplateRef, watch } from 'vue'

import { markAsRead as markAsReadApi } from '@/api/notification'
import { useToast } from '@/composables/useToast'
import { useNotificationStore } from '@/stores/notification'
import { formatDate } from '@/utils/format'

const store = useNotificationStore()
const toast = useToast()
const open = ref(false)
const trigger = useTemplateRef<HTMLButtonElement>('trigger')
const panelStyle = ref<Record<string, string>>({})
const repositionPanel = () => updatePanelPosition()

const unreadCount = computed(() => store.unreadCount)
const items = computed(() => store.notifications)

function updatePanelPosition() {
  if (!trigger.value) return

  const rect = trigger.value.getBoundingClientRect()
  const viewportPadding = 12
  const panelWidth = Math.min(360, window.innerWidth - viewportPadding * 2)
  const left = Math.min(
    Math.max(viewportPadding, rect.right - panelWidth),
    window.innerWidth - panelWidth - viewportPadding,
  )
  const top = rect.bottom + 8

  panelStyle.value = {
    top: `${top}px`,
    left: `${left}px`,
    width: `${panelWidth}px`,
  }
}

function close() {
  open.value = false
}

function toggleOpen() {
  open.value = !open.value
}

watch(open, (isOpen) => {
  if (!isOpen) return

  updatePanelPosition()
  window.addEventListener('resize', repositionPanel)
  window.addEventListener('scroll', repositionPanel, true)

  const cleanup = () => {
    window.removeEventListener('resize', repositionPanel)
    window.removeEventListener('scroll', repositionPanel, true)
  }

  const stop = watch(open, (next) => {
    if (!next) {
      cleanup()
      stop()
    }
  })
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', repositionPanel)
  window.removeEventListener('scroll', repositionPanel, true)
})

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
  close()
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
  close()
}
</script>
