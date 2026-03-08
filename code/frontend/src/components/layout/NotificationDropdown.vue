<template>
  <div class="relative">
    <button
      ref="trigger"
      type="button"
      class="relative inline-flex h-10 w-10 items-center justify-center rounded-xl border border-border bg-surface text-text-secondary transition hover:border-primary/45 hover:bg-elevated hover:text-text-primary"
      aria-label="打开通知中心"
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
          class="fixed z-[130] overflow-hidden rounded-[28px] border border-border bg-surface/96 shadow-[0_32px_80px_var(--color-shadow-strong)] backdrop-blur-xl"
          :style="panelStyle"
          @click.stop
        >
          <div class="pointer-events-none absolute inset-x-0 top-0 h-20 bg-[linear-gradient(180deg,rgba(8,145,178,0.12),transparent)]" />
          <div class="pointer-events-none absolute inset-y-0 left-0 w-px bg-[linear-gradient(180deg,transparent,rgba(255,255,255,0.08),transparent)]" />

          <div class="relative border-b border-border px-4 py-4">
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="text-[11px] font-semibold uppercase tracking-[0.22em] text-text-muted">Notification Hub</div>
                <div class="mt-1 text-base font-semibold text-text-primary">通知中心</div>
                <div class="mt-1 text-xs leading-5 text-text-secondary">
                  最近 {{ items.length }} 条消息，{{ unreadCount }} 条未读
                </div>
              </div>

              <div class="flex items-center gap-2">
                <span
                  class="inline-flex min-h-8 items-center gap-2 rounded-full border px-3 py-1 text-[11px] font-semibold tracking-[0.14em]"
                  :style="statusPillStyle"
                >
                  <span class="inline-flex h-2 w-2 rounded-full" :style="{ backgroundColor: statusMeta.accentColor }" />
                  {{ statusMeta.label }}
                </span>
                <button
                  type="button"
                  class="inline-flex h-8 w-8 items-center justify-center rounded-xl border border-border bg-base/60 text-text-muted transition hover:border-primary/40 hover:text-text-primary"
                  aria-label="关闭通知中心"
                  @click="close"
                >
                  <X class="h-4 w-4" />
                </button>
              </div>
            </div>

            <div class="mt-4 flex flex-wrap items-center gap-2">
              <button
                type="button"
                class="rounded-xl border border-border bg-base/70 px-3 py-2 text-xs font-medium text-text-secondary transition hover:border-primary/45 hover:text-text-primary disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="unreadCount === 0"
                @click="markAllRead"
              >
                全部标记已读
              </button>
              <button
                type="button"
                class="rounded-xl border border-border bg-base/70 px-3 py-2 text-xs font-medium text-text-secondary transition hover:border-primary/45 hover:text-text-primary"
                @click="goToNotifications"
              >
                查看全部通知
              </button>
            </div>
          </div>

          <div class="max-h-[min(520px,calc(100vh-7rem))] overflow-auto px-3 py-3">
            <div v-if="items.length === 0" class="rounded-[24px] border border-dashed border-border bg-base/60 px-5 py-8 text-center">
              <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                <Bell class="h-5 w-5" />
              </div>
              <div class="mt-3 text-sm font-semibold text-text-primary">暂无通知</div>
              <div class="mt-1 text-sm leading-6 text-text-secondary">
                新的系统、训练或竞赛消息会在这里实时出现。
              </div>
            </div>

            <div v-else class="space-y-3">
              <button
                v-for="item in previewItems"
                :key="item.id"
                type="button"
                class="w-full rounded-[24px] border px-4 py-4 text-left transition hover:-translate-y-0.5"
                :style="notificationCardStyle(item.unread)"
                @click="markAsRead(item.id)"
              >
                <div class="flex items-start gap-3">
                  <div
                    class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl border"
                    :style="typeMeta(item.type).iconWrapStyle"
                  >
                    <component :is="typeMeta(item.type).icon" class="h-4 w-4" :style="{ color: typeMeta(item.type).accentColor }" />
                  </div>

                  <div class="min-w-0 flex-1">
                    <div class="flex flex-wrap items-center gap-2">
                      <span
                        class="inline-flex items-center rounded-full border px-2.5 py-1 text-[11px] font-semibold"
                        :style="typeMeta(item.type).badgeStyle"
                      >
                        {{ typeMeta(item.type).label }}
                      </span>
                      <span
                        v-if="item.unread"
                        class="inline-flex items-center rounded-full bg-primary/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.14em] text-primary"
                      >
                        未读
                      </span>
                    </div>

                    <div class="mt-2 flex items-start justify-between gap-3">
                      <div class="min-w-0">
                        <div class="break-words text-sm font-semibold text-text-primary">
                          {{ item.title }}
                        </div>
                        <div v-if="item.content" class="mt-1 line-clamp-2 break-words text-sm leading-6 text-text-secondary">
                          {{ item.content }}
                        </div>
                      </div>
                      <span
                        v-if="item.unread"
                        class="mt-1 inline-flex h-2.5 w-2.5 shrink-0 rounded-full"
                        :style="{ backgroundColor: typeMeta(item.type).accentColor }"
                      />
                    </div>

                    <div class="mt-3 flex items-center justify-between gap-3">
                      <div class="text-xs text-text-muted">{{ formatDate(item.created_at) }}</div>
                      <div class="text-xs font-medium text-text-secondary">
                        {{ item.unread ? '点击标记已读' : '已读消息' }}
                      </div>
                    </div>
                  </div>
                </div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { Bell, Flag, GraduationCap, Info, Trophy, X } from 'lucide-vue-next'
import { computed, onBeforeUnmount, ref, useTemplateRef, watch } from 'vue'
import { useRouter } from 'vue-router'

import { markAsRead as markAsReadApi } from '@/api/notification'
import { useToast } from '@/composables/useToast'
import type { WebSocketStatus } from '@/composables/useWebSocket'
import { useNotificationStore } from '@/stores/notification'
import { formatDate } from '@/utils/format'

interface NotificationTypeMeta {
  icon: Component
  label: string
  accentColor: string
  iconWrapStyle: Record<string, string>
  badgeStyle: Record<string, string>
}

interface StatusMeta {
  label: string
  accentColor: string
}

const props = defineProps<{
  realtimeStatus: WebSocketStatus
}>()

const router = useRouter()
const store = useNotificationStore()
const toast = useToast()
const open = ref(false)
const trigger = useTemplateRef<HTMLButtonElement>('trigger')
const panelStyle = ref<Record<string, string>>({})
const repositionPanel = () => updatePanelPosition()

const unreadCount = computed(() => store.unreadCount)
const items = computed(() => store.notifications)
const previewItems = computed(() => items.value.slice(0, 6))
const statusMeta = computed<StatusMeta>(() => {
  if (props.realtimeStatus === 'open') return { label: '实时在线', accentColor: 'var(--color-success)' }
  if (props.realtimeStatus === 'connecting') return { label: '连接中', accentColor: 'var(--color-warning)' }
  if (props.realtimeStatus === 'error') return { label: '连接异常', accentColor: 'var(--color-danger)' }
  return { label: '手动刷新', accentColor: 'var(--color-text-muted)' }
})
const statusPillStyle = computed<Record<string, string>>(() => ({
  color: statusMeta.value.accentColor,
  borderColor: `color-mix(in srgb, ${statusMeta.value.accentColor} 22%, var(--color-border-default))`,
  backgroundColor: `color-mix(in srgb, ${statusMeta.value.accentColor} 10%, transparent)`,
}))

const typeMap: Record<string, NotificationTypeMeta> = {
  system: {
    icon: Info,
    label: '系统',
    accentColor: 'var(--color-primary)',
    iconWrapStyle: {
      backgroundColor: 'var(--color-primary-soft)',
      borderColor: 'color-mix(in srgb, var(--color-primary) 28%, transparent)',
    },
    badgeStyle: {
      color: 'var(--color-primary)',
      borderColor: 'color-mix(in srgb, var(--color-primary) 22%, transparent)',
      backgroundColor: 'color-mix(in srgb, var(--color-primary) 10%, transparent)',
    },
  },
  contest: {
    icon: Trophy,
    label: '竞赛',
    accentColor: 'var(--color-warning)',
    iconWrapStyle: {
      backgroundColor: 'rgba(210, 153, 34, 0.12)',
      borderColor: 'rgba(210, 153, 34, 0.26)',
    },
    badgeStyle: {
      color: 'var(--color-warning)',
      borderColor: 'rgba(210, 153, 34, 0.22)',
      backgroundColor: 'rgba(210, 153, 34, 0.1)',
    },
  },
  challenge: {
    icon: Flag,
    label: '训练',
    accentColor: 'var(--color-success)',
    iconWrapStyle: {
      backgroundColor: 'rgba(63, 185, 80, 0.12)',
      borderColor: 'rgba(63, 185, 80, 0.26)',
    },
    badgeStyle: {
      color: 'var(--color-success)',
      borderColor: 'rgba(63, 185, 80, 0.22)',
      backgroundColor: 'rgba(63, 185, 80, 0.1)',
    },
  },
  team: {
    icon: GraduationCap,
    label: '团队',
    accentColor: '#8b5cf6',
    iconWrapStyle: {
      backgroundColor: 'rgba(139, 92, 246, 0.12)',
      borderColor: 'rgba(139, 92, 246, 0.26)',
    },
    badgeStyle: {
      color: '#a78bfa',
      borderColor: 'rgba(139, 92, 246, 0.22)',
      backgroundColor: 'rgba(139, 92, 246, 0.1)',
    },
  },
}

const fallbackTypeMeta: NotificationTypeMeta = typeMap.system

function typeMeta(type: string): NotificationTypeMeta {
  return typeMap[type] || fallbackTypeMeta
}

function notificationCardStyle(unread: boolean): Record<string, string> {
  if (unread) {
    return {
      borderColor: 'color-mix(in srgb, var(--color-primary) 22%, var(--color-border-default))',
      background:
        'linear-gradient(180deg, color-mix(in srgb, var(--color-primary) 10%, transparent), rgba(15, 23, 42, 0.38))',
      boxShadow: '0 18px 36px var(--color-shadow-soft)',
    }
  }

  return {
    borderColor: 'var(--color-border-default)',
    backgroundColor: 'color-mix(in srgb, var(--color-bg-base) 62%, transparent)',
  }
}

function updatePanelPosition() {
  if (!trigger.value) return

  const rect = trigger.value.getBoundingClientRect()
  const viewportPadding = 12
  const panelWidth = Math.min(420, window.innerWidth - viewportPadding * 2)
  const left = Math.min(
    Math.max(viewportPadding, rect.right - panelWidth),
    window.innerWidth - panelWidth - viewportPadding,
  )
  const top = rect.bottom + 10

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

function goToNotifications() {
  close()
  void router.push('/notifications')
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
  } catch {
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
