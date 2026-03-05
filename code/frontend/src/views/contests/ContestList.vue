<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">竞赛中心</h1>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="contest in list"
        :key="contest.id"
        class="cursor-pointer rounded-lg border bg-[var(--color-bg-surface)] p-5 transition-colors duration-150"
        :class="getContestBorderClass(contest.status)"
        @click="goToDetail(contest.id)"
      >
        <div class="space-y-3">
          <h3 class="text-xl font-semibold text-[var(--color-text-primary)]">{{ contest.title }}</h3>

          <div class="flex items-center gap-3 text-sm">
            <span
              class="rounded px-2 py-0.5 text-xs font-medium"
              :class="getStatusBadgeClass(contest.status)"
            >
              {{ getStatusLabel(contest.status) }}
            </span>
            <span class="text-[var(--color-text-secondary)]">{{ getModeLabel(contest.mode) }}</span>
          </div>

          <div class="space-y-1 text-sm text-[var(--color-text-secondary)]">
            <div class="font-mono">
              {{ formatTime(contest.starts_at) }} ~ {{ formatTime(contest.ends_at) }}
            </div>
          </div>

          <div class="flex gap-3">
            <button
              class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white transition-colors duration-150 hover:bg-[#06b6d4]"
              @click.stop="handleAction(contest)"
            >
              {{ getActionLabel(contest.status) }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="list.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
        <div class="text-[var(--color-text-muted)]">暂无竞赛</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getContests } from '@/api/contest'
import { usePagination } from '@/composables/usePagination'
import type { ContestStatus, ContestMode } from '@/api/contracts'

const router = useRouter()
const { list, loading, refresh } = usePagination(getContests)

onMounted(() => {
  refresh()
})

function goToDetail(id: string) {
  router.push(`/contests/${id}`)
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
    archived: '已归档'
  }
  return labels[status] || status
}

function getModeLabel(mode: ContestMode): string {
  const labels: Record<ContestMode, string> = {
    jeopardy: 'Jeopardy',
    awd: 'AWD',
    awd_plus: 'AWD+',
    king_of_hill: 'King of the Hill',
  }
  return labels[mode] || mode
}

function getContestBorderClass(status: ContestStatus): string {
  if (status === 'running') return 'border-l-2 border-[#0891b2] border-[var(--color-border-default)]'
  if (status === 'registering') return 'border-l-2 border-[#f59e0b] border-[var(--color-border-default)]'
  if (status === 'ended') return 'border-[var(--color-border-default)] opacity-70'
  return 'border-[var(--color-border-default)] hover:border-[var(--color-primary)]/50'
}

function getStatusBadgeClass(status: ContestStatus): string {
  if (status === 'running') return 'bg-[var(--color-primary)]/10 text-[#06b6d4]'
  if (status === 'registering') return 'bg-[#f59e0b]/10 text-[#f59e0b]'
  if (status === 'ended') return 'bg-[#30363d] text-[var(--color-text-secondary)]'
  return 'bg-[#30363d] text-[var(--color-text-secondary)]'
}

function getActionLabel(status: ContestStatus): string {
  if (status === 'running') return '进入竞赛'
  if (status === 'registering') return '立即报名'
  return '查看详情'
}

function handleAction(contest: any) {
  goToDetail(contest.id)
}

function formatTime(time: string): string {
  return new Date(time).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>
