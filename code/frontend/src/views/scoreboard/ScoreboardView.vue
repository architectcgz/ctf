<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">排行榜</h1>
        <p class="text-sm text-[var(--color-text-muted)]">{{ selectionHint }}</p>
      </div>
      <button
        type="button"
        class="rounded-lg border border-[var(--color-border-default)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)]"
        @click="refresh"
      >
        刷新
      </button>
    </div>

    <div
      v-if="loading && !hasSections"
      class="flex justify-center rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] py-12"
    >
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"
      ></div>
    </div>

    <div
      v-else-if="!hasSections"
      class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-8 text-center text-[var(--color-text-muted)]"
    >
      暂无可查看的竞赛排行榜
    </div>

    <div v-else class="space-y-5">
      <article
        v-for="section in sections"
        :key="section.contest.id"
        data-testid="scoreboard-card"
        class="overflow-hidden rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)]"
      >
        <header
          class="flex flex-col gap-3 border-b border-[var(--color-border-subtle)] px-5 py-4 lg:flex-row lg:items-center lg:justify-between"
        >
          <div class="space-y-2">
            <div class="flex flex-wrap items-center gap-2">
              <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">
                {{ section.contest.title }}
              </h2>
              <span
                class="rounded-full px-2 py-1 text-xs font-medium"
                :class="getStatusBadgeClass(section.contest.status)"
              >
                {{ getStatusLabel(section.contest.status) }}
              </span>
              <span
                v-if="section.frozen"
                class="rounded-full bg-[#f59e0b]/20 px-2 py-1 text-xs font-medium text-[#f59e0b]"
              >
                排行榜已冻结
              </span>
            </div>

            <p class="text-sm text-[var(--color-text-muted)]">
              {{ formatContestWindow(section.contest.starts_at, section.contest.ends_at) }}
            </p>
          </div>

          <div class="text-sm text-[var(--color-text-secondary)]">
            {{ section.rows.length > 0 ? `展示前 ${section.rows.length} 支队伍` : '暂无排行队伍' }}
          </div>
        </header>

        <div
          v-if="section.error"
          class="px-5 py-8 text-center text-sm text-[var(--color-text-muted)]"
        >
          该竞赛排行榜加载失败，请稍后重试
        </div>

        <div
          v-else-if="section.rows.length === 0"
          class="px-5 py-8 text-center text-sm text-[var(--color-text-muted)]"
        >
          暂无排行榜数据
        </div>

        <div v-else class="overflow-x-auto">
          <table class="min-w-full">
            <thead class="bg-[var(--color-bg-elevated)]">
              <tr>
                <th
                  class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]"
                >
                  排名
                </th>
                <th
                  class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]"
                >
                  队伍
                </th>
                <th
                  class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]"
                >
                  得分
                </th>
                <th
                  class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]"
                >
                  解题数
                </th>
                <th
                  class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]"
                >
                  最近得分
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in section.rows"
                :key="`${section.contest.id}-${item.team_id}`"
                class="border-b border-[var(--color-border-subtle)] transition-colors duration-100 hover:bg-[var(--color-bg-elevated)]"
                :class="getRowClass(item.rank)"
              >
                <td class="px-4 py-3 font-mono font-bold text-[var(--color-text-primary)]">
                  {{ item.rank }}
                </td>
                <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">
                  {{ item.team_name }}
                </td>
                <td class="px-4 py-3 font-mono text-sm text-[var(--color-text-primary)]">
                  {{ item.score }}
                </td>
                <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">
                  {{ item.solved_count }}
                </td>
                <td class="px-4 py-3 text-sm text-[var(--color-text-secondary)]">
                  {{ formatDateTime(item.last_submission_at) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ContestStatus } from '@/api/contracts'
import { useScoreboardView } from '@/composables/useScoreboardView'

const { hasSections, loading, refresh, sections, selectionHint } = useScoreboardView()

function formatDateTime(value?: string): string {
  if (!value) {
    return '未记录'
  }

  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function formatContestWindow(startsAt: string, endsAt: string): string {
  return `${formatDateTime(startsAt)} ~ ${formatDateTime(endsAt)}`
}

function getStatusLabel(status: ContestStatus): string {
  const labels: Record<ContestStatus, string> = {
    draft: '草稿',
    published: '已发布',
    registering: '报名中',
    running: '进行中',
    frozen: '已冻结',
    ended: '已结束',
    cancelled: '已取消',
    archived: '已归档',
  }
  return labels[status] || status
}

function getStatusBadgeClass(status: ContestStatus): string {
  if (status === 'running') return 'bg-emerald-500/15 text-emerald-500'
  if (status === 'frozen') return 'bg-[#f59e0b]/20 text-[#f59e0b]'
  if (status === 'ended') return 'bg-slate-500/15 text-slate-400'
  return 'bg-[var(--color-border-subtle)] text-[var(--color-text-secondary)]'
}

function getRowClass(rank: number): string {
  if (rank === 1) return 'bg-amber-500/5 border-l-2 border-amber-400'
  if (rank === 2) return 'bg-slate-300/5 border-l-2 border-slate-400'
  if (rank === 3) return 'bg-orange-400/5 border-l-2 border-orange-400'
  return ''
}
</script>
