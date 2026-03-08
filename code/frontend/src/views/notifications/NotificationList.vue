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

      <AppEmpty
        v-else-if="list.length === 0"
        icon="Inbox"
        title="暂无通知"
        description="新的系统、竞赛、团队和训练消息会在这里汇总展示。"
      />

      <div v-else class="space-y-3">
        <AppCard
          v-for="item in list"
          :key="item.id"
          as="button"
          variant="action"
          :accent="typeAccent(item.type)"
          interactive
          class="cursor-pointer"
          @click="handleMarkAsRead(item)"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="flex min-w-0 items-start gap-3">
              <div
                class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl border"
                :class="
                  typeAccent(item.type) === 'warning'
                    ? 'border-amber-500/20 bg-amber-500/10 text-amber-300'
                    : typeAccent(item.type) === 'success'
                      ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-300'
                      : typeAccent(item.type) === 'violet'
                        ? 'border-violet-500/20 bg-violet-500/10 text-violet-300'
                        : 'border-primary/20 bg-primary/10 text-primary'
                "
              >
                <component :is="typeIcon(item.type)" class="h-4 w-4" />
              </div>

              <div class="min-w-0 space-y-2">
              <div class="flex flex-wrap items-center gap-2">
                <span
                  class="rounded-full border px-2.5 py-1 text-xs font-medium"
                  :class="
                    typeAccent(item.type) === 'warning'
                      ? 'border-amber-500/20 bg-amber-500/10 text-amber-300'
                      : typeAccent(item.type) === 'success'
                        ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-300'
                        : typeAccent(item.type) === 'violet'
                          ? 'border-violet-500/20 bg-violet-500/10 text-violet-300'
                          : 'border-primary/20 bg-primary/10 text-primary'
                  "
                >
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
        </AppCard>
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
import { Flag, GraduationCap, Info, Trophy } from 'lucide-vue-next'

import { getNotifications, markAsRead } from '@/api/notification'
import type { NotificationItem } from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
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

function typeAccent(type: string): 'primary' | 'success' | 'warning' | 'violet' {
  if (type === 'contest') return 'warning'
  if (type === 'challenge') return 'success'
  if (type === 'team') return 'violet'
  return 'primary'
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
