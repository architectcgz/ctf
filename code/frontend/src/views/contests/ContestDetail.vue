<template>
  <div class="space-y-6">
    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[#30363d] border-t-[#0891b2]"></div>
    </div>

    <div v-else-if="contest" class="space-y-6">
      <div class="rounded-lg border border-[#30363d] bg-[#161b22] p-6">
        <h1 class="text-3xl font-bold text-[#e6edf3]">{{ contest.title }}</h1>
        <p class="mt-3 text-[#8b949e]">{{ contest.description }}</p>

        <div class="mt-4 flex items-center gap-4 text-sm">
          <span class="rounded px-2 py-0.5 text-xs font-medium" :class="getStatusBadgeClass(contest.status)">
            {{ getStatusLabel(contest.status) }}
          </span>
          <span class="text-[#8b949e]">{{ getModeLabel(contest.mode) }}</span>
        </div>

        <div class="mt-4 font-mono text-sm text-[#8b949e]">
          {{ formatTime(contest.starts_at) }} ~ {{ formatTime(contest.ends_at) }}
        </div>
      </div>

      <div class="text-center text-[#6e7681]">竞赛详情页面开发中...</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getContestDetail } from '@/api/contest'
import type { ContestDetailData, ContestStatus, ContestMode } from '@/api/contracts'

const route = useRoute()
const contest = ref<ContestDetailData | null>(null)
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    contest.value = await getContestDetail(route.params.id as string)
  } finally {
    loading.value = false
  }
})

function getStatusLabel(status: ContestStatus): string {
  const labels: Record<ContestStatus, string> = {
    draft: '草稿', published: '已发布', registering: '报名中',
    running: '进行中', frozen: '已冻结', ended: '已结束',
    cancelled: '已取消', archived: '已归档'
  }
  return labels[status] || status
}

function getModeLabel(mode: ContestMode): string {
  const labels: Record<ContestMode, string> = {
    jeopardy: 'Jeopardy', awd: 'AWD', mixed: '混合模式'
  }
  return labels[mode] || mode
}

function getStatusBadgeClass(status: ContestStatus): string {
  if (status === 'running') return 'bg-[#0891b2]/10 text-[#06b6d4]'
  if (status === 'registering') return 'bg-[#f59e0b]/10 text-[#f59e0b]'
  return 'bg-[#30363d] text-[#8b949e]'
}

function formatTime(time: string): string {
  return new Date(time).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>
