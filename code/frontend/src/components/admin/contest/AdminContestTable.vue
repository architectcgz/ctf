<script setup lang="ts">
import { computed } from 'vue'

import type { ContestDetailData } from '@/api/contracts'
import { getModeLabel, getStatusBadgeClass, getStatusLabel } from '@/utils/contest'

const props = defineProps<{
  contests: ContestDetailData[]
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  edit: [contest: ContestDetailData]
  export: [contest: ContestDetailData]
  changePage: [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

function formatTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="space-y-5">
    <div class="overflow-hidden rounded-2xl border border-border">
      <table class="min-w-full divide-y divide-border">
        <thead class="bg-surface-alt/70">
          <tr class="text-left text-xs font-semibold uppercase tracking-[0.2em] text-[var(--color-text-muted)]">
            <th class="px-4 py-3">竞赛</th>
            <th class="px-4 py-3">模式</th>
            <th class="px-4 py-3">状态</th>
            <th class="px-4 py-3">时间窗口</th>
            <th class="px-4 py-3">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border bg-surface/70">
          <tr
            v-for="contest in contests"
            :key="contest.id"
            class="transition hover:bg-surface-alt/60"
          >
            <td class="px-4 py-4 align-top">
              <div class="space-y-1">
                <p class="font-medium text-[var(--color-text-primary)]">{{ contest.title }}</p>
                <p class="text-sm text-[var(--color-text-muted)]">
                  {{ contest.description || '当前未填写竞赛描述。' }}
                </p>
              </div>
            </td>
            <td class="px-4 py-4 align-top text-sm text-[var(--color-text-secondary)]">
              {{ getModeLabel(contest.mode) }}
            </td>
            <td class="px-4 py-4 align-top">
              <span
                class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
                :class="getStatusBadgeClass(contest.status)"
              >
                {{ getStatusLabel(contest.status) }}
              </span>
            </td>
            <td class="px-4 py-4 align-top text-sm text-[var(--color-text-secondary)]">
              <p>{{ formatTime(contest.starts_at) }}</p>
              <p class="mt-1 text-[var(--color-text-muted)]">至 {{ formatTime(contest.ends_at) }}</p>
            </td>
            <td class="px-4 py-4 align-top">
              <div class="flex flex-wrap gap-2">
                <button
                  type="button"
                  class="rounded-xl border border-border px-3 py-1.5 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary hover:text-primary"
                  @click="emit('edit', contest)"
                >
                  编辑
                </button>
                <button
                  type="button"
                  class="rounded-xl border border-border px-3 py-1.5 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary hover:text-primary"
                  @click="emit('export', contest)"
                >
                  导出结果
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="flex flex-col gap-3 text-sm text-[var(--color-text-muted)] sm:flex-row sm:items-center sm:justify-between">
      <span>共 {{ total }} 场竞赛</span>
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="rounded-xl border border-border px-3 py-1.5 text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="page <= 1"
          @click="emit('changePage', page - 1)"
        >
          上一页
        </button>
        <span>{{ page }} / {{ totalPages }}</span>
        <button
          type="button"
          class="rounded-xl border border-border px-3 py-1.5 text-[var(--color-text-primary)] transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="page >= totalPages"
          @click="emit('changePage', page + 1)"
        >
          下一页
        </button>
      </div>
    </div>
  </div>
</template>
